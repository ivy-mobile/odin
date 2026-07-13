package locate

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisLocator struct {
	gameRedis       redis.UniversalClient // 游戏 redis
	playerKeyFormat string                // 玩家key format
	gateNodeField   string                // 玩家gate node字段名
}

var _ Locator = (*RedisLocator)(nil)

func NewLocator(redisClient redis.UniversalClient, playerKeyFormat, gateNodeField string) *RedisLocator {
	return &RedisLocator{
		gameRedis:       redisClient,
		playerKeyFormat: playerKeyFormat,
		gateNodeField:   gateNodeField,
	}
}

func (r *RedisLocator) Name() string {
	return "redis"
}

func (r *RedisLocator) BindGateNode(uid int64, node string) error {
	key := fmt.Sprintf(r.playerKeyFormat, uid)
	return r.gameRedis.HMSet(context.Background(), key, r.gateNodeField, node).Err()
}

func (r *RedisLocator) UnBindGateNode(uid int64, node string) error {
	current, err := r.GetGateNode(uid)
	if err != nil {
		return err
	}
	if current == node {
		key := fmt.Sprintf(r.playerKeyFormat, uid)
		if err = r.gameRedis.HMSet(context.Background(), key, r.gateNodeField, "").Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisLocator) GetGateNode(uid int64) (string, error) {
	key := fmt.Sprintf(r.playerKeyFormat, uid)
	node, err := r.gameRedis.HGet(context.Background(), key, r.gateNodeField).Result()
	if err != nil {
		return "", err
	}
	return node, nil
}
