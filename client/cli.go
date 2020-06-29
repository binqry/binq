package client

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

func (c *CLI) Run(args []string) (exit int) {
	prog := filepath.Base(args[0])
	flags := pflag.NewFlagSet(prog, pflag.ContinueOnError)
	flags.SetOutput(c.ErrStream)
	help := flags.BoolP("help", "h", false, "show help")
	version := flags.BoolP("version", "V", false, "show version")
	noExtract := flags.BoolP("no-extract", "z", false, "don't extract archive")
	noExec := flags.BoolP("no-exec", "X", false, "don't care for executable files")
	verbose := flags.BoolP("verbose", "v", false, "verbose output")
	debug := flags.Bool("debug", false, "show debug messages")
	directory := flags.StringP("directory", "d", "", "output directory")
	file := flags.StringP("file", "f", "", "output file name")
	server := flags.StringP("server", "s", "", "index server URL")
	flags.Usage = func() { c.usage(flags, prog) }
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(c.ErrStream, "Error! Parsing arguments failed. %s\n", err)
		return exitNG
	}

	if *help {
		flags.Usage()
		return exitOK
	} else if *version {
		fmt.Fprintf(c.OutStream, "Version: %s\n", binq.Version)
		return exitOK
	}

	if flags.NArg() == 0 {
		fmt.Fprintln(c.ErrStream, "Error! Target is not specified!")
		flags.Usage()
		return exitNG
	}

	mode := ModeDefault
	if *noExtract {
		mode = mode ^ ModeExtract
	}
	if *noExec {
		mode = mode ^ ModeExecutable
	}
	if mode == 0 {
		mode = ModeDLOnly
	}

	source := flags.Arg(0)

	logLevel := logs.Notice
	if *debug {
		logLevel = logs.Debug
	} else if *verbose {
		logLevel = logs.Info
	}
	opts := runOpts{
		mode:     mode,
		source:   source,
		destDir:  *directory,
		destFile: *file,
		outs:     c.ErrStream,
		level:    logLevel,
		server:   *server,
	}
	if err := Run(opts); err != nil {
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
