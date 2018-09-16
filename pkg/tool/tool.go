package tool

import "path/filepath"

// Tool represents a go package of a tool dependency.
type Tool string

// Name returns an executable name.
func (t Tool) Name() string {
	return filepath.Base(string(t))
}
