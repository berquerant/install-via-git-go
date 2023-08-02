package script

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/filepathx"
	"context"
)

type Runner interface {
	Run(ctx context.Context, opt ...execx.ConfigOption) error
}

func NewSerialRunner(executors []execx.Executor) *SerialRunner {
	return &SerialRunner{
		executors: executors,
	}
}

type SerialRunner struct {
	executors []execx.Executor
}

func (r *SerialRunner) Run(ctx context.Context, opt ...execx.ConfigOption) error {
	config := execx.NewConfigBuilder().Dir(filepathx.PWD()).Env(execx.NewEnv()).Build()
	config.Apply(opt...)

	for i, exe := range r.executors {
		_, err := exe.Execute(ctx, execx.WithDir(config.Dir.Get()), execx.WithEnv(config.Env.Get()))
		if err != nil {
			return errorx.Errorf(err, "serial runner index %d", i)
		}
	}
	return nil
}
