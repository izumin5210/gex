package tool

import (
	"sort"
	"strings"

	"github.com/izumin5210/gex/pkg/manager"
)

// Manifest contains tool list
type Manifest struct {
	toolMap          map[string]Tool
	managerType      manager.Type
	defaultBuildMode BuildMode
}

// NewManifest creates a new Manifest instance.
func NewManifest(tools []Tool, mType manager.Type, options ...ManifestOption) *Manifest {
	toolMap := make(map[string]Tool, len(tools))
	for _, t := range tools {
		toolMap[t.Name()] = t
	}
	m := &Manifest{toolMap: toolMap, managerType: mType}
	for _, option := range options {
		option(m)
	}
	if m.defaultBuildMode == BuildModeUnknown {
		m.defaultBuildMode = BuildModeBin
	}
	return m
}

func (m *Manifest) ManagerType() manager.Type { return m.managerType }

func (m *Manifest) DefaultBuildMode() BuildMode { return m.defaultBuildMode }

// AddTool adds a new tool to the manifest.
func (m *Manifest) AddTool(tool Tool) {
	m.toolMap[tool.Name()] = tool
}

// FindTool returns a tool by a name.
func (m *Manifest) FindTool(name string) (t Tool, ok bool) {
	t, ok = m.toolMap[name]
	return
}

// Tools returns a tool list.
func (m *Manifest) Tools() []Tool {
	n := len(m.toolMap)
	ts := make([]Tool, 0, n)
	for _, t := range m.toolMap {
		ts = append(ts, t)
	}
	sort.Slice(ts, func(i int, j int) bool {
		return strings.Compare(ts[i].ImportPath, ts[j].ImportPath) < 0
	})
	return ts
}

func (m *Manifest) addTool(t Tool) {
	m.toolMap[t.Name()] = t
}

type ManifestOption func(m *Manifest)

// WithDefaultBuildMode configures default build mode.
func WithDefaultBuildMode(buildMode BuildMode) ManifestOption {
	return func(m *Manifest) {
		m.defaultBuildMode = buildMode
	}
}
