package snowflake

import (
	"github.com/bwmarrin/snowflake"
)

var (
	defaultSnowflake *snowflake.Node
)

func init() {
	defaultSnowflake, _ = snowflake.NewNode(0)
}

func SetNodeId(id int64) error {
	node, err := snowflake.NewNode(id)
	if err != nil {
		return err
	}
	defaultSnowflake = node
	return nil
}

func NextID() string {
	return defaultSnowflake.Generate().String()
}

func NextIDInt64() int64 {
	return defaultSnowflake.Generate().Int64()
}
