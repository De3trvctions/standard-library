package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Option 配置
type Option struct {
	Dial                 func(address string, opt *Option) (*grpc.ClientConn, error) `json:"-"`
	MaxIdle              int                                                         //最大链接池大小
	MaxActive            int                                                         //在给定时间分配的最大连接数。为0时，池中的连接数没有限制
	MaxConcurrentStreams int                                                         //限制每个连接的并发流数量
	Reuse                bool                                                        //pool在 MaxActive 限制时，为 true，Get() 会返回重用连接，为 false，则创建新链接返回。
	RecycleDur           uint64                                                      //回收间隔时间(s)。最小间隔必须大于10s
	Logger               Logger                                                      //log打印
	DialOptions          []grpc.DialOption                                           //额外的grpc链接设置
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
	return &Option{
		Dial:                 o.Dial,
		MaxIdle:              o.MaxIdle,
		MaxActive:            o.MaxActive,
		MaxConcurrentStreams: o.MaxConcurrentStreams,
		Reuse:                o.Reuse,
		RecycleDur:           o.RecycleDur,
		Logger:               o.Logger,
		DialOptions:          o.DialOptions,
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

// Dial 返回默认配置的 grpc 连接。支持填写IPv4和hostname
func Dial(address string, opt *Option) (*grpc.ClientConn, error) {
	var port = "80"
	addresses := strings.Split(address, ":")
	if len(addresses) > 1 && addresses[1] != "" {
		var err error
		port = addresses[1]
		if err != nil {
			return nil, fmt.Errorf("GRPC invalid address <%s>", address)
		}
		address = addresses[0]
	}
	target := fmt.Sprint(address, ":", port)
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	defer cancel()
	return grpc.DialContext(ctx, target, append(opt.getDialOptions(), grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.DefaultConfig, MinConnectTimeout: MinConnectTimeout}),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxSendMsgSize), grpc.MaxCallRecvMsgSize(MaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                KeepAliveTime,
			Timeout:             KeepAliveTimeout,
			PermitWithoutStream: false,
		}),
		grpc.WithBlock(), //此处添加block防止因为grpc address不可用导致的无限重试问题)
	)...)
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
