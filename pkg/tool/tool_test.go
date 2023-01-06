package tool_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/izumin5210/gex/pkg/tool"
	"testing"
)

func TestTool_Name(t *testing.T) {
	table := []struct {
		Input string
		Want  string
	}{
		{
			Input: "github.com/reviewdog/reviewdog",
			Want:  "reviewdog",
		},
		{
			Input: "github.com/volatiletech/sqlboiler/v4",
			Want:  "sqlboiler",
		},
	}
	for _, testcase := range table {
		t.Run(testcase.Input, func(t *testing.T) {
			name := tool.Tool(testcase.Input).Name()
			if diff := cmp.Diff(name, testcase.Want); diff != "" {
				t.Errorf("tool differs: (-want +got)\n%s", diff)
			}
		})
	}
}
