package conf

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	NotifyServiceStartup DingTalkWebhookConfig `yaml:"notify_service_startup" json:"notify_service_startup" toml:"notify_service_startup"`
}

// DingTalkWebhookConfig 钉钉群自定义机器人 Webhook 配置
type DingTalkWebhookConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled" toml:"enabled"`
	Webhook string `yaml:"webhook" json:"webhook" toml:"webhook"`
	Secret  string `yaml:"secret" json:"secret" toml:"secret"`
}
