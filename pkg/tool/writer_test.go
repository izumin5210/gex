package tool_test

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/spf13/afero"

	"github.com/izumin5210/gex/pkg/manager"
	"github.com/izumin5210/gex/pkg/tool"
)

func TestWriter_Write(t *testing.T) {
	fs := afero.NewMemMapFs()
	writer := tool.NewWriter(fs)

	for _, typ := range []manager.Type{manager.TypeModules, manager.TypeDep} {
		t.Run(typ.String(), func(t *testing.T) {
			in := tool.NewManifest([]tool.Tool{
				"github.com/gogo/protobuf/protoc-gen-gogofast",
				"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway",
				"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger",
				"github.com/volatiletech/sqlboiler",
				"github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql",
			}, typ)
			path := "/home/src/awesomeapp/tools"

			err := writer.Write(path, in)
			if err != nil {
				t.Fatalf("Write() returned an error: %v", err)
			}

			data, err := afero.ReadFile(fs, path)
			if err != nil {
				t.Fatalf("faield to read %s: %v", path, err)
			}

			t.Run("tools.go", func(t *testing.T) {
				cupaloy.SnapshotT(t, string(data))
			})
		})
	}
}
