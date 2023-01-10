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

	testcases := []struct {
		name             string
		managerType      manager.Type
		tools            []tool.Tool
		defaultBuildMode tool.BuildMode
	}{
		{
			name:        "dep",
			managerType: manager.TypeDep,
			tools: []tool.Tool{
				{"github.com/gogo/protobuf/protoc-gen-gogofast", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger", tool.BuildModeUnknown},
				{"github.com/volatiletech/sqlboiler", tool.BuildModeUnknown},
				{"github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql", tool.BuildModeUnknown},
			},
			defaultBuildMode: tool.BuildModeBin,
		},
		{
			name:        "mod",
			managerType: manager.TypeModules,
			tools: []tool.Tool{
				{"github.com/gogo/protobuf/protoc-gen-gogofast", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger", tool.BuildModeBin},
				{"github.com/volatiletech/sqlboiler", tool.BuildModeNoBin},
				{"github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql", tool.BuildModeNoBin},
			},
			defaultBuildMode: tool.BuildModeBin,
		},
		{
			name:        "mod default nobuild",
			managerType: manager.TypeModules,
			tools: []tool.Tool{
				{"github.com/gogo/protobuf/protoc-gen-gogofast", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger", tool.BuildModeBin},
				{"github.com/volatiletech/sqlboiler", tool.BuildModeNoBin},
				{"github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql", tool.BuildModeNoBin},
			},
			defaultBuildMode: tool.BuildModeNoBin,
		},
		{
			name:        "mod all nobuild",
			managerType: manager.TypeModules,
			tools: []tool.Tool{
				{"github.com/gogo/protobuf/protoc-gen-gogofast", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway", tool.BuildModeUnknown},
				{"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger", tool.BuildModeUnknown},
				{"github.com/volatiletech/sqlboiler", tool.BuildModeUnknown},
				{"github.com/volatiletech/sqlboiler/drivers/sqlboiler-psql", tool.BuildModeUnknown},
			},
			defaultBuildMode: tool.BuildModeNoBin,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			in := tool.NewManifest(testcase.tools, testcase.managerType, tool.WithDefaultBuildMode(
				testcase.defaultBuildMode,
			))
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
