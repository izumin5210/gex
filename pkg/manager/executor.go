package manager

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/utils/exec"
)

// Executor is an interface for executing managers.
type Executor interface {
	Exec(ctx context.Context, name string, args ...string) error
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
}

// NewExecutor creates a new Executor instance.
func NewExecutor(execer exec.Interface, outW, errW io.Writer, inR io.Reader, cwd string, log *log.Logger) Executor {
	env := os.Environ()
	for _, e := range env {
		kv := strings.SplitN(e, "=", 2)
		if kv[0] == "PATH" {
			kv[1] = filepath.Join(cwd, "bin") + string(os.PathListSeparator) + kv[1]
		}
		env = append(env, strings.Join(kv, "="))
	}
	return &executorImpl{
		execer: execer,
		outW:   outW,
		errW:   errW,
		inR:    inR,
		cwd:    cwd,
		env:    env,
		log:    log,
	}
}

type executorImpl struct {
	execer     exec.Interface
	outW, errW io.Writer
	inR        io.Reader
	cwd        string
	env        []string
	log        *log.Logger
}

func (e *executorImpl) Exec(ctx context.Context, name string, args ...string) error {
	cmd := e.execer.CommandContext(ctx, name, args...)
	cmd.SetStdout(e.outW)
	cmd.SetStderr(e.errW)
	cmd.SetStdin(e.inR)
	cmd.SetDir(e.cwd)
	cmd.SetEnv(e.env)
	e.log.Println("execute", strings.Join(append([]string{name}, args...), " "))
	return errors.WithStack(cmd.Run())
}

func (e *executorImpl) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := e.execer.CommandContext(ctx, name, args...)
	cmd.SetStderr(e.errW)
	cmd.SetStdin(e.inR)
	cmd.SetDir(e.cwd)
	cmd.SetEnv(e.env)
	e.log.Println("execute", strings.Join(append([]string{name}, args...), " "))
	out, err := cmd.Output()
	return out, errors.WithStack(err)
}
