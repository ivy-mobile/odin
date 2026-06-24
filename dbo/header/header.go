package header

import (
	"context"
	"strconv"

	"dubbo.apache.org/dubbo-go/v3/common/constant"

	"github.com/ivy-mobile/odin/xutil/xconv"
)

const (
	GameID   = "x-id"      // 游戏 ID
	GameName = "x-name"    // 游戏名称
	Env      = "x-env"     // 运行环境
	NodeID   = "x-node-id" // 节点 ID
	Version  = "x-version" // 服务版本
	MsgID    = "x-msg-id"  // 消息 ID
)

// Header 表示写入 Dubbo attachment 的业务请求头
type Header map[string]string

// Service 表示当前服务自身的固定 header 信息
type Service struct {
	GameID   int
	GameName string
	Env      string
	NodeID   string
	Version  string
}

// With 将 header 合并写入 context 中的 Dubbo attachment
func With(ctx context.Context, h Header) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	attachments := cloneAttachments(ctx.Value(constant.AttachmentKey))
	if attachments == nil {
		attachments = make(map[string]any)
	}
	for key, value := range h.Clean() {
		attachments[key] = value
	}
	return context.WithValue(ctx, constant.AttachmentKey, attachments)
}

// FromService 根据服务信息生成 header，零值和空值不会写入
func FromService(s Service) Header {
	h := Header{}
	if s.GameID > 0 {
		h.Set(GameID, strconv.Itoa(s.GameID))
	}
	h.Set(GameName, s.GameName)
	h.Set(Env, s.Env)
	h.Set(NodeID, s.NodeID)
	h.Set(Version, s.Version)
	return h
}

// Clean 返回移除空 key 和空 value 后的新 header
func (h Header) Clean() Header {
	cleaned := make(Header, len(h))
	for key, value := range h {
		if key == "" || value == "" {
			continue
		}
		cleaned[key] = value
	}
	return cleaned
}

// Set 设置非空 header 值
func (h Header) Set(key, value string) {
	if h == nil || key == "" || value == "" {
		return
	}
	h[key] = value
}

// Get 获取非空 header 值
func (h Header) Get(key string) (string, bool) {
	if h == nil || key == "" {
		return "", false
	}
	value, ok := h[key]
	if !ok || value == "" {
		return "", false
	}
	return value, ok
}

// GameID 返回游戏 ID
func (h Header) GameID() int {
	return xconv.Int(h[GameID])
}

// GameName 返回游戏名称
func (h Header) GameName() string {
	return h[GameName]
}

// Env 返回运行环境
func (h Header) Env() string {
	return h[Env]
}

// NodeID 返回当前节点 ID
func (h Header) NodeID() string {
	return h[NodeID]
}

// Version 返回当前服务版本
func (h Header) Version() string {
	return h[Version]
}

// MsgID 返回消息追踪 ID
func (h Header) MsgID() string {
	return h[MsgID]
}

// UserID 获取用户 ID
func (h Header) UserID() string {
	return h[MsgID]
}

// From 从 context 的 Dubbo attachment 中读取 string header
func From(ctx context.Context) Header {
	if ctx == nil {
		return Header{}
	}

	h := Header{}
	for key, value := range cloneAttachments(ctx.Value(constant.AttachmentKey)) {
		switch v := value.(type) {
		case string:
			h.Set(key, v)
		case []string:
			if len(v) > 0 {
				h.Set(key, v[0])
			}
		}
	}
	return h
}

// Add 向 context 的 Dubbo attachment 中追加一个 header
func Add(ctx context.Context, key, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if key == "" || value == "" {
		return ctx
	}
	attachments := cloneAttachments(ctx.Value(constant.AttachmentKey))
	if attachments == nil {
		attachments = make(map[string]any)
	}
	attachments[key] = value
	return context.WithValue(ctx, constant.AttachmentKey, attachments)
}

// AddIfAbsent 在 header 不存在时追加默认值
// 优化：直接检查 attachment 中是否存在指定 key，避免完整遍历
func AddIfAbsent(ctx context.Context, key, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if key == "" || value == "" {
		return ctx
	}

	// 直接检查 key 是否存在，避免 From(ctx) 的完整遍历和拷贝
	raw := ctx.Value(constant.AttachmentKey)
	if raw != nil {
		switch attachments := raw.(type) {
		case map[string]any:
			if v, ok := attachments[key]; ok {
				// 检查值是否非空
				switch val := v.(type) {
				case string:
					if val != "" {
						return ctx
					}
				case []string:
					if len(val) > 0 && val[0] != "" {
						return ctx
					}
				default:
					// 其他类型视为已存在
					return ctx
				}
			}
		case map[string]string:
			if v, ok := attachments[key]; ok && v != "" {
				return ctx
			}
		case Header:
			if v, ok := attachments[key]; ok && v != "" {
				return ctx
			}
		}
	}

	return Add(ctx, key, value)
}

// Clone 返回 header 的浅拷贝
func (h Header) Clone() Header {
	if len(h) == 0 {
		return Header{}
	}

	clone := make(Header, len(h))
	for key, value := range h {
		clone[key] = value
	}
	return clone
}

func cloneAttachments(raw any) map[string]any {
	if raw == nil {
		return nil
	}

	switch attachments := raw.(type) {
	case map[string]any:
		if len(attachments) == 0 {
			return nil
		}
		clone := make(map[string]any, len(attachments))
		for key, value := range attachments {
			clone[key] = value
		}
		return clone
	case map[string]string:
		if len(attachments) == 0 {
			return nil
		}
		clone := make(map[string]any, len(attachments))
		for key, value := range attachments {
			clone[key] = value
		}
		return clone
	case Header:
		if len(attachments) == 0 {
			return nil
		}
		clone := make(map[string]any, len(attachments))
		for key, value := range attachments {
			clone[key] = value
		}
		return clone
	default:
		return nil
	}
}
