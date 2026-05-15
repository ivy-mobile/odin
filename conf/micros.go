package conf

// MicrosConfig 微服务配置
type MicrosConfig struct {
	// 游戏中心
	GameCenter *MicroConfig `yaml:"game_center" json:"game_center" toml:"game_center"`
	// 用户服务 - Java端
	User *MicroConfig `yaml:"user" json:"user" toml:"user"`
	// 房间服务 - Java端
	Room *MicroConfig `yaml:"room" json:"room" toml:"room"`
	// ...
}

type MicroConfig struct {
	Filters string `yaml:"filters" json:"filters" toml:"filters"`
}
