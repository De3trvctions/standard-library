package grpc

import (
	"errors"
	"time"
)

const (
	// DialTimeout 创建连接超时时间。
	DialTimeout = 5 * time.Second

	// BackoffMaxDelay 在连接尝试失败后退出时提供最大延迟。
	BackoffMaxDelay = 120 * time.Second

	// MinConnectTimeout 最小连接超时时间
	MinConnectTimeout = 20 * time.Second

	// KeepAliveTime 一段时间后，如果客户端没有看到任何活动，会 ping 服务器以查看传输是否仍然有效。
	KeepAliveTime = 20 * time.Second

	// KeepAliveTimeout 等待客户端 ping 返回的时间。
	KeepAliveTimeout = 3 * time.Second

	// InitialWindowSize 设置总吞吐量1GB。
	InitialWindowSize = 1 << 30

	// InitialConnWindowSize 设置链接吞吐量1GB.
	InitialConnWindowSize = 1 << 30

	// MaxSendMsgSize 设置发送到服务器的最大 gRPC 请求消息大小。
	// 如果任何请求消息大小大于当前值，gRPC 将报错。
	MaxSendMsgSize = 4 << 30

	// MaxRecvMsgSize 设置从服务器接收到的最大 gRPC 接收消息大小。
	// 如果任何请求消息大小大于当前值，gRPC 将报错。
	MaxRecvMsgSize = 4 << 30

	// RecycleDuration 回收间隔时间
	RecycleDuration = 10 * time.Minute

	//ReconnectDuration 重连间隔时间
	ReconnectDuration = 30 * time.Second
)

// ErrClosed Close grpc连接时的报错。
var ErrClosed = errors.New("pool is closed")
var ErrNoOption = errors.New("no option to connect")
var ErrReconnectCD = errors.New("reconnect colling down")
