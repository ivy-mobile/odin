package webhook

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMarkdownJSON(t *testing.T) {
	msg := NewMarkdown("告警", "### CPU 使用率过高", AtMobiles("13800138000"), AtUserIDs("user1"))

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	want := `{"msgtype":"markdown","markdown":{"title":"告警","text":"### CPU 使用率过高"},"at":{"atMobiles":["13800138000"],"atUserIds":["user1"],"isAtAll":false}}`
	require.Equal(t, want, string(data))
}

func TestMessageValidate(t *testing.T) {
	tests := []struct {
		name string
		msg  *Message
		err  error
	}{
		{
			name: "nil",
			msg:  nil,
			err:  ErrMessageNil,
		},
		{
			name: "empty type",
			msg:  &Message{},
			err:  ErrMessageTypeEmpty,
		},
		{
			name: "empty text",
			msg:  NewText(""),
			err:  ErrMessageContentEmpty,
		},
		{
			name: "empty link",
			msg:  NewLink("", "text", "https://example.com", ""),
			err:  ErrMessageContentEmpty,
		},
		{
			name: "empty markdown",
			msg:  NewMarkdown("title", ""),
			err:  ErrMessageContentEmpty,
		},
		{
			name: "empty action card",
			msg:  &Message{MsgType: MsgTypeActionCard},
			err:  ErrMessageContentEmpty,
		},
		{
			name: "empty single action card button",
			msg:  NewSingleActionCard("title", "text", "", "", BtnVertical),
			err:  ErrMessageContentEmpty,
		},
		{
			name: "empty action card independent button",
			msg: NewActionCard("title", "text", []ActionCardButton{
				{Title: "", ActionURL: "https://example.com"},
			}, BtnVertical),
			err: ErrMessageContentEmpty,
		},
		{
			name: "empty feed card",
			msg:  NewFeedCard(),
			err:  ErrMessageContentEmpty,
		},
		{
			name: "empty feed card link",
			msg: NewFeedCard(FeedCardLink{
				Title:      "title",
				MessageURL: "",
				PicURL:     "https://example.com/pic.png",
			}),
			err: ErrMessageContentEmpty,
		},
		{
			name: "unsupported",
			msg:  &Message{MsgType: "unknown"},
			err:  ErrMessageTypeUnsupported,
		},
		{
			name: "valid text",
			msg:  NewText("hello"),
		},
		{
			name: "valid action card",
			msg: NewActionCard("title", "text", []ActionCardButton{
				{Title: "button", ActionURL: "https://example.com"},
			}, BtnHorizontal),
		},
		{
			name: "valid single action card",
			msg:  NewSingleActionCard("title", "text", "button", "https://example.com", "bad"),
		},
		{
			name: "valid feed card",
			msg: NewFeedCard(FeedCardLink{
				Title:      "title",
				MessageURL: "https://example.com",
				PicURL:     "https://example.com/pic.png",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.validate()
			require.ErrorIs(t, err, tt.err)
		})
	}
}

func TestNormalizeOrientation(t *testing.T) {
	require.Equal(t, BtnHorizontal, normalizeOrientation(BtnHorizontal))
	require.Equal(t, BtnVertical, normalizeOrientation("bad"))
}
