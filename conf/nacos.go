package conf

import "fmt"

// NacosConfig Naco配置
type NacosConfig struct {
	//nolint:revive // 配置字段名保持历史兼容。
	IpAddr string `yaml:"ip_addr" json:"ip_addr" toml:"ip_addr"`
	Port   uint64 `yaml:"port" json:"port" toml:"port"`
	Path   string `yaml:"path" json:"path" toml:"path"`
	//nolint:revive // 配置字段名保持历史兼容。
	DataId         string `yaml:"data_id" json:"data_id" toml:"data_id"`
	Group          string `yaml:"group" json:"group" toml:"group"`
	Namespace      string `yaml:"namespace" json:"namespace" toml:"namespace"`
	RequestTimeout uint64 `yaml:"request_timeout" json:"request_timeout" toml:"request_timeout"`
	LogDir         string `yaml:"log_dir" json:"log_dir" toml:"log_dir"`
	CacheDir       string `yaml:"cache_dir" json:"cache_dir" toml:"cache_dir"`
	LogLevel       string `yaml:"log_level" json:"log_level" toml:"log_level"`
}

func (cfg *NacosConfig) Addr() string {
	return fmt.Sprintf("%s:%d%s", cfg.IpAddr, cfg.Port, cfg.Path)
}
