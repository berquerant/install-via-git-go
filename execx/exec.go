package execx

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
	"bufio"
	"context"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

//go:generate go run github.com/berquerant/goconfig@v0.3.0 -field "Dir filepathx.DirPath|Env Env" -option -output exec_config_generated.go

// Executor is a shell command executor.
type Executor interface {
	Execute(ctx context.Context, opt ...ConfigOption) (Result, error)
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

func (c *Command) Execute(ctx context.Context, opt ...ConfigOption) (result Result, retErr error) {
	config := NewConfigBuilder().Dir(filepathx.PWD()).Env(NewEnv()).Build()
	config.Apply(opt...)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		env          = config.Env.Get().Add(EnvFromEnviron())
		expandedArgs = env.ExpandStrings(c.args)
		envSlice     = env.IntoSlice()
	)

	logx.Info("exec start",
		logx.S("dir", config.Dir.Get().String()),
		logx.SS("args", c.args),
		logx.SS("expanded", expandedArgs),
	)
	logx.Debug("exec start",
		logx.SS("env", envSlice),
	)
	defer func() {
		logx.Info("exec end", logx.Err(retErr))
	}()

	cmd := exec.CommandContext(ctx, expandedArgs[0], expandedArgs[1:]...)
	cmd.Dir = config.Dir.Get().String()
	cmd.Env = envSlice
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		retErr = errorx.Errorf(err, "stdout pipe")
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		retErr = errorx.Errorf(err, "stderr pipe")
		return
	}
	var (
		stdoutBuf strings.Builder
		stderrBuf strings.Builder
		eg, _     = errgroup.WithContext(ctx)
	)

	if err := cmd.Start(); err != nil {
		retErr = errorx.Errorf(err, "command start")
		return
	}
	// read stdout and stderr
	eg.Go(func() error {
		stdoutScanner := bufio.NewScanner(stdout)
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			stdoutBuf.WriteString(line + "\n")
			logx.Raw(line)
		}
		return stdoutScanner.Err()
	})
	eg.Go(func() error {
		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			stderrBuf.WriteString(line + "\n")
			logx.Raw(line)
		}
		return stderrScanner.Err()
	})
	if err := eg.Wait(); err != nil {
		retErr = errorx.Errorf(err, "read wait")
		return
	}

	if err := cmd.Wait(); err != nil {
		retErr = errorx.Errorf(err, "command wait")
		return
	}

	result.Args = expandedArgs
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	return
}
