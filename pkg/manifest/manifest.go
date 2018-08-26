package manifest

import (
	"path/filepath"
	"sort"
)

// Manifest contains tool list
type Manifest struct {
	toolMap map[string]Tool
}

func newManifest() *Manifest {
	return &Manifest{toolMap: make(map[string]Tool)}
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

// Tool represents a go package of a tool dependency.
type Tool string

// Name returns an executable name.
func (t Tool) Name() string {
	return filepath.Base(string(t))
}
