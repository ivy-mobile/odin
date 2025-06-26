package rocketmq

// Option 配置选项类型
type Option func(*options)

type options struct {
	Endpoint      string
	NameSpace     string
	ConsumerGroup string
	AccessKey     string
	SecretKey     string
}

// 默认配置
func defaultOptions() *options {
	return &options{
		Endpoint:      "",
		NameSpace:     "",
		ConsumerGroup: "",
		AccessKey:     "",
		SecretKey:     "",
	}
}

// WithEndpoint 设置 Endpoint,必选
func WithEndpoint(endpoint string) Option {
	return func(o *options) {
		o.Endpoint = endpoint
	}
}

// WithNameSpace 设置命名空间
func WithNameSpace(nameSpace string) Option {
	return func(o *options) {
		o.NameSpace = nameSpace
	}
}

// WithConsumerGroup 设置消费者组 - 必选
func WithConsumerGroup(consumerGroup string) Option {
	return func(o *options) {
		o.ConsumerGroup = consumerGroup
	}
}

// WithAccessKey 设置访问密钥 - 可选
func WithAccessKey(accessKey string) Option {
	return func(o *options) {
		o.AccessKey = accessKey
	}
}

// WithSecretKey 设置密钥 - 可选
func WithSecretKey(secretKey string) Option {
	return func(o *options) {
		o.SecretKey = secretKey
	}
}
