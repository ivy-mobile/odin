package player

import (
	"errors"
	"sync"
	"time"

	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
)

// 定义对象池 - 性能优化
var basePool = sync.Pool{
	New: func() any {
		return &Base{
			actionChan: make(chan func(), 20),
			done:       make(chan struct{}),
		}
	},
}

// Base 基础玩家结构
type Base struct {
	wg            sync.WaitGroup // 等待组
	done          chan struct{}  // 房间关闭信号
	actionChan    chan func()    // 操作通道
	actionTimeout time.Duration  // 玩家操作超时时间

	msgHandler MsgHandler // 给玩家发送消息函数 - 外部传入
	id         int64      // 玩家 ID
	name       string     // 昵称
	avatar     string     // 头像
	roomID     int        // 当前所在房间 ID
	active     time.Time  // 最近活跃时间
	isOffline  bool       // 是否离线
}

var _ Player = (*Base)(nil)

// GetBase 从对象池获取一个 Base 实例
func GetBase(
	msgHandler MsgHandler,
	id int64,
	name string,
	avatar string,
	actionTimeout time.Duration) *Base {
	p := basePool.Get().(*Base)
	p.msgHandler = msgHandler
	p.id = id
	p.name = name
	p.avatar = avatar
	p.roomID = 0
	p.active = time.Now()
	p.isOffline = false
	p.actionTimeout = actionTimeout
	p.serve()
	return p
}

func (b *Base) serve() {

	b.wg.Add(1)
	xgo.Go(func() {
		defer b.wg.Done()
		for {
			select {
			case action := <-b.actionChan:
				action()
			case <-b.done:
				return
			}
		}
	})
}

func (b *Base) ID() int64 {
	return b.id
}

func (b *Base) Name() string {
	return b.name
}

func (b *Base) Avatar() string {
	return b.avatar
}

func (b *Base) RoomID() int {
	return b.roomID
}

func (b *Base) SetRoomID(roomID int) {
	b.roomID = roomID
}

func (b *Base) Active() {
	b.active = time.Now()
}

// SetOffline 设置离线状态
func (b *Base) SetOffline(isOffline bool) {
	b.isOffline = isOffline
}

// IsOffline 是否离线
func (b *Base) IsOffline() bool {
	return b.isOffline
}

// SendMessage 发送消息
func (b *Base) SendMessage(seq uint64, route, version string, msgID uint64, payload any) error {
	if b == nil || b.msgHandler == nil {
		return errors.New("player not found")
	}
	return b.msgHandler.SendMessage(seq, b.id, route, version, msgID, payload)
}

// Go 玩家协程执行操作
func (b *Base) Go(action func()) {
	if action == nil {
		return
	}
	// 玩家未登录时,启动新协程异步处理
	if b == nil {
		xgo.Go(action)
		return
	}
	// 玩家已登录时,将操作放入通道
	select {
	case b.actionChan <- action:
	case <-time.After(b.actionTimeout): // 入列超时
		xlog.Error().Msgf("player action timeout, userId: %v", b.id)
	case <-b.done:
		xlog.Info().Msgf("player closed, userId: %v", b.id)
	}
}

// Close 关闭玩家 释放资源并将实例放回对象池
func (b *Base) Close() {
	if b == nil { // 检查是否为 nil
		return
	}
	close(b.done)
	b.wg.Wait()

	// 重置字段
	b.msgHandler = nil
	b.id = 0
	b.name = ""
	b.avatar = ""
	b.roomID = 0
	b.isOffline = false
	basePool.Put(b)
}
