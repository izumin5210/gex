package manager

import "context"

type Interface interface {
	Build(ctx context.Context, binPath, pkg string, verbose bool) error
	Sync(ctx context.Context, verbose bool) error
}
