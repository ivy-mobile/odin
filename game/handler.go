package game

import (
	"fmt"

	"github.com/ivy-mobile/odin/message"
)

type (
	GameMessageHandler func(g *Game, msg message.Message) error // 游戏消息处理器
	CmdMessageHandler  func(g *Game, msg []byte)                // 指令消息处理器
)

// Handler 消息处理器包装器
// desc: 多包装一层,使用泛型,自动解析业务数据payload,避免重复编码
func Handler[I any](fn func(ctx Context, msg *I)) GameMessageHandler {
	return func(g *Game, msg message.Message) error {
		// 解析业务数据payload到传入的类型I
		var in I
		if len(msg.GetPayload()) > 0 {
			if err := g.opts.codec.Unmarshal(msg.GetPayload(), &in); err != nil {
				return fmt.Errorf("[Handler] unmarshal payload faild: %w", err)
			}
		}
		ctx := newDefaultContext(g, msg)
		ctx.Player().Go(func() {
			// 处理用户请求, fn内部必须同步处理,否则会出现执行结束前上下文被回收
			fn(ctx, &in)
			// 回收上下文资源
			ctx.Close()
		})
		return nil
	}
}
