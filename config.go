package gex

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"k8s.io/utils/exec"

	"github.com/izumin5210/gex/pkg/manager"
	"github.com/izumin5210/gex/pkg/manager/dep"
	"github.com/izumin5210/gex/pkg/manager/mod"
	"github.com/izumin5210/gex/pkg/tool"
)

// Config specifies the configration for managing development tools.
type Config struct {
	OutWriter io.Writer
	ErrWriter io.Writer
	InReader  io.Reader

	FS     afero.Fs
	Execer exec.Interface

	WorkingDir   string
	RootDir      string
	ManifestName string
	BinDirName   string

	Verbose bool
	Logger  *log.Logger

	mode Mode
}

// Default contains default configuration.
var Default = createDefaultConfig()

func createDefaultConfig() *Config {
	wd, _ := os.Getwd()
	if wd == "" {
		wd = "."
	}
	return &Config{
		OutWriter:    os.Stdout,
		ErrWriter:    os.Stderr,
		InReader:     os.Stdin,
		FS:           afero.NewOsFs(),
		Execer:       exec.New(),
		WorkingDir:   wd,
		ManifestName: "tools.go",
		BinDirName:   "bin",
		Logger:       log.New(os.Stdout, "", 0),
	}
}

// Create creates a new instance of tool.Repository to manage developemnt tools.
func (c *Config) Create() (tool.Repository, error) {
	c.setDefaultsIfNeeded()

	manager, err := c.createManager()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tool.NewRepository(manager, manager, manager, &tool.Config{
		FS:           c.FS,
		WorkingDir:   c.WorkingDir,
		RootDir:      c.RootDir,
		ManifestName: c.ManifestName,
		BinDirName:   c.BinDirName,
		Verbose:      c.Verbose,
		Log:          c.Logger,
	}), nil
}

func (c *Config) setDefaultsIfNeeded() {
	d := createDefaultConfig()

	if c.OutWriter == nil {
		c.OutWriter = d.OutWriter
	}
	if c.ErrWriter == nil {
		c.ErrWriter = d.ErrWriter
	}
	if c.InReader == nil {
		c.InReader = d.InReader
	}
	if c.FS == nil {
		c.FS = d.FS
	}
	if c.Execer == nil {
		c.Execer = d.Execer
	}
	if c.WorkingDir == "" {
		c.WorkingDir = d.WorkingDir
	}
	if c.ManifestName == "" {
		c.ManifestName = d.ManifestName
	}
	if c.BinDirName == "" {
		c.BinDirName = d.BinDirName
	}
	if c.Logger == nil {
		c.Logger = d.Logger
	}

	if c.RootDir == "" {
		c.RootDir, _ = c.findRoot(c.ManifestName)
	}
}

func (c *Config) createManager() (
	interface {
		manager.Adder
		manager.Builder
		manager.Executor
	},
	error,
) {
	executor := manager.NewExecutor(c.Execer, c.OutWriter, c.ErrWriter, c.InReader, c.WorkingDir, c.Verbose, c.Logger)
	var (
		builder manager.Builder
		adder   manager.Adder
	)

	switch c.DetectMode() {
	case ModeModules:
		builder = mod.NewBuilder(executor)
		adder = mod.NewAdder(executor)
	case ModeDep:
		builder = dep.NewBuilder(executor, c.RootDir, c.WorkingDir)
		adder = dep.NewAdder(executor)
	default:
		return nil, errors.New("failed to detect a dependencies management tool")
	}

	return &struct {
		manager.Adder
		manager.Builder
		manager.Executor
	}{
		Adder:    adder,
		Builder:  builder,
		Executor: executor,
	}, nil
}

// Mode represents the dependencies management tool that is used.
type Mode int

// Mode values
const (
	ModeUnknown Mode = iota
	ModeModules
	ModeDep
)

// DetectMode returns a current Mode.
func (c *Config) DetectMode() (m Mode) {
	if c.mode != ModeUnknown {
		m = c.mode
		return
	}

	defer func() { c.mode = m }()

	out, err := c.Execer.Command("go", "env", "GOMOD").Output()
	if err == nil && len(bytes.TrimRight(out, "\n")) > 0 {
		m = ModeModules
		return
	}

	_, err = c.findRoot("Gopkg.toml")
	if err == nil {
		m = ModeDep
		return
	}

	return
}

func (c *Config) findRoot(manifest string) (string, error) {
	from := c.WorkingDir
	for {
		if ok, err := afero.Exists(c.FS, filepath.Join(from, manifest)); ok {
			return from, nil
		} else if err != nil {
			return "", errors.WithStack(err)
		}

		parent := filepath.Dir(from)
		if parent == from {
			return "", errors.Errorf("could not find %s", c.ManifestName)
		}
		from = parent
	}
}
