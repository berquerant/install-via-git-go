package git

import (
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/filepathx"
	"context"
	"errors"
	"fmt"
	"strings"
)

type CLI interface {
	Execute(ctx context.Context, args ...string) (string, error)
	Command() string
	Dir() filepathx.DirPath
	Env() execx.Env
}

func NewCLI(dir filepathx.DirPath, env execx.Env, command string) *CLIImpl {
	return &CLIImpl{
		dir:     dir,
		command: command,
		env:     env,
	}
}

type CLIImpl struct {
	command string
	dir     filepathx.DirPath
	env     execx.Env
}

var (
	ErrCLI = errors.New("GitCLI")
)

func (c CLIImpl) Env() execx.Env {
	return c.env
}

func (c CLIImpl) Command() string {
	return c.command
}

func (c CLIImpl) Dir() filepathx.DirPath {
	return c.dir
}

func (c CLIImpl) Execute(ctx context.Context, args ...string) (string, error) {
	r, err := execx.NewCommand(
		append([]string{c.command}, args...)...,
	).
		Execute(ctx, execx.WithDir(c.dir), execx.WithEnv(c.env))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(r.Stdout), nil
}

type Command interface {
	Clone(ctx context.Context, repo string) error
	GetCommitHash(ctx context.Context) (string, error)
	Fetch(ctx context.Context) error
	Checkout(ctx context.Context, commit string) error
	ResetHard(ctx context.Context, commit string) error
	PullForce(ctx context.Context, repo string) error
}

func NewCommand(cli CLI) *CommandImpl {
	return &CommandImpl{
		cli: cli,
	}
}

type CommandImpl struct {
	cli CLI
}

func (c CommandImpl) GetCommitHash(ctx context.Context) (string, error) {
	return c.cli.Execute(ctx, "rev-parse", "HEAD")
}

func (c CommandImpl) Clone(ctx context.Context, repo string) error {
	_, err := execx.NewCommand(
		c.cli.Command(),
		"clone",
		repo,
		c.cli.Dir().Tail(),
	).Execute(ctx, execx.WithDir(c.cli.Dir().Parent().DirPath()), execx.WithEnv(c.cli.Env()))
	return err
}

func (c CommandImpl) Fetch(ctx context.Context) error {
	_, err := c.cli.Execute(ctx, "fetch", "--prune")
	return err
}

func (c CommandImpl) Checkout(ctx context.Context, commit string) error {
	_, err := c.cli.Execute(ctx, "checkout", commit)
	return err
}

func (c CommandImpl) ResetHard(ctx context.Context, commit string) error {
	_, err := c.cli.Execute(ctx, "reset", "--hard", fmt.Sprintf("origin/%s", commit))
	return err
}

func (c CommandImpl) PullForce(ctx context.Context, repo string) error {
	_, err := c.cli.Execute(ctx, "pull", "--prune", "--force", "origin", repo)
	return err
}
