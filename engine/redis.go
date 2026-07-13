package engine

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/ivy-mobile/odin/conf"
)

// NewRedisClient 新建redis客户端
func NewRedisClient(cfg conf.RedisConfig) (redis.UniversalClient, error) {
	rc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      []string{cfg.Addr},
		ClientName: cfg.ClientName,
		Username:   cfg.Username,
		Password:   cfg.Password,
		DB:         cfg.DB,
	})
	if err := rc.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %v", err)
	}
	return rc, nil
}
