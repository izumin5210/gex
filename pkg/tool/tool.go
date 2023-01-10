package tool

import "path/filepath"

// Tool represents a go package of a tool dependency.
type Tool struct {
	ImportPath string
	BuildMode  BuildMode
}

// Name returns an executable name.
func (t Tool) Name() string {
	return filepath.Base(t.ImportPath)
}

// NeedBuild determines if the tool should be built before run.
func (t Tool) NeedBuild(defaultBuildMode BuildMode) bool {
	buildMode := t.BuildMode
	if buildMode == BuildModeUnknown {
		buildMode = defaultBuildMode
	}
	return buildMode == BuildModeBin
}

// BuildMode records whether the package should be build ahead of time.
type BuildMode int

const (
	BuildModeUnknown BuildMode = iota
	BuildModeBin
	BuildModeNoBin
)
