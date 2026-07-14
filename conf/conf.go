package conf

var cfg Config

// Config 总配置
type Config struct {
	Application  ApplicationConfig `yaml:"application" json:"application" toml:"application"`       // 应用
	ConfigCenter NacosConfig       `yaml:"config_center" json:"config_center" toml:"config_center"` // 配置中心
	Registry     NacosConfig       `yaml:"registry" json:"registry" toml:"registry"`                // 注册中心
	Log          LogConfig         `yaml:"log" json:"log" toml:"log"`                               // 日志
	Redis        RedisConfig       `yaml:"redis" json:"redis" toml:"redis"`                         // 游戏 redis
	MQ           RocketMQConfig    `yaml:"mq" json:"mq" toml:"mq"`                                  // 消息队列
	Micros       MicrosConfig      `yaml:"micros" json:"micros" toml:"micros"`                      // 微服务配置(调用其它服务)
	DingTalk     DingTalkConfig    `yaml:"dingtalk" json:"dingtalk" toml:"dingtalk"`                // 钉钉
}

// Cfg 获取完整配置
func Cfg() Config {
	return cfg
}

// Application 应用配置
func Application() ApplicationConfig {
	return cfg.Application
}

// ConfigCenter 配置中心
func ConfigCenter() NacosConfig {
	return cfg.ConfigCenter
}

// Registry 注册中心
func Registry() NacosConfig {
	return cfg.Registry
}

// Log 日志
func Log() LogConfig {
	return cfg.Log
}

// Redis redis
func Redis() RedisConfig {
	return cfg.Redis
}

// MQ 消息队列配置
func MQ() RocketMQConfig {
	return cfg.MQ
}

// Micros 微服务配置(调用其它服务)
func Micros() MicrosConfig {
	return cfg.Micros
}

// DingTalk 钉钉配置
func DingTalk() DingTalkConfig {
	return cfg.DingTalk
}
