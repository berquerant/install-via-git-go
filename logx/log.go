package logx

import (
	"sync"
)

var (
	logger Logger

	enableDebug     bool
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

func Error(msg string, attrs ...Attr) {
	get().Error(msg, attrs...)
}

func Info(msg string, attrs ...Attr) {
	get().Info(msg, attrs...)
}

func Debug(msg string, attrs ...Attr) {
	get().Debug(msg, attrs...)
}

func Raw(msg string) {
	get().Raw(msg)
}

func Sync() error {
	return get().Sync()
}
