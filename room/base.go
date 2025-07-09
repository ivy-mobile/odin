package room

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ivy-mobile/odin/player"
	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
)

// Base 基础房间 - 适用于单个房间20人
// 提供房间的基础能力，如房间ID、房间名、房间状态、房间内玩家管理、房间内消息广播、房间内玩家操作等
type Base struct {
	opts *options

	state           RoomState   // 当前房间状态
	stateTimer      *time.Timer // 房间状态计时器
	stateChangeTime time.Time   // 状态变更时间

	players *player.Manager // 房间内玩家管理

	wg         sync.WaitGroup // 等待组
	done       chan struct{}  // 房间关闭信号
	actionChan chan *Action   // 操作通道
}

func NewBaseRoom(opts ...Option) (*Base, error) {

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	if o.id <= 0 {
		return nil, errors.New("room id must gt 0")
	}
	if o.name == "" {
		return nil, errors.New("room name must not empty")
	}
	if o.maxPlayerCount <= 0 {
		return nil, errors.New("max player count must gt 0")
	}
	if o.actionTimeout <= 0 {
		return nil, errors.New("action timeout must gt 0")
	}
	if o.stateTimeoutHandler == nil {
		return nil, errors.New("state timeout handler must not nil")
	}
	// 状态计时器
	stateTimer := time.NewTimer(time.Second)
	stateTimer.Stop()

	b := &Base{
		opts:            o,
		state:           o.idleState,
		stateTimer:      stateTimer,
		done:            make(chan struct{}),
		stateChangeTime: time.Now(),
		actionChan:      make(chan *Action, 100),
		players:         player.NewManager(),
	}
	return b, nil
}

func (b *Base) Serve() {

	b.wg.Add(2)

	// 单一协程处理玩家操作
	xgo.Go(func() {
		defer b.wg.Done()
		for {
			select {
			case act := <-b.actionChan:
				if act != nil && act.fn != nil {
					code, err := act.fn() // 执行
					if act.result != nil {
						act.result <- ActResult{code: code, err: err} // 结果信号
					}
				}
			case <-b.done:
				return
			}
		}
	})

	// 单一协程处理房间状态变更
	xgo.Go(func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.stateTimer.C:
				// 实现房间状态超时逻辑 - 同样需要放入房间队列中执行
				if b.opts.stateTimeoutHandler != nil {
					<-b.Go(b.opts.stateTimeoutHandler)
				}
			case <-b.done:
				return
			}
		}
	})
}

func (b *Base) ID() int {
	return b.opts.id
}

func (b *Base) Name() string {
	return b.opts.name
}

func (b *Base) State() RoomState {
	return b.state
}

func (b *Base) SetState(state RoomState) {
	fmt.Printf("%v,  进入%s, 倒计时:%v , duration: %v \n", time.Now().Format(time.StampMilli), state.Name(), state.Timeout().Seconds(), time.Since(b.stateChangeTime))
	b.state = state
	b.stateChangeTime = time.Now()
	b.stateTimer.Reset(state.Timeout())
}

// 将房间操作放入队列中执行
// 如需等待结果 需等待 action.result 通道返回 <- ActResult
func (b *Base) Go(fn func() (uint16, error)) <-chan ActResult {
	resultChan := make(chan ActResult, 1)
	action := &Action{
		fn:     fn,
		result: resultChan,
	}
	select {
	case b.actionChan <- action:
		return resultChan
	case <-time.After(b.opts.actionTimeout):
		resultChan <- ActResult{code: 0, err: errors.New("room action timeout")}
		return resultChan
	case <-b.done:
		resultChan <- ActResult{code: 0, err: errors.New("room closed")}
		return resultChan
	}
}

// PlayerIn 玩家是否在房间内
func (b *Base) PlayerIn(uid int64) bool {
	_, ok := b.players.Get(uid)
	return ok
}

// Close 关闭房间
func (b *Base) Close() {
	close(b.done)
	b.wg.Wait()
	close(b.actionChan)
	b.stateTimer.Stop()
}

// Broadcast 房间内广播消息
func (b *Base) Broadcast(seq uint64, route, version string, msgId uint64, msg any) {
	b.players.Range(func(_ int64, p player.Player) bool {
		if err := p.SendMessage(seq, route, version, msgId, msg); err != nil {
			xlog.Error().Int64("Uid", p.ID()).Msgf("[Broadcast.SendMessage] send message error: %v", err)
		}
		return true
	})
}
