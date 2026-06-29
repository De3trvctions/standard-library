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
	if option == nil {
		option = DefaultOptions.Copy()
	} else {
		option = option.Copy()
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
	index                 uint32 //随机获取连接
	current               int32  //当前的实际连接数
	ref                   int32  //当前的逻辑链接 logic connection = physical connection * MaxConcurrentStreams
	opt                   *Option
	conns                 []*conn    //全部实际链接
	dirty                 *list.List //要被清理的链接
	address               string     //服务器地址
	closed                int32      //pool关闭标识位
	ring                  *ring.Ring
	ringMu                sync.Mutex
	usage                 atomic.Value
	windowSize            int32
	lowStreak             int32
	overflowInFlight      int32
	lastOverflowDialNs    int64
	overflowMinIntervalNs int64
	overflowMax           int32
	lastGrowNs            int64
	growCooldownNs        int64
	growStepMax           int32
	shrinkLowRate         float64
	shrinkLowStreak       int32
	ctx                   context.Context
	cancel                context.CancelFunc
	sync.RWMutex
}

// create 创建链接池
func create(address string, option *Option) (Pool, error) {
	if address == "" {
		return nil, errors.New("invalid address settings")
	}
	if option == nil {
		return nil, errors.New("invalid option settings")
	}
	option = option.Copy()
	if option.Dial == nil {
		option.Dial = Dial
	}
	if option.MaxIdle <= 0 || option.MaxActive <= 0 || option.MaxIdle > option.MaxActive {
		return nil, errors.New("invalid maximum settings")
	}
	if option.MaxConcurrentStreams <= 0 {
		return nil, errors.New("invalid maximum settings")
	}
	if option.GrowStepMax <= 0 {
		option.GrowStepMax = 32
		if option.GrowStepMax > option.MaxActive {
			option.GrowStepMax = option.MaxActive
		}
	}
	if option.GrowCooldownMs == 0 {
		option.GrowCooldownMs = 200
	}
	if option.OverflowMax <= 0 {
		def := option.MaxActive / 8
		if def > 8 {
			def = 8
		}
		if def < 1 {
			def = 1
		}
		option.OverflowMax = def
	}
	if option.OverflowDialPerSecond == 0 {
		option.OverflowDialPerSecond = 50
	}
	if option.ShrinkLowRate <= 0 {
		option.ShrinkLowRate = 0.15
	}
	if option.ShrinkLowStreak <= 0 {
		option.ShrinkLowStreak = 3
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
	p.windowSize = int32(p.ring.Len())
	p.overflowMax = int32(option.OverflowMax)
	p.growStepMax = int32(option.GrowStepMax)
	p.growCooldownNs = int64(option.GrowCooldownMs) * int64(time.Millisecond)
	if option.OverflowDialPerSecond > 0 {
		p.overflowMinIntervalNs = int64(time.Second) / int64(option.OverflowDialPerSecond)
	}
	p.shrinkLowRate = option.ShrinkLowRate
	p.shrinkLowStreak = int32(option.ShrinkLowStreak)
	p.ctx, p.cancel = context.WithCancel(context.TODO())
	r := p.ring
	for i := 0; i < r.Len(); i++ {
		r.Value = new(int64)
		r = r.Next()
	}
	p.usage.Store(p.ring.Value.(*int64))
	for i := 0; i < p.opt.MaxIdle; i++ {
		c, err := p.opt.Dial(address, p.opt)
		if err != nil {
			_ = p.Close()
			return nil, fmt.Errorf("dial is not able to fill the pool: %s", err)
		}
		p.conns[i] = p.wrapConn(c, false, false)
	}
	p.opt.Logger.Printf("create pool success: %v\n", p.Status())
	go p.autoRecycle()
	go p.recycle()
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
	if v, ok := servicesOpt.Load(serviceName); ok {
		ro, ok := v.(*reloadOpt)
		if !ok || ro == nil {
			return ErrNoOption
		}
		if time.Since(ro.reloadTime) < ReconnectDuration {
			return ErrReconnectCD
		}
		ro.reloadTime = time.Now()
		servicesOpt.Store(serviceName, ro)
		pool, err := create(ro.address, ro.opt)
		if err != nil {
			return err
		}
		services.Store(serviceName, pool)
		return nil
	}
	return ErrNoOption
}

func (p *pool) getPooled(current int32) (*conn, error) {
	next := atomic.AddUint32(&p.index, 1) % uint32(current)
	atomic.AddInt64(p.usage.Load().(*int64), 1)
	p.RLock()
	c := p.conns[next]
	p.RUnlock()
	if c == nil {
		return nil, ErrClosed
	}
	return c, nil
}

// Get 从pool中获取连接
func (p *pool) Get() (Conn, error) {
	if atomic.LoadInt32(&p.closed) != 0 || atomic.LoadInt32(&p.current) == 0 {
		return nil, ErrClosed
	}
	nextRef := p.incrRef()
	current := atomic.LoadInt32(&p.current)
	if current == 0 {
		p.decrRef()
		return nil, ErrClosed
	}
	streams := int32(p.opt.MaxConcurrentStreams)
	if streams <= 0 {
		p.decrRef()
		return nil, errors.New("invalid maximum settings")
	}
	required := (nextRef + streams - 1) / streams
	if required <= current {
		c, err := p.getPooled(current)
		if err != nil {
			p.decrRef()
			return nil, err
		}
		return c, nil
	}

	// 连接数达到maxActive
	if current == int32(p.opt.MaxActive) {
		if p.opt.Reuse {
			c, err := p.getPooled(current)
			if err != nil {
				p.decrRef()
				return nil, err
			}
			return c, nil
		}
		if p.overflowMax <= 0 {
			c, err := p.getPooled(current)
			if err != nil {
				p.decrRef()
				return nil, err
			}
			return c, nil
		}
		if atomic.AddInt32(&p.overflowInFlight, 1) > p.overflowMax {
			atomic.AddInt32(&p.overflowInFlight, -1)
			c, err := p.getPooled(current)
			if err != nil {
				p.decrRef()
				return nil, err
			}
			return c, nil
		}
		if p.overflowMinIntervalNs > 0 {
			nowNs := time.Now().UnixNano()
			for {
				last := atomic.LoadInt64(&p.lastOverflowDialNs)
				if nowNs-last < p.overflowMinIntervalNs {
					atomic.AddInt32(&p.overflowInFlight, -1)
					c, err := p.getPooled(current)
					if err != nil {
						p.decrRef()
						return nil, err
					}
					return c, nil
				}
				if atomic.CompareAndSwapInt64(&p.lastOverflowDialNs, last, nowNs) {
					break
				}
			}
		}
		cc, err := p.opt.Dial(p.address, p.opt)
		if err != nil {
			atomic.AddInt32(&p.overflowInFlight, -1)
			c, er := p.getPooled(current)
			if er != nil {
				p.decrRef()
				return nil, er
			}
			return c, nil
		}
		atomic.AddInt64(p.usage.Load().(*int64), 1)
		return p.wrapConn(cc, true, true), nil
	}

	// 创建新的连接返回给pool(小步扩容)
	p.Lock()
	current = atomic.LoadInt32(&p.current)
	if current < int32(p.opt.MaxActive) && required > current {
		nowNs := time.Now().UnixNano()
		lastGrowNs := atomic.LoadInt64(&p.lastGrowNs)
		if p.growCooldownNs == 0 || nowNs-lastGrowNs >= p.growCooldownNs {
			atomic.StoreInt64(&p.lastGrowNs, nowNs)
			need := required - current
			if need < 1 {
				need = 1
			}
			quarter := current / 4
			if quarter < 1 {
				quarter = 1
			}
			increment := need
			if increment > quarter {
				increment = quarter
			}
			if p.growStepMax > 0 && increment > p.growStepMax {
				increment = p.growStepMax
			}
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
				p.conns[current+i] = p.wrapConn(c, false, false)
			}
			newCurrent := current + i
			if i > 0 {
				p.opt.Logger.Printf("grow pool: %d ---> %d, increment: %d, maxActive: %d\n",
					current, newCurrent, increment, p.opt.MaxActive)
				atomic.StoreInt32(&p.current, newCurrent)
				current = newCurrent
			}
			if err != nil {
				p.Unlock()
				p.decrRef()
				return nil, err
			}
		}
	}
	p.Unlock()
	c, err := p.getPooled(current)
	if err != nil {
		p.decrRef()
		return nil, err
	}
	return c, nil
}

// Close see Pool interface.
func (p *pool) Close() error {
	p.cancel()
	p.Lock()
	atomic.StoreInt32(&p.closed, 1)
	atomic.StoreUint32(&p.index, 0)
	atomic.StoreInt32(&p.current, 0)
	atomic.StoreInt32(&p.ref, 0)
	p.deleteFrom(0)
	p.Unlock()
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
	ticker := time.NewTicker(p.opt.GetRecycleDur())
	defer ticker.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			current := atomic.LoadInt32(&p.current)
			if current <= 0 {
				continue
			}
			bucket := p.usage.Load().(*int64)
			usage := atomic.SwapInt64(bucket, 0)
			sum := usage
			p.ringMu.Lock()
			r := p.ring
			for i := 0; i < r.Len(); i++ {
				v := r.Value.(*int64)
				if v != bucket {
					sum += atomic.LoadInt64(v)
				}
				r = r.Next()
			}
			p.ring = p.ring.Next()
			p.usage.Store(p.ring.Value.(*int64))
			p.ringMu.Unlock()
			windowSize := p.windowSize
			if windowSize <= 0 {
				windowSize = 1
			}
			denom := float64(current) * float64(p.opt.MaxConcurrentStreams) * float64(windowSize)
			rate := float64(sum) / denom
			p.opt.Logger.Printf("calculate utilisation rate：%d/%d=%.3f\n",
				sum, current*int32(p.opt.MaxConcurrentStreams)*windowSize, rate)
			if rate <= p.shrinkLowRate {
				if atomic.AddInt32(&p.lowStreak, 1) >= p.shrinkLowStreak {
					atomic.StoreInt32(&p.lowStreak, 0)
					if current > int32(p.opt.MaxIdle) {
						shrink := current / 2
						if shrink < int32(p.opt.MaxIdle) {
							shrink = int32(p.opt.MaxIdle)
						}
						atomic.StoreInt32(&p.current, shrink)
						atomic.StoreUint32(&p.index, 0)
						p.Lock()
						p.removeFrom(int(shrink))
						p.Unlock()
						p.opt.Logger.Printf("shrink pool: %d ---> %d, maxActive: %d\n",
							current, shrink, p.opt.MaxActive)
					}
				}
			} else {
				atomic.StoreInt32(&p.lowStreak, 0)
			}
		}
	}
}

func (p *pool) recycle() {
	ticker := time.NewTicker(p.opt.GetRecycleDur() + time.Second*60)
	defer ticker.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
		}
		var toReset []*conn
		p.Lock()
		for e := p.dirty.Front(); e != nil; {
			next := e.Next()
			if v, ok := e.Value.(*conn); ok && v != nil {
				toReset = append(toReset, v)
			}
			p.dirty.Remove(e)
			e = next
		}
		remaining := p.dirty.Len()
		p.Unlock()
		for _, c := range toReset {
			_ = c.reset()
		}
		p.opt.Logger.Printf("Clean dirty connecting Done.conn remaining <%d>\n", remaining)
	}
}
