package manager

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"k8s.io/utils/exec"
)

// Type represents the dependencies management tool that is used.
type Type int

// Type values
const (
	TypeUnknown Type = iota
	TypeModules
	TypeDep
)

// DetectType detects a current Mode and sets a root directory.
func DetectType(workDir string, fs afero.Fs, execer exec.Interface) (t Type, rootDir string) {
	root, err := FindRoot(workDir, fs, "Gopkg.toml")
	if err == nil {
		return TypeDep, root
	}

	dir, ok := lookupMod(workDir, fs, execer)
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

func lookupMod(workDir string, fs afero.Fs, execer exec.Interface) (string, bool) {
	out, err := execer.Command("go", "env", "GOMOD").CombinedOutput()
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
