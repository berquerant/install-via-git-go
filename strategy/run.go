package strategy

import (
	"context"
	"errors"
)

type Runner interface {
	Run(ctx context.Context) error
}

//go:generate go tool dataclass -type "RunnerConfig" -field "Repo string|Branch string|Pair *lock.Pair|Command git.Command" -output run_dataclass_generated.go

var (
	ErrNoopStrategy    = errors.New("NoopStrategy")
	ErrUnknownStrategy = errors.New("UnknownStrategy")
	ErrNoLock          = errors.New("NoLock")
)

func NewUpdateToLatestWithLock(c RunnerConfig) *UpdateToLatestWithLock {
	return &UpdateToLatestWithLock{
		c: c,
	}
}

type UpdateToLatestWithLock struct {
	c RunnerConfig
}

func (r *UpdateToLatestWithLock) Run(ctx context.Context) error {
	current := r.c.Pair().Current
	if current == "" {
		return ErrNoLock
	}
	_ = r.c.Command().Checkout(ctx, r.c.Branch())
	if err := r.c.Command().Fetch(ctx); err != nil {
		return err
	}
	if err := r.c.Command().PullForce(ctx, r.c.Branch()); err != nil {
		return err
	}
	if err := r.c.Command().Checkout(ctx, r.c.Branch()); err != nil {
		return err
	}

	next, err := r.c.Command().GetCommitHash(ctx)
	if err != nil {
		return err
	}
	r.c.Pair().Next = next
	return nil
}

func NewUpdateToLockRunner(c RunnerConfig) *UpdateToLockRunner {
	return &UpdateToLockRunner{
		c: c,
	}
}

type UpdateToLockRunner struct {
	c RunnerConfig
}

func (r *UpdateToLockRunner) Run(ctx context.Context) error {
	current := r.c.Pair().Current
	if current == "" {
		return ErrNoLock
	}
	_ = r.c.Command().Checkout(ctx, r.c.Branch())
	if err := r.c.Command().Fetch(ctx); err != nil {
		return err
	}
	if err := r.c.Command().PullForce(ctx, r.c.Branch()); err != nil {
		return err
	}

	repoCurrent, err := r.c.Command().GetCommitHash(ctx)
	if err != nil {
		return err
	}
	if current == repoCurrent {
		return nil
	}

	return r.c.Command().Checkout(ctx, current)
}

func NewCreateLatestLockRunner(c RunnerConfig) *CreateLatestLockRunner {
	return &CreateLatestLockRunner{
		c: c,
	}
}

type CreateLatestLockRunner struct {
	c RunnerConfig
}

func (r *CreateLatestLockRunner) Run(ctx context.Context) error {
	current, err := r.c.Command().GetCommitHash(ctx)
	if err != nil {
		return err
	}
	r.c.Pair().Current = current

	_ = r.c.Command().Checkout(ctx, r.c.Branch())
	if err := r.c.Command().Fetch(ctx); err != nil {
		return err
	}
	if err := r.c.Command().PullForce(ctx, r.c.Branch()); err != nil {
		return err
	}
	if err := r.c.Command().Checkout(ctx, r.c.Branch()); err != nil {
		return err
	}

	next, err := r.c.Command().GetCommitHash(ctx)
	if err != nil {
		return err
	}
	r.c.Pair().Next = next
	return nil
}

func NewCreateLockRunner(c RunnerConfig) *CreateLockRunner {
	return &CreateLockRunner{
		c: c,
	}
}

type CreateLockRunner struct {
	c RunnerConfig
}

func (r *CreateLockRunner) Run(ctx context.Context) error {
	_ = r.c.Command().Checkout(ctx, r.c.Branch())
	if err := r.c.Command().Fetch(ctx); err != nil {
		return err
	}
	if err := r.c.Command().PullForce(ctx, r.c.Branch()); err != nil {
		return err
	}
	if err := r.c.Command().Checkout(ctx, r.c.Branch()); err != nil {
		return err
	}
	current, err := r.c.Command().GetCommitHash(ctx)
	if err != nil {
		return err
	}
	r.c.Pair().Next = current
	return nil
}

func NewInitFromEmptyToLatestRunner(c RunnerConfig) *InitFromEmptyToLatestRunner {
	return &InitFromEmptyToLatestRunner{
		c: c,
	}
}

type InitFromEmptyToLatestRunner struct {
	c RunnerConfig
}

func (r *InitFromEmptyToLatestRunner) Run(ctx context.Context) error {
	if err := r.c.Command().Clone(ctx, r.c.Repo()); err != nil {
		return err
	}
	if err := r.c.Command().PullForce(ctx, r.c.Branch()); err != nil {
		return err
	}
	if err := r.c.Command().Checkout(ctx, r.c.Branch()); err != nil {
		return err
	}

	next, err := r.c.Command().GetCommitHash(ctx)
	if err != nil {
		return err
	}
	r.c.Pair().Next = next
	return nil
}

func NewInitFromEmptyToLockRunner(c RunnerConfig) *InitFromEmptyToLockRunner {
	return &InitFromEmptyToLockRunner{
		c: c,
	}
}

type InitFromEmptyToLockRunner struct {
	c RunnerConfig
}

func (r *InitFromEmptyToLockRunner) Run(ctx context.Context) error {
	if r.c.Pair().Current == "" {
		return ErrNoLock
	}
	if err := r.c.Command().Clone(ctx, r.c.Repo()); err != nil {
		return err
	}
	if err := r.c.Command().PullForce(ctx, r.c.Branch()); err != nil {
		return err
	}

	return r.c.Command().Checkout(ctx, r.c.Pair().Current)
}

func NewInitFromEmptyRunner(c RunnerConfig) *InitFromEmptyRunner {
	return &InitFromEmptyRunner{
		c: c,
	}
}

type InitFromEmptyRunner struct {
	c RunnerConfig
}

func (r *InitFromEmptyRunner) Run(ctx context.Context) error {
	if err := r.c.Command().Clone(ctx, r.c.Repo()); err != nil {
		return err
	}
	if err := r.c.Command().PullForce(ctx, r.c.Branch()); err != nil {
		return err
	}
	if err := r.c.Command().Checkout(ctx, r.c.Branch()); err != nil {
		return err
	}
	next, err := r.c.Command().GetCommitHash(ctx)
	if err != nil {
		return err
	}

	r.c.Pair().Next = next
	return nil
}

type RetryRunner struct{}

func NewRetryRunner() *RetryRunner {
	return &RetryRunner{}
}

func (*RetryRunner) Run(_ context.Context) error {
	return nil
}

func NewNoUpdateRunner() *NoUpdateRunner {
	return &NoUpdateRunner{}
}

type NoUpdateRunner struct {
}

func (*NoUpdateRunner) Run(_ context.Context) error {
	return nil
}

func NewNoopRunner() *NoopRunner {
	return &NoopRunner{}
}

type NoopRunner struct{}

func (*NoopRunner) Run(_ context.Context) error {
	return ErrNoopStrategy
}

func NewUnknownRunner() *UnknownRunner {
	return &UnknownRunner{}
}

type UnknownRunner struct{}

func (*UnknownRunner) Run(_ context.Context) error {
	return ErrUnknownStrategy
}

func NewRemoveRunner(c RunnerConfig) *RemoveRunner {
	return &RemoveRunner{
		c: c,
	}
}

type RemoveRunner struct {
	c RunnerConfig
}

func (r *RemoveRunner) Run(_ context.Context) error {
	return r.c.Command().CLI().Dir().Remove()
}
