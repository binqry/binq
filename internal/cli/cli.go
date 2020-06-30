package cli

import (
	"io"
	"path/filepath"

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

	common := &commonCmd{outs: c.OutStream, errs: c.ErrStream, prog: prog, name: "install"}
	installer := newInstallCmd(common)

	if len(args) == 1 {
		installer.usage()
		return exitNG
	}

	return installer.run(args[1:])
}
