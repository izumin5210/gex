package dep

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// NewManager creates a manager.Interface instance to manage tools vendored with dep.
func NewManager(executor manager.Executor, rootDir, workingDir string) manager.Interface {
	return &managerImpl{
		executor:   executor,
		rootDir:    rootDir,
		workingDir: workingDir,
	}
}

type managerImpl struct {
	executor   manager.Executor
	rootDir    string
	workingDir string
}

func (m *managerImpl) Add(ctx context.Context, pkgs []string, verbose bool) error {
	args := []string{"ensure"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "-add")
	args = append(args, pkgs...)
	return errors.WithStack(m.executor.Exec(ctx, "dep", args...))
}

func (m *managerImpl) Build(ctx context.Context, binPath, pkg string, verbose bool) error {
	target, err := filepath.Rel(m.workingDir, m.rootDir)
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
	return errors.WithStack(m.executor.Exec(ctx, "go", args...))
}
