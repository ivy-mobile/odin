package conf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/encoding/toml"
	"github.com/ivy-mobile/odin/encoding/yaml"
	"github.com/ivy-mobile/odin/xutil/xconf"
)

var (
	buildConfigClientFunc = buildConfigClient
	cfgClient             config_client.IConfigClient
)

// 业务配置选项
type options struct {
	filename string
	target   any
	watch    bool
}

// Option 加载配置选项
type Option func(*options)

// Load 加载统一配置.
//
// 参数:
//   - filename: 服务必备的系统配置文件路径, 如 "Config/Config.yaml", 不可为空.
//   - opts: 可选加载项. 需要加载业务配置时传入 WithBusiness.
//
// 加载规则:
//   - 系统配置始终从 filename 指定的本地文件读取, 并写入 pkg/conf 内部配置.
//   - 未传 WithBusiness 时, 只加载系统配置.
//   - 传入 WithBusiness 后, 如果系统配置中的 ConfigCenter 有效, 则从 Nacos
//     获取业务配置并反序列化到调用方传入的业务结构体.
//   - 如果 ConfigCenter 无效, 则从 WithBusiness 指定的本地业务配置文件读取.
//
// 使用示例:
//
//	var business BusinessConfig
//	if err := conf.Load(
//		"Config/Config.yaml",
//		conf.WithBusiness("Config/business.yaml", &business, true),
//	); err != nil {
//		panic(err)
//	}
func Load(filename string, opts ...Option) (func(), error) {
	ops := &options{}
	for _, opt := range opts {
		if opt != nil {
			opt(ops)
		}
	}

	if err := validateLoadOptions(filename, ops); err != nil {
		return func() {}, err
	}

	raw, err := os.ReadFile(filename)
	if err != nil {
		return func() {}, fmt.Errorf("read system Config %q: %w", filename, err)
	}
	expanded, err := envsubst.Bytes(raw)
	if err != nil {
		return func() {}, fmt.Errorf("expand env in system Config %q: %w", filename, err)
	}
	if err := unmarshalConfig(filename, expanded, &cfg); err != nil {
		return func() {}, fmt.Errorf("parse system Config %q: %w", filename, err)
	}
	if ops.target == nil {
		return func() {}, nil
	}
	if !validConfigCenter(cfg.ConfigCenter) {
		if err := xconf.LoadConfigFromFile(ops.filename, ops.target, ops.watch); err != nil {
			return func() {}, fmt.Errorf("load business Config %q: %w", ops.filename, err)
		}
		return func() {}, nil
	}

	return closeClient, loadBusinessFromNacos(cfg.ConfigCenter, ops.filename, ops.target, ops.watch)
}

// WithBusiness 设置业务配置加载参数.
//
// 参数:
//   - filename: 本地业务配置文件路径, 如 "Config/business.yaml".
//     当 ConfigCenter 无效时会读取该文件; 当 ConfigCenter 有效时仅作为必填参数,
//     实际业务配置来源为 Nacos.
//   - target: 业务配置对应的结构体指针, Nacos 或本地文件内容会反序列化到该对象.
//   - watch: 是否监听业务配置变化. 本地文件模式下监听文件变化; Nacos 模式下监听
//     ConfigCenter.DataId/Group 对应的配置变化.
//
// filename 和 target 必须同时传入; target 必须是非 nil 指针.
// 支持 yaml 配置
func WithBusiness(filename string, target any, watch bool) Option {
	return func(opts *options) {
		opts.filename = filename
		opts.target = target
		opts.watch = watch
	}
}

// 释放资源
func closeClient() {
	if cfgClient != nil {
		cfgClient.CloseClient()
		cfgClient = nil
	}
}

// 构建配置中心客户端
func buildConfigClient(nacosCfg *NacosConfig) (config_client.IConfigClient, error) {
	clientConfig := constant.NewClientConfig(
		constant.WithNamespaceId(nacosCfg.Namespace),
		constant.WithTimeoutMs(nacosCfg.RequestTimeout),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir(nacosCfg.LogDir),
		constant.WithCacheDir(nacosCfg.CacheDir),
		constant.WithLogLevel(nacosCfg.LogLevel),
	)
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      nacosCfg.IpAddr,
			Port:        nacosCfg.Port,
			ContextPath: nacosCfg.Path,
		},
	}
	return clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
}

// 验证加载选项
func validateLoadOptions(sysFilename string, opts *options) error {
	if strings.TrimSpace(sysFilename) == "" {
		return errors.New("system Config filename is required")
	}

	hasBusinessFilename := strings.TrimSpace(opts.filename) != ""
	hasBusinessTarget := opts.target != nil
	if hasBusinessFilename != hasBusinessTarget {
		return errors.New("business Config filename and target must be provided together")
	}

	if hasBusinessTarget {
		value := reflect.ValueOf(opts.target)
		if value.Kind() != reflect.Ptr || value.IsNil() {
			return errors.New("business Config target must be a non-nil pointer")
		}
	}
	return nil
}

// 验证配置中心
func validConfigCenter(nacosCfg NacosConfig) bool {
	return strings.TrimSpace(nacosCfg.IpAddr) != "" &&
		nacosCfg.Port != 0 &&
		strings.TrimSpace(nacosCfg.DataId) != "" &&
		strings.TrimSpace(nacosCfg.Group) != ""
}

// 从配置中心加载业务配置
func loadBusinessFromNacos(nacosCfg NacosConfig, filename string, target any, watch bool) error {
	client, err := buildConfigClientFunc(&nacosCfg)
	if err != nil {
		return fmt.Errorf("create nacos Config client: %w", err)
	}
	if !watch {
		defer client.CloseClient()
	}

	param := vo.ConfigParam{
		DataId: nacosCfg.DataId,
		Group:  nacosCfg.Group,
	}
	content, err := client.GetConfig(param)
	if err != nil {
		return fmt.Errorf("get nacos business Config dataId=%q group=%q: %w", nacosCfg.DataId, nacosCfg.Group, err)
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("get nacos business Config dataId=%q group=%q: empty content", nacosCfg.DataId, nacosCfg.Group)
	}
	if err := unmarshalConfig(filename, []byte(content), target); err != nil {
		return fmt.Errorf("unmarshal nacos business Config dataId=%q group=%q: %w", nacosCfg.DataId, nacosCfg.Group, err)
	}
	if watch {
		param.OnChange = func(_, _, _, data string) {
			_ = unmarshalConfig(filename, []byte(data), target)
		}
		if err := client.ListenConfig(param); err != nil {
			client.CloseClient()
			return fmt.Errorf("listen nacos business Config dataId=%q group=%q: %w", nacosCfg.DataId, nacosCfg.Group, err)
		}
		if cfgClient != nil {
			cfgClient.CloseClient()
		}
		cfgClient = client
	}
	return nil
}

// 反序列化配置
// 支持 json/toml/yaml，默认为 yaml
func unmarshalConfig(filename string, content []byte, target any) error {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return json.Unmarshal(content, target)
	case ".toml":
		return toml.Unmarshal(content, target)
	default:
		return yaml.Unmarshal(content, target)
	}
}
