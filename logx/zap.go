package logx

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getZapConfig(debug bool) zap.Config {
	if debug {
		return zap.NewDevelopmentConfig()
	}
	return zap.NewProductionConfig()
}

func setupZap(debug bool) Logger {
	config := getZapConfig(debug)
	config.Encoding = "console"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	config.EncoderConfig.CallerKey = zapcore.OmitKey
	zapLogger, _ := config.Build()
	return &ZapLogger{l: zapLogger}
}

type ZapLogger struct {
	l *zap.Logger
}

func (l *ZapLogger) Info(msg string, data ...Data) {
	l.l.Info(msg, intoZapFields(data)...)
}

func (l *ZapLogger) Error(msg string, data ...Data) {
	l.l.Error(msg, intoZapFields(data)...)
}

func (l *ZapLogger) Debug(msg string, data ...Data) {
	l.l.Debug(msg, intoZapFields(data)...)
}

// Raw is an alias of Info.
func (l *ZapLogger) Raw(msg string) {
	l.l.Info(msg)
}

func (l *ZapLogger) Sync() error {
	return l.l.Sync()
}

func intoZapFields(dataList []Data) []zap.Field {
	result := make([]zap.Field, len(dataList))
	for i, data := range dataList {
		result[i] = data.intoZapField()
	}
	return result
}

func (d Data) intoZapField() zap.Field {
	switch value := d.value.(type) {
	case string:
		return zap.String(d.key, value)
	case []string:
		return zap.Strings(d.key, value)
	case bool:
		return zap.Bool(d.key, value)
	case error:
		return zap.Error(value)
	default:
		return zap.Any(d.key, value)
	}
}
