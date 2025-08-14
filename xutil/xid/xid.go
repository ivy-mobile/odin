package xid

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ivy-mobile/odin/xutil/xid/snowflake"
	"github.com/ivy-mobile/odin/xutil/xid/sonyflake"
	"github.com/ivy-mobile/odin/xutil/xid/ulid"
)

// Snowflake 雪花ID (字符串)
// 需维护 nodeId, SetNodeId(), 默认 nodeId=0
func Snowflake() string {
	return snowflake.NextID()
}

// SnowflakeInt64 雪花ID (int64)
// 需维护 nodeId, SetNodeId(), 默认 nodeId=0
func SnowflakeInt64() int64 {
	return snowflake.NextIDInt64()
}

// Sonyflake 索尼flake ID (int64)
// 有序，雪花算法升级版
// 需维护 machineId, SetSettings()， 默认根据机器 IP 生成
func Sonyflake() (int64, error) {
	return sonyflake.NextID()
}

// SonyflakeStr 索尼flake ID (字符串)
// 有序，雪花算法升级版， 默认根据机器 IP 生成
func SonyflakeStr() (string, error) {
	id, err := sonyflake.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

// ULID ulid
// 有序
func ULID() (string, error) {
	return ulid.NextID()
}

// UUID 标准UUID
func UUID() string {
	return uuid.New().String()
}

// UUIDX UUID 无横杠
func UUIDX() string {
	str := uuid.New().String()
	return strings.Replace(str, "-", "", -1)
}
