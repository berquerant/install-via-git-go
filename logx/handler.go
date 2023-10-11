package logx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

type LevelHandler struct {
	level   slog.Leveler
	handler slog.Handler
}

func NewLevelHandler(level slog.Leveler, handler slog.Handler) *LevelHandler {
	return &LevelHandler{
		level:   level,
		handler: handler,
	}
}

func (h *LevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithAttrs(attrs))
}

func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithGroup(name))
}

type BlockHandler struct {
	w io.Writer
}

const (
	LevelRaw = slog.Level(2)
)

func NewBlockHandler(w io.Writer) *BlockHandler {
	return &BlockHandler{
		w: w,
	}
}

func (h *BlockHandler) Handle(_ context.Context, r slog.Record) error {
	var (
		b     bytes.Buffer
		write = func(format string, v ...any) {
			_, _ = b.WriteString(fmt.Sprintf(format, v...))
		}
	)
	// ignore groups
	switch r.Level {
	case LevelRaw:
		write(r.Message)
	default:
		write("%s\t%s\t%s\n", r.Time.Format(time.DateTime), r.Level, r.Message)
		var (
			i  int
			ss = make([]string, r.NumAttrs())
		)
		r.Attrs(func(attr slog.Attr) bool {
			ss[i] = fmt.Sprintf("%s=%s", attr.Key, attr.Value.String())
			i++
			return true
		})
		write("%s\n----------------------------------------",
			strings.Join(ss, "\n"),
		)
	}

	write("\n")
	_, _ = h.w.Write(b.Bytes())
	return nil
}

func (h *BlockHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *BlockHandler) WithAttrs(_ []slog.Attr) slog.Handler         { return h }
func (h *BlockHandler) WithGroup(_ string) slog.Handler              { return h }
