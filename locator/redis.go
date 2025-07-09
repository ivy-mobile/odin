package locator

import (
	"context"
	"errors"
	"fmt"
	"slices"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	locatorUserGateKey = "%s:locator:user:%d:gate" // string
	locatorUserGameKey = "%s:locator:user:%d:game" // hash
	locatorEventKey    = "%s:locator:%s:event"     // string, 用于广播事件
)

// 基于redis实现的定位器
type redisLocator struct {
	ctx    context.Context // 上下文
	cancel context.CancelFunc
	sfg    singleflight.Group

	opts             *options
	gameNodes        *lru.Cache[string, string] // 用户Game节点映射, key: uid-gameName value: gameID
	gateNodes        *lru.Cache[int64, string]  // 用户Gate节点-LRU缓存
	unbindGateScript *redis.Script              // Lua脚本，用于解绑网关节点
	unbindGameScript *redis.Script              // Lua脚本，用于解绑游戏节点
}

func New(opts ...Option) Locator {

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	if o.prefix == "" {
		o.prefix = defaultPrefix
	}
	if o.lruSize <= 0 {
		o.lruSize = defaultLRUSize
	}
	if o.client == nil {
		o.client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      o.addrs,
			DB:         o.db,
			Username:   o.username,
			Password:   o.password,
			MaxRetries: o.maxRetries,
		})
	}

	c, cancel := context.WithCancel(o.ctx)

	gateNodes, _ := lru.New[int64, string](o.lruSize)
	gameNodes, _ := lru.New[string, string](o.lruSize)

	return &redisLocator{
		ctx:              c,
		cancel:           cancel,
		opts:             o,
		gameNodes:        gameNodes,
		gateNodes:        gateNodes,
		unbindGateScript: redis.NewScript(unbindGateScript),
		unbindGameScript: redis.NewScript(unbindGameScript),
	}
}

// BindGate 绑定网关节点
func (l *redisLocator) BindGate(ctx context.Context, uid int64, gateID string) error {

	key := fmt.Sprintf(locatorUserGateKey, l.opts.prefix, uid)
	// 更新 Redis
	if err := l.opts.client.Set(ctx, key, gateID, redis.KeepTTL).Err(); err != nil {
		return fmt.Errorf("BindGate failed: %v, uid: %d, gateID: %s", err, uid, gateID)
	}

	// 广播绑定事件给其他服务
	if err := l.publishEvent(ctx, EventType_BindGate, uid, gateID); err != nil {
		return err
	}

	// 更新内存
	l.gateNodes.Add(uid, gateID)
	return nil
}

// BindGame 绑定游戏节点
func (l *redisLocator) BindGame(ctx context.Context, uid int64, gameName, gameID string) error {

	key := fmt.Sprintf(locatorUserGameKey, l.opts.prefix, uid)
	// 更新 Redis
	if err := l.opts.client.HSet(ctx, key, gameName, gameID).Err(); err != nil {
		return fmt.Errorf("BindGame failed: %v, uid: %d, gameName: %s, gameID: %s", err, uid, gameName, gameID)
	}

	// 广播绑定事件给其他服务
	if err := l.publishEvent(ctx, EventType_BindGame, uid, gameID, gameName); err != nil {
		return err
	}

	// 更新内存
	l.gameNodes.Add(fmt.Sprintf("%d-%s", uid, gameName), gameID)
	return nil
}

// UnbindGate 解绑网关节点
func (l *redisLocator) UnbindGate(ctx context.Context, uid int64, gateID string) error {

	// 更新 Redis
	key := fmt.Sprintf(locatorUserGateKey, l.opts.prefix, uid)
	// 使用 Lua 脚本进行原子操作
	// 脚本会检查 key 是否存在，并且 gateID 是否匹配
	rst, err := l.unbindGateScript.Run(ctx, l.opts.client, []string{key}, gateID).StringSlice()
	if err != nil {
		return err
	}
	// 解绑失败，可能是因为 key 不存在或 gateID 不匹配
	if rst[0] == "NO" {
		return fmt.Errorf("UnbindGate failed, uid: %d, gateID: %s", uid, gateID)
	}
	// 广播绑定事件给其他服务
	if err := l.publishEvent(ctx, EventType_UnbindGate, uid, gateID); err != nil {
		return err
	}
	// 更新内存
	l.gateNodes.Remove(uid)
	return nil
}

// UnbindGame 解绑游戏节点
func (l *redisLocator) UnbindGame(ctx context.Context, uid int64, gameName, gameID string) error {

	// 更新 Redis
	key := fmt.Sprintf(locatorUserGameKey, l.opts.prefix, uid)
	// 使用 Lua 脚本进行原子操作
	// 脚本会检查 key 是否存在，并且 gameID 是否匹配
	rst, err := l.unbindGameScript.Run(ctx, l.opts.client, []string{key}, gameName, gameID).StringSlice()
	if err != nil {
		return err
	}
	// 解绑失败，可能是因为 key 不存在或 gateID 不匹配
	if rst[0] == "NO" {
		return fmt.Errorf("UnbindGame failed: %v, uid: %d, gameName: %s, gameID: %s", err, uid, gameName, gameID)
	}

	// 广播绑定事件给其他服务
	if err := l.publishEvent(ctx, EventType_UnbindGame, uid, gameID, gameName); err != nil {
		return err
	}

	// 更新内存
	l.gameNodes.Remove(fmt.Sprintf("%d-%s", uid, gameName))
	return nil
}

// GetGateNode 获取用户所在的 Gate 节点
func (l *redisLocator) GetGateNode(ctx context.Context, uid int64) (string, error) {

	// 先从内存获取
	value, ok := l.gateNodes.Get(uid)
	if ok {
		return value, nil
	}

	// 使用 singleflight 防止缓存击穿
	key := fmt.Sprintf("%d:%s", uid, NodeTypeGate)
	result, err, _ := l.sfg.Do(key, func() (any, error) {

		rkey := fmt.Sprintf(locatorUserGateKey, l.opts.prefix, uid)
		// 从 Redis 获取
		gateID, err := l.opts.client.Get(ctx, rkey).Result()
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("GetGateNode failed: %v, uid: %d", err, uid)
		}
		// 更新内存
		l.gateNodes.Add(uid, gateID)
		return gateID, nil
	})
	if err != nil {
		return "", err
	}
	// 如果 result 是 nil，表示没有找到对应的节点
	if result == nil {
		return "", nil
	}
	return result.(string), nil
}

// GetGameNode 获取用户所在的 Game 节点
func (l *redisLocator) GetGameNode(ctx context.Context, uid int64, gameName string) (string, error) {

	// 先从内存获取
	value, ok := l.gameNodes.Get(fmt.Sprintf("%d-%s", uid, gameName))
	if ok {
		return value, nil
	}

	// 使用 singleflight 防止缓存击穿
	key := fmt.Sprintf("%d:%s:%s", uid, NodeTypeGame, gameName)
	result, err, _ := l.sfg.Do(key, func() (any, error) {

		// 从 Redis 获取
		rkey := fmt.Sprintf(locatorUserGameKey, l.opts.prefix, uid)
		gameId, err := l.opts.client.HGet(ctx, rkey, gameName).Result()
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("GetGameNode failed: %v, uid: %d, gameName: %s", err, uid, gameName)
		}
		// 更新内存
		l.gameNodes.Add(fmt.Sprintf("%d-%s", uid, gameName), gameId)
		return gameId, nil
	})
	if err != nil {
		return "", err
	}
	// 如果 result 是 nil，表示没有找到对应的节点
	if result == nil {
		return "", nil
	}
	return result.(string), nil
}

// 发布事件到redis，其它服务节点订阅消息 更新本地缓存
func (l *redisLocator) publishEvent(ctx context.Context, eventType EventType, uid int64, nodeID string, nodeNames ...string) error {

	data := Event{
		EventType: eventType,
		Uid:       uid,
		NodeID:    nodeID,
	}
	if len(nodeNames) > 0 {
		data.NodeName = nodeNames[0]
	}

	var nodeType NodeType
	switch eventType {
	case EventType_BindGate, EventType_UnbindGate:
		nodeType = NodeTypeGate
	case EventType_BindGame, EventType_UnbindGame:
		nodeType = NodeTypeGame
	}
	// 发布事件到 Redis
	if err := l.opts.client.Publish(ctx, fmt.Sprintf(locatorEventKey, l.opts.prefix, nodeType), data).Err(); err != nil {
		return fmt.Errorf("eventPublish failed: %v, nodeType: %s, data: %v", err, nodeType, data)
	}
	return nil
}

func (l *redisLocator) WatchChange(ctx context.Context, channels ...EventChannel) {

	if len(channels) == 0 {
		return
	}
	// 监听指定的事件频道 (Gate和Game订阅channel分离)
	cs := make([]string, 0, 2)
	if slices.Contains(channels, EventChannel_Game) {
		cs = append(cs, fmt.Sprintf(locatorEventKey, l.opts.prefix, NodeTypeGame))
	}
	if slices.Contains(channels, EventChannel_Gate) {
		cs = append(cs, fmt.Sprintf(locatorEventKey, l.opts.prefix, NodeTypeGate))
	}
	sub := l.opts.client.Subscribe(ctx, cs...)

	go func() {
		defer sub.Close() // 确保订阅关闭

		ch := sub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				e := &Event{}
				err := e.UnmarshalBinary([]byte(msg.Payload))
				if err != nil {
					fmt.Println("invalid payload, ", msg.Payload)
					continue
				}
				fmt.Printf("Received event: %v\n", e)
				// 根据事件类型更新本地缓存
				switch e.EventType {
				case EventType_BindGate, EventType_UnbindGate:
					l.gateNodes.Remove(e.Uid)
				case EventType_BindGame, EventType_UnbindGame:
					l.gameNodes.Remove(fmt.Sprintf("%d-%s", e.Uid, e.NodeName))
				}
			}
		}
	}()
}
