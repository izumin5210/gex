package mod

import (
	"context"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// NewManager creates a manager.Interface instance to build tools vendored with Modules.
func NewManager(executor manager.Executor) manager.Interface {
	return &managerImpl{
		executor: executor,
	}
}

type managerImpl struct {
	executor manager.Executor
}

func (m *managerImpl) Add(ctx context.Context, pkgs []string, verbose bool) error {
	args := []string{"get"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, pkgs...)
	return errors.WithStack(m.executor.Exec(ctx, "go", args...))
}

func (m *managerImpl) Build(ctx context.Context, binPath, pkg string, verbose bool) error {
	args := []string{"build", "-o", binPath}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, pkg)
	return errors.WithStack(m.executor.Exec(ctx, "go", args...))
}

func (m *managerImpl) RunInPlace(ctx context.Context, pkg string, verbose bool, commandArgs ...string) error {
	args := []string{"run"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, pkg)
	args = append(args, commandArgs...)
	return errors.WithStack(m.executor.Exec(ctx, "go", args...))
}

func (m *managerImpl) Sync(ctx context.Context, verbose bool) error {
	args := []string{"mod", "tidy"}
	if verbose {
		args = append(args, "-v")
	}
	return errors.WithStack(m.executor.Exec(ctx, "go", args...))
}
