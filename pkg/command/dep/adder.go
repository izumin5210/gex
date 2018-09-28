package dep

import (
	"context"

	"github.com/izumin5210/gex/pkg/command"
)

// NewAdder creates a command.Adder instance to add tools to dep.
func NewAdder(executor command.Executor) command.Adder {
	return &adderImpl{
		executor: executor,
	}
}

type adderImpl struct {
	executor command.Executor
}

func (a *adderImpl) Add(ctx context.Context, pkgs []string, verbose bool) error {
	args := []string{"ensure"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "-add")
	args = append(args, pkgs...)
	return a.executor.Exec(ctx, "dep", args...)
}
