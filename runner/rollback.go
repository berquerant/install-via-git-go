package runner

import (
	"berquerant/install-via-git-go/config"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/gitlock"
	"berquerant/install-via-git-go/logx"
	"context"
)

type Rollback struct {
	keeper       *gitlock.GitKeeper
	cfg          *config.Config
	localRepoDir filepathx.DirPath
	env          execx.Env
	noupdate     bool
	shell        []string
}

func NewRollback(
	keeper *gitlock.GitKeeper,
	noupdate bool,
	cfg *config.Config,
	localRepoDir filepathx.DirPath,
	env execx.Env,
	shell []string,
) *Rollback {
	return &Rollback{
		keeper:       keeper,
		cfg:          cfg,
		localRepoDir: localRepoDir,
		env:          env,
		noupdate:     noupdate,
		shell:        shell,
	}
}

func (r *Rollback) Run(ctx context.Context) {
	if r.noupdate {
		logx.Info("skip rollback repo and lockfile")
		if _, err := execx.NewExecutorFromStrings(r.cfg.Steps.Rollback, r.shell...).
			Execute(ctx, execx.WithDir(r.localRepoDir), execx.WithEnv(r.env)); err != nil {
			logx.Error("run rollback", logx.Err(err))
		}
		return
	}

	logx.Error("rollback")
	if err := r.keeper.Rollback(ctx); err != nil {
		logx.Error("rollback error", logx.Err(err))
	}
	if _, err := execx.NewExecutorFromStrings(r.cfg.Steps.Rollback, r.shell...).
		Execute(ctx, execx.WithDir(r.localRepoDir), execx.WithEnv(r.env)); err != nil {
		logx.Error("run rollback", logx.Err(err))
	}
}
