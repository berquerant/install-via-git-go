package runner

import (
	"berquerant/install-via-git-go/config"
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
	"berquerant/install-via-git-go/strategy"
	"context"
	"errors"
)

type Strategy struct {
	cfg          *config.Config
	localRepoDir filepathx.DirPath
	env          execx.Env
	runner       strategy.Runner
	shell        []string
}

func NewStrategy(
	runner strategy.Runner,
	cfg *config.Config,
	localRepoDir filepathx.DirPath,
	env execx.Env,
	shell []string,
) *Strategy {
	return &Strategy{
		cfg:          cfg,
		env:          env,
		localRepoDir: localRepoDir,
		runner:       runner,
		shell:        shell,
	}
}

func (s *Strategy) Run(ctx context.Context) error {
	if err := s.runner.Run(ctx); err != nil {
		if !errors.Is(err, strategy.ErrNoopStrategy) {
			return errorx.Errorf(err, "run strategy")
		}

		logx.Info("skip")
		if _, err := execx.NewExecutorFromStrings(s.cfg.Steps.Skip, s.shell...).
			Execute(ctx, execx.WithDir(s.localRepoDir), execx.WithEnv(s.env)); err != nil {
			return errorx.Errorf(err, "run skip")
		}
		return nil
	}

	logx.Info("install")
	if _, err := execx.NewExecutorFromStrings(s.cfg.Steps.Install, s.shell...).
		Execute(ctx, execx.WithDir(s.localRepoDir), execx.WithEnv(s.env)); err != nil {
		return errorx.Errorf(err, "run install")
	}
	return nil
}
