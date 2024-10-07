// Package grpc GRPC操作封装
package grpc

import "google.golang.org/grpc"

// Conn GRPC连接结构体
type Conn interface {
	Conn() *grpc.ClientConn
	Close() error
}

type conn struct {
	cli  *grpc.ClientConn
	pool *pool
	once bool
}

// Conn 返回GRPC cli
func (c *conn) Conn() *grpc.ClientConn {
	return c.cli
}

// Close 关闭连接池并释放连接
func (c *conn) Close() error {
	c.pool.decrRef()
	if c.once {
		return c.reset()
	}
	return nil
}

func (c *conn) reset() error {
	cc := c.cli
	c.cli = nil
	c.once = false
	if cc != nil {
		return cc.Close()
	}
	return nil
}

func (p *pool) wrapConn(cli *grpc.ClientConn, once bool) *conn {
	return &conn{
		cli:  cli,
		pool: p,
		once: once,
	}
}
