package game

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/eventbus"
	"github.com/ivy-mobile/odin/message"
	"github.com/ivy-mobile/odin/xutil/xos"

	msgjson "github.com/ivy-mobile/odin/message/json"
	msgproto "github.com/ivy-mobile/odin/message/proto"

	"sync"

	"github.com/ivy-mobile/odin/player"
	"github.com/ivy-mobile/odin/room"
	"github.com/ivy-mobile/odin/topic"
	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
)

const GatewayName = "game-gateway"
const defaultChanSize = 1024

// Game 游戏封装
type Game struct {
	ctx    context.Context
	cancel context.CancelFunc

	opts *options
	mux  sync.RWMutex

	routeHandlers map[string]GameMessageHandler // 路由处理器 key=version.route

	gateChan chan []byte // 网关消息通道

	rooms   *room.Manager   // 房间管理器
	players *player.Manager // 玩家管理器
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
		return fmt.Errorf("game id is empty")
	}
	if g.opts.name == "" {
		return fmt.Errorf("game name is empty")
	}
	if g.opts.codec == nil {
		return fmt.Errorf("game message codec is nil")
	}
	return nil
}

// Start 启动游戏
func (g *Game) Start() {

	// 1. 验证选项
	if err := g.validateOptions(); err != nil {
		panic(fmt.Errorf("game start failed, validate options, err: %w", err))
	}

	// 2. 监听外部消息
	g.listenMessage()

	// 3. 等待系统信号
	xos.WaitSysSignal(func(sig os.Signal) {
		xlog.Info().Msgf("game %s received signal: %v, exiting...", g.opts.name, sig)
	})

	// 4. 释放资源
	g.shutdown()
}

// RegisterRouter 注册路由
func (g *Game) RegisterRouter(version, route string, handler GameMessageHandler) {
	g.routeHandlers[routeKey(version, route)] = handler
}

// 网关消息写入通道
func (g *Game) writeGateMsg(msg []byte) {
	g.gateChan <- msg
}

// 模拟接收网关数据
func (g *Game) MockReciveGateMessage(msg []byte) {
	g.writeGateMsg(msg)
}

// 模拟接收网关数据
func (g *Game) MockReciveGateMessagex(msg []byte) {
	err := g.EventBus().Publish(context.Background(), topic.Gate2GameTopic("game-gateway", "test"), msg)
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
	handler(g, msg)
}

// 从事件总线中订阅消息
func (g *Game) listenMessage() {

	// 订阅网关消息
	g.subscribeGateMessage()

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
	topic := topic.Gate2GameTopic(GatewayName, g.opts.name)
	err := eb.Subscribe(context.Background(), topic, func(data []byte) {
		g.writeGateMsg(data)
	})
	if err != nil {
		xlog.Error().Msgf("[listenGateMessage] failed, subscribe topic: %s, err: %s", topic, err.Error())
	}
}

// 订阅后台指令消息
func (g *Game) subscribeAdminCmdMessage() {

	eb := g.opts.eventbus
	if eb == nil {
		xlog.Error().Msgf("[listenAdminCmdMessage] failed, eventbus is nil")
		return
	}
	topic := topic.Admin2GameTopic(g.opts.name)
	// 订阅
	err := eb.Subscribe(context.Background(), topic, func(data []byte) {
		if g.opts.adminCmdHandler != nil {
			g.opts.adminCmdHandler(data)
		}
	})
	if err != nil {
		xlog.Error().Msgf("[listenAdminCmdMessage] failed, subscribe topic: %s, err: %s", topic, err.Error())
	}
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
			return fmt.Errorf("marshal json payload failed, err: %w", err)
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
			return fmt.Errorf("marshal json message failed, err: %w", err)
		}
	case proto.Name:
		data, err := proto.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal proto payload failed, err: %w", err)
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
			return fmt.Errorf("marshal proto message failed, err: %w", err)
		}
	default:
		return fmt.Errorf("codec not support, codec: %s", g.opts.codec.Name())
	}

	return g.opts.eventbus.Publish(g.ctx, topic.Game2GateTopic(GatewayName, g.opts.name), bytes)
}

func routeKey(version, route string) string {
	if version == "" {
		return route
	}
	return version + "." + route
}
