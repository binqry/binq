package cli

import (
	"io"

	"github.com/progrhyme/binq/internal/logs"
	"github.com/spf13/pflag"
)

type runner interface {
	getOuts() io.Writer
	getErrs() io.Writer
}

type commonCmd struct {
	outs, errs io.Writer
	prog, name string
	flags      *pflag.FlagSet
}

type flavor interface {
	getHelp() *bool
	getVerbose() *bool
	getDebug() *bool
}

type commonOpts struct {
	help, verbose, debug *bool
}

func (cmd *commonCmd) getOuts() io.Writer {
	return cmd.outs
}

func (cmd *commonCmd) getErrs() io.Writer {
	return cmd.errs
}

func (opt *commonOpts) getHelp() *bool {
	return opt.help
}

func (opt *commonOpts) getVerbose() *bool {
	return opt.verbose
}

func (opt *commonOpts) getDebug() *bool {
	return opt.debug
}

func newCommonOpts(fs *pflag.FlagSet) *commonOpts {
	return &commonOpts{
		help:    fs.BoolP("help", "h", false, "# Show help"),
		verbose: fs.BoolP("verbose", "v", false, "# Show verbose messages"),
		debug:   fs.Bool("debug", false, "# Show debug messages"),
	}
}

func setLogLevelByOption(opt flavor) {
	if *opt.getDebug() {
		logs.SetLevel(logs.Debug)
	} else if *opt.getVerbose() {
		logs.SetLevel(logs.Info)
	}
}
