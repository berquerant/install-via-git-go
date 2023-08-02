package execx

import (
	"berquerant/install-via-git-go/filepathx"
	"context"
	"io"
	"os"
)

func NewRawScript(content string) *RawScript {
	return &RawScript{
		content: content,
	}
}

type RawScript struct {
	content string
}

func (s *RawScript) Execute(ctx context.Context, opt ...ConfigOption) (result Result, retErr error) {
	f, err := os.CreateTemp("", "install_via_git_execx")
	if err != nil {
		retErr = err
		return
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()

	if _, err := io.WriteString(f, s.content); err != nil {
		retErr = err
		return
	}
	if err := os.Chmod(f.Name(), 0755); err != nil {
		retErr = err
		return
	}

	p, _ := filepathx.NewPath(f.Name())
	result, retErr = NewScript(p.FilePath()).Execute(ctx, opt...)
	return
}

// NewScript returns a new Command that executes the script by bash.
func NewScript(path filepathx.FilePath) *Command {
	return &Command{
		args: []string{"bash", path.String()},
	}
}
