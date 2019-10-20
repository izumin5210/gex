package manager_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/izumin5210/execx"
	"github.com/spf13/afero"

	"github.com/izumin5210/gex/pkg/manager"
)

func TestDetectType(t *testing.T) {
	wd := "/go/src/awesomeapp/foobar"

	if v, ok := os.LookupEnv("GO111MODULE"); ok {
		defer func() { os.Setenv("GO111MODULE", v) }()
		os.Unsetenv("GO111MODULE")
	}

	checkGoEnvCmd := func(t *testing.T, args []string) {
		t.Helper()
		if diff := cmp.Diff(args, []string{"go", "env", "GOMOD"}); diff != "" {
			t.Errorf("args differs: (-want +got)\n%s", diff)
		}
	}
	createExec := func(t *testing.T, out string) *execx.Executor {
		t.Helper()
		return execx.New(
			execx.WithFakeProcess(func(_ context.Context, cmd *exec.Cmd) error {
				checkGoEnvCmd(t, cmd.Args)
				fmt.Fprintln(cmd.Stdout, out+"\n")
				return nil
			}),
		)
	}
	dieIf := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("should not be error, got %v", err)
		}
	}
	createFS := func(t *testing.T) afero.Fs {
		t.Helper()
		fs := afero.NewMemMapFs()
		dieIf(t, fs.MkdirAll(wd, 0755))
		return fs
	}

	cases := []struct {
		test   string
		fs     afero.Fs
		execer *execx.Executor
		typ    manager.Type
		root   string
	}{
		{
			test:   "modules",
			fs:     createFS(t),
			execer: createExec(t, filepath.Join(wd, "go.mod")),
			typ:    manager.TypeModules,
			root:   wd,
		},
		{
			test:   "modules from subdirectory",
			fs:     createFS(t),
			execer: createExec(t, filepath.Join(filepath.Dir(wd), "go.mod")),
			typ:    manager.TypeModules,
			root:   filepath.Dir(wd),
		},
		{
			test: "dep",
			fs: func() afero.Fs {
				fs := createFS(t)
				path := filepath.Join(wd, "Gopkg.toml")
				dieIf(t, afero.WriteFile(fs, path, []byte(""), 0644))
				return fs
			}(),
			execer: createExec(t, ""),
			typ:    manager.TypeDep,
			root:   wd,
		},
		{
			test: "dep from subdirectory",
			fs: func() afero.Fs {
				fs := createFS(t)
				path := filepath.Join(filepath.Dir(wd), "Gopkg.toml")
				dieIf(t, afero.WriteFile(fs, path, []byte(""), 0644))
				return fs
			}(),
			execer: createExec(t, ""),
			typ:    manager.TypeDep,
			root:   filepath.Dir(wd),
		},
		{
			test:   "unknown",
			fs:     createFS(t),
			execer: createExec(t, ""),
			typ:    manager.TypeUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			typ, root := manager.DetectType(wd, tc.fs, tc.execer)

			if got, want := typ, tc.typ; got != want {
				t.Errorf("Detected mode is %v, want %v", got, want)
			}

			if got, want := root, tc.root; got != want {
				t.Errorf("Detected root is %s, want %s", got, want)
			}
		})
	}
}
