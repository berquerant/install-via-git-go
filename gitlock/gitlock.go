package gitlock

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/git"
	"berquerant/install-via-git-go/lock"
	"berquerant/install-via-git-go/logx"
	"context"
)

// Keeper manages commit hashes and local repo.
type Keeper interface {
	Locker() lock.Keeper
	// Commit writes next hash.
	Commit() error
	// Rollback writes current hash and rollback repo.
	Rollback(ctx context.Context) error
}

func NewGitKeeper(locker lock.Keeper, command git.Command) *GitKeeper {
	return &GitKeeper{
		locker:  locker,
		command: command,
	}
}

type GitKeeper struct {
	locker  lock.Keeper
	command git.Command
}

func (g *GitKeeper) Locker() lock.Keeper {
	return g.locker
}

func (g *GitKeeper) Commit() error {
	logx.Debug("gitlock commit", logx.S("hash", g.locker.Pair().Next))
	if err := g.locker.Commit(); err != nil {
		return errorx.Errorf(err, "gitlock commit")
	}
	return nil
}

func (g *GitKeeper) Rollback(ctx context.Context) error {
	logx.Debug("gitlock rollback", logx.S("hash", g.locker.Pair().Current))
	if err := g.locker.Rollback(); err != nil {
		return errorx.Errorf(err, "gitlock rollback")
	}
	current := g.locker.Pair().Current
	if current == "" {
		return nil
	}
	if err := g.command.ResetHard(ctx, current); err != nil {
		return errorx.Errorf(err, "gitlock rollback")
	}
	return nil
}
