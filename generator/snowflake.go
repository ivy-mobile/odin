package generator

import (
	"fmt"
	"sync"
	"time"
)

const (
	workerIDBits       = 5                           // 机器ID所占位数
	sequenceBits       = 7                           // 序列号所占位数
	workerIDShift      = sequenceBits                // 机器ID左移位数
	timestampLeftShift = sequenceBits + workerIDBits // 时间戳左移位数
	sequenceMask       = -1 ^ (-1 << sequenceBits)   // 序列号掩码
	twepoch            = 1672531200000               // 起始时间戳（2023-01-01）
)

// SnowflakeIDGenerator 雪花算法ID生成器
type SnowflakeIDGenerator struct {
	mu            sync.Mutex
	workerID      int64
	lastTimestamp int64
	sequence      int64
}

// NewSnowflakeIDGenerator 创建新的雪花算法ID生成器
func NewSnowflakeIDGenerator(workerID int64) *SnowflakeIDGenerator {
	// 确保workerID在有效范围内
	if workerID > ((1<<workerIDBits)-1) || workerID < 0 {
		panic(fmt.Sprintf("Worker ID must be between 0 and %d", (1<<workerIDBits)-1))
	}

	return &SnowflakeIDGenerator{
		workerID:      workerID,
		lastTimestamp: -1,
		sequence:      0,
	}
}

// GenRoomID 生成ID并转换为6位字符串
func (g *SnowflakeIDGenerator) GenRoomID() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 获取当前时间戳（毫秒）
	now := time.Now().UnixMilli()

	// 处理时钟回拨
	if now < g.lastTimestamp {
		// 等待时钟追上
		for now < g.lastTimestamp {
			now = time.Now().UnixMilli()
		}
	}

	// 如果是同一时间生成的，则进行序列号自增
	if now == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & sequenceMask
		// 序列号溢出
		if g.sequence == 0 {
			// 等待下一毫秒
			for now <= g.lastTimestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 时间戳改变，序列号重置
		g.sequence = 0
	}

	g.lastTimestamp = now

	// 生成ID
	id := ((now - twepoch) << timestampLeftShift) |
		(g.workerID << workerIDShift) |
		g.sequence

	// 将ID转换为6位字符串
	return fmt.Sprintf("%06d", id%1000000)
}
