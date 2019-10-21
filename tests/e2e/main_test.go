package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	gexCmd string
	debug  bool
)

func TestMain(m *testing.M) {
	debug = os.Getenv("DEBUG") == "1"

	_, filename, _, _ := runtime.Caller(0)
	projRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	binDir := filepath.Join(projRoot, "bin")

	gexCmd = filepath.Join(binDir, "gex")

	cmd := exec.Command("go", "build", "-v", "-o", gexCmd, "./cmd/gex")
	cmd.Dir = projRoot
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stdout = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()
	defer os.Exit(exitCode)
}

func TestGex_Mod(t *testing.T) {
	invokeE2ETest(t, TestModeMod)
}

func TestGex_Dep(t *testing.T) {
	invokeE2ETest(t, TestModeDep)
}

func invokeE2ETest(t *testing.T, tm TestMode) {
	t.Helper()

	tc := CreateTestContext(t, tm, debug)
	defer tc.Close(t)

	t.Run("add first tool", func(t *testing.T) {
		tc.ExecCmd(t, gexCmd, "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway")
		tc.SnapshotManifest(t)
	})

	t.Run("add 2 tools", func(t *testing.T) {
		tc.ExecCmd(t, gexCmd, "--add", "github.com/srvc/wraperr/cmd/wraperr", "--add", "golang.org/x/lint/golint")
		tc.SnapshotManifest(t)
	})

	t.Run("add a tool that has already been added", func(t *testing.T) {
		tc.ExecCmd(t, gexCmd, "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway")
		tc.SnapshotManifest(t)
	})

	t.Run("add tools that the tool has the same package has already been added", func(t *testing.T) {
		tc.ExecCmd(t, gexCmd, "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger")
		tc.SnapshotManifest(t)
	})

	t.Run("add tools included in the same package", func(t *testing.T) {
		tc.ExecCmd(t, gexCmd, "--add", "github.com/gogo/protobuf/protoc-gen-gogo", "--add", "github.com/gogo/protobuf/protoc-gen-gogofast")
		tc.SnapshotManifest(t)
	})

	t.Run("add tools that its root proejct has been added", func(t *testing.T) {
		tc.ExecCmd(t, gexCmd, "--add", "github.com/golang/mock/mockgen")
		tc.SnapshotManifest(t)
	})

	t.Run("generated binaries with `gex --add`", func(t *testing.T) {
		tc.CheckBinaries(t, []string{"protoc-gen-grpc-gateway", "wraperr", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast", "mockgen"})
	})

	t.Run("generated binaries with `go generate`", func(t *testing.T) {
		tc.RemoveBinaries(t)
		tc.ExecCmd(t, "go", "generate", "tools.go")
		tc.CheckBinaries(t, []string{"protoc-gen-grpc-gateway", "wraperr", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast", "mockgen"})
	})
}
