package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/progrhyme/binq"
	"github.com/progrhyme/binq/client"
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
	help := flags.BoolP("help", "h", false, "Show help")
	version := flags.BoolP("version", "V", false, "Show version")
	target := flags.StringP("target", "t", "", "Target Item (Name or URL)")
	directory := flags.StringP("directory", "d", "", "Output Directory")
	file := flags.StringP("file", "f", "", "Output File name")
	server := flags.StringP("server", "s", "", "Index Server URL")
	noExtract := flags.BoolP("no-extract", "z", false, "Don't extract archive")
	noExec := flags.BoolP("no-exec", "X", false, "Don't care for executable files")
	verbose := flags.BoolP("verbose", "v", false, "Verbose output")
	debug := flags.Bool("debug", false, "Show debug messages")
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

	if *target == "" && flags.NArg() == 0 {
		fmt.Fprintln(c.ErrStream, "Error! Target is not specified!")
		flags.Usage()
		return exitNG
	}

	mode := client.ModeDefault
	if *noExtract {
		mode = mode ^ client.ModeExtract
	}
	if *noExec {
		mode = mode ^ client.ModeExecutable
	}
	if mode == 0 {
		mode = client.ModeDLOnly
	}

	var source string
	if *target != "" {
		source = *target
	} else {
		source = flags.Arg(0)
	}

	logLevel := logs.Notice
	if *debug {
		logLevel = logs.Debug
	} else if *verbose {
		logLevel = logs.Info
	}
	opts := client.RunOption{
		Mode:      mode,
		Source:    source,
		DestDir:   *directory,
		DestFile:  *file,
		Output:    c.ErrStream,
		LogLevel:  logLevel,
		ServerURL: *server,
	}
	if err := client.Run(opts); err != nil {
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
