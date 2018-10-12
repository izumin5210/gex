package tool

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/izumin5210/gex/pkg/manager"
)

// Repository is an interface for managing and operating tools
type Repository interface {
	Add(ctx context.Context, pkgs ...string) error
	Build(ctx context.Context, t Tool) (string, error)
	BuildAll(ctx context.Context) error
	Run(ctx context.Context, name string, args ...string) error
}

type repositoryImpl struct {
	*Config
	parser   Parser
	writer   Writer
	executor manager.Executor
	builder  manager.Builder
	adder    manager.Adder
}

// NewRepository creates a new Repository instance.
func NewRepository(executor manager.Executor, builder manager.Builder, adder manager.Adder, cfg *Config) Repository {
	return &repositoryImpl{
		Config:   cfg,
		parser:   NewParser(cfg.FS),
		writer:   NewWriter(cfg.FS),
		executor: executor,
		builder:  builder,
		adder:    adder,
	}
}

func (r *repositoryImpl) Add(ctx context.Context, pkgs ...string) error {
	if r.Verbose {
		r.Log.Printf("  --> Add %s", strings.Join(pkgs, ", "))
	}
	err := r.adder.Add(ctx, pkgs, r.Verbose)
	if err != nil {
		return errors.Wrap(err, "failed to add tools")
	}

	m, err := r.parser.Parse(r.ManifestPath())
	if err != nil {
		m = NewManifest([]Tool{})
	}

	for _, pkg := range pkgs {
		pkg = strings.SplitN(pkg, "@", 2)[0]
		t := Tool(pkg)
		m.AddTool(t)
		_, err = r.Build(ctx, t)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = r.writer.Write(r.ManifestPath(), m)
	if err != nil {
		return errors.Wrap(err, "failed to write a manifest file")
	}

	return nil
}

func (r *repositoryImpl) Build(ctx context.Context, t Tool) (string, error) {
	binPath := r.BinPath(t.Name())

	if st, err := r.FS.Stat(binPath); err != nil {
		if r.Verbose {
			r.Log.Printf("  --> Build %s\n", t)
		}
		err := r.builder.Build(ctx, binPath, string(t), r.Verbose)
		if err != nil {
			return "", errors.Wrapf(err, "failed to build %s", t)
		}
	} else if st.IsDir() {
		return "", errors.Errorf("%q is a directory", t.Name())
	}

	return binPath, nil
}

func (r *repositoryImpl) BuildAll(ctx context.Context) error {
	if err := r.RequireManifest(); err != nil {
		return errors.WithStack(err)
	}

	m, err := r.parser.Parse(r.ManifestPath())
	if err != nil {
		return errors.Wrap(err, "failed to parse the manifest file")
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
	if err := r.RequireManifest(); err != nil {
		return errors.WithStack(err)
	}

	m, err := r.parser.Parse(r.ManifestPath())
	if err != nil {
		return errors.Wrap(err, "failed to parse the manifest file")
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
