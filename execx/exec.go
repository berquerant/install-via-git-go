package execx

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
	"context"
	"io"
	"strings"

	ex "github.com/berquerant/execx"
)

//go:generate go run github.com/berquerant/goconfig@v0.3.0 -field "Dir filepathx.DirPath|Env Env" -option -output exec_config_generated.go

// Executor is a shell command executor.
type Executor interface {
	Execute(ctx context.Context, opt ...ConfigOption) (Result, error)
}

func NewExecutorFromStrings(scripts []string, shell ...string) Executor {
	if len(scripts) == 0 {
		return NewNoopExecutor()
	}
	return NewRawScript(
		"set -ex\n"+strings.Join(scripts, "\n"),
		shell...,
	)
}

func NewNoopExecutor() *NoopExecutor {
	return &NoopExecutor{}
}

type NoopExecutor struct{}

func (*NoopExecutor) Execute(_ context.Context, _ ...ConfigOption) (Result, error) {
	return Result{}, nil
}

func NewCommand(args ...string) *Command {
	return &Command{
		args: args,
	}
}

type Command struct {
	args []string
}

type Result struct {
	Args   []string
	Stderr string
	Stdout string
}

func (c *Command) Execute(ctx context.Context, opt ...ConfigOption) (Result, error) {
	config := NewConfigBuilder().Dir(filepathx.PWD()).Env(NewEnv()).Build()
	config.Apply(opt...)

	cmd := ex.New(c.args[0], c.args[1:]...)
	cmd.Dir = config.Dir.Get().String()
	cmd.Env.Merge(ex.EnvFromEnviron())
	cmd.Env.Merge(config.Env.Get())

	return run(ctx, cmd)
}

func run(ctx context.Context, cmd *ex.Cmd) (result Result, retErr error) {
	logx.Info("exec start",
		logx.S("dir", cmd.Dir),
		logx.SS("args", cmd.Args),
	)
	logx.Debug("exec start",
		logx.SS("env", cmd.Env.IntoSlice()),
	)
	defer func() {
		logx.Info("exec end", logx.Err(retErr))
	}()

	r, err := cmd.Run(
		ctx,
		ex.WithStdoutConsumer(logx.Raw),
		ex.WithStderrConsumer(logx.Raw),
	)
	if err != nil {
		retErr = errorx.Errorf(err, "exec command")
		return
	}

	stdout, err := io.ReadAll(r.Stdout)
	if err != nil {
		retErr = errorx.Errorf(err, "exec read stdout")
		return
	}
	stderr, err := io.ReadAll(r.Stderr)
	if err != nil {
		retErr = errorx.Errorf(err, "exec read stderr")
		return
	}

	result.Args = r.ExpandedArgs
	result.Stdout = string(stdout)
	result.Stderr = string(stderr)
	return
}
