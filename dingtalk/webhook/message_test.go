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

	want := `{"msgtype":"markdown","markdown":{"title":"告警","text":"### 告警\n\n### CPU 使用率过高\n\n@13800138000 @user1"},"at":{"atMobiles":["13800138000"],"atUserIds":["user1"],"isAtAll":false}}`
	require.Equal(t, want, string(data))
}

func TestNewActionCardWithAtJSON(t *testing.T) {
	msg := NewActionCard("告警", "CPU 使用率过高", []ActionCardButton{
		{Title: "查看详情", ActionURL: "https://example.com"},
	}, BtnHorizontal, AtMobiles("13800138000"))

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	want := `{"msgtype":"actionCard","actionCard":{"title":"告警","text":"CPU 使用率过高\n\n@13800138000","btnOrientation":"1","btns":[{"title":"查看详情","actionURL":"https://example.com"}]},"at":{"atMobiles":["13800138000"],"isAtAll":false}}`
	require.Equal(t, want, string(data))
}

func TestNewSingleActionCardWithAtAllJSON(t *testing.T) {
	msg := NewSingleActionCard("告警", "CPU 使用率过高", "查看详情", "https://example.com", BtnVertical, AtAll())

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	want := `{"msgtype":"actionCard","actionCard":{"title":"告警","text":"CPU 使用率过高\n\n@所有人","btnOrientation":"0","singleTitle":"查看详情","singleURL":"https://example.com"},"at":{"isAtAll":true}}`
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
			name: "empty action card independent button title",
			msg: NewActionCard("title", "text", []ActionCardButton{
				{Title: "", ActionURL: "https://example.com"},
			}, BtnVertical),
			err: ErrActionCardButtonTitleEmpty,
		},
		{
			name: "empty action card independent button action url",
			msg: NewActionCard("title", "text", []ActionCardButton{
				{Title: "button"},
			}, BtnVertical),
			err: ErrActionCardButtonActionURLEmpty,
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

func TestMarkdownText(t *testing.T) {
	tests := []struct {
		name string
		msg  *Message
		want string
	}{
		{
			name: "title and mobile",
			msg:  NewMarkdown("告警", "CPU 使用率过高", AtMobiles("13800138000")),
			want: "### 告警\n\nCPU 使用率过高\n\n@13800138000",
		},
		{
			name: "title already in text",
			msg:  NewMarkdown("告警", "### 告警\n\nCPU 使用率过高", AtMobiles("13800138000")),
			want: "### 告警\n\nCPU 使用率过高\n\n@13800138000",
		},
		{
			name: "mention already in text",
			msg:  NewMarkdown("告警", "CPU 使用率过高 @13800138000", AtMobiles("13800138000")),
			want: "### 告警\n\nCPU 使用率过高 @13800138000",
		},
		{
			name: "at all",
			msg:  NewMarkdown("告警", "CPU 使用率过高", AtAll()),
			want: "### 告警\n\nCPU 使用率过高\n\n@所有人",
		},
		{
			name: "empty text keeps empty",
			msg:  NewMarkdown("告警", "", AtAll()),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.msg.Markdown.Text)
		})
	}
}

func TestActionCardText(t *testing.T) {
	tests := []struct {
		name string
		msg  *Message
		want string
	}{
		{
			name: "mobile",
			msg: NewActionCard("告警", "CPU 使用率过高", []ActionCardButton{
				{Title: "查看详情", ActionURL: "https://example.com"},
			}, BtnVertical, AtMobiles("13800138000")),
			want: "CPU 使用率过高\n\n@13800138000",
		},
		{
			name: "mention already in text",
			msg: NewActionCard("告警", "CPU 使用率过高 @13800138000", []ActionCardButton{
				{Title: "查看详情", ActionURL: "https://example.com"},
			}, BtnVertical, AtMobiles("13800138000")),
			want: "CPU 使用率过高 @13800138000",
		},
		{
			name: "at all",
			msg:  NewSingleActionCard("告警", "CPU 使用率过高", "查看详情", "https://example.com", BtnVertical, AtAll()),
			want: "CPU 使用率过高\n\n@所有人",
		},
		{
			name: "empty text keeps empty",
			msg:  NewSingleActionCard("告警", "", "查看详情", "https://example.com", BtnVertical, AtAll()),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.msg.ActionCard.Text)
		})
	}
}
