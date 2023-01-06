//go:build go1.13
// +build go1.13

package main

import (
	"errors"

	"github.com/izumin5210/gex/pkg/tool"
)

func asBuildErrors(err error) *tool.BuildErrors {
	var errs *tool.BuildErrors
	if errors.As(err, &errs) {
		return errs
	}
	return nil
}
