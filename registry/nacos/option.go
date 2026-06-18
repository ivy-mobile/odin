package nacos

import "github.com/nacos-group/nacos-sdk-go/v2/common/constant"

const (
	DefaultWeight  = 100
	DefaultCluster = "DEFAULT"
	DefaultGroup   = constant.DEFAULT_GROUP
	DefaultKind    = "ws"
)

type options struct {
	weight  float64
	cluster string
	group   string
	kind    string
}

// Option 是 nacos 配置项
type Option func(o *options)

func defaultOptions() options {
	return options{
		weight:  DefaultWeight,
		cluster: DefaultCluster,
		group:   DefaultGroup,
		kind:    DefaultKind,
	}
}

// Weight 设置权重
func Weight(weight float64) Option {
	return func(o *options) { o.weight = weight }
}

// Cluster 设置集群名
func Cluster(cluster string) Option {
	return func(o *options) { o.cluster = cluster }
}

// Group 设置分组
func Group(group string) Option {
	return func(o *options) { o.group = group }
}

// Kind 设置协议类型
func Kind(kind string) Option {
	return func(o *options) { o.kind = kind }
}
