package execx

import (
	"berquerant/install-via-git-go/filepathx"
	"context"
	"io"
	"os"
)

func NewRawScript(content, shell string) *RawScript {
	return &RawScript{
		content: content,
		shell:   shell,
	}
}

type RawScript struct {
	content string
	shell   string
}

func (s *RawScript) Execute(ctx context.Context, opt ...ConfigOption) (result Result, retErr error) {
	config := NewConfigBuilder().Dir(filepathx.PWD()).Env(NewEnv()).Build()
	config.Apply(opt...)

	f, err := os.CreateTemp("", "install_via_git_execx")
	if err != nil {
		retErr = err
		return
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()

	var (
		env      = config.Env.Get().Add(EnvFromEnviron())
		expanded = env.Expand(s.content)
	)

	if _, err := io.WriteString(f, expanded); err != nil {
		retErr = err
		return
	}
	if err := os.Chmod(f.Name(), 0755); err != nil {
		retErr = err
		return
	}

	p, _ := filepathx.NewPath(f.Name())
	result, retErr = NewScript(p.FilePath(), s.shell).Execute(ctx, opt...)
	return
}

// NewScript returns a new Command that executes the script.
func NewScript(path filepathx.FilePath, shell string) *Command {
	return &Command{
		args: []string{shell, path.String()},
	}
}
