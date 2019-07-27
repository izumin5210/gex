package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

func TestGex_Add_Mod_Go112(t *testing.T) {
	t.Parallel()
	testGex_Add(t, TestModeMod, "1.12")
}

func TestGex_Add_Dep_Go112(t *testing.T) {
	t.Parallel()
	testGex_Add(t, TestModeDep, "1.12")
}

func TestGex_Add_Mod_Go111(t *testing.T) {
	t.Parallel()
	testGex_Add(t, TestModeMod, "1.11")
}

func TestGex_Add_Dep_Go111(t *testing.T) {
	t.Parallel()
	testGex_Add(t, TestModeDep, "1.11")
}

func testGex_Add(t *testing.T, mode TestMode, goVersion string) {
	if os.Getenv("E2E") != "1" {
		t.Skip("E2E tests are skipped. If you want to run them, you should set `E2E=1`.")
	}

	tc := CreateTestContainer(t, mode, goVersion)
	if os.Getenv("DEBUG") != "1" {
		defer tc.Close(t)
	}

	t.Run("add first tool", func(t *testing.T) {
		tc.CheckCmd(t, []string{"gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"})
		tc.SnapshotManifest(t)
	})

	t.Run("add 2 tools", func(t *testing.T) {
		tc.CheckCmd(t, []string{"gex", "--add", "github.com/srvc/wraperr/cmd/wraperr", "--add", "golang.org/x/lint/golint"})
		tc.SnapshotManifest(t)
	})

	t.Run("add a tool that has already been added", func(t *testing.T) {
		tc.CheckCmd(t, []string{"gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"})
		tc.SnapshotManifest(t)
	})

	t.Run("add tools that the tool has the same package has already been added", func(t *testing.T) {
		tc.CheckCmd(t, []string{"gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"})
		tc.SnapshotManifest(t)
	})

	t.Run("add tools included in the same package", func(t *testing.T) {
		tc.CheckCmd(t, []string{"gex", "--add", "github.com/gogo/protobuf/protoc-gen-gogo", "--add", "github.com/gogo/protobuf/protoc-gen-gogofast"})
		tc.SnapshotManifest(t)
	})

	t.Run("add tools that its root proejct has been added", func(t *testing.T) {
		tc.CheckCmd(t, []string{"gex", "--add", "github.com/golang/mock/mockgen"})
		tc.SnapshotManifest(t)
	})

	checkBinaries := func(t *testing.T, bins []string) {
		t.Helper()
		buf := new(bytes.Buffer)
		tc.ExecCmd(t, []string{"ls", "-1", "./bin"}, buf, ioutil.Discard)
		var gotBins []string
		for _, b := range strings.Split(buf.String(), "\n") {
			if len(b) > 0 && b != "." {
				gotBins = append(gotBins, b)
			}
		}
		sort.Strings(gotBins)
		wantBins := bins[:]
		sort.Strings(wantBins)

		if got, want := gotBins, wantBins; !reflect.DeepEqual(got, want) {
			t.Errorf("generated bins list is %v, want %v", got, want)
		}
	}

	t.Run("generated binaries with `gex --add`", func(t *testing.T) {
		checkBinaries(t, []string{"protoc-gen-grpc-gateway", "wraperr", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast", "mockgen"})
	})

	t.Run("generated binaries with `go generate`", func(t *testing.T) {
		tc.CheckCmd(t, []string{"rm", "-vrf", "./bin"})
		tc.CheckCmd(t, []string{"go", "generate", "tools.go"})
		checkBinaries(t, []string{"protoc-gen-grpc-gateway", "wraperr", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast", "mockgen"})
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
	TestModeMod TestMode = iota
	TestModeDep
)

type TestContainer struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
	cupaloy  *cupaloy.Config
}

func CreateTestContainer(t *testing.T, mode TestMode, goVersion string) *TestContainer {
	wd, err := os.Getwd()
	checkErr(t, err)

	pool, err := dockertest.NewPool("")
	checkErr(t, err)

	imageTag := "go-" + goVersion
	imageName := "github.com/izumin5210/gex/tests/e2e"

	buildArgs := []docker.BuildArg{{Name: "GO_VERSION", Value: goVersion}}

	switch mode {
	case TestModeMod:
		buildArgs = append(buildArgs, docker.BuildArg{Name: "GO111MODULE", Value: "on"})
	case TestModeDep:
		buildArgs = append(buildArgs, docker.BuildArg{Name: "GO111MODULE", Value: "off"})
	default:
		panic("unreachable")
	}

	err = pool.Client.BuildImage(docker.BuildImageOptions{
		Name:         imageName + ":" + imageTag,
		Dockerfile:   filepath.Join("tests", "e2e", "Dockerfile"),
		ContextDir:   filepath.Join(wd, "..", ".."),
		BuildArgs:    buildArgs,
		OutputStream: NewTestWriter(t),
		ErrorStream:  NewTestWriter(t),
	})
	checkErr(t, err)

	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: imageName,
		Tag:        imageTag,
		Cmd:        []string{"tail", "-f", "/dev/null"},
	})
	checkErr(t, err)

	tc := &TestContainer{
		pool:     pool,
		resource: res,
		cupaloy:  cupaloy.Global,
	}

	switch mode {
	case TestModeMod:
		tc.CheckCmd(t, []string{"go", "mod", "init"})
	case TestModeDep:
		tc.CheckCmd(t, []string{"sh", "-c", "curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh"})
		tc.CheckCmd(t, []string{"dep", "init", "-v"})
	default:
		panic("unreachable")
	}

	return tc
}

func (tc *TestContainer) Close(t *testing.T) {
	t.Helper()
	checkErr(t, tc.resource.Close())
}

func (tc *TestContainer) ExecCmd(t *testing.T, cmd []string, outW, errW io.Writer) {
	t.Helper()

	exec, err := tc.pool.Client.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Container:    tc.resource.Container.ID,
		Cmd:          cmd,
	})
	checkErr(t, err)

	err = tc.pool.Client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: outW,
		ErrorStream:  errW,
	})
	checkErr(t, err)

	resp, err := tc.pool.Client.InspectExec(exec.ID)
	checkErr(t, err)

	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %v", resp.ExitCode, cmd)
	}
}

func (tc *TestContainer) CheckCmd(t *testing.T, cmd []string) {
	t.Helper()
	tc.ExecCmd(t, cmd, NewTestWriter(t), NewTestWriter(t))
}

func (tc *TestContainer) SnapshotManifest(t *testing.T) {
	t.Helper()
	t.Run("tools.go", func(t *testing.T) {
		buf := new(bytes.Buffer)
		tc.ExecCmd(t, []string{"cat", "/go/src/myapp/tools.go"}, buf, buf)
		tc.cupaloy.SnapshotT(t, buf.String())
	})
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
