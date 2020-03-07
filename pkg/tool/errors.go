package tool

import (
	"fmt"
	"strings"
	"sync"
)

type BuildError struct {
	Tool Tool
	Err  error
}

func (e *BuildError) Unwrap() error { return e.Err }

func (e *BuildError) Error() string {
	return fmt.Sprintf("failed to build %s: %s", e.Tool.Name(), e.Err)
}

type BuildErrors struct {
	sync.RWMutex
	Errs []*BuildError
}

func (e *BuildErrors) Unwrap() error {
	e.RLock()
	defer e.RUnlock()

	return e.Errs[0]
}

func (e *BuildErrors) Error() string {
	e.RLock()
	defer e.RUnlock()

	var b strings.Builder
	b.WriteString("failed to build ")
	b.WriteString(e.Errs[0].Tool.Name())
	if n := len(e.Errs); n > 1 {
		b.WriteString(fmt.Sprintf(" (and %d tool(s))", n))
	}
	b.WriteString(": ")
	b.WriteString(e.Errs[0].Err.Error())
	return b.String()
}

func (e *BuildErrors) Append(t Tool, err error) {
	e.Lock()
	defer e.Unlock()
	e.Errs = append(e.Errs, &BuildError{Tool: t, Err: err})
}

func (e *BuildErrors) Empty() bool {
	return len(e.Errs) == 0
}
