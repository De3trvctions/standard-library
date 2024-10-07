package grpc

import (
	"container/list"
	"container/ring"
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var services sync.Map    //map[serviceName]Pool
var servicesOpt sync.Map //map[serviceName]*reloadOpt

type reloadOpt struct {
	opt        *Option
	address    string
	reloadTime time.Time
}

// Get 获取grpc client
func Get(serviceName string) (Conn, error) {
	if serviceName == "" {
		serviceName = "default"
	}
	service, ok := services.Load(serviceName)
	if !ok {
		err := reconnect(serviceName)
		if err == nil {
			return Get(serviceName)
		}
		return nil, fmt.Errorf("unregistired service <%s>", serviceName)
	}
	v, ok := service.(Pool)
	if !ok {
		return nil, fmt.Errorf("uninitializated service <%s>", serviceName)
	}
	return v.Get()
}

// Register 注册GRPC连接
func Register(serviceName string, address string, option *Option) error {
	if serviceName == "" {
		serviceName = "default"
	}
	servicesOpt.Store(serviceName, &reloadOpt{option, address, time.Now()})
	pool, err := create(address, option)
	if err != nil {
		return err
	}
	services.Store(serviceName, pool)
	return nil
}

// Pool 链接池
type Pool interface {
	// Get 从pool中返回一个新连接。关闭连接将其放回pool中。
	// 当pool被销毁或满时关闭它将报错。
	// 当 cli 不为 nil 时，cli.Conn() 一定不为 nil。
	Get() (Conn, error)

	// Close 销毁pool及其所有连接。在 Close() 之后，pool不再可用。
	// 不要同时调用 Close 和 Get 方法。会panic
	Close() error

	// Status 返回pool的当前状态.
	Status() string
}

type pool struct {
	index   uint32 //随机获取连接
	current int32  //当前的实际连接数
	ref     int32  //当前的逻辑链接 logic connection = physical connection * MaxConcurrentStreams
	opt     *Option
	conns   []*conn    //全部实际链接
	dirty   *list.List //要被清理的链接
	address string     //服务器地址
	ip      string     //服务器Ip，解析address
	port    string     //服务器端口，解析address
	closed  int32      //pool关闭标识位
	ring    *ring.Ring
	ctx     context.Context
	cancel  context.CancelFunc
	sync.RWMutex
}

// create 创建链接池
func create(address string, option *Option) (Pool, error) {
	if address == "" {
		return nil, errors.New("invalid address settings")
	}
	if option.Dial == nil {
		option.Dial = Dial
	}
	if option.MaxIdle <= 0 || option.MaxActive <= 0 || option.MaxIdle > option.MaxActive {
		return nil, errors.New("invalid maximum settings")
	}
	if option.MaxConcurrentStreams <= 0 {
		return nil, errors.New("invalid maximum settings")
	}

	p := &pool{
		index:   0,
		current: int32(option.MaxIdle),
		ref:     0,
		opt:     option,
		conns:   make([]*conn, option.MaxActive),
		dirty:   list.New(),
		address: address,
		closed:  0,
		ring:    ring.New(4),
	}
	p.ctx, p.cancel = context.WithCancel(context.TODO())
	zero := int64(0)
	p.ring.Value = &zero
	for i := 0; i < p.opt.MaxIdle; i++ {
		c, err := p.opt.Dial(address, p.opt)
		if err != nil {
			_ = p.Close()
			return nil, fmt.Errorf("dial is not able to fill the pool: %s", err)
		}
		p.conns[i] = p.wrapConn(c, false)
	}
	p.opt.Logger.Printf("create pool success: %v\n", p.Status())
	go p.autoRecycle() //TODO 测试用
	go p.recycle()     //TODO 测试用
	return p, nil
}

func (p *pool) incrRef() int32 {
	newRef := atomic.AddInt32(&p.ref, 1)
	if newRef == math.MaxInt32 {
		panic(fmt.Sprintf("overflow ref: %d", newRef))
	}
	return newRef
}

func (p *pool) decrRef() {
	newRef := atomic.AddInt32(&p.ref, -1)
	if newRef < 0 && atomic.LoadInt32(&p.closed) == 0 {
		panic(fmt.Sprintf("negative ref: %d", newRef))
	}
	if newRef == 0 && atomic.LoadInt32(&p.current) > int32(p.opt.MaxIdle) {
		p.Lock()
		if atomic.LoadInt32(&p.ref) == 0 {
			//log.Printf("shrink pool: %d ---> %d, decrement: %d, maxActive: %d\n", p.current, p.opt.MaxIdle, p.current-int32(p.opt.MaxIdle), p.opt.MaxActive)
			atomic.StoreInt32(&p.current, int32(p.opt.MaxIdle))
			p.deleteFrom(p.opt.MaxIdle)
		}
		p.Unlock()
	}
}

func (p *pool) reset(index int) {
	conn := p.conns[index]
	if conn == nil {
		return
	}
	_ = conn.reset()
	p.conns[index] = nil
}

func (p *pool) deleteFrom(begin int) {
	for i := begin; i < p.opt.MaxActive; i++ {
		p.reset(i)
	}
}

// 为自动缩减提供异步清理支持
func (p *pool) removeFrom(begin int) {
	for i := begin; i < p.opt.MaxActive; i++ {
		if p.conns[i] == nil {
			continue
		}
		p.dirty.PushBack(p.conns[i])
		p.conns[i] = nil
	}
}

// Reconnect 重连机制
// WARNING 仅针对服务未注册情况，若添加新服务，该方法会因为无法获取到新增的配置而无效。
func reconnect(serviceName string) error {
	if t, ok := servicesOpt.Load(serviceName); ok {
		if time.Now().Sub(t.(*reloadOpt).reloadTime).Nanoseconds() < int64(ReconnectDuration) {
			return ErrReconnectCD
		}
		if opt, ok := servicesOpt.Load(serviceName); ok {
			option := opt.(*reloadOpt)
			pool, err := create(option.address, option.opt)
			if err != nil {
				return err
			}
			services.Store(serviceName, pool)
			return nil
		}
	}
	return ErrNoOption
}

// Get 从pool中获取连接
func (p *pool) Get() (Conn, error) {
	//从创建的连接中选择第一个
	nextRef := p.incrRef()
	p.RLock()
	current := atomic.LoadInt32(&p.current)
	p.RUnlock()
	if current == 0 {
		return nil, ErrClosed
	}
	if nextRef <= current*int32(p.opt.MaxConcurrentStreams) {
		next := atomic.AddUint32(&p.index, 1) % uint32(current)
		atomic.AddInt64(p.ring.Value.(*int64), 1)
		return p.conns[next], nil
	}

	// 连接数达到maxActive
	if current == int32(p.opt.MaxActive) {
		// 如果Reuse为true，从pool的连接中选择
		if p.opt.Reuse {
			next := atomic.AddUint32(&p.index, 1) % uint32(current)
			atomic.AddInt64(p.ring.Value.(*int64), 1)
			return p.conns[next], nil
		}
		// 否则新建
		c, err := p.opt.Dial(p.address, p.opt)
		atomic.AddInt64(p.ring.Value.(*int64), 1)
		return p.wrapConn(c, true), err
	}

	// 创建新的连接返回给pool
	p.Lock()
	current = atomic.LoadInt32(&p.current)
	if current < int32(p.opt.MaxActive) && nextRef > current*int32(p.opt.MaxConcurrentStreams) {
		// 2倍增量或保持增量
		increment := current
		if current+increment > int32(p.opt.MaxActive) {
			increment = int32(p.opt.MaxActive) - current
		}
		var i int32
		var err error
		for i = 0; i < increment; i++ {
			c, er := p.opt.Dial(p.address, p.opt)
			if er != nil {
				err = er
				break
			}
			p.reset(int(current + i))
			p.conns[current+i] = p.wrapConn(c, false)
		}
		current += i
		p.opt.Logger.Printf("grow pool: %d ---> %d, increment: %d, maxActive: %d\n",
			p.current, current, increment, p.opt.MaxActive)
		atomic.StoreInt32(&p.current, current)
		if err != nil {
			p.Unlock()
			return nil, err
		}
	}
	p.Unlock()
	next := atomic.AddUint32(&p.index, 1) % uint32(current)
	atomic.AddInt64(p.ring.Value.(*int64), 1)
	return p.conns[next], nil
}

// Close see Pool interface.
func (p *pool) Close() error {
	p.cancel()
	atomic.StoreInt32(&p.closed, 1)
	atomic.StoreUint32(&p.index, 0)
	atomic.StoreInt32(&p.current, 0)
	atomic.StoreInt32(&p.ref, 0)
	p.deleteFrom(0)
	p.opt.Logger.Printf("close pool success: %v\n", p.Status())
	return nil
}

// Status see Pool interface.
func (p *pool) Status() string {
	return fmt.Sprintf("address:%s, index:%d, current:%d, ref:%d. option:%+v",
		p.address, p.index, p.current, p.ref, p.opt)
}

// 自动回收连接池的连接
func (p *pool) autoRecycle() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-time.After(p.opt.GetRecycleDur()):
			usage := atomic.LoadInt64(p.ring.Value.(*int64))
			rate := float64(usage) / (float64(p.current) * float64(p.opt.MaxConcurrentStreams))
			p.opt.Logger.Printf("calculate utilisation rate：%d/%d=%.3f\n",
				usage, p.current*int32(p.opt.MaxConcurrentStreams), rate)
			atomic.StoreInt64(p.ring.Value.(*int64), 0)
			p.ring.Next()
			if rate <= 0.2 {
				curr := atomic.LoadInt32(&p.current)
				shrink := curr / 2
				if curr != int32(p.opt.MaxIdle) {
					if shrink < int32(p.opt.MaxIdle) {
						shrink = int32(p.opt.MaxIdle)
					}
					atomic.StoreInt32(&p.current, shrink)
					atomic.StoreUint32(&p.index, 0)
					p.Lock()
					p.removeFrom(int(shrink))
					p.Unlock()
					p.opt.Logger.Printf("shrink pool: %d ---> %d, decrement: %d, maxActive: %d\n",
						curr, shrink, p.current-int32(p.opt.MaxIdle), p.opt.MaxActive)
				}
			}
		}
	}
}

func (p *pool) recycle() {
	for range time.Tick(p.opt.GetRecycleDur() + time.Second*60) {
		var o *list.Element
		for e := p.dirty.Front(); e != nil; e = o {
			if e.Value == nil {
				o = e.Next()
				p.dirty.Remove(e)
				continue
			}
			v, ok := e.Value.(*conn)
			if !ok {
				o = e.Next()
				p.dirty.Remove(e)
				continue
			}

			_ = v.reset()
			o = e.Next()
			p.dirty.Remove(e)
		}
		p.opt.Logger.Printf("Clean dirty connecting Done.conn remaining <%d>\n", p.dirty.Len())
	}
}
