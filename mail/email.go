package mail

import (
	"errors"
	"fmt"
	"net/smtp"
	"sync"
)

const defaultRFC822 = `From: %s
To: %s
Subject: %s

%s`

type AuthMethod int

const (
	MethodPlainAuth = AuthMethod(iota)
	MethodLoginAuth
)

var clients sync.Map //map[string]*Client

type Client struct {
	opt *Option
}

type Option struct {
	AliasName  string
	Address    string     //smtp服务器地址
	AuthMethod AuthMethod //身份验证方式
	Auth       Auth
	// auth       smtp.Auth
}

type Auth struct {
	Identity string
	Username string
	Password string
	Host     string
	Secret   string
}

// New 创建Email客户端
func New(opt ...*Option) {
	for _, o := range opt {
		clients.Store(o.getAliseName(), &Client{opt: o})
	}
}

// Cli 使用 AliseName 获取发送Client
func Cli(aliseName ...string) *Client {
	if len(aliseName) == 0 {
		aliseName = append(aliseName, "default")
	}
	v, ok := clients.Load(aliseName[0])
	if ok {
		return v.(*Client)
	}
	return nil
}

// 可以使用 RFC822 来组织默认格式化的邮件
func (c *Client) Send(from string, to []string, msg []byte) error {
	return smtp.SendMail(c.opt.Address, c.opt.getAuth(), from, to, msg)
}

// RFC822 组装标准 RFC822 格式邮件内容
// 规定：
// to 必须是收件人的地址
func RFC822(from, to, subject, body string) []byte {
	return []byte(fmt.Sprintf(defaultRFC822, from, to, subject, body))
}

func (o *Option) getAliseName() string {
	if o.AliasName == "" {
		o.AliasName = "default"
	}
	return o.AliasName
}

func (c *Client) Address() string {
	return c.opt.Auth.Username
}

type loginAuth struct {
	username, password string
}

func (l *loginAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	return "LOGIN", []byte(l.username), nil
}

func (l *loginAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(l.username), nil
		case "Password:":
			return []byte(l.password), nil
		default:
			return nil, errors.New("unknown from server")
		}
	}
	return nil, nil
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (o *Option) getAuth() smtp.Auth {
	switch o.AuthMethod {
	case MethodPlainAuth:
		return smtp.PlainAuth(o.Auth.Identity, o.Auth.Username, o.Auth.Password, o.Auth.Host)
	case MethodLoginAuth:
		return LoginAuth(o.Auth.Username, o.Auth.Password)
	default:
		return nil
	}
}
