package logx

import (
	"sync"
)

var (
	logger      Logger
	enableDebug bool

	setupLoggerOnce sync.Once
)

func Setup(debug bool) {
	setupLoggerOnce.Do(func() {
		setup(debug)
	})
}

func setup(debug bool) {
	enableDebug = debug
	logger = setupBlock(debug)
}

func get() Logger {
	Setup(enableDebug)
	return logger
}

func Error(msg string, fields ...Data) {
	get().Error(msg, fields...)
}

func Info(msg string, fields ...Data) {
	get().Info(msg, fields...)
}

func Debug(msg string, fields ...Data) {
	get().Debug(msg, fields...)
}

func Raw(msg string) {
	get().Raw(msg)
}

func Sync() error {
	return get().Sync()
}
