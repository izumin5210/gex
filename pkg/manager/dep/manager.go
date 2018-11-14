package dep

import (
	"context"
	"encoding/json"
	"path"
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
	var err error
	pkgs, err = m.pickNewPackages(ctx, pkgs)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(pkgs) == 0 {
		return nil
	}

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

func (m *managerImpl) Sync(ctx context.Context, verbose bool) error {
	args := []string{"ensure"}
	if verbose {
		args = append(args, "-v")
	}
	return errors.WithStack(m.executor.Exec(ctx, "dep", args...))
}

func (m *managerImpl) pickNewPackages(ctx context.Context, pkgs []string) ([]string, error) {
	pkgSet, err := m.getExistingPackageSet(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(pkgSet) == 0 {
		return pkgs, nil
	}

	result := make([]string, 0, len(pkgs))

	for _, pkg := range pkgs {
		var skipped bool
		for pkg := pkg; pkg != "."; pkg = path.Dir(pkg) {
			if _, ok := pkgSet[pkg]; ok {
				skipped = true
				break
			}
		}
		if !skipped {
			result = append(result, pkg)
		}
	}

	return result, nil
}

func (m *managerImpl) getExistingPackageSet(ctx context.Context) (map[string]struct{}, error) {
	out, err := m.executor.Output(ctx, "dep", "status", "-json")
	if err != nil {
		return make(map[string]struct{}), nil
	}
	pkgs := []struct{ ProjectRoot string }{}
	err = json.Unmarshal(out, &pkgs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pkgRoots := make(map[string]struct{}, len(pkgs))
	for _, pkg := range pkgs {
		pkgRoots[pkg.ProjectRoot] = struct{}{}
	}

	return pkgRoots, nil
}
