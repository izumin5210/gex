package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/izumin5210/gex"
	"github.com/izumin5210/gex/pkg/tool"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	cliName = "gex"
)

var (
	pkgsToBeAdded []string
	flagBuild     bool
	flagInit      bool
	flagRegen     bool
	flagVersion   bool
	flagVerbose   bool
	flagHelp      bool
)

func init() {
	pflag.SetInterspersed(false)
	pflag.StringArrayVar(&pkgsToBeAdded, "add", []string{}, "Add new tools")
	pflag.BoolVar(&flagInit, "init", false, "Initialize tools manifest")
	pflag.BoolVar(&flagBuild, "build", false, "Build all tools")
	pflag.BoolVar(&flagRegen, "regen", false, "Regenerate manifest")
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

	var cfg gex.Config
	if flagVerbose {
		cfg.Verbose = true
		cfg.Logger = log.New(os.Stderr, "", 0)
	}
	toolRepo, err := cfg.Create()
	if err != nil {
		return errors.WithStack(err)
	}

	ctx := context.TODO()

	switch {
	case len(pkgsToBeAdded) > 0:
		err = toolRepo.Add(ctx, pkgsToBeAdded...)
	case flagVersion:
		fmt.Fprintf(os.Stdout, "%s %s\n", cliName, gex.Version)
	case flagHelp:
		printHelp(os.Stdout)
	case flagBuild:
		err = toolRepo.BuildAll(ctx)
		if errs := asBuildErrors(err); errs != nil {
			for _, err := range errs.Errs {
				fmt.Fprintln(os.Stdout, err.Error())
			}
			return errors.New("failed to build tools")
		}
		return err
	case flagInit:
		err = toolRepo.Add(ctx, "github.com/izumin5210/gex/cmd/gex")
	case flagRegen:
		path := filepath.Join(cfg.RootDir, cfg.ManifestName)
		m, err := tool.NewParser(cfg.FS, cfg.ManagerType).Parse(path)
		if err != nil {
			return errors.Wrapf(err, "%s was not found", path)
		}
		err = tool.NewWriter(cfg.FS).Write(path, m)
		if err != nil {
			return errors.WithStack(err)
		}
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
  gex --init
  gex --add [packages...]   Add new tool dependencies
  go generate ./tools.go    Build tools
  gex [command] [args]      Execute a tool

Flags:`
)
