package locate

import (
	"testing"

	"github.com/ivy-mobile/odin/conf"
	"github.com/ivy-mobile/odin/engine"
)

func TestRedisLocator(t *testing.T) {
	r, err := engine.NewRedisClient(conf.RedisConfig{
		Addr:       "127.0.0.1:6379",
		ClientName: "test-test",
		Username:   "",
		Password:   "",
		DB:         0,
	})
	if err != nil {
		t.Fatal(err)
	}
	s := NewLocator(r, "players:%d", "gate_node")
	if err = s.BindGateNode(1000, "asdahsdghhj"); err != nil {
		t.Fatal(err)
	}
	if err = s.UnBindGateNode(1000, "asdahsdghhj"); err != nil {
		t.Fatal(err)
	}
	node, err := s.GetGateNode(1000)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gateNode:%v", node)
}
