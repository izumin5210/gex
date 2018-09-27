package command

import "context"

// Builder is an interface to build vendored development tools.
type Builder interface {
	Build(ctx context.Context, binPath, pkg string, verbose bool) error
}
