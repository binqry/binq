package cli

import (
	"io"

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

type commonOpts struct {
	help, verbose, debug *bool
}

func (cmd *commonCmd) getOuts() io.Writer {
	return cmd.outs
}

func (cmd *commonCmd) getErrs() io.Writer {
	return cmd.errs
}

func newCommonOpts(fs *pflag.FlagSet) *commonOpts {
	return &commonOpts{
		help:    fs.BoolP("help", "h", false, "# Show help"),
		verbose: fs.BoolP("verbose", "v", false, "# Show verbose messages"),
		debug:   fs.Bool("debug", false, "# Show debug messages"),
	}
}
