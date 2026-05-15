package conf

// RedisConfig redis 配置
type RedisConfig struct {
	Addr       string `yaml:"addr" json:"addr" toml:"addr"`
	ClientName string `yaml:"client_name" json:"client_name" toml:"client_name"`
	Username   string `yaml:"username" json:"username" toml:"username"`
	Password   string `yaml:"password" json:"password" toml:"password"`
	DB         int    `yaml:"db" json:"db" toml:"db"`
}
