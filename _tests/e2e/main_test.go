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
		checkCmd(t, exec.Command("gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"))
		snapshotManifest(t)
	})

	t.Run("add 2 tools", func(t *testing.T) {
		checkCmd(t, exec.Command("gex", "--add", "github.com/haya14busa/reviewdog/cmd/reviewdog", "--add", "golang.org/x/lint/golint"))
		snapshotManifest(t)
	})

	t.Run("add a tool that has already been added", func(t *testing.T) {
		checkCmd(t, exec.Command("gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"))
		snapshotManifest(t)
	})

	t.Run("add tools that the tool has the same package has already been added", func(t *testing.T) {
		checkCmd(t, exec.Command("gex", "--add", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"))
		snapshotManifest(t)
	})

	t.Run("add tools included in the same package", func(t *testing.T) {
		checkCmd(t, exec.Command("gex", "--add", "github.com/gogo/protobuf/protoc-gen-gogo@v1.1.1", "--add", "github.com/gogo/protobuf/protoc-gen-gogofast"))
		snapshotManifest(t)
	})

	t.Run("add tools that its root proejct has been added", func(t *testing.T) {
		tt := os.Getenv("TEST_TARGET")
		switch tt {
		case "dep":
			checkCmd(t, exec.Command("dep", "ensure", "-add", "github.com/golang/mock/gomock"))
		case "mod":
			checkCmd(t, exec.Command("go", "get", "github.com/golang/mock/gomock"))
		default:
			t.Fatalf("unknown TEST_TARGET=%s", tt)
		}
		checkErr(t, ioutil.WriteFile("import.go", []byte(`package main

import _ "github.com/golang/mock/gomock"
`), 0755))
		checkCmd(t, exec.Command("gex", "--add", "github.com/golang/mock/mockgen"))
		snapshotManifest(t)
		checkErr(t, os.Remove("import.go"))
	})

	gotBins, err := filepath.Glob("./bin/*")
	checkErr(t, err)
	for i, b := range gotBins {
		gotBins[i] = filepath.Base(b)
	}
	sort.Strings(gotBins)
	wantBins := []string{"protoc-gen-grpc-gateway", "reviewdog", "golint", "protoc-gen-swagger", "protoc-gen-gogo", "protoc-gen-gogofast", "mockgen"}
	sort.Strings(wantBins)

	if got, want := gotBins, wantBins; !reflect.DeepEqual(got, want) {
		t.Errorf("generated bins list is %v, want %v", got, want)
	}
}

func checkCmd(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Errorf("unexpected error: %v", err)
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
