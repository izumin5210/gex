package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/pflag"

	"github.com/izumin5210/gex/pkg/command"
	"github.com/izumin5210/gex/pkg/tool"
)

const (
	manifestName = "tools.go"
	binDirName   = "bin"
	orgName      = "github.com/izumin5210"
	cliName      = "gex"
	version      = "v0.2.0"
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
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	pflag.Parse()
	args := pflag.Args()

	ctx := context.TODO()
	cmdExecutor := command.NewExecutor(os.Stdout, os.Stderr, os.Stdin)
	toolRepo := tool.NewRepository(afero.NewOsFs(), cmdExecutor, &tool.Config{
		WorkingDir:   workingDir,
		ManifestName: manifestName,
		BinDirName:   binDirName,
		Verbose:      flagVerbose,
	})

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

	return err
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
