package command

import "context"

// Adder is an interface to add new development tools.
type Adder interface {
	Add(ctx context.Context, pkgs []string, verbose bool) error
}
