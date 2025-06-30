package main

import (
	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/eventbus/redis"
	"github.com/ivy-mobile/odin/gate"
	"github.com/ivy-mobile/odin/registry/nacos"
	"github.com/ivy-mobile/odin/xutil/xlog"
)

func main() {

	xlog.Init("debug", "", 3, "gate", "dev")

	g := gate.New(
		gate.WithCodec(json.DefaultCodec),
		gate.WithID("1"),
		gate.WithName("Gate"),
		gate.WithPattern("/ws"),
		gate.WithPort(":18080"),
		gate.WithEventbus(redis.NewEventbus(
			redis.WithAddrs("127.0.0.1:6379"),
			redis.WithPassword(""),
		)),
		gate.WithRegistry(nacos.NewRegistry(
			nacos.WithUrls("http://43.153.4.107:18848/nacos"),
			nacos.WithGroupName("party-pop-games"),
			nacos.WithNamespaceId("zhaobin"),
		)),
	)
	g.Start()
}
