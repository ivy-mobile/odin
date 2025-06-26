package player

// Player 统一玩家接口
type Player interface {
	// ID 玩家ID
	ID() int64
	// Name 玩家名
	Name() string
	// Avatar 玩家头像
	Avatar() string
	// RoomID 玩家房间ID
	RoomID() int
	// SetRoomID 设置玩家房间ID
	SetRoomID(roomID int)
	// Active 设置活跃时间
	Active()
	// SetOffline 设置离线状态
	SetOffline(bool)
	// IsOffline 是否离线
	IsOffline() bool
	// Go 玩家协程执行操作
	Go(action func())
	// SendMessage 发送消息
	SendMessage(seq uint64, route, version string, msgID uint64, payload any) error
	// Close 关闭玩家
	Close()
}
