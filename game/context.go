package game

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ivy-mobile/odin/message"
	"github.com/ivy-mobile/odin/player"
	"github.com/ivy-mobile/odin/room"
)

var ctxPool = sync.Pool{
	New: func() any {
		return &defaultContext{}
	},
}

// Context 游戏上下文
type Context interface {

	// Seq 消息序列
	Seq() uint64
	// MsgID 消息ID
	MsgID() uint64
	// Uid 消息中用户ID
	Uid() int64
	// Route 消息中路由ID
	Route() string
	// Game 消息中游戏服务名
	Game() string
	// Version 版本号
	Version() string
	// Timestamp 消息发送时间戳
	Timestamp() int64
	// Player 当前玩家
	Player() player.Player
	// Login 玩家登录，添加Payer到玩家管理器和当前上下文
	Login(p player.Player)
	// Resp 响应消息
	Resp(data any) error
	// Push 推送消息
	Push(seq uint64, route string, msgId uint64, msg any) error
	// GetRoom 根据ID获取房间
	GetRoom(roomId int) (room.Room, bool)
	// CreateRoom 创建房间
	CreateRoom(ro room.Room) error
	// PushToRoom 推送消息至房间内所有玩家
	PushToRoom(seq uint64, route string, msgId uint64, msg any) error

	// Close 关闭上下文
	Close()
}

type defaultContext struct {
	// Game
	g *Game

	// Player 在玩家未登录游戏的情况在 可能会nil
	p player.Player

	// message
	seq       uint64 // 消息序列号
	uid       int64  // 用户ID
	route     string // 路由ID
	game      string // 游戏服务名
	msgID     uint64 // 消息ID
	timestamp int64  // 时间戳
	version   string // 版本号
}

var _ Context = (*defaultContext)(nil)

func newDefaultContext(g *Game, msg message.Message) Context {

	ctx := ctxPool.Get().(*defaultContext)

	p, _ := g.PlayerManager().Get(msg.GetUid())

	ctx.g = g
	ctx.p = p
	ctx.seq = msg.GetSeq()
	ctx.uid = msg.GetUid()
	ctx.route = msg.GetRoute()
	ctx.game = msg.GetGame()
	ctx.msgID = msg.GetMsgId()
	ctx.timestamp = msg.GetTimestamp()
	ctx.version = msg.GetVersion()
	return ctx
}

// Seq 消息序列
func (c *defaultContext) Seq() uint64 {
	return c.seq
}

// Uid 用户ID
func (c *defaultContext) Uid() int64 {
	return c.uid
}

// Player 玩家, 可能为nil(在玩家未登录的情况下)
func (c *defaultContext) Player() player.Player {
	return c.p
}

// Login 玩家登录，添加Payer到玩家管理器和当前上下文
func (c *defaultContext) Login(p player.Player) {
	if p == nil {
		return
	}
	if p.MsgHandler() == nil {
		p.SetMsgHandler(c.g)
	}
	// 添加玩家到玩家管理器
	c.g.PlayerManager().Add(p)
	// 设置玩家到上下文
	c.p = p
}

// GetRoom 根据ID获取房间
func (c *defaultContext) GetRoom(roomId int) (room.Room, bool) {
	return c.g.RoomManager().Get(roomId)
}

// CreateRoom 创建房间
func (c *defaultContext) CreateRoom(ro room.Room) error {
	if ro == nil {
		return errors.New("room is nil")
	}
	_, exist := c.g.RoomManager().Get(ro.ID())
	if exist {
		return fmt.Errorf("room exist, roomId: %v", ro.ID())
	}
	c.g.RoomManager().Add(ro)
	return nil
}

// Route 路由ID
func (c *defaultContext) Route() string {
	return c.route
}

// Game 游戏服务名
func (c *defaultContext) Game() string {
	return c.game
}

// MsgID 消息ID
func (c *defaultContext) MsgID() uint64 {
	return c.msgID
}

// Timestamp 消息发送时间戳
func (c *defaultContext) Timestamp() int64 {
	return c.timestamp
}

// Resp 响应消息
func (c *defaultContext) Resp(data any) error {
	if c.p == nil {
		return errors.New("player not found")
	}
	return c.p.SendMessage(c.Seq(), c.Route(), c.Version(), c.MsgID(), data)
}

// Push 推送消息
func (c *defaultContext) Push(seq uint64, route string, msgId uint64, msg any) error {
	if c.p == nil {
		return errors.New("player not found")
	}
	return c.p.SendMessage(seq, route, c.Version(), msgId, msg)
}

// PushToRoom 推送消息至房间内所有玩家
func (c *defaultContext) PushToRoom(seq uint64, route string, msgId uint64, msg any) error {
	p := c.Player()
	if p == nil {
		return errors.New("player not found")
	}
	roomId := p.RoomID()
	if roomId <= 0 {
		return errors.New("room not found")
	}
	ro, ok := c.g.RoomManager().Get(roomId)
	if !ok || ro == nil {
		return fmt.Errorf("room %d not found", roomId)
	}
	ro.Broadcast(seq, route, c.Version(), msgId, msg) // TODO
	return nil
}

// Version 版本号
func (c *defaultContext) Version() string {
	return c.version
}

// Close 关闭上下文
func (c *defaultContext) Close() {
	c.g = nil
	c.p = nil
	c.seq = 0
	c.uid = 0
	c.route = ""
	c.game = ""
	c.msgID = 0
	c.timestamp = 0
	c.version = ""
	ctxPool.Put(c)
}
