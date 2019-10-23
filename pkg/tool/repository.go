package tool

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// Repository is an interface for managing and operating tools
type Repository interface {
	List(ctx context.Context) ([]Tool, error)
	Add(ctx context.Context, pkgs ...string) error
	Build(ctx context.Context, t Tool) (string, error)
	BuildAll(ctx context.Context) error
	Run(ctx context.Context, name string, args ...string) error
}

type repositoryImpl struct {
	*Config
	parser      Parser
	writer      Writer
	executor    manager.Executor
	manager     manager.Interface
	managerType manager.Type
}

// NewRepository creates a new Repository instance.
func NewRepository(executor manager.Executor, manager manager.Interface, managerType manager.Type, cfg *Config) Repository {
	return &repositoryImpl{
		Config:      cfg,
		parser:      NewParser(cfg.FS, managerType),
		writer:      NewWriter(cfg.FS),
		executor:    executor,
		manager:     manager,
		managerType: managerType,
	}
}

func (r *repositoryImpl) List(ctx context.Context) ([]Tool, error) {
	m, err := r.getManifest()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return m.Tools(), nil
}

func (r *repositoryImpl) Add(ctx context.Context, pkgs ...string) error {
	r.Log.Println("add", strings.Join(pkgs, ", "))

	for _, pkg := range pkgs {
		if strings.Contains(pkg, "@") {
			err := r.manager.Add(ctx, pkgs, r.Verbose)
			if err != nil {
				return errors.Wrap(err, "failed to add tools")
			}
			break
		}
	}

	m, err := r.parser.Parse(r.ManifestPath())
	if err != nil {
		m = NewManifest([]Tool{}, r.managerType)
	}

	tools := make([]Tool, len(pkgs))

	for i, pkg := range pkgs {
		pkg = strings.SplitN(pkg, "@", 2)[0]
		t := Tool(pkg)
		m.AddTool(t)
		tools[i] = t
	}

	err = r.writer.Write(r.ManifestPath(), m)
	if err != nil {
		return errors.Wrap(err, "failed to write a manifest file")
	}

	err = r.manager.Sync(ctx, r.Verbose)
	if err != nil {
		return errors.Wrap(err, "failed to sync packages")
	}

	for _, t := range tools {
		_, err = r.Build(ctx, t)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (r *repositoryImpl) Build(ctx context.Context, t Tool) (string, error) {
	binPath := r.BinPath(t.Name())

	if st, err := r.FS.Stat(binPath); err != nil {
		r.Log.Println("build", t)
		err := r.manager.Build(ctx, binPath, string(t), r.Verbose)
		if err != nil {
			return "", errors.Wrapf(err, "failed to build %s", t)
		}
	} else if st.IsDir() {
		return "", errors.Errorf("%q is a directory", t.Name())
	}

	return binPath, nil
}

func (r *repositoryImpl) BuildAll(ctx context.Context) error {
	m, err := r.getManifest()
	if err != nil {
		return errors.WithStack(err)
	}

	for _, t := range m.Tools() {
		_, err = r.Build(ctx, t)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *repositoryImpl) Run(ctx context.Context, name string, args ...string) error {
	m, err := r.getManifest()
	if err != nil {
		return errors.WithStack(err)
	}

	t, ok := m.FindTool(name)
	if !ok {
		return errors.Errorf("failed to find the tool %q", name)
	}

	bin, err := r.Build(ctx, t)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(r.executor.Exec(ctx, bin, args...))
}

func (r *repositoryImpl) getManifest() (*Manifest, error) {
	if err := r.RequireManifest(); err != nil {
		return nil, errors.WithStack(err)
	}

	m, err := r.parser.Parse(r.ManifestPath())
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse the manifest file")
	}

	return m, nil
}
