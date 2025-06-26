package state

import (
	"time"

	"github.com/ivy-mobile/odin/room"
)

// 游戏中状态
var GamingState = &Gaming{}

type Gaming struct{}

var _ room.RoomState = (*Gaming)(nil)

func (i *Gaming) ID() uint8 {
	return 2
}
func (i *Gaming) Name() string {
	return "游戏中"
}
func (i *Gaming) Timeout() time.Duration {
	return time.Second * 5
}
