package state

import (
	"time"

	"github.com/ivy-mobile/odin/room"
)

// 结束状态
var EndState = &End{}

type End struct{}

var _ room.RoomState = (*End)(nil)

func (i *End) ID() uint8 {
	return 3
}
func (i *End) Name() string {
	return "结算"
}
func (i *End) Timeout() time.Duration {
	return time.Second * 5
}
