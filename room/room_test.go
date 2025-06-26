package room_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ivy-mobile/odin/player"
	"github.com/ivy-mobile/odin/room"
)

var (
	Idle = &IdleState{}
	Game = &GameState{}
	End  = &EndState{}
)

type IdleState struct{}

func (*IdleState) ID() uint8 {
	return 1
}
func (*IdleState) Name() string {
	return "空闲状态"
}
func (*IdleState) Timeout() time.Duration {
	return time.Second * 1
}

type GameState struct{}

func (*GameState) ID() uint8 {
	return 2
}
func (*GameState) Name() string {
	return "游戏状态"
}
func (*GameState) Timeout() time.Duration {
	return time.Second * 1
}

type EndState struct{}

func (*EndState) ID() uint8 {
	return 3
}
func (*EndState) Name() string {
	return "结束状态"
}
func (*EndState) Timeout() time.Duration {
	return time.Second * 1
}

// uno 牌桌模拟
type UnoRoom struct {
	*room.Base
}

var _ room.Room = (*UnoRoom)(nil)

func NewUnoRoom(id int, name string) (*UnoRoom, error) {
	uno := &UnoRoom{}
	b, err := room.NewBaseRoom(
		room.With(id, name),
		room.WithIdleState(Idle),
		room.WithStateTimeoutHandler(uno.StateTimeoutHandler),
	)
	if err != nil {
		return nil, fmt.Errorf("new base room error: %w", err)
	}
	uno.Base = b
	return uno, nil
}

func (u *UnoRoom) StateTimeoutHandler() (uint16, error) {
	switch u.State() {
	case Idle:
		u.SetState(Game)
	case Game:
		u.SetState(End)
	case End:
		u.SetState(Idle)
	}
	return 0, nil
}

func (u *UnoRoom) Join(p player.Player) error {
	// TODO 加入房间
	return nil
}

func (u *UnoRoom) Exit(p player.Player) error {
	// TODO 退出房间
	return nil
}

func (u *UnoRoom) Start() {
	<-u.Go(func() (uint16, error) {
		u.SetState(Idle)
		return 0, nil
	})
}

func TestRoom(t *testing.T) {

	uno, err := NewUnoRoom(1, "999")
	if err != nil {
		t.Fatal(err)
	}
	uno.Serve()
	time.Sleep(time.Second)
	now := time.Now()
	uno.Start()
	time.Sleep(time.Second * 20)
	t.Logf("cost: %v", time.Since(now))
}

func TestMorePlayerRoom(t *testing.T) {
	uno, err := NewUnoRoom(1, "999")
	if err != nil {
		t.Fatal(err)
	}
	uno.Serve()
	time.Sleep(time.Second)

	uno.Start()
	// 模拟延时操作
	for i := 0; i < 10; i++ {
		result := <-uno.Go(func() (uint16, error) {
			time.Sleep(500 * time.Millisecond)
			fmt.Println("----------")
			return 0, nil
		})
		if result.OK() {
			fmt.Println("ok")
		}
	}
	time.Sleep(time.Second * 25)
}
