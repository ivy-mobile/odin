package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ivy-mobile/odin/dingtalk/internal/signature"
)

// Response 表示钉钉机器人接口响应
type Response struct {
	// ErrCode 钉钉业务错误码，0 表示成功
	ErrCode int `json:"errcode"`

	// ErrMsg 钉钉业务错误信息
	ErrMsg string `json:"errmsg"`

	// Raw 原始响应体，不参与 JSON 编解码
	Raw []byte `json:"-"`
}

// Send 发送钉钉自定义机器人消息
func Send(ctx context.Context, webhook string, msg *Message, opts ...Option) (*Response, error) {
	return send(ctx, webhook, msg, opts...)
}

// SendText 发送文本消息
func SendText(ctx context.Context, webhook, content string, opts ...Option) error {
	_, err := send(ctx, webhook, NewText(content, opts...), opts...)
	return err
}

// SendMarkdown 发送 markdown 消息
func SendMarkdown(ctx context.Context, webhook, title, text string, opts ...Option) error {
	_, err := send(ctx, webhook, NewMarkdown(title, text, opts...), opts...)
	return err
}

// SendLink 发送链接消息
func SendLink(ctx context.Context, webhook, title, text, messageURL, picURL string, opts ...Option) error {
	_, err := send(ctx, webhook, NewLink(title, text, messageURL, picURL), opts...)
	return err
}

// SendSingleActionCard 发送单按钮 actionCard 消息
func SendSingleActionCard(ctx context.Context, webhook, title, text, singleTitle, singleURL, orientation string, opts ...Option) error {
	_, err := send(ctx, webhook, NewSingleActionCard(title, text, singleTitle, singleURL, orientation), opts...)
	return err
}

// SendActionCard 发送独立跳转按钮 actionCard 消息
func SendActionCard(ctx context.Context, webhook, title, text string, btns []ActionCardButton, orientation string, opts ...Option) error {
	_, err := send(ctx, webhook, NewActionCard(title, text, btns, orientation), opts...)
	return err
}

// SendFeedCard 发送 feedCard 消息
func SendFeedCard(ctx context.Context, webhook string, links []FeedCardLink, opts ...Option) error {
	_, err := send(ctx, webhook, NewFeedCard(links...), opts...)
	return err
}

func send(ctx context.Context, webhook string, msg *Message, opts ...Option) (*Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	u, err := parseWebhook(webhook)
	if err != nil {
		return nil, err
	}
	if err = msg.validate(); err != nil {
		return nil, err
	}

	op := applyOptions(opts...)
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("dingtalk/webhook: marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL(u, op), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("dingtalk/webhook: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := op.client().Do(req)
	if err != nil {
		return nil, sanitizeRequestError(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("dingtalk/webhook: read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: raw}
	}

	var result Response
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("dingtalk/webhook: decode response: %w", err)
	}
	result.Raw = raw
	if result.ErrCode != 0 {
		return &result, &APIError{Code: result.ErrCode, Message: result.ErrMsg}
	}
	return &result, nil
}

func parseWebhook(webhook string) (*url.URL, error) {
	if strings.TrimSpace(webhook) == "" {
		return nil, ErrWebhookEmpty
	}

	u, err := url.Parse(webhook)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, ErrWebhookInvalid
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrWebhookInvalid
	}
	return u, nil
}

func requestURL(webhook *url.URL, opts *options) string {
	u := *webhook
	q := u.Query()
	if opts.secret != "" {
		timestamp := opts.now().UnixMilli()
		q.Set("timestamp", strconv.FormatInt(timestamp, 10))
		q.Set("sign", signature.Sign(timestamp, opts.secret))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func sanitizeRequestError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("dingtalk/webhook: request failed: %w", urlErr.Err)
	}
	return fmt.Errorf("dingtalk/webhook: request failed: %w", err)
}
