package execx

import (
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
	"context"

	ex "github.com/berquerant/execx"
)

func NewRawScript(content string, shell ...string) *RawScript {
	return &RawScript{
		content: content,
		shell:   shell,
	}
}

type RawScript struct {
	content string
	shell   []string
}

func (s *RawScript) Execute(ctx context.Context, opt ...ConfigOption) (Result, error) {
	config := NewConfigBuilder().Dir(filepathx.PWD()).Env(NewEnv()).Build()
	config.Apply(opt...)

	script := ex.NewScript(s.content, s.shell[0], s.shell[1:]...)
	script.Env.Merge(ex.EnvFromEnviron())
	script.Env.Merge(config.Env.Get())

	var (
		result Result
		err    error
	)
	err = script.Runner(func(cmd *ex.Cmd) error {
		logx.Debug("exec script")
		logx.DebugRaw(s.content)

		cmd.Dir = config.Dir.Get().String()
		r, err := run(ctx, cmd)
		if err != nil {
			return err
		}
		result = r
		return nil
	})
	return result, err
}
