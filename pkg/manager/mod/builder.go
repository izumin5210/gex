package mod

import (
	"context"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// NewBuilder creates a manager.Builder instance to build tools vendored with Modules.
func NewBuilder(executor manager.Executor) manager.Builder {
	return &builderImpl{
		executor: executor,
	}
}

type builderImpl struct {
	executor manager.Executor
}

func (b *builderImpl) Build(ctx context.Context, binPath, pkg string, verbose bool) error {
	args := []string{"build", "-o", binPath}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, pkg)
	return errors.WithStack(b.executor.Exec(ctx, "go", args...))
}
