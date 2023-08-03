package logx

import (
	"fmt"
	"strings"
	"time"
)

func setupBlock(_ bool) Logger {
	return &BlockLogger{
		w: func(v string) {
			fmt.Println(v)
		},
	}
}

type BlockLogger struct {
	w func(string)
}

func (l *BlockLogger) Info(msg string, data ...Data) {
	l.write("INFO", msg, data...)
}

func (l *BlockLogger) Error(msg string, data ...Data) {
	l.write("ERROR", msg, data...)
}

func (l *BlockLogger) Debug(msg string, data ...Data) {
	if enableDebug {
		l.write("DEBUG", msg, data...)
	}
}

// Raw writes msg as is.
func (l *BlockLogger) Raw(msg string) {
	l.w(msg)
}

func (*BlockLogger) Sync() error {
	return nil
}

func (l *BlockLogger) write(level, msg string, data ...Data) {
	var b strings.Builder
	_, _ = b.WriteString(fmt.Sprintf("%s\t%s\t%s\n", time.Now().Format(time.DateTime), level, msg))
	for _, field := range intoBlockFields(data) {
		b.WriteString(field + "\n")
	}
	_, _ = b.WriteString("----------------------------------------")
	l.w(b.String())
}

func intoBlockFields(dataList []Data) []string {
	result := make([]string, len(dataList))
	for i, data := range dataList {
		result[i] = data.intoBlockField()
	}
	return result
}

func (d *Data) intoBlockField() string {
	return fmt.Sprintf("%s=%v", d.key, d.value)
}
