package runner

import (
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/gitlock"
	"berquerant/install-via-git-go/logx"
	"context"
)

type Rollback struct {
	*Argument
	keeper   *gitlock.GitKeeper
	noupdate bool
}

func NewRollback(
	argument *Argument,
	keeper *gitlock.GitKeeper,
	noupdate bool,
) *Rollback {
	return &Rollback{
		Argument: argument,
		keeper:   keeper,
		noupdate: noupdate,
	}
}

func (r *Rollback) Run(ctx context.Context) error {
	if r.noupdate {
		logx.Info("skip rollback repo and lockfile")
		if _, err := execx.NewExecutorFromStrings(r.Config.Steps.Rollback, r.Shell...).
			Execute(ctx, execx.WithDir(r.LocalRepoDir), execx.WithEnv(r.Env)); err != nil {
			logx.Error("run rollback", logx.Err(err))
		}
		return nil
	}

	logx.Error("rollback")
	if err := r.keeper.Rollback(ctx); err != nil {
		logx.Error("rollback error", logx.Err(err))
	}
	if _, err := execx.NewExecutorFromStrings(r.Config.Steps.Rollback, r.Shell...).
		Execute(ctx, execx.WithDir(r.LocalRepoDir), execx.WithEnv(r.Env)); err != nil {
		logx.Error("run rollback", logx.Err(err))
	}
	return nil
}
