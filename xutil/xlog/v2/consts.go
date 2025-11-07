package v2

import "time"

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

const (
	ModeConsole = "console"
	ModeFile    = "file"
)

const (
	// default value
	defaultLevel            = LevelDebug
	defaultLevelFieldName   = "level"
	defaultTimeFieldName    = "time"
	defaultMessageFieldName = "message"
	defaultTimeFormat       = time.RFC3339
	defaultMode             = ModeConsole

	// default value about file
	defaultFilename   = "./logs.log"
	defaultMaxSize    = 100 // MB
	defaultMaxBackups = 10
	defaultMaxAge     = 30
	defaultCompress   = false
	defaultLocalTime  = false
)
