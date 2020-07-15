package cli

import (
	"io"

	"github.com/progrhyme/go-lv"
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
	getLogLevel() *string
}

type commonOpts struct {
	help  *bool
	logLv *string
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

func (opt *commonOpts) getLogLevel() *string {
	return opt.logLv
}

func newCommonOpts(fs *pflag.FlagSet) *commonOpts {
	return &commonOpts{
		help:  fs.BoolP("help", "h", false, "# Show help"),
		logLv: fs.StringP("log-level", "L", "", "# Log level (debug,info,notice,warn,error)"),
	}
}

func logLevelByOption(opt flavor) (level lv.Level) {
	if *opt.getLogLevel() == "" {
		return defaultLogLevel
	}

	level = lv.WordToLevel(*opt.getLogLevel())
	if level == 0 {
		lv.Warnf("Unknown log level: %s. Use default", *opt.getLogLevel())
		return defaultLogLevel
	}
	return level
}

func setLogLevelByOption(opt flavor) {
	lv.SetLevel(logLevelByOption(opt))
}
