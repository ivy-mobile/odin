package tcp

import (
	"net"
	"sync/atomic"

	"github.com/ivy-mobile/odin/xutil/xnet"
)

type client struct {
	opts              *clientOptions         // 配置
	id                int64                  // 连接ID
	connectHandler    xnet.ConnectHandler    // 连接打开hook函数
	disconnectHandler xnet.DisconnectHandler // 连接关闭hook函数
	receiveHandler    xnet.ReceiveHandler    // 接收消息hook函数
}

var _ xnet.Client = &client{}

func NewClient(opts ...ClientOption) xnet.Client {
	o := defaultClientOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &client{opts: o}
}

// Dial 拨号连接
func (c *client) Dial(addr ...string) (xnet.Conn, error) {
	var address string
	if len(addr) > 0 && addr[0] != "" {
		address = addr[0]
	} else {
		address = c.opts.addr
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout(tcpAddr.Network(), tcpAddr.String(), c.opts.timeout)
	if err != nil {
		return nil, err
	}

	return newClientConn(c, atomic.AddInt64(&c.id, 1), conn), nil
}

// Protocol 协议
func (c *client) Protocol() string {
	return protocol
}

// OnConnect 监听连接打开
func (c *client) OnConnect(handler xnet.ConnectHandler) {
	c.connectHandler = handler
}

// OnDisconnect 监听连接关闭
func (c *client) OnDisconnect(handler xnet.DisconnectHandler) {
	c.disconnectHandler = handler
}

// OnReceive 监听接收到消息
func (c *client) OnReceive(handler xnet.ReceiveHandler) {
	c.receiveHandler = handler
}
