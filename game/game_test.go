package game_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/eventbus/redis"
	"github.com/ivy-mobile/odin/game"
	msgjson "github.com/ivy-mobile/odin/message/json"
)

const (
	Version100 = "1.0.0"
	Version200 = "2.0.0"
)

type LoginRequest struct {
	Msg string
}

func TestRegister(t1 *testing.T) {

	g := game.New(
		game.WithID("1"),
		game.WithName("test"),
		game.WithEventbus(redis.NewEventbus(
			redis.WithAddrs("localhost:6379"),
			redis.WithPassword(""),
		)),
		game.WithCodec(json.DefaultCodec),
		game.WithAdminCmdHandler(func(data []byte) {

		}),
	)

	// test 注册路由
	g.RegisterRouter(Version100, "Heartbeat", game.Handler(Heartbeat))
	// g.RegisterRouter(Version200, "Heartbeat", game.Handler(Heartbeat))
	g.RegisterRouter(Version100, "Login", game.Handler(Login))
	// g.RegisterRouter(Version200, "Login", game.Handler(Login))

	seq := uint64(0)
	go func() {
		time.Sleep(time.Second * 2)
		// 登录
		req := LoginRequest{
			Msg: "hello",
		}
		paylod, _ := json.Marshal(req)
		data := msgjson.JsonMessage{
			Seq:     atomic.AddUint64(&seq, 1),
			Uid:     1,
			Route:   "Login",
			Version: Version100,
			MsgID:   1,
			Payload: paylod,
		}
		bytes, _ := json.Marshal(data)
		g.MockReciveGateMessagex(bytes)
		fmt.Println("Login Publish success, data: ", string(bytes))

		// 心跳
		hbdata := msgjson.JsonMessage{
			Seq:     atomic.AddUint64(&seq, 1),
			Uid:     1,
			Version: Version100,
			Route:   "Heartbeat",
			MsgID:   1,
		}
		hbbytes, _ := json.Marshal(hbdata)
		g.MockReciveGateMessagex(hbbytes)
		fmt.Println("Heartbeat Publish success, data: ", string(hbbytes))
	}()
	g.Start()
}

func Login(ctx game.Context, req *LoginRequest) {
	time.Sleep(time.Second * 2)
	fmt.Printf("time: %v, Seq: %v, Uid: %v, Route: %v, Login: %v\n", time.Now().Format(time.DateTime), ctx.Seq(), ctx.Uid(), ctx.Route(), req)
}

func Heartbeat(ctx game.Context, req *struct{}) {
	time.Sleep(time.Second * 4)
	fmt.Printf("time: %v, Seq: %v, Uid: %v, Route: %v, Heartbeat: %v\n", time.Now().Format(time.DateTime), ctx.Seq(), ctx.Uid(), ctx.Route(), req)
}
