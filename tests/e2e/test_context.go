package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"golang.org/x/tools/go/packages/packagestest"
)

func CreateTestContext(t *testing.T, mode TestMode, debug bool) *TestContext {
	t.Helper()
	tc := &TestContext{mode: mode, debug: debug}
	tc.setupSandbox(t)
	return tc
}

type TestContext struct {
	mode     TestMode
	exported *packagestest.Exported
	debug    bool
	closers  []func(t *testing.T)
}

func (tc *TestContext) setupSandbox(t *testing.T) {
	t.Helper()

	tc.exported = packagestest.Export(t, tc.mode.Exporter(), []packagestest.Module{
		{Name: "sampleapp", Files: map[string]interface{}{".keep": "", "main.go": "package main"}},
	})
	tc.closers = append(tc.closers, func(t *testing.T) {
		if tc.debug {
			t.Log("Keep the test environment on debug mode")
			return
		}
		tc.exported.Cleanup()
	})

	if tc.mode == TestModeDep {
		tc.exported.Config.Dir = filepath.Join(tc.exported.Config.Dir, "sampleapp")
	}

	t.Logf("root directory: %s", tc.rootDir())

	switch tc.mode {
	case TestModeMod:
		// no-op
	case TestModeDep:
		tc.ExecCmd(t, "dep", "init", "-v")
	default:
		panic("unreachable")
	}
}

func (tc *TestContext) Close(t *testing.T) {
	for i := len(tc.closers) - 1; i >= 0; i-- {
		tc.closers[i](t)
	}
}

func (tc *TestContext) rootDir() string {
	return tc.exported.Config.Dir
}

func (tc *TestContext) binDir() string {
	return filepath.Join(tc.rootDir(), "bin")
}

func (tc *TestContext) environ() []string {
	env := make([]string, 0, len(tc.exported.Config.Env))
	for _, kv := range tc.exported.Config.Env {
		if strings.HasPrefix(kv, "GOPROXY=") {
			continue
		}
		if tc.mode == TestModeDep && runtime.GOOS == "darwin" && strings.HasPrefix(kv, "GOPATH=/var") {
			kv = strings.Replace(kv, "GOPATH=/var", "GOPATH=/private/var", 1)
		}
		env = append(env, kv)
	}
	return env
}

func (tc *TestContext) SnapshotManifest(t *testing.T) {
	t.Helper()
	t.Run("tools.go", func(t *testing.T) {
		data, err := ioutil.ReadFile(filepath.Join(tc.rootDir(), "tools.go"))
		tc.checkErr(t, err)
		cupaloy.SnapshotT(t, string(data))
	})
}

func (tc *TestContext) CheckBinaries(t *testing.T, wantBins []string) {
	files, err := ioutil.ReadDir(tc.binDir())
	tc.checkErr(t, err)
	var gotBins []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		gotBins = append(gotBins, f.Name())
	}
	sort.Strings(gotBins)
	sort.Strings(wantBins)
	if got, want := gotBins, wantBins; !reflect.DeepEqual(got, want) {
		t.Errorf("generated bins list is %v, want %v", got, want)
	}
}

func (tc *TestContext) RemoveBinaries(t *testing.T) {
	tc.checkErr(t, os.RemoveAll(tc.binDir()))
}

func (tc *TestContext) Bin(name string) string {
	return filepath.Join(tc.binDir(), name)
}

func (tc *TestContext) ExecCmd(t *testing.T, name string, args ...string) {
	tc.ExecCmdWithOut(t, name, args, NewTestWriter(t), NewTestWriter(t))
}

func (tc *TestContext) ExecCmdWithOut(t *testing.T, name string, args []string, outW, errW io.Writer) {
	cmd := exec.Command(name, args...)
	cmd.Dir = tc.rootDir()
	cmd.Env = tc.environ()
	cmd.Stdout = outW
	cmd.Stderr = errW
	tc.checkErr(t, cmd.Run())
}

func (tc *TestContext) checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
