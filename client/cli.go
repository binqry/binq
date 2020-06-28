package client

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/progrhyme/dlx"
	"github.com/progrhyme/dlx/internal/logs"
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

func (c *CLI) Run(args []string) (exit int) {
	prog := filepath.Base(args[0])
	flags := pflag.NewFlagSet(prog, pflag.ContinueOnError)
	flags.SetOutput(c.ErrStream)
	help := flags.BoolP("help", "h", false, "show help")
	version := flags.BoolP("version", "V", false, "show version")
	verbose := flags.BoolP("verbose", "v", false, "verbose output")
	debug := flags.Bool("debug", false, "show debug messages")
	directory := flags.StringP("directory", "d", "", "output directory")
	file := flags.StringP("file", "f", "", "output file name")
	flags.Usage = func() { c.usage(flags, prog) }
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(c.ErrStream, "Error! Parsing arguments failed. %s\n", err)
		return exitNG
	}

	if *help {
		flags.Usage()
		return exitOK
	} else if *version {
		fmt.Fprintf(c.OutStream, "Version: %s\n", dlx.Version)
		return exitOK
	}

	if flags.NArg() == 0 {
		fmt.Fprintln(c.ErrStream, "Error! Target is not specified!")
		flags.Usage()
		return exitNG
	}

	source := flags.Arg(0)
	logLevel := logs.Notice
	if *debug {
		logLevel = logs.Debug
	} else if *verbose {
		logLevel = logs.Info
	}
	if err := Run(source, *directory, *file, c.ErrStream, logLevel); err != nil {
		fmt.Fprintf(c.ErrStream, "Error! %v\n", err)
		return exitNG
	}

	return exitOK
}

func (c *CLI) usage(fs *pflag.FlagSet, prog string) {
	fmt.Fprintf(c.ErrStream, `Summary:
  "%s" does download & extract binary or archive via HTTP; then locate executable files into target
  directory.

Options:
`, prog)
	fs.PrintDefaults()
}
