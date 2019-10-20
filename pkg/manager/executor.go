package manager

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/izumin5210/execx"
	"github.com/pkg/errors"
)

// Executor is an interface for executing managers.
type Executor interface {
	Exec(ctx context.Context, name string, args ...string) error
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
}

// NewExecutor creates a new Executor instance.
func NewExecutor(exec *execx.Executor, outW, errW io.Writer, inR io.Reader, cwd string, log *log.Logger) Executor {
	env := os.Environ()
	for _, e := range env {
		kv := strings.SplitN(e, "=", 2)
		if kv[0] == "PATH" {
			kv[1] = filepath.Join(cwd, "bin") + string(os.PathListSeparator) + kv[1]
		}
		env = append(env, strings.Join(kv, "="))
	}
	return &executorImpl{
		exec: exec,
		outW: outW,
		errW: errW,
		inR:  inR,
		cwd:  cwd,
		env:  env,
		log:  log,
	}
}

type executorImpl struct {
	exec       *execx.Executor
	outW, errW io.Writer
	inR        io.Reader
	cwd        string
	env        []string
	log        *log.Logger
}

func (e *executorImpl) Exec(ctx context.Context, name string, args ...string) error {
	cmd := e.exec.CommandContext(ctx, name, args...)
	cmd.Stdout = e.outW
	cmd.Stderr = e.errW
	cmd.Stdin = e.inR
	cmd.Dir = e.cwd
	cmd.Env = e.env
	e.log.Println("execute", strings.Join(append([]string{name}, args...), " "))
	return errors.WithStack(cmd.Run())
}

func (e *executorImpl) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := e.exec.CommandContext(ctx, name, args...)
	cmd.Stderr = e.errW
	cmd.Stdin = e.inR
	cmd.Dir = e.cwd
	cmd.Env = e.env
	e.log.Println("execute", strings.Join(append([]string{name}, args...), " "))
	out, err := cmd.Output()
	return out, errors.WithStack(err)
}
