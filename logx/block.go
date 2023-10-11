package logx

import (
	"os"

	"golang.org/x/exp/slog"
)

func setupBlock(debug bool) Logger {
	blockHandler := NewBlockHandler(os.Stdout)
	level := func() slog.Level {
		if debug {
			return slog.LevelDebug
		}
		return slog.LevelInfo
	}()
	levelHandler := NewLevelHandler(level, blockHandler)
	logger := slog.New(levelHandler)
	return &BlockLogger{
		logger,
	}
}

type BlockLogger struct {
	*slog.Logger
}

func (l *BlockLogger) logAttrs(level slog.Level, msg string, attrs ...Attr) {
	rawAttrs := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		rawAttrs[i] = slog.Attr(attr)
	}
	l.LogAttrs(nil, level, msg, rawAttrs...)
}

func (l *BlockLogger) Info(msg string, attrs ...Attr) {
	l.logAttrs(slog.LevelInfo, msg, attrs...)
}

func (l *BlockLogger) Error(msg string, attrs ...Attr) {
	l.logAttrs(slog.LevelError, msg, attrs...)
}

func (l *BlockLogger) Debug(msg string, attrs ...Attr) {
	l.logAttrs(slog.LevelDebug, msg, attrs...)
}

// Raw writes msg as is.
func (l *BlockLogger) Raw(msg string) {
	l.logAttrs(LevelRaw, msg)
}

func (*BlockLogger) Sync() error {
	return nil
}
