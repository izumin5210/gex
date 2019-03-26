package gex

import (
	"bytes"
	"io"
	"io/ioutil"
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
	Mode         Mode

	Verbose bool
	Logger  *log.Logger
}

// Default contains default configuration.
var Default = createDefaultConfig()

func createDefaultConfig() *Config {
	wd, _ := os.Getwd()
	if wd == "" {
		wd = "."
	}
	cfg := &Config{
		OutWriter:    os.Stdout,
		ErrWriter:    os.Stderr,
		InReader:     os.Stdin,
		FS:           afero.NewOsFs(),
		Execer:       exec.New(),
		WorkingDir:   wd,
		ManifestName: "tools.go",
		BinDirName:   "bin",
		Logger:       log.New(ioutil.Discard, "", 0),
	}
	cfg.DetectMode()
	return cfg
}

// Create creates a new instance of tool.Repository to manage developemnt tools.
func (c *Config) Create() (tool.Repository, error) {
	c.setDefaultsIfNeeded()

	manager, executor, err := c.createManager()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tool.NewRepository(executor, manager, &tool.Config{
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

	if c.Mode == ModeUnknown {
		c.DetectMode()
	}

	if rootDir, err := c.findRoot(c.ManifestName); err == nil {
		if len(rootDir) > len(c.RootDir) {
			c.RootDir = rootDir
		}
	}
}

func (c *Config) createManager() (
	manager.Interface,
	manager.Executor,
	error,
) {
	executor := manager.NewExecutor(c.Execer, c.OutWriter, c.ErrWriter, c.InReader, c.WorkingDir, c.Logger)
	var (
		manager manager.Interface
	)

	switch c.Mode {
	case ModeModules:
		manager = mod.NewManager(executor)
	case ModeDep:
		manager = dep.NewManager(executor, c.RootDir, c.WorkingDir)
	default:
		return nil, nil, errors.New("failed to detect a dependencies management tool")
	}

	return manager, executor, nil
}

// Mode represents the dependencies management tool that is used.
type Mode int

// Mode values
const (
	ModeUnknown Mode = iota
	ModeModules
	ModeDep
)

// DetectMode detects a current Mode and sets a root directory.
func (c *Config) DetectMode() {
	root, err := c.findRoot("Gopkg.toml")
	if err == nil {
		c.Mode = ModeDep
		if c.RootDir == "" {
			c.RootDir = root
		}
		return
	}

	dir, ok := c.lookupMod()
	if ok {
		c.RootDir = dir
		c.Mode = ModeModules
		return
	}

	c.Mode = ModeUnknown

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

func (c *Config) lookupMod() (string, bool) {
	out, err := c.Execer.Command("go", "env", "GOMOD").CombinedOutput()
	if err == nil && len(bytes.TrimRight(out, "\n")) > 0 {
		return filepath.Dir(string(out)), true
	}

	dir, err := c.findRoot("go.mod")
	if err == nil {
		return dir, true
	}

	if os.Getenv("GO111MODULE") == "on" {
		return c.WorkingDir, true
	}

	return "", false
}
