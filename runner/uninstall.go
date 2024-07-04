package runner

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/logx"
	"berquerant/install-via-git-go/strategy"
	"context"
	"errors"
)

type Uninstall struct {
	*Argument
	runner strategy.Runner
}

func NewUninstall(
	argument *Argument,
	runner strategy.Runner,
) *Uninstall {
	return &Uninstall{
		Argument: argument,
		runner:   runner,
	}
}

func (u *Uninstall) Run(ctx context.Context) error {
	if u.LocalRepoDir.Exist() {
		logx.Info("uninstall")
		if _, err := execx.NewExecutorFromStrings(u.Config.Steps.Uninstall, u.Shell...).
			Execute(ctx, execx.WithDir(u.LocalRepoDir), execx.WithEnv(u.Env)); err != nil {
			return errorx.Errorf(err, "run uninstall")
		}
	}

	if err := u.runner.Run(ctx); err != nil {
		if !errors.Is(err, strategy.ErrNoopStrategy) {
			return errorx.Errorf(err, "run strategy")
		}
	}
	return nil
}
