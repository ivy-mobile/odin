package conf

// ApplicationConfig 应用
type ApplicationConfig struct {
	ID        int    `yaml:"id" json:"id" toml:"id"`
	Name      string `yaml:"name" json:"name" toml:"name"`
	Env       string `yaml:"env" json:"env" toml:"env"`
	WsPath    string `yaml:"ws_path" json:"ws_path" toml:"ws_path"`
	WsPort    string `yaml:"ws_port" json:"ws_port" toml:"ws_port"`          // ws 端口, 例 :8080
	PprofPort string `yaml:"pprof_port" json:"pprof_port" toml:"pprof_port"` // pprof 端口， 例  :6060
}
