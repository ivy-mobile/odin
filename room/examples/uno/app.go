package main

import (
	"fmt"
	"time"

	"github.com/ivy-mobile/odin/player"
	"github.com/ivy-mobile/odin/room"
	"github.com/ivy-mobile/odin/room/examples/uno/state"
)

// UnoRoom UNO游戏房间模拟
type UnoRoom struct {
	// 基础房间 - 提供基础能力
	*room.Base

	// UNO 个性化数据
	// ...
}

var _ room.Room = (*UnoRoom)(nil)

func NewUnoRoom(id int, name string) (*UnoRoom, error) {
	uno := &UnoRoom{}

	// 基础房间 Base
	base, err := room.NewBaseRoom(
		room.With(id, name),
		room.WithIdleState(state.IdleState),
		room.WithStateTimeoutHandler(uno.StateTimeoutHandler),
	)
	if err != nil {
		return nil, err
	}
	uno.Base = base
	uno.Base.Serve()
	return uno, nil
}

func (u *UnoRoom) ID() int {
	return u.Base.ID()
}

// Name 房间名
func (u *UnoRoom) Name() string {
	return u.Base.Name()
}

// State 房间状态
func (u *UnoRoom) State() room.RoomState {
	return u.Base.State()
}

// PlayerIn 是否在房间内
func (u *UnoRoom) PlayerIn(uid int64) bool {
	return u.Base.PlayerIn(uid)
}

// StateTimeoutHandler 状态超时函数
// 当状态倒计时结束时执行
func (u *UnoRoom) StateTimeoutHandler() (uint16, error) {
	switch u.State() {
	case state.IdleState:
		u.Base.SetState(state.GamingState)
	case state.GamingState:
		u.Base.SetState(state.EndState)
	case state.EndState:
		u.Base.SetState(state.IdleState)
	}
	return 0, nil
}

func (u *UnoRoom) start() {
	<-u.Base.Go(func() (uint16, error) {
		u.Base.SetState(state.GamingState)
		return 0, nil
	})
}

func (u *UnoRoom) Join(p player.Player) error {
	return nil
}

func (u *UnoRoom) Exit(p player.Player) error {
	return nil
}

func main() {

	room, err := NewUnoRoom(123456, "123456")
	if err != nil {
		panic(fmt.Errorf("new uno room err: %w", err))
	}
	room.start()
	time.Sleep(time.Second * 25)
}
