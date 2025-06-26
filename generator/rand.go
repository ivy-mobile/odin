package generator

import (
	"fmt"
	"math/rand"
	"strings" // 新增 strings 包
	"sync"
	"time"
)

// RandGenerator 牌桌ID生成器
type RandGenerator struct {
	rand          *rand.Rand
	lastTimestamp int64
	counter       int32
	mutex         sync.Mutex
}

// NewRandGenerator 创建新的牌桌ID生成器
func NewRandGenerator() *RandGenerator {
	return &RandGenerator{
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
		lastTimestamp: 0,
		counter:       0,
	}
}

// GenRoomID 生成6位牌桌ID
func (g *RandGenerator) GenRoomID() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// 获取当前时间戳（秒）
	now := time.Now().UnixMilli()

	// 如果是同一秒内生成的ID，增加计数器
	if now == g.lastTimestamp {
		g.counter++
		// 同一秒内最多支持999个ID
		if g.counter > 999 {
			// 等待下一秒
			for now <= g.lastTimestamp {
				now = time.Now().UnixMilli()
			}
			g.counter = 0
		}
	} else {
		g.lastTimestamp = now
		g.counter = 0
	}

	// 生成随机数部分 (0-99)
	randomPart := g.rand.Intn(100)

	// 组合ID: 时间戳(3位) + 计数器(3位) + 随机数(2位)
	// 取时间戳的后3位
	timestampPart := now % 1000

	// 拼接ID
	id := fmt.Sprintf("%03d%03d%d", timestampPart, g.counter, randomPart)
	// 确保ID长度为6位
	if len(id) > 6 {
		id = strings.TrimPrefix(id, id[:len(id)-6]) // 使用 strings 包截取字符串
	}
	return id
}
