package webhook

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewMarkdownJSON(t *testing.T) {
	msg := NewMarkdown("告警", "### CPU 使用率过高", AtMobiles("13800138000"), AtUserIDs("user1"))

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"msgtype":"markdown","markdown":{"title":"告警","text":"### CPU 使用率过高"},"at":{"atMobiles":["13800138000"],"atUserIds":["user1"],"isAtAll":false}}`
	if string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}
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
			if !errors.Is(err, tt.err) {
				t.Fatalf("validate() error = %v, want %v", err, tt.err)
			}
		})
	}
}
