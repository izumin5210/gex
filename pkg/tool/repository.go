package tool

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/izumin5210/gex/pkg/command"
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
	fs       afero.Fs
	parser   Parser
	writer   Writer
	executor command.Executor
	builder  command.Builder
	adder    command.Adder
}

// NewRepository creates a new Repository instance.
func NewRepository(fs afero.Fs, executor command.Executor, builder command.Builder, adder command.Adder, cfg *Config) Repository {
	return &repositoryImpl{
		Config:   cfg,
		fs:       fs,
		parser:   NewParser(fs),
		writer:   NewWriter(fs),
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

	if st, err := r.fs.Stat(binPath); err != nil {
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
