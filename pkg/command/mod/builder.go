package mod

import (
	"context"

	"github.com/izumin5210/gex/pkg/command"
)

// NewBuilder creates a command.Builder instance to build tools vendored with Modules.
func NewBuilder(executor command.Executor) command.Builder {
	return &builderImpl{
		executor: executor,
	}
}

type builderImpl struct {
	executor command.Executor
}

func (b *builderImpl) Build(ctx context.Context, binPath, pkg string, verbose bool) error {
	args := []string{"build", "-o", binPath}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, pkg)
	return b.executor.Exec(ctx, "go", args...)
}
