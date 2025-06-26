package tcp

import (
	"net"
	"time"

	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
	"github.com/ivy-mobile/odin/xutil/xnet"
	"github.com/ivy-mobile/odin/xutil/xos"
)

type server struct {
	opts              *serverOptions         // 配置
	listener          net.Listener           // 监听器
	connMgr           *serverConnMgr         // 连接管理器
	startHandler      xnet.StartHandler      // 服务器启动hook函数
	stopHandler       xnet.CloseHandler      // 服务器关闭hook函数
	connectHandler    xnet.ConnectHandler    // 连接打开hook函数
	disconnectHandler xnet.DisconnectHandler // 连接关闭hook函数
	receiveHandler    xnet.ReceiveHandler    // 接收消息hook函数
}

var _ xnet.Server = &server{}

func NewServer(opts ...ServerOption) xnet.Server {
	o := defaultServerOptions()
	for _, opt := range opts {
		opt(o)
	}

	s := &server{}
	s.opts = o
	s.connMgr = newServerConnMgr(s)

	return s
}

// Addr 监听地址
func (s *server) Addr() string {
	return s.opts.addr
}

// Start 启动服务器
func (s *server) Start() error {
	if err := s.init(); err != nil {
		return err
	}

	if s.startHandler != nil {
		s.startHandler()
	}

	// core
	xgo.Go(s.serve)

	// 优雅关闭
	xos.WaitSysSignal()

	return nil
}

// Stop 关闭服务器
func (s *server) Stop() error {
	if err := s.listener.Close(); err != nil {
		return err
	}

	s.connMgr.close()

	if s.stopHandler != nil {
		s.stopHandler()
	}

	return nil
}

// Protocol 协议
func (s *server) Protocol() string {
	return protocol
}

// OnStart 监听服务器启动
func (s *server) OnStart(handler xnet.StartHandler) {
	s.startHandler = handler
}

// OnStop 监听服务器关闭
func (s *server) OnStop(handler xnet.CloseHandler) {
	s.stopHandler = handler
}

// OnConnect 监听连接打开
func (s *server) OnConnect(handler xnet.ConnectHandler) {
	s.connectHandler = handler
}

// OnDisconnect 监听连接关闭
func (s *server) OnDisconnect(handler xnet.DisconnectHandler) {
	s.disconnectHandler = handler
}

// OnReceive 监听接收到消息
func (s *server) OnReceive(handler xnet.ReceiveHandler) {
	s.receiveHandler = handler
}

// 初始化TCP服务器
func (s *server) init() error {
	addr, err := net.ResolveTCPAddr("tcp", s.opts.addr)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		return err
	}

	s.listener = ln

	return nil
}

// 等待连接
func (s *server) serve() {
	var tempDelay time.Duration

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				xlog.Warn().Msgf("tcp accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}

			xlog.Warn().Msgf("tcp accept error: %v", err)
			return
		}

		tempDelay = 0

		if err = s.connMgr.allocate(conn); err != nil {
			xlog.Error().Msgf("connection allocate error: %v", err)
			_ = conn.Close()
		}
	}
}
