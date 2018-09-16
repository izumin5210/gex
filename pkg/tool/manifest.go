package tool

import (
	"sort"
)

// Manifest contains tool list
type Manifest struct {
	toolMap map[string]Tool
}

// NewManifest creates a new Manifest instance.
func NewManifest(tools []Tool) *Manifest {
	toolMap := make(map[string]Tool, len(tools))
	for _, t := range tools {
		toolMap[t.Name()] = t
	}
	return &Manifest{toolMap: toolMap}
}

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
	s := make([]string, 0, n)
	for _, t := range m.toolMap {
		s = append(s, string(t))
	}
	sort.StringSlice(s).Sort()
	ts := make([]Tool, n, n)
	for i, t := range s {
		ts[i] = Tool(t)
	}
	return ts
}

func (m *Manifest) addTool(t Tool) {
	m.toolMap[t.Name()] = t
}
