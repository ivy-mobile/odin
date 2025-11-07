package v2

// Logger 日志接口, 链式调用
type Logger interface {
	With(string, string) Logger // 添加字段,返回新的Logger
	Debug() Entry
	Info() Entry
	Warn() Entry
	Error() Entry
}

func New(opts ...Option) Logger {
	return newDefaultLog(opts...)
}
