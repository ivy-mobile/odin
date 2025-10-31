package xlog

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ivy-mobile/odin/xutil/xfile"

	"github.com/rs/zerolog"
)

var logger *XLogger

type XLogger struct {
	zerolog.Logger
	mux          sync.Mutex
	interval     time.Duration // 日志切割时间间隔, 单位:h
	lastFileTime time.Time     // 上次log文件创建时间
	path         string        // 日志文件存放路径
	env          string        // 环境
	serviceName  string        // 服务名
	node         string        // 节点
	ip           string        // ip
}

func Init(level, pathname string, interval time.Duration, serviceName, env, node, ip string) {
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // 默认debug级别
	}

	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix // 更快更小
	//zerolog.TimeFieldFormat = "2006-01-02 15:04:05" // 秒级,不带时区
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z700" // 毫秒级,带时区
	//zerolog.TimestampFieldName = "timestamp"
	//zerolog.LevelFieldName = "Level"
	//zerolog.MessageFieldName = "msg"

	logger = &XLogger{
		Logger:       zerolog.New(newOutput(pathname, node)).With().Logger(),
		interval:     interval,
		mux:          sync.Mutex{},
		lastFileTime: time.Now(),
		path:         pathname,
		env:          env,
		serviceName:  serviceName,
		node:         node,
		ip:           ip,
	}
}

// 获取输出 控制台/文件
// 当pathname为空时,输出到控制台
func newOutput(pathname, node string) io.Writer {

	// 1. 默认标准输出
	// 2. 文件夹设置不为空时,写入文件
	if pathname != "" {
		now := time.Now().Format("20060102_15:04:05")
		var filename = fmt.Sprintf("%s.log", now)
		if node != "" {
			filename = fmt.Sprintf("%s_%s.log", node, now)
		}
		// 文件夹不存在,则创建
		if !xfile.IsExist(pathname) {
			err := os.MkdirAll(pathname, os.ModePerm)
			if err != nil {
				fmt.Println("MkdirAll path[", pathname, "] error:", err.Error())
			}
		}
		file, err := os.Create(path.Join(pathname, filename))
		if err == nil {
			return file
		} else {
			fmt.Println("create file[", filename, "] error:", err.Error())
		}
	}
	return os.Stdout
}

func Debug() *zerolog.Event {
	return newEvent(zerolog.DebugLevel)
}

func Info() *zerolog.Event {
	return newEvent(zerolog.InfoLevel)
}

func Error() *zerolog.Event {
	return newEvent(zerolog.ErrorLevel)
}

func Warn() *zerolog.Event {
	return newEvent(zerolog.WarnLevel)
}

// Fatal Fatal消息打印 (程序终止)
func Fatal() *zerolog.Event {
	return newEvent(zerolog.FatalLevel)
}

// Panic Panic消息打印 (程序不会终止)
func Panic() *zerolog.Event {
	return newEvent(zerolog.PanicLevel)
}

func log() *zerolog.Event {
	return newEvent(zerolog.NoLevel)
}

func newEvent(level zerolog.Level) *zerolog.Event {
	// 1.检测logger引擎是否初始化、是否要切割
	logger.check()
	// 2.根据level返回Event
	switch level {
	case zerolog.DebugLevel:
		return logger.Logger.Debug().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	case zerolog.InfoLevel:
		return logger.Logger.Info().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	case zerolog.WarnLevel:
		return logger.Logger.Warn().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	case zerolog.ErrorLevel:
		return logger.Logger.Error().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	case zerolog.FatalLevel:
		return logger.Logger.Fatal().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	case zerolog.PanicLevel:
		return logger.Logger.Panic().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	case zerolog.NoLevel:
		return logger.Logger.Log().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	default:
		return logger.Logger.Debug().
			Str("env", logger.env).
			Str("service", logger.serviceName).
			Str("ip", logger.ip).
			Str("node", logger.node).
			Timestamp()
	}
}

func (log *XLogger) check() {
	if log == nil {
		Init("debug", "", 0, "unknown", "unknown", "", "0.0.0.0")
	} else {
		// 日志文件切割
		if log.interval > 0 && time.Now().Add(-log.interval).After(log.lastFileTime) {
			log.mux.Lock()
			Init("debug", log.path, log.interval, log.serviceName, log.env, log.node, log.ip)
			log.mux.Unlock()
		}
	}
}
