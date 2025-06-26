package state

import (
	"time"

	"github.com/ivy-mobile/odin/room"
)

// 空闲状态
var IdleState = &Idle{}

type Idle struct{}

var _ room.RoomState = (*Idle)(nil)

func (i *Idle) ID() uint8 {
	return 1
}
func (i *Idle) Name() string {
	return "空闲"
}
func (i *Idle) Timeout() time.Duration {
	return time.Second * 5
}
