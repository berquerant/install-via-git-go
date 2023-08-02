package logx

import (
	"sync"
)

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=Type -output type_stringer_generated.go

type Type int

const (
	Tunknown Type = iota
	Tzap
	Tblock
)

var (
	loggerType  Type
	logger      Logger
	enableDebug bool

	setupLoggerOnce sync.Once
)

func Setup(debug bool, loggerType Type) {
	setupLoggerOnce.Do(func() {
		setup(debug, loggerType)
	})
}

func setup(debug bool, loggerType Type) {
	enableDebug = debug
	switch loggerType {
	case Tzap:
		logger = setupZap(debug)
	default:
		logger = setupBlock(debug)
	}
}

func get() Logger {
	Setup(enableDebug, loggerType)
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

func Sync() error {
	return get().Sync()
}
