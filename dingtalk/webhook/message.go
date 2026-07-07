package webhook

import "strings"

const (
	// MsgTypeText 文本消息类型
	MsgTypeText = "text"

	// MsgTypeLink 链接消息类型
	MsgTypeLink = "link"

	// MsgTypeMarkdown markdown 消息类型
	MsgTypeMarkdown = "markdown"

	// MsgTypeActionCard actionCard 消息类型
	MsgTypeActionCard = "actionCard"

	// MsgTypeFeedCard feedCard 消息类型
	MsgTypeFeedCard = "feedCard"
)

const (
	// BtnVertical actionCard 按钮竖向排列
	BtnVertical = "0"

	// BtnHorizontal actionCard 按钮横向排列
	BtnHorizontal = "1"
)

// Message 表示钉钉自定义机器人消息体
type Message struct {
	// MsgType 消息类型，对应钉钉协议中的 msgtype
	MsgType string `json:"msgtype"`

	// Text 文本消息内容，仅在 MsgTypeText 时使用
	Text *Text `json:"text,omitempty"`

	// Link 链接消息内容，仅在 MsgTypeLink 时使用
	Link *Link `json:"link,omitempty"`

	// Markdown markdown 消息内容，仅在 MsgTypeMarkdown 时使用
	Markdown *Markdown `json:"markdown,omitempty"`

	// ActionCard actionCard 消息内容，仅在 MsgTypeActionCard 时使用
	ActionCard *ActionCard `json:"actionCard,omitempty"`

	// FeedCard feedCard 消息内容，仅在 MsgTypeFeedCard 时使用
	FeedCard *FeedCard `json:"feedCard,omitempty"`

	// At @ 设置，仅 text 和 markdown 消息支持
	At *At `json:"at,omitempty"`
}

// Text 表示文本消息内容
type Text struct {
	// Content 文本正文
	Content string `json:"content"`
}

// Link 表示链接消息内容
type Link struct {
	// Text 链接消息正文
	Text string `json:"text"`

	// Title 链接消息标题
	Title string `json:"title"`

	// PicURL 链接消息图片地址，可为空
	PicURL string `json:"picUrl,omitempty"`

	// MessageURL 点击消息后跳转的链接地址
	MessageURL string `json:"messageUrl"`
}

// Markdown 表示 markdown 消息内容
type Markdown struct {
	// Title 首屏会话透出的消息标题
	Title string `json:"title"`

	// Text markdown 格式正文
	Text string `json:"text"`
}

// ActionCard 表示 actionCard 消息内容
type ActionCard struct {
	// Title 首屏会话透出的消息标题
	Title string `json:"title"`

	// Text markdown 格式正文
	Text string `json:"text"`

	// BtnOrientation 按钮排列方向，取值为 BtnVertical 或 BtnHorizontal
	BtnOrientation string `json:"btnOrientation,omitempty"`

	// SingleTitle 单按钮模式下的按钮标题
	SingleTitle string `json:"singleTitle,omitempty"`

	// SingleURL 单按钮模式下的跳转地址
	SingleURL string `json:"singleURL,omitempty"`

	// Btns 独立跳转模式下的按钮列表
	Btns []ActionCardButton `json:"btns,omitempty"`
}

// ActionCardButton 表示 actionCard 独立跳转按钮
type ActionCardButton struct {
	// Title 按钮标题
	Title string `json:"title"`

	// ActionURL 点击按钮后的跳转地址
	ActionURL string `json:"actionURL"`
}

// FeedCard 表示 feedCard 消息内容
type FeedCard struct {
	// Links feedCard 链接列表
	Links []FeedCardLink `json:"links"`
}

// FeedCardLink 表示 feedCard 中的一条链接
type FeedCardLink struct {
	// Title 单条链接标题
	Title string `json:"title"`

	// MessageURL 点击单条链接后的跳转地址
	MessageURL string `json:"messageURL"`

	// PicURL 单条链接图片地址
	PicURL string `json:"picURL"`
}

// NewText 创建文本消息
func NewText(content string, opts ...Option) *Message {
	return &Message{
		MsgType: MsgTypeText,
		Text:    &Text{Content: content},
		At:      atFromOptions(opts...),
	}
}

// NewLink 创建链接消息
func NewLink(title, text, messageURL, picURL string) *Message {
	return &Message{
		MsgType: MsgTypeLink,
		Link: &Link{
			Title:      title,
			Text:       text,
			MessageURL: messageURL,
			PicURL:     picURL,
		},
	}
}

// NewMarkdown 创建 markdown 消息
func NewMarkdown(title, text string, opts ...Option) *Message {
	return &Message{
		MsgType:  MsgTypeMarkdown,
		Markdown: &Markdown{Title: title, Text: text},
		At:       atFromOptions(opts...),
	}
}

// NewSingleActionCard 创建单按钮 actionCard 消息
func NewSingleActionCard(title, text, singleTitle, singleURL, orientation string) *Message {
	return &Message{
		MsgType: MsgTypeActionCard,
		ActionCard: &ActionCard{
			Title:          title,
			Text:           text,
			SingleTitle:    singleTitle,
			SingleURL:      singleURL,
			BtnOrientation: normalizeOrientation(orientation),
		},
	}
}

// NewActionCard 创建独立跳转按钮 actionCard 消息
func NewActionCard(title, text string, btns []ActionCardButton, orientation string) *Message {
	return &Message{
		MsgType: MsgTypeActionCard,
		ActionCard: &ActionCard{
			Title:          title,
			Text:           text,
			Btns:           btns,
			BtnOrientation: normalizeOrientation(orientation),
		},
	}
}

// NewFeedCard 创建 feedCard 消息
func NewFeedCard(links ...FeedCardLink) *Message {
	return &Message{
		MsgType:  MsgTypeFeedCard,
		FeedCard: &FeedCard{Links: links},
	}
}

func (m *Message) validate() error {
	if m == nil {
		return ErrMessageNil
	}
	if strings.TrimSpace(m.MsgType) == "" {
		return ErrMessageTypeEmpty
	}

	switch m.MsgType {
	case MsgTypeText:
		if m.Text == nil || strings.TrimSpace(m.Text.Content) == "" {
			return ErrMessageContentEmpty
		}
	case MsgTypeLink:
		if m.Link == nil ||
			strings.TrimSpace(m.Link.Title) == "" ||
			strings.TrimSpace(m.Link.Text) == "" ||
			strings.TrimSpace(m.Link.MessageURL) == "" {
			return ErrMessageContentEmpty
		}
	case MsgTypeMarkdown:
		if m.Markdown == nil ||
			strings.TrimSpace(m.Markdown.Title) == "" ||
			strings.TrimSpace(m.Markdown.Text) == "" {
			return ErrMessageContentEmpty
		}
	case MsgTypeActionCard:
		return validateActionCard(m.ActionCard)
	case MsgTypeFeedCard:
		return validateFeedCard(m.FeedCard)
	default:
		return ErrMessageTypeUnsupported
	}

	return nil
}

func validateActionCard(card *ActionCard) error {
	if card == nil ||
		strings.TrimSpace(card.Title) == "" ||
		strings.TrimSpace(card.Text) == "" {
		return ErrMessageContentEmpty
	}
	if len(card.Btns) == 0 {
		if strings.TrimSpace(card.SingleTitle) == "" || strings.TrimSpace(card.SingleURL) == "" {
			return ErrMessageContentEmpty
		}
		return nil
	}
	for _, btn := range card.Btns {
		if strings.TrimSpace(btn.Title) == "" || strings.TrimSpace(btn.ActionURL) == "" {
			return ErrMessageContentEmpty
		}
	}
	return nil
}

func validateFeedCard(card *FeedCard) error {
	if card == nil || len(card.Links) == 0 {
		return ErrMessageContentEmpty
	}
	for _, link := range card.Links {
		if strings.TrimSpace(link.Title) == "" ||
			strings.TrimSpace(link.MessageURL) == "" ||
			strings.TrimSpace(link.PicURL) == "" {
			return ErrMessageContentEmpty
		}
	}
	return nil
}

func normalizeOrientation(orientation string) string {
	if orientation == BtnHorizontal {
		return BtnHorizontal
	}
	return BtnVertical
}
