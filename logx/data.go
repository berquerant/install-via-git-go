package logx

import "golang.org/x/exp/slog"

type Logger interface {
	Error(msg string, attrs ...Attr)
	Info(msg string, attrs ...Attr)
	Debug(msg string, attrs ...Attr)
	Raw(msg string)
	Sync() error
}

type Attr slog.Attr

func S(key, value string) Attr {
	return Attr(slog.String(key, value))
}

func SS(key string, value []string) Attr {
	return Attr(slog.Any(key, value))
}

func B(key string, value bool) Attr {
	return Attr(slog.Bool(key, value))
}

func Err(err error) Attr {
	return Attr(slog.Any("err", err))
}
