package timer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Start 原子替换同键旧项并写入新到期时间
func (r *RedisTimer) Start(tableID int, timerID string, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}

	timerKey := r.getTimerKey()
	expireTime := time.Now().Add(duration).UnixMilli()

	tm := Message{
		TableID:   tableID,
		TimerID:   timerID,
		Timestamp: time.Now(),
	}

	tmBytes, err := json.Marshal(tm)
	if err != nil {
		return fmt.Errorf("marshal timer message: %w", err)
	}

	ctx := context.Background()

	// 使用 Lua 脚本原子性地删除旧计时器并添加新计时器
	// 这样可以避免并发问题：多个节点同时 Start 同一个计时器时，不会产生重复
	err = r.startTimerAtomically(ctx, timerKey, tableID, timerID, string(tmBytes), expireTime)
	if err != nil {
		return fmt.Errorf("start timer atomically: %w", err)
	}

	// r.config.Logger.Info().Int("Room", tableID).Str("TimerId", timerID).
	// Msgf("Start timer success, duration: %v, expireTime: %d", duration, expireTime)

	return nil
}

// Stop 查找并删除匹配的 ZSet 成员
func (r *RedisTimer) Stop(tableID int, timerID string) error {
	timerKey := r.getTimerKey()
	ctx := context.Background()

	// 获取所有成员，找到匹配的计时器
	allMembers, err := r.config.RedisClient.ZRange(ctx, timerKey, 0, -1).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil // 没有计时器，直接返回
		}
		return fmt.Errorf("ZRange failed: %w", err)
	}

	// 查找匹配的成员
	var targetMember string
	for _, m := range allMembers {
		var tm Message
		if unmarshalErr := json.Unmarshal([]byte(m), &tm); unmarshalErr != nil {
			continue
		}
		if tm.TableID == tableID && tm.TimerID == timerID {
			targetMember = m
			break
		}
	}

	if targetMember == "" {
		return nil // 没有找到匹配的计时器
	}

	// 使用原子操作删除
	_, err = r.stopTimerAtomically(ctx, timerKey, targetMember)
	if err != nil {
		return fmt.Errorf("stop timer atomically: %w", err)
	}

	return nil
}

// Lua：按桌与计时器 id 去掉旧成员再 ZADD。
func (r *RedisTimer) startTimerAtomically(ctx context.Context, timerKey string, tableID int, timerID string, newMember string, expireTime int64) error {
	// 使用 Lua 脚本原子性地完成所有操作：
	// 1. 遍历所有成员，使用字符串匹配查找匹配 tableID 和 timerID 的旧计时器
	// 2. 删除找到的旧计时器
	// 3. 添加新计时器
	// 注意：使用字符串匹配来查找 JSON 中的 tableID 和 timerID
	// 这依赖于 JSON 格式的稳定性，通常格式为: {"table_id":123,"timer_id":"xxx",...}
	script := `
		local key = KEYS[1]
		local tableId = tonumber(ARGV[1])
		local timerId = ARGV[2]
		local newMember = ARGV[3]
		local expireTime = tonumber(ARGV[4])

		-- 获取所有成员
		local members = redis.call('ZRANGE', key, 0, -1)

		-- 删除匹配的旧计时器
		-- 使用字符串匹配查找: "table_id":tableId 和 "timer_id":"timerId"
		local tableIdPattern = '"table_id":' .. tableId
		local timerIdPattern = '"timer_id":"' .. timerId .. '"'

		for i = 1, #members do
			local member = members[i]
			if string.find(member, tableIdPattern, 1, true) and
			   string.find(member, timerIdPattern, 1, true) then
				redis.call('ZREM', key, member)
				break  -- 通常只有一个匹配的计时器
			end
		end

		-- 添加新计时器
		redis.call('ZADD', key, expireTime, newMember)
		return 1
	`

	_, err := r.config.RedisClient.Eval(ctx, script, []string{timerKey}, tableID, timerID, newMember, expireTime).Result()
	if err != nil {
		return err
	}
	return nil
}

// Lua：ZREM 指定 member；返回是否删掉一条。
func (r *RedisTimer) stopTimerAtomically(ctx context.Context, timerKey, member string) (bool, error) {
	// 使用 Lua 脚本原子性地删除
	script := `
		local removed = redis.call('ZREM', KEYS[1], ARGV[1])
		return removed
	`
	result, err := r.config.RedisClient.Eval(ctx, script, []string{timerKey}, member).Result()
	if err != nil {
		return false, err
	}
	removed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type: %T", result)
	}
	return removed == 1, nil
}

// Lua：score 仍与预期一致时才 ZREM，避免误删被续期的项。
func (r *RedisTimer) removeExpiredTimerAtomically(ctx context.Context, timerKey, member string, expectedScore float64) (bool, error) {
	// 使用 Lua 脚本保证原子性：
	// 1. 检查 member 的 score 是否匹配
	// 2. 如果匹配，删除并返回 1；否则返回 0
	script := `
		local score = redis.call('ZSCORE', KEYS[1], ARGV[1])
		if score == false or tonumber(score) ~= tonumber(ARGV[2]) then
			return 0
		end
		redis.call('ZREM', KEYS[1], ARGV[1])
		return 1
	`
	result, err := r.config.RedisClient.Eval(ctx, script, []string{timerKey}, member, expectedScore).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	removed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type: %T", result)
	}
	return removed == 1, nil
}

// Listen 定时拉取到期区间并逐条安全删除后回调
func (r *RedisTimer) Listen(ctx context.Context, callback Handler) error {
	timerKey := r.getTimerKey()
	ticker := time.NewTicker(r.config.ScanInterval)
	defer ticker.Stop()

	// r.config.Logger.Info().Msg("Start listening expired timers")

	for {
		select {
		case <-ctx.Done():
			// r.config.Logger.Info().Msg("Stop listening expired timers")
			return nil
		case <-ticker.C:
			r.processExpiredTimers(ctx, timerKey, callback)
		}
	}
}

// 批量处理当前时刻前到期的 ZSet 项
func (r *RedisTimer) processExpiredTimers(ctx context.Context, timerKey string, callback Handler) {
	now := time.Now().UnixMilli()

	// 使用 ZRANGEBYSCORE 获取所有到期的计时器 (score <= now)
	results, err := r.config.RedisClient.ZRangeByScoreWithScores(ctx, timerKey, &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%d", now),
		Count: int64(r.config.BatchSize),
	}).Result()
	if err != nil {
		return
	}

	if len(results) == 0 {
		return
	}

	// r.config.Logger.Debug().Int("count", len(results)).
	// Msgf("[processExpiredTimers] found %d expired timers", len(results))

	// 处理每个到期的计时器
	for _, z := range results {
		var tm Message
		memberStr, ok := z.Member.(string)
		if !ok {
			// r.config.Logger.Error().Msg("[processExpiredTimers] invalid member type")
			// 删除无效的计时器
			r.config.RedisClient.ZRem(ctx, timerKey, z.Member)
			continue
		}

		if err := json.Unmarshal([]byte(memberStr), &tm); err != nil {
			// r.config.Logger.Error().Err(err).Str("member", memberStr).
			// Msg("[processExpiredTimers] unmarshal timer message failed")
			// 删除无效的计时器
			r.config.RedisClient.ZRem(ctx, timerKey, z.Member)
			continue
		}

		tableID, timerID := tm.TableID, tm.TimerID

		// r.config.Logger.Info().Int("Room", tableID).Str("TimerId", timerID).
		// Msgf("Process expired timer, score: %v", z.Score)

		// 使用原子操作：先检查 score 是否匹配，然后删除
		// 只有成功删除的节点才调用 callback，确保同一计时器只被一个节点处理
		// 使用 Lua 脚本保证原子性
		removed, err := r.removeExpiredTimerAtomically(ctx, timerKey, memberStr, z.Score)
		if err != nil {
			// r.config.Logger.Error().Int("Room", tableID).
			// Msgf("Remove expired timer atomically failed, err: %v", err)
			continue
		}

		// 只有成功删除的节点才调用回调函数
		// 这样可以确保多个节点同时扫描到同一个到期计时器时，只有一个节点会处理它
		if removed && callback != nil {
			callback(tableID, timerID, tm.Timestamp)
		}
	}
}
