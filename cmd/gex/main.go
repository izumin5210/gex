package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/izumin5210/gex"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	cliName = "gex"
	version = "v0.3.1"
)

var (
	pkgsToBeAdded []string
	flagBuild     bool
	flagVersion   bool
	flagVerbose   bool
	flagHelp      bool
)

func init() {
	pflag.SetInterspersed(false)
	pflag.StringArrayVar(&pkgsToBeAdded, "add", []string{}, "Add new tools")
	pflag.BoolVar(&flagBuild, "build", false, "Build all tools")
	pflag.BoolVar(&flagVersion, "version", false, "Print the CLI version")
	pflag.BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose level output")
	pflag.BoolVarP(&flagHelp, "help", "h", false, "Help for the CLI")
}

func main() {
	var exitCode int

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func run() error {
	pflag.Parse()
	args := pflag.Args()

	toolRepo, err := gex.Default.Create()
	if err != nil {
		return errors.WithStack(err)
	}

	ctx := context.TODO()

	switch {
	case len(pkgsToBeAdded) > 0:
		err = toolRepo.Add(ctx, pkgsToBeAdded...)
	case flagVersion:
		fmt.Fprintf(os.Stdout, "%s %s\n", cliName, version)
	case flagHelp:
		printHelp(os.Stdout)
	case flagBuild:
		err = toolRepo.BuildAll(ctx)
	case len(args) > 0:
		err = toolRepo.Run(ctx, args[0], args[1:]...)
	default:
		printHelp(os.Stdout)
	}

	return errors.WithStack(err)
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, helpText)
	pflag.PrintDefaults()
}

var (
	helpText = `The implementation of clarify best practice for tool dependencies.

See https://github.com/golang/go/issues/25922#issuecomment-412992431

Usage:
  gex [command] [args]
  gex --add [packages...]

Flags:`
)
