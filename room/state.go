package room

import "time"

// RoomState 房间状态统一接口
type RoomState interface {
	ID() uint8
	Name() string
	Timeout() time.Duration
}
