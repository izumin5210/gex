package mod

import (
	"context"

	"github.com/izumin5210/gex/pkg/command"
)

// NewAdder creates a command.Adder instance to add tools to Modules.
func NewAdder(executor command.Executor) command.Adder {
	return &adderImpl{
		executor: executor,
	}
}

type adderImpl struct {
	executor command.Executor
}

func (a *adderImpl) Add(ctx context.Context, pkgs []string, verbose bool) error {
	args := []string{"get"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, pkgs...)
	return a.executor.Exec(ctx, "go", args...)
}
