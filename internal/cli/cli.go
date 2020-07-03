package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/progrhyme/binq"
	"github.com/progrhyme/binq/internal/logs"
	"github.com/spf13/pflag"
)

const (
	exitOK = iota
	exitNG
)

type CLI struct {
	OutStream, ErrStream io.Writer
}

func NewCLI(outs, errs io.Writer) *CLI {
	return &CLI{OutStream: outs, ErrStream: errs}
}

type commonCmd struct {
	outs, errs io.Writer
	prog, name string
	flags      *pflag.FlagSet
}

type commonOpts struct {
	help, verbose, debug *bool
}

func (c *CLI) Run(args []string) (exit int) {
	prog := filepath.Base(args[0])

	logs.Configure(c.ErrStream, logs.Notice, 0)
	common := &commonCmd{outs: c.OutStream, errs: c.ErrStream, prog: prog, name: "install"}
	installer := newInstallCmd(common)

	if len(args) == 1 {
		installer.usage(false)
		return exitNG
	}

	switch args[1] {
	case "install":
		return installer.run(args[1:])
	case "new":
		creator := newCreateCmd(common)
		creator.name = "new"
		return creator.run(args[2:])
	case "revise":
		revisor := newReviseCmd(common)
		revisor.name = "revise"
		return revisor.run(args[2:])
	case "version":
		fmt.Fprintf(c.OutStream, "Version: %s\n", binq.Version)
		return exitOK
	}

	return installer.run(args[1:])
}

func newCommonOpts(fs *pflag.FlagSet) *commonOpts {
	return &commonOpts{
		help:    fs.BoolP("help", "h", false, "# Show help"),
		verbose: fs.BoolP("verbose", "v", false, "# Show verbose messages"),
		debug:   fs.Bool("debug", false, "# Show debug messages"),
	}
}
