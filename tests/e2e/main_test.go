package main

import (
	"fmt"
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
	"unicode"

	"github.com/bradleyjkemp/cupaloy/v2"
	"golang.org/x/tools/go/packages/packagestest"
)

func TestGex_Add(t *testing.T) {
	if os.Getenv("E2E") != "1" {
		t.Skip("E2E tests are skipped. If you want to run them, you should set `E2E=1`.")
	}

	mode := TestModeMod
	if v := os.Getenv("MODE"); v != "" {
		checkErr(t, mode.UnmarshalText([]byte(v)))
	}
	t.Logf("test mode: %s", mode)

	debug := os.Getenv("DEBUG") == "1"
	t.Logf("debug: %t", debug)

	tc := CreateTestContext(t, mode, debug)
	defer tc.Close(t)

	invokeE2ETest(t, tc)
}

func invokeE2ETest(t *testing.T, tc *TestContext) {
	t.Run("add first tool", func(t *testing.T) {
		tc.ExecCmd(t, "gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway")
		tc.SnapshotManifest(t)
	})

	t.Run("add 2 tools", func(t *testing.T) {
		tc.ExecCmd(t, "gex", "--add", "github.com/srvc/wraperr/cmd/wraperr", "--add", "golang.org/x/lint/golint")
		tc.SnapshotManifest(t)
	})

	t.Run("add a tool that has already been added", func(t *testing.T) {
		tc.ExecCmd(t, "gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway")
		tc.SnapshotManifest(t)
	})

	t.Run("add tools that the tool has the same package has already been added", func(t *testing.T) {
		tc.ExecCmd(t, "gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger")
		tc.SnapshotManifest(t)
	})

	t.Run("add tools included in the same package", func(t *testing.T) {
		tc.ExecCmd(t, "gex", "--add", "github.com/gogo/protobuf/protoc-gen-gogo", "--add", "github.com/gogo/protobuf/protoc-gen-gogofast")
		tc.SnapshotManifest(t)
	})

	t.Run("add tools that its root proejct has been added", func(t *testing.T) {
		tc.ExecCmd(t, "gex", "--add", "github.com/golang/mock/mockgen")
		tc.SnapshotManifest(t)
	})

	t.Run("generated binaries with `gex --add`", func(t *testing.T) {
		tc.CheckBinaries(t, []string{"protoc-gen-grpc-gateway", "wraperr", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast", "mockgen"})
	})

	t.Run("generated binaries with `go generate`", func(t *testing.T) {
		tc.ExecCmd(t, "rm", "-vrf", "./bin")
		tc.ExecCmd(t, "go", "generate", "tools.go")
		tc.CheckBinaries(t, []string{"protoc-gen-grpc-gateway", "wraperr", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast", "mockgen"})
	})
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type TestMode int

const (
	TestModeUnknown TestMode = iota
	TestModeMod
	TestModeDep
)

func (tm TestMode) String() string {
	switch tm {
	case TestModeMod:
		return "mod"
	case TestModeDep:
		return "dep"
	default:
		return "unknown"
	}
}

func (tm *TestMode) UnmarshalText(text []byte) error {
	switch string(text) {
	case "mod":
		*tm = TestModeMod
	case "dep":
		*tm = TestModeDep
	default:
		return fmt.Errorf("unknown mode: %s", text)
	}
	return nil
}

func (tm TestMode) Exporter() packagestest.Exporter {
	switch tm {
	case TestModeMod:
		return packagestest.Modules
	case TestModeDep:
		return packagestest.GOPATH
	default:
		panic("unreachable")
	}
}

func CreateTestContext(t *testing.T, mode TestMode, debug bool) *TestContext {
	t.Helper()
	tc := &TestContext{mode: mode, debug: debug}
	tc.initialize(t)
	return tc
}

type TestContext struct {
	mode      TestMode
	exported  *packagestest.Exported
	debug     bool
	gexBinDir string
	closers   []func(t *testing.T)
}

func (tc *TestContext) initialize(t *testing.T) {
	t.Helper()

	// build gex
	{
		_, filename, _, _ := runtime.Caller(0)
		projRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
		binDir := filepath.Join(projRoot, "bin")
		gexCmd := filepath.Join(binDir, "gex")
		cmd := exec.Command("go", "build", "-v", "-o", gexCmd, "./cmd/gex")
		cmd.Dir = projRoot
		cmd.Stdout = NewTestWriter(t)
		cmd.Stderr = NewTestWriter(t)
		checkErr(t, cmd.Run())
		tc.gexBinDir = binDir

		oldPath := os.Getenv("PATH")
		checkErr(t, os.Setenv("PATH", binDir+string(filepath.ListSeparator)+oldPath))
		tc.closers = append(tc.closers, func(t *testing.T) { checkErr(t, os.Setenv("PATH", oldPath)) })

		t.Logf("gex binary: %s", gexCmd)
	}

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
		checkErr(t, err)
		tc.snapshot(t, string(data))
	})
}

func (tc *TestContext) CheckBinaries(t *testing.T, wantBins []string) {
	dir := filepath.Join(tc.rootDir(), "bin")
	files, err := ioutil.ReadDir(dir)
	checkErr(t, err)
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

func (tc *TestContext) snapshot(t *testing.T, v ...interface{}) {
	t.Helper()
	cupaloy.Global.
		WithOptions(
			cupaloy.SnapshotSubdirectory(".snapshots_"+tc.mode.String()),
		).
		SnapshotT(t, v...)
}

func (tc *TestContext) ExecCmd(t *testing.T, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = tc.rootDir()
	cmd.Env = tc.environ()
	cmd.Stdout = NewTestWriter(t)
	cmd.Stderr = NewTestWriter(t)
	checkErr(t, cmd.Run())
}

func NewTestWriter(t *testing.T) io.Writer {
	return &TestWriter{t: t}
}

type TestWriter struct {
	t *testing.T
}

func (w *TestWriter) Write(p []byte) (n int, err error) {
	w.t.Helper()
	s := string(p)
	n = len(s)
	s = strings.TrimRightFunc(string(p), unicode.IsSpace)
	w.t.Log(s)
	return
}
