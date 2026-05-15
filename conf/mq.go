package conf

// RocketMQConfig rocketMQ 配置
type RocketMQConfig struct {
	Endpoint      string `yaml:"endpoint" json:"endpoint" toml:"endpoint"`
	Namespace     string `yaml:"namespace" json:"namespace" toml:"namespace"`
	Group         string `yaml:"group" json:"group" toml:"group"`
	AccessKey     string `yaml:"access_key" json:"access_key" toml:"access_key"`
	SecretKey     string `yaml:"secret_key" json:"secret_key" toml:"secret_key"`
	SecurityToken string `yaml:"security_token" json:"security_token" toml:"security_token"`
}
