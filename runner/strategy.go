package runner

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/logx"
	"berquerant/install-via-git-go/strategy"
	"context"
	"errors"
)

type Strategy struct {
	*Argument
	runner strategy.Runner
}

func NewStrategy(
	argument *Argument,
	runner strategy.Runner,
) *Strategy {
	return &Strategy{
		Argument: argument,
		runner:   runner,
	}
}

func (s *Strategy) Run(ctx context.Context) error {
	if err := s.runner.Run(ctx); err != nil {
		if !errors.Is(err, strategy.ErrNoopStrategy) {
			return errorx.Errorf(err, "run strategy")
		}

		logx.Info("skip")
		if _, err := execx.NewExecutorFromStrings(s.Config.Steps.Skip, s.Shell...).
			Execute(ctx, execx.WithDir(s.LocalRepoDir), execx.WithEnv(s.Env)); err != nil {
			return errorx.Errorf(err, "run skip")
		}
		return nil
	}

	logx.Info("install")
	if _, err := execx.NewExecutorFromStrings(s.Config.Steps.Install, s.Shell...).
		Execute(ctx, execx.WithDir(s.LocalRepoDir), execx.WithEnv(s.Env)); err != nil {
		return errorx.Errorf(err, "run install")
	}
	return nil
}
