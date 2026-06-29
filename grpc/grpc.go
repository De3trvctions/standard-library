package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Option 配置
type Option struct {
	Dial                  func(address string, opt *Option) (*grpc.ClientConn, error) `json:"-"`
	MaxIdle               int                                                         //最大链接池大小
	MaxActive             int                                                         //在给定时间分配的最大连接数。为0时，池中的连接数没有限制
	MaxConcurrentStreams  int                                                         //限制每个连接的并发流数量
	Reuse                 bool                                                        //pool在 MaxActive 限制时，为 true，Get() 会返回重用连接，为 false，则创建新链接返回。
	RecycleDur            uint64                                                      //回收间隔时间(s)。最小间隔必须大于10s
	GrowStepMax           int                                                         //单次扩容最大增量(连接数)
	GrowCooldownMs        uint32                                                      //扩容冷却时间(ms)
	OverflowMax           int                                                         //MaxActive 达到后允许的溢出(once)连接并发上限
	OverflowDialPerSecond uint32                                                      //溢出连接拨号频率上限(每秒)，0 表示不限速
	ShrinkLowRate         float64                                                     //缩容低水位阈值(滑动窗口利用率)
	ShrinkLowStreak       int                                                         //连续低水位次数达到后才缩容
	Logger                Logger                                                      //log打印
	DialOptions           []grpc.DialOption                                           //额外的grpc链接设置
}

// DefaultOptions 默认配置
var DefaultOptions = Option{
	Dial:                 Dial,
	MaxIdle:              8,
	MaxActive:            64,
	MaxConcurrentStreams: 64,
	Reuse:                true,
	RecycleDur:           600,
	Logger:               Logger{true},
	DialOptions:          []grpc.DialOption{},
}

// Copy 拷贝配置，防止指针传递后被修改
func (o *Option) Copy() *Option {
	copiedDialOptions := append([]grpc.DialOption(nil), o.DialOptions...)
	return &Option{
		Dial:                  o.Dial,
		MaxIdle:               o.MaxIdle,
		MaxActive:             o.MaxActive,
		MaxConcurrentStreams:  o.MaxConcurrentStreams,
		Reuse:                 o.Reuse,
		RecycleDur:            o.RecycleDur,
		GrowStepMax:           o.GrowStepMax,
		GrowCooldownMs:        o.GrowCooldownMs,
		OverflowMax:           o.OverflowMax,
		OverflowDialPerSecond: o.OverflowDialPerSecond,
		ShrinkLowRate:         o.ShrinkLowRate,
		ShrinkLowStreak:       o.ShrinkLowStreak,
		Logger:                o.Logger,
		DialOptions:           copiedDialOptions,
	}
}

// GetRecycleDur 获取回收间隔时间
func (o *Option) GetRecycleDur() time.Duration {
	if o.RecycleDur == 0 || o.RecycleDur < 10 {
		return RecycleDuration
	}
	return time.Duration(o.RecycleDur) * time.Second
}

// WithDialOptions 添加额外的grpc链接设置
func (o *Option) WithDialOptions(options ...grpc.DialOption) {
	o.DialOptions = options
}

func (o *Option) getDialOptions() []grpc.DialOption {
	return o.DialOptions
}

func parseTarget(address string, defaultPort string) (string, error) {
	if address == "" {
		return "", fmt.Errorf("GRPC invalid address <%s>", address)
	}

	host := address
	port := defaultPort

	if h, p, err := net.SplitHostPort(address); err == nil {
		host = h
		if p != "" {
			port = p
		}
	} else if strings.Count(address, ":") == 1 {
		parts := strings.SplitN(address, ":", 2)
		if len(parts) == 2 {
			host = parts[0]
			if parts[1] != "" {
				port = parts[1]
			}
		}
	}

	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") && !strings.HasSuffix(host, "]") {
		host = "[" + host + "]"
	}
	return host + ":" + port, nil
}

func defaultConnectParams() grpc.ConnectParams {
	cfg := backoff.DefaultConfig
	cfg.MaxDelay = BackoffMaxDelay
	return grpc.ConnectParams{
		Backoff:           cfg,
		MinConnectTimeout: MinConnectTimeout,
	}
}

func defaultDialOptions(opt *Option) []grpc.DialOption {
	base := append([]grpc.DialOption(nil), opt.getDialOptions()...)
	return append(base,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(defaultConnectParams()),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxSendMsgSize), grpc.MaxCallRecvMsgSize(MaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                KeepAliveTime,
			Timeout:             KeepAliveTimeout,
			PermitWithoutStream: false,
		}),
		grpc.WithBlock(),
	)
}

// Dial 返回默认配置的 grpc 连接。支持填写IPv4和hostname
func Dial(address string, opt *Option) (*grpc.ClientConn, error) {
	target, err := parseTarget(address, "80")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	defer cancel()
	return grpc.DialContext(ctx, target, defaultDialOptions(opt)...)
}

// Server 对grpc package server结构的封装
type Server struct {
	Srv *grpc.Server
}

// NewServer 创建新的GRPC服务
func NewServer(opt ...grpc.ServerOption) *Server {
	return &Server{grpc.NewServer(opt...)}
}

//封装 grpc.DialOption

// WithUnaryClientInterceptor 封装gRPC WithUnaryInterceptor
func WithUnaryClientInterceptor(interceptor grpc.UnaryClientInterceptor) grpc.DialOption {
	return grpc.WithUnaryInterceptor(interceptor)
}

// WithChainUnaryClientInterceptor 封装gRPC WithChainUnaryInterceptor
func WithChainUnaryClientInterceptor(interceptor ...grpc.UnaryClientInterceptor) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(interceptor...)
}
