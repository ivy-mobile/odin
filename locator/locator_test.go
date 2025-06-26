package locator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	ctx = context.Background()
	// loc 是一个 Locator 实例，用于测试
	loc = New(
		WithAddrs("10.80.40.36:6379"),
		WithDB(0),
		WithMaxCacheSize(1000),
		WithPrefix("ivy"),
		WithContext(context.Background()),
	)
)

func TestGate(t *testing.T) {

	// 绑定网关节点
	t.Logf("Binding gate node...")
	err := loc.BindGate(ctx, 1001, "gate-1")
	if err != nil {
		t.Fatalf("BindGate failed: %v", err)
	}
	gate, err := loc.GetGateNode(ctx, 1001)
	if err != nil {
		t.Fatalf("GetGateNode failed: %v", err)
	}
	assert.Equal(t, "gate-1", gate)

	// 解绑网关节点
	t.Logf("Unbinding gate node...")
	err = loc.UnbindGate(ctx, 1001, gate)
	if err != nil {
		t.Fatalf("UnbindGate failed: %v", err)
	}
	gate, err = loc.GetGateNode(ctx, 1001)
	if err != nil {
		t.Fatalf("After UnbindGate - GetGateNode failed: %v", err)
	}
	assert.Equal(t, "", gate)
	t.Logf("test success!!")
}

func TestUnbindGate(t *testing.T) {

	err := loc.UnbindGate(ctx, 1001, "gate-1")
	if err != nil {
		t.Fatalf("UnbindGate failed: %v", err)
	}

	gate, err := loc.GetGateNode(ctx, 1001)
	if err != nil {
		t.Fatalf("GetGateNode failed: %v", err)
	}
	assert.Equal(t, "", gate)
}

func TestGame(t *testing.T) {

	// 绑定游戏节点
	t.Logf("Binding game node...")

	err := loc.BindGame(ctx, 1001, "uno", "1")
	if err != nil {
		t.Fatalf("BindGame failed: %v", err)
	}
	gameID, err := loc.GetGameNode(ctx, 1001, "uno")
	if err != nil {
		t.Fatalf("After BindGame - GetGameNode failed: %v", err)
	}
	assert.Equal(t, "1", gameID)

	// 解绑游戏节点
	t.Logf("Unbinding game node...")

	err = loc.UnbindGame(ctx, 1001, "uno", gameID)
	if err != nil {
		t.Fatalf("UnbindGame failed: %v", err)
	}

	newGameID, err := loc.GetGameNode(ctx, 1001, "uno")
	if err != nil {
		t.Fatalf("After UnbindGame - GetGameNode failed: %v", err)
	}
	assert.Equal(t, "", newGameID)

	t.Logf("test success!!")
}

func TestWatch(t *testing.T) {

	t.Logf("Starting watch...")

	go loc.WatchChange(ctx)
	go func() {
		time.Sleep(2 * time.Second)
		err := loc.BindGame(ctx, 1001, "uno", "1")
		if err != nil {
			t.Logf("BindGame failed: %v", err)
		}
		time.Sleep(2 * time.Second)
		err = loc.BindGate(ctx, 1001, "gate-1")
		if err != nil {
			t.Logf("BindGate failed: %v", err)
		}
	}()
	time.Sleep(10 * time.Second)
}

func TestWatchChange_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	loc.WatchChange(ctx, EventChannel_Gate)
	time.Sleep(2 * time.Second)
	// cancel() // 取消监听
}
