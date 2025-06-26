package gate

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
	"github.com/ivy-mobile/odin/xutil/xos"

	"github.com/olahol/melody"
)

// Gate websocket网关
// desc: 区别于流量网关，本网关侧重于业务，可理解为业务网关，服务统一的用户链接管理和消息转发
type Gate struct {
	// options
	opts *options
	// websocket server
	wsServer *melody.Melody
	// http server
	httpServer *http.Server
	// 会话管理器
	sessions *Sessions
}

func New(opts ...Option) *Gate {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Gate{
		opts:     o,
		wsServer: melody.New(),
		sessions: NewSessions(),
	}
}

// Name 组件名称
func (g *Gate) Name() string {
	return g.opts.name
}

// InstanceID 实例ID
// 实例ID = 组件名称 + 实例ID
func (g *Gate) InstanceID() string {
	return fmt.Sprintf("%s-%s", g.opts.name, g.opts.id)
}

// Start 启动
func (g *Gate) Start() {

	// 验证选项
	if err := g.validateOptions(); err != nil {
		xlog.Error().Msgf("start faild: %s", err.Error())
		return
	}
	// 启动websocket服务
	sign := <-g.startWsServer()
	if sign != nil {
		xlog.Error().Msgf("start faild: %s", sign.Error())
		return
	}

	xlog.Info().Msgf("websocket server started success ... on: %s", g.opts.port+g.opts.pattern)

	// 等待信号
	xos.WaitSysSignal(func(s os.Signal) {
		xlog.Info().Msgf("Received signal: %s, shutting down server...", s.String())
	})
	// 释放资源
	g.shutdown()
}

// validateOptions 验证选项
func (g *Gate) validateOptions() error {

	if g.opts.id == "" {
		return fmt.Errorf("gate id is empty")
	}
	if g.opts.name == "" {
		return fmt.Errorf("gate name is empty")
	}
	if g.opts.port == "" {
		return fmt.Errorf("gate port is empty")
	}
	if g.opts.codec == nil {
		return fmt.Errorf("codec is unset, use WithCodec(...) to set")
	}
	if !strings.HasPrefix(g.opts.pattern, "/") {
		return fmt.Errorf("gate pattern must start with /")
	}
	if g.wsServer == nil {
		return fmt.Errorf("websocket server is nil")
	}
	return nil
}

// 关闭服务
func (g *Gate) shutdown() {

	// 创建一个带超时的上下文用于关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 关闭http
	if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
		xlog.Error().Msgf("Failed to shutdown HTTP server")
	}
	// 关闭websocket [melody]
	if err := g.wsServer.Close(); err != nil {
		xlog.Error().Msgf("Failed to close Websocket server")
	}
	xlog.Info().Msgf("Gate Server shutdown completed")
}

// 启动websocket服务
func (g *Gate) startWsServer() chan error {

	startSign := make(chan error, 1)

	g.wsServer.HandleConnect(g.handleConnect)
	g.wsServer.HandleDisconnect(g.handleDisconnect)
	g.wsServer.HandleMessage(g.handleMessage)
	g.wsServer.HandleMessageBinary(g.handleMessageBinary)

	// 创建 http server
	g.httpServer = &http.Server{
		Addr: g.opts.port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := g.wsServer.HandleRequest(w, r); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}),
	}

	xgo.Go(func() {

		xlog.Info().Msgf("Gate server starting on %s", g.opts.port)

		listener, err := net.Listen("tcp", g.opts.port)
		if err != nil {
			startSign <- fmt.Errorf("server start failed: %w", err)
			return
		}

		// 通知服务器已成功启动
		startSign <- nil

		// Serve 会一直在协程内阻塞，直到服务器关闭
		if err := g.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			startSign <- fmt.Errorf("server start failed: %w", err)
			return
		}
	})

	return startSign
}
