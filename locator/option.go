package locator

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	defaultAddr       = "127.0.0.1:6379" // redis连接地址，默认
	defaultDB         = 0                // redis数据库，默认0
	defaultMaxRetries = 3                // 最大重试次数
	defaultPrefix     = "ivy"            // key前缀，默认ivy
	defaultLRUSize    = 1000             // 默认LRU缓存大小
)

type Option func(o *options)

type options struct {
	ctx        context.Context
	addrs      []string // 客户端连接地址，默认[]string{"127.0.0.1:6379"}
	db         int      // redis数据库，默认为0
	username   string   // 用户名，默认为空
	password   string   // 密码，默认为空
	maxRetries int      // 最大重试次数，默认3次

	lruSize int // LRU缓存大小，默认1000

	client redis.UniversalClient // 客户端，存在外部客户端时，优先使用外部客户端，默认为nil
	prefix string                // key前缀，默认为ivy
}

func defaultOptions() *options {
	return &options{
		ctx:        context.Background(),
		addrs:      []string{defaultAddr},
		db:         defaultDB,
		maxRetries: defaultMaxRetries,
		prefix:     defaultPrefix,
		lruSize:    defaultLRUSize,
	}
}

// WithMaxCacheSize 设置LRU缓存大小
func WithMaxCacheSize(size int) Option {
	return func(o *options) { o.lruSize = size }
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// WithAddrs 设置连接地址
func WithAddrs(addrs ...string) Option {
	return func(o *options) { o.addrs = addrs }
}

// WithDB 设置数据库号
func WithDB(db int) Option {
	return func(o *options) { o.db = db }
}

// WithUsername 设置用户名
func WithUsername(username string) Option {
	return func(o *options) { o.username = username }
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(o *options) { o.password = password }
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(o *options) { o.maxRetries = maxRetries }
}

// WithClient 设置外部客户端
func WithClient(client redis.UniversalClient) Option {
	return func(o *options) { o.client = client }
}

// WithPrefix 设置前缀
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}
