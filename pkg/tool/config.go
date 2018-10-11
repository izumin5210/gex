package tool

import (
	"log"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Config contains configurations to manage development tools.
type Config struct {
	FS           afero.Fs
	WorkingDir   string
	RootDir      string
	ManifestName string
	BinDirName   string
	Verbose      bool
	Log          *log.Logger
}

// RequireManifest returns an error if the manifest file does not exist.
func (c *Config) RequireManifest() error {
	if ok, err := afero.Exists(c.FS, c.ManifestPath()); err != nil {
		return errors.WithStack(err)
	} else if !ok {
		return errors.Errorf("could not find %s", c.ManifestPath())
	}
	return nil
}

func (c *Config) ManifestPath() string {
	return filepath.Join(c.baseDir(), c.ManifestName)
}

func (c *Config) BinDir() string {
	return filepath.Join(c.baseDir(), c.BinDirName)
}

func (c *Config) BinPath(bin string) string {
	return filepath.Join(c.BinDir(), bin)
}

func (c *Config) baseDir() (dir string) {
	dir = c.RootDir
	if dir == "" {
		dir = c.WorkingDir
	}
	return
}
