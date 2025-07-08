package game

import (
	"errors"
	"fmt"

	"github.com/ivy-mobile/odin/message"
	"github.com/ivy-mobile/odin/player"
)

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
	// Player 玩家
	Player() player.Player
	// RoomID 房间ID
	RoomID() int
	// Resp 响应消息
	Resp(data any) error
	// Push 推送消息
	Push(seq uint64, route string, msgId uint64, msg any) error
	// PushToRoom 推送消息至房间内所有玩家
	PushToRoom(seq uint64, route string, msgId uint64, msg any) error
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

	p, _ := g.players.Get(msg.GetUid())

	return &defaultContext{
		g:         g,
		p:         p,
		seq:       msg.GetSeq(),
		uid:       msg.GetUid(),
		route:     msg.GetRoute(),
		game:      msg.GetGame(),
		msgID:     msg.GetMsgId(),
		timestamp: msg.GetTimestamp(),
	}
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

// RoomID 当前所在房间ID
func (c *defaultContext) RoomID() int {
	if c.p == nil {
		return 0
	}
	return c.p.RoomID()
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
	return c.p.SendMessage(seq, route, c.version, msgId, msg)
}

// PushToRoom 推送消息至房间内所有玩家
func (c *defaultContext) PushToRoom(seq uint64, route string, msgId uint64, msg any) error {
	roomId := c.RoomID()
	if roomId <= 0 {
		return errors.New("room not found")
	}
	room, ok := c.g.rooms.Get(roomId)
	if !ok {
		return fmt.Errorf("room %d not found", roomId)
	}
	room.Broadcast(seq, route, c.Version(), msgId, msg) // TODO
	return nil
}

// Version 版本号
func (c *defaultContext) Version() string {
	return c.version
}
