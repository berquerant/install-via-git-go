package runner

import (
	"berquerant/install-via-git-go/config"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/filepathx"
	"context"
)

type Runner interface {
	Run(ctx context.Context) error
}

type Argument struct {
	Config       *config.Config
	Env          execx.Env
	Shell        []string
	LocalRepoDir filepathx.DirPath
}
