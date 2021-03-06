package cli

import (
	"fmt"
	"os"
	"text/template"

	"github.com/binqry/binq"
	"github.com/binqry/binq/install"
	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

type installCmd struct {
	*commonCmd
	option *installOpts
}

type installOpts struct {
	target, directory, file, server *string
	noExtract, noExec               *bool
	*commonOpts
}

func newInstallCmd(common *commonCmd) (self *installCmd) {
	self = &installCmd{commonCmd: common}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &installOpts{
		target:     fs.StringP("target", "t", "", "# Target Item (Name or URL)"),
		directory:  fs.StringP("directory", "d", "", "# Output Directory"),
		file:       fs.StringP("file", "f", "", "# Output File name"),
		server:     fs.StringP("server", "s", "", "# Index Server URL"),
		noExtract:  fs.BoolP("no-extract", "z", false, "# Don't extract archive"),
		noExec:     fs.BoolP("no-exec", "X", false, "# Don't care for executable files"),
		commonOpts: newCommonOpts(fs),
	}
	fs.Usage = func() { self.usage(true) }
	self.flags = fs

	return self
}

func (cmd *installCmd) usage(subcommand bool) {
	if subcommand {
		const help = `Summary:
  Download & extract binary or archive via HTTP; then locate executable files into target directory.

Syntax:
  {{.prog}} [{{.name}}] [-t|--target] SOURCE[@VERSION]
    [-d|--dir OUTPUT_DIR] [-f|--file OUTFILE] \
    [-s|--server SERVER] \
    [-z|--no-extract] [-X|--no-exec] \
    [GENERAL_OPTIONS]

Examples:
  # With full URL
  {{.prog}} https://github.com/peco/peco/releases/download/v0.5.7/peco_darwin_amd64.zip \
    -d path/to/bin
  {{.prog}} {{.name}} -t https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 \
    -d path/to/bin -f jq

  # With index server which defaults to https://binqry.github.io/index/
  {{.prog}} {{.name}} peco -d path/to/bin
  export BINQ_BIN_DIR=path/to/bin
  {{.prog}} jq@1.6

  # Specify index server
  binq -s https://your-index-server/ peco@0.5.7
  export BINQ_SERVER="https://your-index-server/"
  binq jq

Options:
`

		t := template.Must(template.New("usage").Parse(help))
		t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})

		cmd.flags.PrintDefaults()
	} else {
		// As root command
		const help = `Summary:
  "{{.prog}}" does download & extract binary or archive via HTTP; then locate executable files into target
  directory.

Usage:
  {{.prog}} [install] [arguments...] [options...]
  {{.prog}} COMMAND [arguments...] [options...]
  {{.prog}} -h|--help

Available Commands:
  install (Default)  # Install binary or archive Item
  index              # List Items on Index Server
  new                # Create Item Manifest
  revise             # Add/Edit/Delete a version in Item Manifest
  verify             # Verify checksum of a version in Item Manifest
  register           # Register or Update Item Manifest onto Local Index Dataset
  modify             # Modify Item properties on Local Index Dataset
  deregister         # Deregister Item from Local Index
  self-upgrade       # Upgrade {{.prog}} binary itself
  version            # Show {{.prog}} version

Run "{{.prog}} COMMAND -h|--help" to see usage of each command.

General Options:
  -h|--help                # Show help
  -L, --log-level string   # Log level (debug,info,notice,warn,error)
`
		t := template.Must(template.New("usage").Parse(help))
		t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})
	}
}

func (cmd *installCmd) run(args []string) (exit int) {
	var subcommand bool
	if args[0] == "install" {
		subcommand = true
		args = args[1:]
	}
	if err := cmd.flags.Parse(args); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Parsing arguments failed. %s\n", err)
		cmd.usage(subcommand)
		return exitNG
	}

	opt := cmd.option
	if *opt.help {
		cmd.usage(subcommand)
		return exitOK
	}

	if *opt.target == "" && cmd.flags.NArg() == 0 {
		fmt.Fprintln(cmd.errs, "Error! Target is not specified!")
		cmd.usage(subcommand)
		return exitNG
	}

	mode := install.ModeDefault
	if *opt.noExtract {
		mode = mode ^ install.ModeExtract
	}
	if *opt.noExec {
		mode = mode ^ install.ModeExecutable
	}
	if mode == 0 {
		mode = install.ModeDLOnly
	}

	var source string
	if *opt.target != "" {
		source = *opt.target
	} else {
		source = cmd.flags.Arg(0)
	}
	setLogLevelByOption(opt)

	dir := os.Getenv(binq.EnvKeyBinDir)
	if *opt.directory != "" {
		dir = *opt.directory
	}
	opts := install.RunOption{
		Mode:      mode,
		Source:    source,
		DestDir:   dir,
		DestFile:  *opt.file,
		Output:    cmd.errs,
		LogLevel:  lv.GetLevel(),
		ServerURL: *opt.server,
	}
	if err := install.Run(opts); err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}

	return exitOK
}
