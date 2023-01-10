package tool

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/izumin5210/gex/pkg/manager"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Parser retrieve tool packages from given paths.
type Parser interface {
	Parse(path string) (*Manifest, error)
}

// NewParser creates a new parser instance.
func NewParser(fs afero.Fs, mType manager.Type) Parser {
	return &parserImpl{
		fs:    fs,
		mType: mType,
	}
}

type parserImpl struct {
	fs    afero.Fs
	mType manager.Type
}

func (p *parserImpl) Parse(path string) (*Manifest, error) {
	data, err := afero.ReadFile(p.fs, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %q", path)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", string(data), parser.ImportsOnly|parser.ParseComments)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %q", path)
	}

	tools := make([]Tool, 0, len(f.Imports))

	comments := commentTraverser{comments: f.Comments}
	fileComments := comments.consume(f.Pos())
	fileBuildMode := parseBuildMode(fileComments)
	for _, s := range f.Imports {
		if pkg, err := strconv.Unquote(s.Path.Value); err == nil {
			importComments := comments.consume(s.Pos())
			buildMode := parseBuildMode(importComments)
			tools = append(tools, Tool{ImportPath: pkg, BuildMode: buildMode})
		}
	}

	return NewManifest(tools, p.mType, WithDefaultBuildMode(fileBuildMode)), nil
}

type commentTraverser struct {
	comments []*ast.CommentGroup
	index    int
}

func (t *commentTraverser) consume(pos token.Pos) []*ast.CommentGroup {
	start := t.index
	for t.index < len(t.comments) && t.comments[t.index].Pos() < pos {
		t.index++
	}
	return t.comments[start:t.index]
}

func parseBuildMode(groups []*ast.CommentGroup) BuildMode {
	for _, group := range groups {
		for _, element := range group.List {
			if strings.Contains(element.Text, "gex:bin") {
				return BuildModeBin
			}
			if strings.Contains(element.Text, "gex:nobin") {
				return BuildModeNoBin
			}
		}
	}
	return BuildModeUnknown
}
