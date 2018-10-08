package mod

import (
	"context"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// NewAdder creates a manager.Adder instance to add tools to Modules.
func NewAdder(executor manager.Executor) manager.Adder {
	return &adderImpl{
		executor: executor,
	}
}

type adderImpl struct {
	executor manager.Executor
}

func (a *adderImpl) Add(ctx context.Context, pkgs []string, verbose bool) error {
	args := []string{"get"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, pkgs...)
	return errors.WithStack(a.executor.Exec(ctx, "go", args...))
}
