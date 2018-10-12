package gex_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"k8s.io/utils/exec"
	"k8s.io/utils/exec/testing"

	"github.com/izumin5210/gex"
)

func TestConfig_DetectMode(t *testing.T) {
	wd := "/go/src/awesomeapp/foobar"

	checkGoEnvCmd := func(t *testing.T, cmd string, args []string) {
		t.Helper()
		if diff := cmp.Diff(append([]string{cmd}, args...), []string{"go", "env", "GOMOD"}); diff != "" {
			t.Errorf("execed differs: (-want +got)\n%s", diff)
		}
	}
	createFakeCmd := func(t *testing.T, out string) exec.Cmd {
		t.Helper()
		return &testingexec.FakeCmd{
			CombinedOutputScript: []testingexec.FakeCombinedOutputAction{
				func() ([]byte, error) { return []byte(out + "\n"), nil },
			},
		}
	}
	createExecer := func(t *testing.T, out string) exec.Interface {
		t.Helper()
		return &testingexec.FakeExec{
			CommandScript: []testingexec.FakeCommandAction{
				func(cmd string, args ...string) exec.Cmd {
					checkGoEnvCmd(t, cmd, args)
					return createFakeCmd(t, out)
				},
			},
		}
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
		execer exec.Interface
		mode   gex.Mode
		root   string
	}{
		{
			test:   "modules",
			fs:     createFS(t),
			execer: createExecer(t, filepath.Join(wd, "go.mod")),
			mode:   gex.ModeModules,
			root:   wd,
		},
		{
			test:   "modules from subdirectory",
			fs:     createFS(t),
			execer: createExecer(t, filepath.Join(filepath.Dir(wd), "go.mod")),
			mode:   gex.ModeModules,
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
			execer: createExecer(t, ""),
			mode:   gex.ModeDep,
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
			execer: createExecer(t, ""),
			mode:   gex.ModeDep,
			root:   filepath.Dir(wd),
		},
		{
			test:   "unknown",
			fs:     createFS(t),
			execer: createExecer(t, ""),
			mode:   gex.ModeUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			cfg := gex.Config{WorkingDir: wd, FS: tc.fs, Execer: tc.execer}
			cfg.DetectMode()

			if got, want := cfg.Mode, tc.mode; got != want {
				t.Errorf("Detected mode is %v, want %v", got, want)
			}

			if got, want := cfg.RootDir, tc.root; got != want {
				t.Errorf("Detected root is %s, want %s", got, want)
			}
		})
	}
}
