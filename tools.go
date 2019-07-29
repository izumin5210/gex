// Code generated by github.com/izumin5210/gex. DO NOT EDIT.

// +build tools

package tools

// tool dependencies
import (
	_ "github.com/haya14busa/reviewdog/cmd/reviewdog"
	_ "github.com/srvc/wraperr/cmd/wraperr"
	_ "golang.org/x/lint/golint"
)

// If you want to use tools, please run the following command:
//  go generate ./tools.go
//
//go:generate go build -v -o=./bin/reviewdog github.com/haya14busa/reviewdog/cmd/reviewdog
//go:generate go build -v -o=./bin/wraperr github.com/srvc/wraperr/cmd/wraperr
//go:generate go build -v -o=./bin/golint golang.org/x/lint/golint
