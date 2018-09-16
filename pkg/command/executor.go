package command

import (
	"context"
	"io"
	"os/exec"
)

// Executor is an interface for executing commands.
type Executor interface {
	Exec(ctx context.Context, name string, args ...string) error
}

// NewExecutor creates a new Executor instance.
func NewExecutor(outW, errW io.Writer, inR io.Reader) Executor {
	return &executorImpl{
		outW: outW,
		errW: errW,
		inR:  inR,
	}
}

type executorImpl struct {
	outW, errW io.Writer
	inR        io.Reader
}

func (e *executorImpl) Exec(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = e.outW
	cmd.Stderr = e.errW
	cmd.Stdin = e.inR
	return cmd.Run()
}
