package webhook

import (
	"errors"
	"fmt"
)

var (
	// ErrWebhookEmpty webhook 地址为空
	ErrWebhookEmpty = errors.New("dingtalk/webhook: webhook can not be empty")

	// ErrWebhookInvalid webhook 地址格式不合法
	ErrWebhookInvalid = errors.New("dingtalk/webhook: invalid webhook")

	// ErrMessageNil 消息为空
	ErrMessageNil = errors.New("dingtalk/webhook: message can not be nil")

	// ErrMessageTypeEmpty 消息类型为空
	ErrMessageTypeEmpty = errors.New("dingtalk/webhook: message type can not be empty")

	// ErrMessageTypeUnsupported 消息类型不支持
	ErrMessageTypeUnsupported = errors.New("dingtalk/webhook: unsupported message type")

	// ErrMessageContentEmpty 消息内容为空或缺少必填字段
	ErrMessageContentEmpty = errors.New("dingtalk/webhook: message content can not be empty")
)

// APIError 表示钉钉业务错误响应
type APIError struct {
	// Code 钉钉业务错误码
	Code int

	// Message 钉钉业务错误信息
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("dingtalk/webhook: api error code=%d message=%s", e.Code, e.Message)
}

// HTTPError 表示非 2xx HTTP 响应
type HTTPError struct {
	// StatusCode HTTP 状态码
	StatusCode int

	// Body 原始响应体
	Body []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("dingtalk/webhook: http status %d", e.StatusCode)
}
