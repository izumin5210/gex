// +build !go1.13

package main

import (
	"github.com/izumin5210/gex/pkg/tool"
)

func asBuildErrors(err error) *tool.BuildErrors {
	if errs, ok := err.(*tool.BuildErrors); ok {
		return errs
	}
	return nil
}
