package dep

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// NewBuilder creates a manager.Builder instance to build tools vendored with dep.
func NewBuilder(executor manager.Executor, rootDir, workingDir string) manager.Builder {
	return &builderImpl{
		executor:   executor,
		rootDir:    rootDir,
		workingDir: workingDir,
	}
}

type builderImpl struct {
	executor   manager.Executor
	rootDir    string
	workingDir string
}

func (b *builderImpl) Build(ctx context.Context, binPath, pkg string, verbose bool) error {
	target, err := filepath.Rel(b.workingDir, b.rootDir)
	if err != nil {
		return errors.WithStack(err)
	}
	target = filepath.Join(target, "vendor", pkg)
	if !strings.HasPrefix(target, "..") {
		target = "." + string(filepath.Separator) + target
	}
	args := []string{"build", "-o", binPath}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, target)
	return errors.WithStack(b.executor.Exec(ctx, "go", args...))
}
