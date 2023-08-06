package inspect

import (
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/git"
	"berquerant/install-via-git-go/strategy"
	"context"
)

func RepoStatus(ctx context.Context, command git.Command, lockFile filepathx.FilePath) strategy.RepoStatus {
	if !lockFile.Exist() {
		return strategy.RSunknown
	}
	commit, err := lockFile.Read()
	if err != nil {
		return strategy.RSunknown
	}
	current, err := command.GetCommitHash(ctx)
	if err != nil {
		return strategy.RSunknown
	}
	if current == commit {
		return strategy.RSmatch
	}
	return strategy.RSconflict
}

func UpdateSpec(update, retry, noupdate bool) strategy.UpdateSpec {
	switch {
	case noupdate:
		return strategy.USnoupdate
	case retry:
		return strategy.USretry
	case update:
		return strategy.USforce
	default:
		return strategy.USunspec
	}
}

func LockExistence(lockFile filepathx.FilePath) strategy.LockExistence {
	if !lockFile.Exist() {
		return strategy.LEnone
	}
	content, err := lockFile.Read()
	if err != nil || content == "" {
		return strategy.LEnone
	}
	return strategy.LEexist
}

func RepoExistence(ctx context.Context, command git.Command) strategy.RepoExistence {
	if _, err := command.GetCommitHash(ctx); err != nil {
		return strategy.REnone
	}
	return strategy.REexist
}
