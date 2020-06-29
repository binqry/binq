package cli

import (
	"fmt"

	"github.com/progrhyme/binq"
	"github.com/progrhyme/binq/client"
	"github.com/progrhyme/binq/internal/logs"
	"github.com/spf13/pflag"
)

type installCmd struct {
	*commonCmd
	option *installOpts
}

type installOpts struct {
	target, directory, file, server *string
	version, noExtract, noExec      *bool
	*commonOpts
}

func newInstallCmd(common *commonCmd) (self *installCmd) {
	self = &installCmd{commonCmd: common}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &installOpts{
		version:   fs.BoolP("version", "V", false, "Show version"),
		target:    fs.StringP("target", "t", "", "Target Item (Name or URL)"),
		directory: fs.StringP("directory", "d", "", "Output Directory"),
		file:      fs.StringP("file", "f", "", "Output File name"),
		server:    fs.StringP("server", "s", "", "Index Server URL"),
		noExtract: fs.BoolP("no-extract", "z", false, "Don't extract archive"),
		noExec:    fs.BoolP("no-exec", "X", false, "Don't care for executable files"),
		commonOpts: &commonOpts{
			help:    fs.BoolP("help", "h", false, "Show help"),
			verbose: fs.BoolP("verbose", "v", false, "Verbose output"),
			debug:   fs.Bool("debug", false, "Show debug messages"),
		},
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *installCmd) usage() {
	fmt.Fprintf(cmd.errs, `Summary:
  "%s" does download & extract binary or archive via HTTP; then locate executable files into target
  directory.

Options:
`, cmd.prog)
	cmd.flags.PrintDefaults()
}

func (cmd *installCmd) run(args []string) (exit int) {
	if err := cmd.flags.Parse(args); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Parsing arguments failed. %s\n", err)
		return exitNG
	}

	opt := cmd.option
	if *opt.help {
		cmd.usage()
		return exitOK
	} else if *opt.version {
		fmt.Fprintf(cmd.outs, "Version: %s\n", binq.Version)
		return exitOK
	}

	if *opt.target == "" && cmd.flags.NArg() == 0 {
		fmt.Fprintln(cmd.errs, "Error! Target is not specified!")
		cmd.usage()
		return exitNG
	}

	mode := client.ModeDefault
	if *opt.noExtract {
		mode = mode ^ client.ModeExtract
	}
	if *opt.noExec {
		mode = mode ^ client.ModeExecutable
	}
	if mode == 0 {
		mode = client.ModeDLOnly
	}

	var source string
	if *opt.target != "" {
		source = *opt.target
	} else {
		source = cmd.flags.Arg(0)
	}

	logLevel := logs.Notice
	if *opt.debug {
		logLevel = logs.Debug
	} else if *opt.verbose {
		logLevel = logs.Info
	}
	opts := client.RunOption{
		Mode:      mode,
		Source:    source,
		DestDir:   *opt.directory,
		DestFile:  *opt.file,
		Output:    cmd.errs,
		LogLevel:  logLevel,
		ServerURL: *opt.server,
	}
	if err := client.Run(opts); err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}

	return exitOK
}
