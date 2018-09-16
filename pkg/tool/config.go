package tool

import "path/filepath"

type Config struct {
	WorkingDir   string
	ManifestName string
	BinDirName   string
	Verbose      bool
}

func (c *Config) ManifestPath() string {
	return filepath.Join(c.WorkingDir, c.ManifestName)
}

func (c *Config) BinDir() string {
	return filepath.Join(c.WorkingDir, c.BinDirName)
}

func (c *Config) BinPath(bin string) string {
	return filepath.Join(c.BinDir(), bin)
}
