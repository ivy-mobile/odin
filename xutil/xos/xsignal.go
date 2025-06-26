package xos

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// WaitSysSignal 等待系统信号
func WaitSysSignal(afterHandler ...func(s os.Signal)) {
	sig := make(chan os.Signal)

	switch runtime.GOOS {
	case `windows`:
		signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	default:
		signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGKILL, syscall.SIGTERM)
	}
	s := <-sig
	signal.Stop(sig)

	if len(afterHandler) != 0 && afterHandler[0] != nil {
		afterHandler[0](s)
	}
}
