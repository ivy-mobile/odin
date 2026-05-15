package conf

// LogConfig 日志配置
type LogConfig struct {
	Level            string         `yaml:"level" json:"level" toml:"level"`
	LevelFieldName   string         `yaml:"level_field_name" json:"level_field_name" toml:"level_field_name"`
	TimeFieldName    string         `yaml:"time_field_name" json:"time_field_name" toml:"time_field_name"`
	MessageFieldName string         `yaml:"message_field_name" json:"message_field_name" toml:"message_field_name"`
	TimeFormat       string         `yaml:"time_format" json:"time_format" toml:"time_format"`
	Mode             string         `yaml:"mode" json:"mode" toml:"mode"`
	File             *LogFileConfig `yaml:"file" json:"file" toml:"file"`
}

// LogFileConfig 日志-文件 配置
type LogFileConfig struct {
	Filename   string `yaml:"filename" json:"filename" toml:"filename"`
	MaxSize    int    `yaml:"max_size" json:"max_size" toml:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups" toml:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age" toml:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress" toml:"compress"`
	LocalTime  bool   `yaml:"local_time" json:"local_time" toml:"local_time"`
}
