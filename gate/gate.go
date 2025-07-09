package gate

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/enum"
	msgjson "github.com/ivy-mobile/odin/message/json"
	msgproto "github.com/ivy-mobile/odin/message/proto"
	"github.com/ivy-mobile/odin/registry"
	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
	"github.com/ivy-mobile/odin/xutil/xnet"
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
	// 订阅游戏列表
	subGames sync.Map
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

func (g *Gate) ID() string {
	return g.opts.id
}

// Name 组件名称
func (g *Gate) Name() string {
	return g.opts.name
}

// GameServiceName 游戏服务名称
func (g *Gate) GameServiceName() string {
	return g.opts.gameServiceName
}

// Port 端口
func (g *Gate) Port() string {
	return g.opts.port
}

// Pattern 路由路径
func (g *Gate) Pattern() string {
	return g.opts.pattern
}

// Start 启动
func (g *Gate) Start() {

	// 1. 验证选项
	if err := g.validateOptions(); err != nil {
		xlog.Error().Msgf("start faild: %s", err.Error())
		return
	}

	// 2. 注册服务
	if err := g.registerService(); err != nil {
		xlog.Error().Msgf("register service faild: %s", err.Error())
		return
	}
	// 3. 监听服务
	if err := g.watchGameService(); err != nil {
		xlog.Error().Msgf("watch game service faild: %s", err.Error())
		return

	}

	// 3. 启动websocket服务
	sign := <-g.startWsServer()
	if sign != nil {
		xlog.Error().Msgf("start faild: %s", sign.Error())
		return
	}

	xlog.Info().Msgf("websocket server started success ... on: %s", g.Port()+g.Pattern())

	// 4. 等待信号
	xos.WaitSysSignal(func(s os.Signal) {
		xlog.Info().Msgf("Received signal: %s, shutting down server...", s.String())
	})
	// 5. 释放资源
	g.shutdown()
}

// validateOptions 验证选项
func (g *Gate) validateOptions() error {

	if g.ID() == "" {
		return fmt.Errorf("gate id is empty")
	}
	if g.Name() == "" {
		return fmt.Errorf("gate name is empty")
	}
	if g.Port() == "" {
		return fmt.Errorf("gate port is empty")
	}
	if g.opts.codec == nil {
		return fmt.Errorf("codec is unset, use WithCodec(...) to set")
	}
	if !strings.HasPrefix(g.Pattern(), "/") {
		return fmt.Errorf("gate pattern must start with /")
	}
	if g.wsServer == nil {
		return fmt.Errorf("websocket server is nil")
	}
	if g.opts.registry == nil {
		return fmt.Errorf("registry is unset, use WithRegistry(...) to set")
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
	// 关闭事件总线
	if err := g.opts.eventbus.Close(); err != nil {
		xlog.Error().Msgf("Failed to close eventbus")
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
		Addr: g.Port(),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := g.wsServer.HandleRequest(w, r); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}),
	}

	xgo.Go(func() {

		xlog.Info().Msgf("Gate server starting on %s", g.Port())

		listener, err := net.Listen("tcp", g.Port())
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

// 注册服务
func (g *Gate) registerService() error {
	host, err := xnet.ExternalIP()
	if err != nil {
		return fmt.Errorf("get external ip failed: %w", err)
	}
	return g.opts.registry.Register(g.opts.ctx, &registry.ServiceInstance{
		ID:       g.ID(),
		Name:     g.Name(),
		Kind:     enum.NodeType_Gate,
		Endpoint: fmt.Sprintf("ws://%s%s%s", host, g.Port(), g.Pattern()),
		State:    enum.NodeState_Work,
		Weight:   100,
	})
}

// 监听游戏服务
// 所有游戏注册到服务中心的服务名都统一(如: game-service)，游戏之间根据alias别名进行区分.
func (g *Gate) watchGameService() error {

	reg := g.opts.registry
	if reg == nil {
		return fmt.Errorf("watch game service failed, registry is nil")
	}
	// 所有Game服务节点
	services, err := reg.Services(g.opts.ctx, g.GameServiceName())
	if err != nil {
		return fmt.Errorf("watch game service failed, GetServices err: %v", err)
	}
	// 订阅所有Game服务节点
	for _, service := range services {
		xlog.Info().Msgf("[watchGameService] service: %s", enum.GameNodeName(service.Name, service.ID, service.Alias))
		g.subscribeGame(service.Name, service.ID, service.Alias)
	}
	// 监听Game服务
	w, err := reg.Watch(g.opts.ctx, g.GameServiceName())
	if err != nil {
		return fmt.Errorf("watch game service failed, Watch err: %v", err)
	}
	xgo.Go(func() {
		for {
			services, err := w.Next()
			if err != nil {
				xlog.Error().Msgf("[watchGameService] watch game service failed, Next err: %v", err)
				continue
			}
			for _, service := range services {
				// 更新订阅信息
				xlog.Info().Msgf("[watchGameService] game service: %s changed!", enum.GameNodeName(service.Name, service.ID, service.Alias))
				g.subscribeGame(service.Name, service.ID, service.Alias)
			}
			// 清理无效的旧订阅信息
			g.cleanSubscribe(services)
		}
	})
	return nil
}

// 订阅游戏服务
func (g *Gate) subscribeGame(gameServiceName, id, alias string) {

	// 当前网关节点
	gate := enum.GateNodeName(g.Name(), g.ID())
	// 目标游戏节点
	game := enum.GameNodeName(gameServiceName, id, alias)
	// topic
	topic := enum.Game2GateTopic(gate, game)
	// 订阅
	err := g.opts.eventbus.Subscribe(context.Background(), topic, g.dispatchToSession)
	if err != nil {
		xlog.Error().Msgf("[subscribeGame] failed, subscribe topic: %s, err: %s", topic, err.Error())
		return
	}
	g.subGames.Store(topic, struct{}{})

	xlog.Info().Msgf("[subscribeGame] success, subscribe topic: %s", topic)
}

// 分发消息到玩家
func (g *Gate) dispatchToSession(data []byte) {

	codec := g.opts.codec
	switch codec.Name() {
	case json.Name:
		var jsonMsg msgjson.JsonMessage
		if err := json.Unmarshal(data, &jsonMsg); err != nil {
			xlog.Error().Msgf("[dispatchToSession] failed, unmarshal faild, err: %s", err.Error())
			return
		}
		g.sessions.SendText(data, jsonMsg.GetUid())
	case proto.Name:
		var protoMsg msgproto.Message
		if err := proto.Unmarshal(data, &protoMsg); err != nil {
			xlog.Error().Msgf("[dispatchToSession] failed, unmarshal faild, err: %s", err.Error())
			return
		}
		g.sessions.Send(data, protoMsg.GetUid())
	}
}

// 清理订阅信息
// svs 为最新的Game服务实例列表
func (g *Gate) cleanSubscribe(svs []*registry.ServiceInstance) {
	gate := enum.GateNodeName(g.Name(), g.ID())

	// 最新的所有Game服务实例对应topic
	newTopics := make([]string, 0, len(svs))
	for _, sv := range svs {
		game := enum.GameNodeName(g.GameServiceName(), sv.ID, sv.Alias)
		tp := enum.Game2GateTopic(gate, game)
		newTopics = append(newTopics, tp)
	}
	// 已有的topic如果不在新的topic列表中，则取消订阅
	g.subGames.Range(func(key, value interface{}) bool {
		topic := key.(string)
		if !slices.Contains(newTopics, topic) {
			if err := g.opts.eventbus.Unsubscribe(context.Background(), topic); err != nil {
				xlog.Error().Msgf("[cleanSubscribe] unsubscribe topic: %s, err: %s", topic, err.Error())
			} else {
				g.subGames.Delete(topic)
				xlog.Info().Msgf("[cleanSubscribe] unsubscribe topic: %s, success", topic)
			}
		}
		return true
	})

}
