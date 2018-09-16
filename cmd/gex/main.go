package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/pflag"

	"github.com/izumin5210/gex/pkg/tool"
)

const (
	manifestName = "tools.go"
	binDirName   = "bin"
	orgName      = "github.com/izumin5210"
	cliName      = "gex"
	version      = "v0.1.0"
)

var (
	pkgsToBeAdded []string
	flagVersion   bool
	flagVerbose   bool
	flagHelp      bool
)

func init() {
	pflag.SetInterspersed(false)
	pflag.StringArrayVar(&pkgsToBeAdded, "add", []string{}, "Add new tools")
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

type config struct {
	outW, errW   io.Writer
	inR          io.Reader
	fs           afero.Fs
	workingDir   string
	manifestName string
	binDirName   string
	pkgName      string
	verbose      bool
}

func (c *config) ManifestPath() string {
	return filepath.Join(c.workingDir, c.manifestName)
}

func (c *config) BinDir() string {
	return filepath.Join(c.workingDir, c.binDirName)
}

func run() error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	pflag.Parse()
	args := pflag.Args()

	cfg := &config{
		outW:         os.Stdout,
		errW:         os.Stderr,
		inR:          os.Stdin,
		fs:           afero.NewOsFs(),
		workingDir:   workingDir,
		manifestName: manifestName,
		binDirName:   binDirName,
		pkgName:      filepath.Join(orgName, cliName),
		verbose:      flagVerbose,
	}

	ctx := context.TODO()

	switch {
	case len(pkgsToBeAdded) > 0:
		err = addTool(ctx, pkgsToBeAdded, cfg)
	case flagVersion:
		fmt.Fprintf(cfg.outW, "%s %s\n", cliName, version)
	case flagHelp:
		printHelp(cfg)
	case len(args) > 0:
		err = runTool(ctx, args[0], args[1:], cfg)
	default:
		printHelp(cfg)
	}

	return err
}

func addTool(ctx context.Context, pkgs []string, cfg *config) (err error) {
	args := []string{"get"}
	if cfg.verbose {
		args = append(args, "-v")
	}
	args = append(args, pkgs...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = cfg.outW
	cmd.Stderr = cfg.errW
	cmd.Stdin = cfg.inR
	err = cmd.Run()
	if err != nil {
		return err
	}

	p := tool.NewParser(cfg.fs)
	m, err := p.Parse(cfg.ManifestPath())
	if err != nil {
		m = tool.NewManifest([]tool.Tool{})
	}

	for _, pkg := range pkgs {
		pkg = strings.SplitN(pkg, "@", 2)[0]
		m.AddTool(tool.Tool(pkg))
	}

	err = tool.NewWriter(cfg.fs).Write(cfg.ManifestPath(), m)
	if err != nil {
		return err
	}

	return nil
}

func runTool(ctx context.Context, name string, args []string, cfg *config) error {
	p := tool.NewParser(cfg.fs)
	m, err := p.Parse(cfg.ManifestPath())
	if err != nil {
		return err
	}

	t, ok := m.FindTool(name)
	if !ok {
		return fmt.Errorf("failed to find the tool %q", name)
	}

	bin := filepath.Join(cfg.BinDir(), name)

	if st, err := cfg.fs.Stat(bin); err != nil {
		// build
		args := []string{"build", "-o", bin}
		if cfg.verbose {
			args = append(args, "-v")
		}
		args = append(args, string(t))
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Stdout = cfg.outW
		cmd.Stderr = cfg.errW
		cmd.Stdin = cfg.inR
		err = cmd.Run()
		if err != nil {
			return err
		}
	} else if st.IsDir() {
		return fmt.Errorf("%q is a directory", bin)
	}

	// exec
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = cfg.outW
	cmd.Stderr = cfg.errW
	cmd.Stdin = cfg.inR
	return cmd.Run()
}

func printHelp(cfg *config) {
	fmt.Fprintln(cfg.outW, helpText)
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
