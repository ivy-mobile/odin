package header

import (
	"context"
	"testing"

	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromService(t *testing.T) {
	t.Run("builds service header", func(t *testing.T) {
		h := FromService(Service{
			GameID:   1001,
			GameName: "sword-ball",
			Env:      "dev",
			NodeID:   "node-1",
			Version:  "v1.2.3",
		})

		assert.Equal(t, "1001", h[GameID])
		assert.Equal(t, "sword-ball", h[GameName])
		assert.Equal(t, "dev", h[Env])
		assert.Equal(t, "node-1", h[NodeID])
		assert.Equal(t, "v1.2.3", h[Version])
	})

	t.Run("skips zero and empty values", func(t *testing.T) {
		h := FromService(Service{
			GameID:   0,
			GameName: "",
			Env:      "prod",
		})

		assert.NotContains(t, h, GameID)
		assert.NotContains(t, h, GameName)
		assert.Equal(t, Header{Env: "prod"}, h)
	})
}

func TestWithAndFrom(t *testing.T) {
	existing := map[string]any{
		"custom":    "keep",
		"slice":     []string{"first", "second"},
		"ignored":   123,
		GameName:    "old-name",
		"empty-key": "",
	}
	ctx := context.WithValue(context.Background(), constant.AttachmentKey, existing)

	ctx = With(ctx, Header{
		GameName: "new-name",
		NodeID:   "node-1",
		Env:      "",
		"":       "bad",
	})

	h := From(ctx)
	assert.Equal(t, "keep", h["custom"])
	assert.Equal(t, "first", h["slice"])
	assert.Equal(t, "new-name", h[GameName])
	assert.Equal(t, "node-1", h[NodeID])
	assert.NotContains(t, h, "ignored")
	assert.NotContains(t, h, Env)
	assert.NotContains(t, h, "")

	assert.Equal(t, "old-name", existing[GameName])
	assert.NotContains(t, existing, NodeID)
}

func TestWithAcceptsNilContext(t *testing.T) {
	ctx := With(context.TODO(), Header{NodeID: "node-1"})

	require.NotNil(t, ctx)
	assert.Equal(t, "node-1", From(ctx).NodeID())
}

func TestAdd(t *testing.T) {
	t.Run("adds one header and preserves existing attachments", func(t *testing.T) {
		ctx := With(context.Background(), Header{NodeID: "node-1"})
		ctx = Add(ctx, MsgID, "msg-1")

		h := From(ctx)
		assert.Equal(t, "node-1", h.NodeID())
		assert.Equal(t, "msg-1", h.MsgID())
	})

	t.Run("accepts nil context", func(t *testing.T) {
		require.NotPanics(t, func() {
			ctx := Add(context.TODO(), NodeID, "node-1")
			assert.Equal(t, "node-1", From(ctx).NodeID())
		})
	})

	t.Run("skips empty key or value", func(t *testing.T) {
		ctx := Add(context.Background(), "", "value")
		ctx = Add(ctx, "empty-value", "")

		h := From(ctx)
		assert.Empty(t, h)
	})
}

func TestAddIfAbsent(t *testing.T) {
	t.Run("keeps existing string value", func(t *testing.T) {
		ctx := With(context.Background(), Header{MsgID: "custom-msg"})
		ctx = AddIfAbsent(ctx, MsgID, "generated-msg")

		assert.Equal(t, "custom-msg", From(ctx).MsgID())
	})

	t.Run("keeps existing slice value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), constant.AttachmentKey, map[string]any{
			MsgID: []string{"custom-msg"},
		})
		ctx = AddIfAbsent(ctx, MsgID, "generated-msg")

		assert.Equal(t, "custom-msg", From(ctx).MsgID())
	})

	t.Run("adds missing value", func(t *testing.T) {
		ctx := AddIfAbsent(context.Background(), MsgID, "generated-msg")

		assert.Equal(t, "generated-msg", From(ctx).MsgID())
	})
}

func TestHeaderMethods(t *testing.T) {
	h := Header{}
	h.Set(GameID, "1001")
	h.Set(GameName, "sword-ball")
	h.Set(Env, "dev")
	h.Set(NodeID, "node-1")
	h.Set(Version, "v1.2.3")
	h.Set(MsgID, "msg-1")
	h.Set("", "bad")
	h.Set("empty", "")

	assert.Equal(t, 1001, h.GameID())
	assert.Equal(t, "sword-ball", h.GameName())
	assert.Equal(t, "dev", h.Env())
	assert.Equal(t, "node-1", h.NodeID())
	assert.Equal(t, "v1.2.3", h.Version())
	assert.Equal(t, "msg-1", h.MsgID())
	assert.NotContains(t, h, "")
	assert.NotContains(t, h, "empty")

	value, ok := h.Get(NodeID)
	assert.True(t, ok)
	assert.Equal(t, "node-1", value)

	_, ok = Header{"empty": ""}.Get("empty")
	assert.False(t, ok)

	cleaned := Header{"": "bad", "empty": "", NodeID: "node-1"}.Clean()
	assert.Equal(t, Header{NodeID: "node-1"}, cleaned)
}

func TestClone(t *testing.T) {
	h := Header{NodeID: "node-1"}
	cloned := h.Clone()
	cloned.Set(NodeID, "node-2")

	assert.Equal(t, "node-1", h.NodeID())
	assert.Equal(t, "node-2", cloned.NodeID())
	assert.NotSame(t, &h, &cloned)
}

// BenchmarkAddIfAbsent 测试 AddIfAbsent 在不同场景下的性能
func BenchmarkAddIfAbsent(b *testing.B) {
	b.Run("empty context", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = AddIfAbsent(ctx, MsgID, "test-id")
		}
	})

	b.Run("key exists", func(b *testing.B) {
		ctx := With(context.Background(), Header{
			GameID:   "1001",
			GameName: "sword-ball",
			Env:      "dev",
			NodeID:   "node-1",
			Version:  "v1.2.3",
			MsgID:    "existing-msg",
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = AddIfAbsent(ctx, MsgID, "new-msg")
		}
	})

	b.Run("key missing", func(b *testing.B) {
		ctx := With(context.Background(), Header{
			GameID:   "1001",
			GameName: "sword-ball",
			Env:      "dev",
			NodeID:   "node-1",
			Version:  "v1.2.3",
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = AddIfAbsent(ctx, MsgID, "new-msg")
		}
	})
}

func BenchmarkWith(b *testing.B) {
	h := FromService(Service{
		GameID:   1001,
		GameName: "sword-ball",
		Env:      "dev",
		NodeID:   "node-1",
		Version:  "v1.2.3",
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = With(context.Background(), h)
	}
}

func BenchmarkFrom(b *testing.B) {
	ctx := With(context.Background(), Header{
		GameID:   "1001",
		GameName: "sword-ball",
		Env:      "dev",
		NodeID:   "node-1",
		Version:  "v1.2.3",
		MsgID:    "msg-1",
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = From(ctx)
	}
}
