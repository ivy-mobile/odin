package webhook

import "strings"

// At 表示钉钉消息 @ 设置
type At struct {
	// AtMobiles 需要 @ 的用户手机号列表
	AtMobiles []string `json:"atMobiles,omitempty"`

	// AtUserIDs 需要 @ 的钉钉用户 ID 列表
	AtUserIDs []string `json:"atUserIds,omitempty"`

	// IsAtAll 是否 @ 所有群成员
	IsAtAll bool `json:"isAtAll"`
}

// AtMobiles 通过手机号 @ 用户
func AtMobiles(mobiles ...string) Option {
	return func(o *options) {
		at := o.ensureAt()
		at.AtMobiles = appendNonEmpty(at.AtMobiles, mobiles...)
	}
}

// AtUserIDs 通过钉钉用户 ID @ 用户
func AtUserIDs(userIDs ...string) Option {
	return func(o *options) {
		at := o.ensureAt()
		at.AtUserIDs = appendNonEmpty(at.AtUserIDs, userIDs...)
	}
}

// AtAll @ 所有群成员
func AtAll() Option {
	return func(o *options) {
		o.ensureAt().IsAtAll = true
	}
}

func appendNonEmpty(dst []string, vals ...string) []string {
	for _, val := range vals {
		if val != "" {
			dst = append(dst, val)
		}
	}
	return dst
}

func (at *At) visibleMentions(text string) string {
	if at == nil {
		return ""
	}
	mentions := make([]string, 0, len(at.AtMobiles)+len(at.AtUserIDs)+1)
	for _, mobile := range at.AtMobiles {
		mention := "@" + mobile
		if !strings.Contains(text, mention) {
			mentions = append(mentions, mention)
		}
	}
	for _, userID := range at.AtUserIDs {
		mention := "@" + userID
		if !strings.Contains(text, mention) {
			mentions = append(mentions, mention)
		}
	}
	if at.IsAtAll && !strings.Contains(text, "@所有人") && !strings.Contains(text, "@all") {
		mentions = append(mentions, "@所有人")
	}
	return strings.Join(mentions, " ")
}
