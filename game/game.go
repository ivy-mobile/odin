package game

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/enum"
	"github.com/ivy-mobile/odin/eventbus"
	"github.com/ivy-mobile/odin/message"
	"github.com/ivy-mobile/odin/registry"
	"github.com/ivy-mobile/odin/xutil/xnet"
	"github.com/ivy-mobile/odin/xutil/xos"

	msgjson "github.com/ivy-mobile/odin/message/json"
	msgproto "github.com/ivy-mobile/odin/message/proto"

	"sync"

	"github.com/ivy-mobile/odin/player"
	"github.com/ivy-mobile/odin/room"
	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
)

const defaultChanSize = 1024

// Game 游戏封装
type Game struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts *options
	mux  sync.RWMutex

	routeHandlers map[string]GameMessageHandler // 路由处理器 key=version.route

	gateChan chan []byte // 网关消息通道

	rooms    *room.Manager   // 房间管理器
	players  *player.Manager // 玩家管理器
	subGates sync.Map        // 订阅网关列表
}

// New 创建游戏
func New(opts ...Option) *Game {

	ctx, cancel := context.WithCancel(context.Background())

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Game{
		ctx:           ctx,
		cancel:        cancel,
		opts:          o,
		routeHandlers: make(map[string]GameMessageHandler),
		rooms:         room.NewManager(),
		players:       player.NewManager(),
		gateChan:      make(chan []byte, defaultChanSize),
	}
}

// 验证选项
func (g *Game) validateOptions() error {
	if g.opts.id == "" {
		return errors.New("game id is empty")
	}
	if g.opts.name == "" {
		return errors.New("game name is empty")
	}
	if g.opts.codec == nil {
		return errors.New("game message codec is nil")
	}
	if g.opts.eventbus == nil {
		return errors.New("game eventbus is nil")
	}
	if g.opts.registry == nil {
		return errors.New("game registry is nil")
	}
	if g.opts.serviceName == "" {
		return errors.New("game public service name is empty")
	}
	return nil
}

// Start 启动游戏
func (g *Game) Start() {

	defer g.shutdown()

	// 1. 验证选项
	if err := g.validateOptions(); err != nil {
		xlog.Error().Msgf("game start failed, validate options, err: %v", err)
		return
	}

	// 2. 监听外部消息
	g.listenMessage()

	// 3. 注册服务
	if err := g.registerService(); err != nil {
		xlog.Error().Msgf("game start failed, register service, err: %v", err)
		return
	}

	// 4. 监听网关服务
	if err := g.watchGateService(); err != nil {
		xlog.Error().Msgf("game start failed, watch gate service, err: %v", err)
		return
	}

	xlog.Error().Msgf("Game start success, nodeId: %s, nodeName: %s, serviceName: %s", g.opts.id, g.opts.name, g.opts.serviceName)

	// 5. 等待系统信号
	xos.WaitSysSignal(func(sig os.Signal) {
		xlog.Info().Msgf("game %s received signal: %v, exiting...", g.opts.name, sig)
	})

}

// RegisterRouter 注册路由
func (g *Game) RegisterRouter(version, route string, handler GameMessageHandler) {
	g.routeHandlers[routeKey(version, route)] = handler
}

// RegisterCmdHandler 注册后台指令处理器
func (g *Game) RegisterCmdHandler(cmdHandler CmdMessageHandler) {
	g.opts.adminCmdHandler = cmdHandler
}

// 网关消息写入通道
func (g *Game) writeGateMsg(msg []byte) {
	g.gateChan <- msg
}

// MockReceiveGateMessage 模拟接收网关数据
func (g *Game) MockReceiveGateMessage(msg []byte) {
	err := g.EventBus().Publish(context.Background(), enum.Gate2GameTopic(g.opts.gateServiceName, "test"), msg)
	if err != nil {
		fmt.Printf("Publish error: %v", err)
		return
	}
}

// 处理网关消息
func (g *Game) loopGateMsg() {
	xgo.Go(func() {
		for {
			select {
			case <-g.ctx.Done():
				return
			case msg := <-g.gateChan:
				g.handlerGateMessage(msg)
			}
		}
	})
}

// 处理网关消息
func (g *Game) handlerGateMessage(data []byte) {

	var msg message.Message
	codec := g.opts.codec
	switch codec.Name() {
	case json.Name:
		var jsonMsg msgjson.JsonMessage
		if err := json.Unmarshal(data, &jsonMsg); err != nil {
			xlog.Error().Msgf("[handlerGateMessage] failed, unmarshal faild, err: %s", err.Error())
			return
		}
		msg = &jsonMsg
	case proto.Name:
		var protoMsg msgproto.Message
		if err := proto.Unmarshal(data, &protoMsg); err != nil {
			xlog.Error().Msgf("[handlerGateMessage] failed, unmarshal faild, err: %s", err.Error())
			return
		}
		msg = &protoMsg
	default:
		xlog.Error().Msgf("[handlerGateMessage] failed, codec not support, codec: %s", codec.Name())
		return
	}

	route, version := msg.GetRoute(), msg.GetVersion()
	if route == "" {
		xlog.Error().Msgf("[handlerGateMessage] failed, route is empty")
		return
	}
	handler, ok := g.routeHandlers[routeKey(version, route)]
	if !ok {
		xlog.Error().Msgf("[handlerGateMessage] failed, not found handler, route: %s", routeKey(version, route))
		return
	}
	err := handler(g, msg)
	if err != nil {
		xlog.Error().Msgf("[handlerGateMessage] failed, handler route: %s, err: %s", routeKey(version, route), err.Error())
	}
}

// 从事件总线中订阅消息
func (g *Game) listenMessage() {

	// 订阅网关消息 - 移至watchGateService中实现(从服务注册中心动态订阅)
	//g.subscribeGateMessage()

	// 订阅后台指令消息
	g.subscribeAdminCmdMessage()

	// 处理网关消息
	g.loopGateMsg()
}

// 订阅网关消息
func (g *Game) subscribeGateMessage() {

	eb := g.EventBus()
	if eb == nil {
		xlog.Error().Msgf("[listenGateMessage] failed, eventbus is nil")
		return
	}
	topic := enum.Gate2GameTopic(g.opts.gateServiceName, g.opts.name)
	err := eb.Subscribe(context.Background(), topic, g.writeGateMsg)
	if err != nil {
		xlog.Error().Msgf("[listenGateMessage] failed, subscribe topic: %s, err: %s", topic, err.Error())
		return
	}
	xlog.Info().Msgf("[subscribeGateMessage] success, subscribe topic: %s", topic)
}

// 订阅后台指令消息
func (g *Game) subscribeAdminCmdMessage() {

	eb := g.opts.eventbus
	if eb == nil {
		xlog.Error().Msgf("[listenAdminCmdMessage] failed, eventbus is nil")
		return
	}
	topic := enum.Admin2GameTopic(g.opts.name)
	// 订阅
	err := eb.Subscribe(context.Background(), topic, func(data []byte) {
		if g.opts.adminCmdHandler != nil {
			g.opts.adminCmdHandler(g, data)
		}
	})
	if err != nil {
		xlog.Error().Msgf("[listenAdminCmdMessage] failed, subscribe topic: %s, err: %s", topic, err.Error())
		return
	}
	xlog.Info().Msgf("[subscribeAdminCmdMessage] success, subscribe topic: %s", topic)
}

// 关闭游戏
func (g *Game) shutdown() {
	// 1. 取消上下文
	g.cancel()
	// 2. 关闭事件总线
	if g.opts.eventbus != nil {
		err := g.opts.eventbus.Close()
		if err != nil {
			xlog.Error().Msgf("[shutdown] failed, close eventbus, err: %s", err.Error())
		}
	}
	// 3. 关闭网关消息通道
	close(g.gateChan)

	xlog.Info().Msgf("Game Server shutdown completed")
}

// RoomManager 房间管理器
func (g *Game) RoomManager() *room.Manager {
	return g.rooms
}

// PlayerManager 玩家管理器
func (g *Game) PlayerManager() *player.Manager {
	return g.players
}

// EventBus 事件总线
func (g *Game) EventBus() eventbus.Eventbus {
	return g.opts.eventbus
}

// SendMessage 发送消息至Gate
func (g *Game) SendMessage(seq uint64, uid int64, route, version string, msgID uint64, payload any) error {

	var bytes []byte

	switch g.opts.codec.Name() {
	case json.Name:
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal json payload failed, err: %v", err)
		}
		msg := msgjson.JsonMessage{
			Seq:       seq,
			Uid:       uid,
			Route:     route,
			Game:      g.opts.name,
			MsgID:     msgID,
			Timestamp: time.Now().UnixMilli(),
			Version:   version,
			Payload:   data,
		}
		bytes, err = json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal json message failed, err: %v", err)
		}
	case proto.Name:
		data, err := proto.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal proto payload failed, err: %v", err)
		}
		msg := &msgproto.Message{
			Seq:       seq,
			Uid:       uid,
			Route:     route,
			Game:      g.opts.name,
			MsgId:     msgID,
			Version:   version,
			Timestamp: time.Now().UnixMilli(),
			Payload:   data,
		}
		bytes, err = proto.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal proto message failed, err: %v", err)
		}
	default:
		return fmt.Errorf("codec not support, codec: %s", g.opts.codec.Name())
	}

	return g.opts.eventbus.Publish(g.ctx, enum.Game2GateTopic(g.opts.gateServiceName, g.opts.name), bytes)
}

func routeKey(version, route string) string {
	if version == "" {
		return route
	}
	return version + "." + route
}

// 注册服务
func (g *Game) registerService() error {
	host, err := xnet.ExternalIP()
	if err != nil {
		return fmt.Errorf("[registerService] get external ip failed: %v", err)
	}
	return g.opts.registry.Register(g.ctx, &registry.ServiceInstance{
		ID:       g.opts.id,
		Name:     g.opts.serviceName,
		Alias:    g.opts.name,
		Kind:     enum.NodeType_Game,
		Endpoint: fmt.Sprintf("http://%s:8888", host), // 游戏服不暴露对外接口,固定任意值即可
		State:    enum.NodeState_Work,
		Weight:   100, // Todo
	})
}

// 监听网关服务
func (g *Game) watchGateService() error {

	reg := g.opts.registry
	if reg == nil {
		return fmt.Errorf("watch game service failed, registry is nil")
	}
	// 所有Game服务节点
	services, err := reg.Services(g.ctx, g.opts.gateServiceName)
	if err != nil {
		return fmt.Errorf("watch gate service failed, GetServices err: %v", err)
	}
	// 订阅所有Game服务节点
	for _, service := range services {
		xlog.Info().Msgf("[watchGameService] service: %s", enum.GateNodeName(service.Name, service.ID))
		g.subscribeGate(service.Name, service.ID)
	}
	// 监听Gate服务
	w, err := reg.Watch(g.ctx, g.opts.gateServiceName)
	if err != nil {
		return fmt.Errorf("watch game service failed, Watch err: %v", err)
	}
	xgo.Go(func() {
		for {
			services, err := w.Next()
			if err != nil {
				xlog.Error().Msgf("[watchGameService] watch gate service failed, Next err: %v", err)
				continue
			}
			for _, service := range services {
				// 更新订阅信息s
				xlog.Info().Msgf("[watchGameService] gates service: %s changed!", enum.GateNodeName(service.Name, service.ID))
				g.subscribeGate(service.Name, service.ID)
			}
			// 清理无效的旧订阅信息
			g.cleanSubscribe(services)
		}
	})
	return nil
}

// 订阅网关服务
func (g *Game) subscribeGate(gateServiceName, id string) {

	// 当前网关节点
	game := enum.GameNodeName(g.opts.serviceName, g.opts.id, g.opts.name)
	// 目标游戏节点
	gate := enum.GateNodeName(gateServiceName, id)
	// topic
	topic := enum.Gate2GameTopic(gate, game)
	// 订阅
	err := g.opts.eventbus.Subscribe(context.Background(), topic, g.writeGateMsg)
	if err != nil {
		xlog.Error().Msgf("[subscribeGame] failed, subscribe topic: %s, err: %s", topic, err.Error())
		return
	}
	g.subGates.Store(topic, struct{}{})
	xlog.Info().Msgf("[subscribeGame] success, subscribe topic: %s", topic)
}

// 清理订阅信息
// svs 为最新的Gate服务实例列表
func (g *Game) cleanSubscribe(svs []*registry.ServiceInstance) {
	game := enum.GameNodeName(g.opts.serviceName, g.opts.id, g.opts.name)

	// 最新的所有Gate服务实例对应topic
	newTopics := make([]string, 0, len(svs))
	for _, sv := range svs {
		gate := enum.GateNodeName(g.opts.gateServiceName, sv.ID)
		tp := enum.Gate2GameTopic(gate, game)
		newTopics = append(newTopics, tp)
	}
	// 已有的topic如果不在新的topic列表中，则取消订阅
	g.subGates.Range(func(key, value interface{}) bool {
		topic := key.(string)
		if !slices.Contains(newTopics, topic) {
			if err := g.opts.eventbus.Unsubscribe(context.Background(), topic); err != nil {
				xlog.Error().Msgf("[cleanSubscribe] unsubscribe topic: %s, err: %s", topic, err.Error())
			} else {
				g.subGates.Delete(topic)
				xlog.Info().Msgf("[cleanSubscribe] unsubscribe topic: %s, success", topic)
			}
		}
		return true
	})

}
