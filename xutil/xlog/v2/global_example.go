package v2

// 全局日志示例
// odin 不推荐采用全局日志的方式,而是采用依赖注入的方式,将日志组件注入到各个层级,各层级可自定义个性化字段信息
// 如需全局日志组件,请自行在项目中实现,参考如下案例:

//var globalLogger Logger
//
//// Init 手动初始化全局日志
//func Init(logger Logger) {
//	globalLogger = logger
//}
//
//func Debug() Entry {
//	return globalLogger.Debug()
//}
//
//func Info() Entry {
//	return globalLogger.Info()
//}
//
//func Warn() Entry {
//	return globalLogger.Warn()
//}
//
//func Error() Entry {
//	return globalLogger.Error()
//}
