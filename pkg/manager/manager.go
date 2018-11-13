package manager

import "context"

type Interface interface {
	Add(ctx context.Context, pkgs []string, verbose bool) error
	Build(ctx context.Context, binPath, pkg string, verbose bool) error
}
