package room

import "github.com/ivy-mobile/odin/player"

type Room interface {
	// ID 房间ID
	ID() int
	// Name 房间名
	Name() string
	// State 房间状态
	State() RoomState
	// PlayerIn 是否在房间内
	PlayerIn(uid int64) bool
	// Join 加入房间
	Join(p player.Player) error
	// Exit 退出房间
	Exit(p player.Player) error
	// Broadcast 房间内广播消息
	Broadcast(seq uint64, route, version string, msgId uint64, msg any)
}
