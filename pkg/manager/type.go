package manager

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/izumin5210/execx"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Type represents the dependencies management tool that is used.
type Type int

// Type values
const (
	TypeUnknown Type = iota
	TypeModules
	TypeDep
)

func (t Type) Vendor() bool { return t == TypeDep }

func (t Type) String() string {
	switch t {
	case TypeModules:
		return "mod"
	case TypeDep:
		return "dep"
	default:
		return "unknown"
	}
}

// DetectType detects a current Mode and sets a root directory.
func DetectType(workDir string, fs afero.Fs, exec *execx.Executor) (t Type, rootDir string) {
	root, err := FindRoot(workDir, fs, "Gopkg.toml")
	if err == nil {
		return TypeDep, root
	}

	dir, ok := lookupMod(workDir, fs, exec)
	if ok {
		return TypeModules, dir
	}

	return TypeUnknown, ""
}

// FindRoot gets a manifest file path.
func FindRoot(from string, fs afero.Fs, manifest string) (string, error) {
	for {
		if ok, err := afero.Exists(fs, filepath.Join(from, manifest)); ok {
			return from, nil
		} else if err != nil {
			return "", errors.WithStack(err)
		}

		parent := filepath.Dir(from)
		if parent == from {
			return "", errors.New("could not find manifest")
		}
		from = parent
	}
}

func lookupMod(workDir string, fs afero.Fs, exec *execx.Executor) (string, bool) {
	out, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
	if err == nil && len(bytes.TrimRight(out, "\n")) > 0 {
		return filepath.Dir(string(out)), true
	}

	dir, err := FindRoot(workDir, fs, "go.mod")
	if err == nil {
		return dir, true
	}

	if os.Getenv("GO111MODULE") == "on" {
		return workDir, true
	}

	return "", false
}
