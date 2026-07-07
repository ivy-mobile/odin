package webhook

import (
	"net/http"
	"time"
)

// defaultTimeout 默认请求超时时间
const defaultTimeout = 5 * time.Second

// options 表示发送函数内部配置
type options struct {
	// secret 加签密钥，为空时不启用加签
	secret string

	// httpClient 发送请求使用的 HTTP 客户端
	httpClient *http.Client

	// timeout 默认 HTTP 客户端超时时间
	timeout time.Duration

	// now 生成加签时间戳使用的时钟函数
	now func() time.Time

	// at 消息 @ 设置，为空时不携带 at 字段
	at *At
}

// Option 表示发送配置选项
type Option func(*options)

func defaultOptions() *options {
	return &options{
		timeout: defaultTimeout,
		now:     time.Now,
	}
}

// WithSecret 启用钉钉机器人加签认证
func WithSecret(secret string) Option {
	return func(o *options) {
		o.secret = secret
	}
}

// WithHTTPClient 设置发送请求使用的 HTTP 客户端
func WithHTTPClient(client *http.Client) Option {
	return func(o *options) {
		if client != nil {
			o.httpClient = client
		}
	}
}

// WithTimeout 设置默认 HTTP 客户端超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout > 0 {
			o.timeout = timeout
		}
	}
}

func withClock(now func() time.Time) Option {
	return func(o *options) {
		if now != nil {
			o.now = now
		}
	}
}

func applyOptions(opts ...Option) *options {
	op := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(op)
		}
	}
	return op
}

func (o *options) client() *http.Client {
	if o.httpClient != nil {
		return o.httpClient
	}
	return &http.Client{Timeout: o.timeout}
}

func (o *options) ensureAt() *At {
	if o.at == nil {
		o.at = &At{}
	}
	return o.at
}

func atFromOptions(opts ...Option) *At {
	op := applyOptions(opts...)
	if op.at == nil {
		return nil
	}
	return &At{
		AtMobiles: append([]string(nil), op.at.AtMobiles...),
		AtUserIDs: append([]string(nil), op.at.AtUserIDs...),
		IsAtAll:   op.at.IsAtAll,
	}
}
