package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

var ss = cupaloy.Global

func init() {
	if dir, ok := os.LookupEnv("SNAPSHOT_DIR"); ok {
		ss = ss.WithOptions(cupaloy.SnapshotSubdirectory(dir))
	}
}

func TestGex_Add(t *testing.T) {
	t.Run("add first tool", func(t *testing.T) {
		checkErr(t, exec.Command("gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway").Run())
		snapshotManifest(t)
	})

	t.Run("add 2 tools", func(t *testing.T) {
		checkErr(t, exec.Command("gex", "--add", "github.com/haya14busa/reviewdog/cmd/reviewdog", "--add", "golang.org/x/lint/golint").Run())
		snapshotManifest(t)
	})

	t.Run("add a tool that has already been added", func(t *testing.T) {
		checkErr(t, exec.Command("gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway").Run())
		snapshotManifest(t)
	})

	t.Run("add tools that the tool has the same package has already been added", func(t *testing.T) {
		checkErr(t, exec.Command("gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger").Run())
		snapshotManifest(t)
	})

	t.Run("add tools included in the same package", func(t *testing.T) {
		checkErr(t, exec.Command("gex", "--add", "github.com/gogo/protobuf/protoc-gen-gogo", "--add", "github.com/gogo/protobuf/protoc-gen-gogofast").Run())
		snapshotManifest(t)
	})

	gotBins, err := filepath.Glob("./bin/*")
	checkErr(t, err)
	for i, b := range gotBins {
		gotBins[i] = filepath.Base(b)
	}
	sort.Strings(gotBins)
	wantBins := []string{"protoc-gen-grpc-gateway", "reviewdog", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast"}
	sort.Strings(wantBins)

	if got, want := gotBins, wantBins; !reflect.DeepEqual(got, want) {
		t.Errorf("generated bins list is %v, want %v", got, want)
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func snapshotManifest(t *testing.T) {
	t.Helper()
	t.Run("tools.go", func(t *testing.T) {
		data, err := ioutil.ReadFile("tools.go")
		checkErr(t, err)
		ss.SnapshotT(t, string(data))
	})
}
