package gate

import (
	"testing"
	"time"

	"github.com/ivy-mobile/odin/encoding/json"
)

func TestGate_Start(t *testing.T) {
	gate := New(
		WithID("test"),
		WithName("test"),
		WithPort(":8080"),
		WithCodec(json.DefaultCodec),
		WithPattern("/ws"),
		WithWriteWait(10*time.Second),
		WithPongWait(60*time.Second),
		WithPingPeriod(30*time.Second),
		WithMaxMessageSize(1024),
	)
	gate.Start()
}
