package tool

import (
	"go/parser"
	"go/token"
	"strconv"

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
	f, err := parser.ParseFile(fset, "", string(data), parser.ImportsOnly)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %q", path)
	}

	tools := make([]Tool, 0, len(f.Imports))

	for _, s := range f.Imports {
		if pkg, err := strconv.Unquote(s.Path.Value); err == nil {
			tools = append(tools, Tool(pkg))
		}
	}

	return NewManifest(tools, p.mType), nil
}
