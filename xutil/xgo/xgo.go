package xgo

import (
	"runtime"

	"github.com/ivy-mobile/odin/xutil/xlog"
)

// Recover 捕获异常
// panicHandler: 自定义Panic信息处理
func Recover(panicHandler ...func(err any)) {
	if err := recover(); err != nil {
		// 支持自定义Panic信息处理
		if len(panicHandler) > 0 && panicHandler[0] != nil {
			panicHandler[0](err)
			return
		}

		// 默认处理方式
		xlog.Error().Msgf("panic: %v", err)
		for i := 0; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			xlog.Error().Msgf("%s: %d", file, line)
		}
	}
}

// Go 协程安全的执行函数
// panicHandler: 自定义Panic信息处理
func Go(fn func(), panicHandler ...func(err any)) {
	if fn == nil {
		return
	}
	go func() {
		defer Recover(panicHandler...)
		fn()
	}()
}
