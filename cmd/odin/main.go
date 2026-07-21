package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	// 命令层关闭 Cobra 的自动错误输出，确保每个错误只打印一次并返回非零状态码。
	if err := newRootCommand(defaultDependencies()).ExecuteContext(context.Background()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
