package tool

import (
	"path/filepath"
	"strconv"
	"strings"
)

// Tool represents a go package of a tool dependency.
type Tool string

// Name returns an executable name.
func (t Tool) Name() string {
	lastSegment := filepath.Base(string(t))
	if isVersionLike(lastSegment) {
		// Likely a major version suffix
		return filepath.Base(filepath.Dir(string(t)))
	}
	return lastSegment
}

func isVersionLike(segment string) bool {
	if !strings.HasPrefix(segment, "v") {
		return false
	}
	versionPart := strings.TrimPrefix(segment, "v")
	if _, err := strconv.ParseUint(versionPart, 10, 64); err == nil {
		return true
	}
	return false
}
