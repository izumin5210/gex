package gex

import (
	"io"
	"io/ioutil"
	"log"
	"os"

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
	ManagerType  manager.Type

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
	cfg.ManagerType, cfg.RootDir = manager.DetectType(cfg.WorkingDir, cfg.FS, cfg.Execer)
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

	if c.ManagerType == manager.TypeUnknown {
		c.ManagerType, c.RootDir = manager.DetectType(c.WorkingDir, c.FS, c.Execer)
	}

	if rootDir, err := manager.FindRoot(c.WorkingDir, c.FS, c.ManifestName); err == nil {
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
		m manager.Interface
	)

	switch c.ManagerType {
	case manager.TypeModules:
		m = mod.NewManager(executor)
	case manager.TypeDep:
		m = dep.NewManager(executor, c.RootDir, c.WorkingDir)
	default:
		return nil, nil, errors.New("failed to detect a dependencies management tool")
	}

	return m, executor, nil
}
