package sonyflake

import "github.com/sony/sonyflake/v2"

var (
	defaultSonyflake *sonyflake.Sonyflake
)

func init() {
	defaultSonyflake, _ = sonyflake.New(sonyflake.Settings{})
}

// SetSettings 自定义 Sonyflake 配置
func SetSettings(settings sonyflake.Settings) {
	defaultSonyflake, _ = sonyflake.New(settings)
}

// NextID 生成下一个 ID
func NextID() (int64, error) {
	return defaultSonyflake.NextID()
}
