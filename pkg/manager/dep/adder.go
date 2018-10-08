package dep

import (
	"context"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// NewAdder creates a manager.Adder instance to add tools to dep.
func NewAdder(executor manager.Executor) manager.Adder {
	return &adderImpl{
		executor: executor,
	}
}

type adderImpl struct {
	executor manager.Executor
}

func (a *adderImpl) Add(ctx context.Context, pkgs []string, verbose bool) error {
	args := []string{"ensure"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "-add")
	args = append(args, pkgs...)
	return errors.WithStack(a.executor.Exec(ctx, "dep", args...))
}
