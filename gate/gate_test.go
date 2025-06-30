package gate

import (
	"testing"
	"time"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/eventbus/redis"
	"github.com/ivy-mobile/odin/registry/nacos"
)

func TestGate_Start(t *testing.T) {
	gate := New(
		WithID("test"),
		WithName("test"),
		WithPort(":18080"),
		WithCodec(json.DefaultCodec),
		WithPattern("/ws"),
		WithWriteWait(10*time.Second),
		WithPongWait(60*time.Second),
		WithPingPeriod(30*time.Second),
		WithMaxMessageSize(1024),
		WithEventbus(redis.NewEventbus(
			redis.WithAddrs("127.0.0.1:6379"),
			redis.WithPassword(""),
		)),
		WithRegistry(nacos.NewRegistry(
			nacos.WithUrls("http://43.153.4.107:18848/nacos"),
			nacos.WithGroupName("party-pop-games"),
			nacos.WithNamespaceId("zhaobin"),
		)),
	)
	gate.Start()
}
