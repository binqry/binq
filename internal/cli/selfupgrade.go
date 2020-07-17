package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/binqry/binq"
	"github.com/binqry/binq/install"
	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

type selfUpgradeCmd struct {
	progPath string
	*commonCmd
	option *selfUpgradeOpts
}

type selfUpgradeOpts struct {
	identifier *string
	*clientOpts
}

func (cmd *selfUpgradeCmd) getClientOpts() clientFlavor {
	return cmd.option
}

func newSelfUpgradeCmd(common *commonCmd, prog string) (self *selfUpgradeCmd) {
	self = &selfUpgradeCmd{
		progPath:  prog,
		commonCmd: common,
	}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &selfUpgradeOpts{
		identifier: fs.StringP("ident", "i", "", "# binq identifying name or path in the index server"),
		clientOpts: newClientOpts(fs),
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *selfUpgradeCmd) usage() {
	const help = `Summary:
  Upgrade {{.prog}} binary itself.

Usage:
  {{.prog}} {{.name}} [OPTIONS] [GENERAL_OPTIONS]

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})
	cmd.flags.PrintDefaults()
}

func (cmd *selfUpgradeCmd) run(args []string) (exit int) {
	if err := cmd.flags.Parse(args); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Parsing arguments failed. %s\n", err)
		return exitNG
	}

	opt := cmd.option
	if *opt.help {
		cmd.usage()
		return exitOK
	}
	setLogLevelByOption(opt)

	ident := "binq"
	if *opt.identifier != "" {
		ident = *opt.identifier
	}

	tmpdir, err := ioutil.TempDir(os.TempDir(), "binq.*")
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to create tempdir. %v", err)
		return exitNG
	}
	defer func() {
		os.RemoveAll(tmpdir)
	}()

	fmt.Fprintf(cmd.errs, "Check and fetch latest %s ...\n", ident)
	logDest := &strings.Builder{}
	opts := install.RunOption{
		Source:    ident,
		DestDir:   tmpdir,
		Output:    logDest,
		LogLevel:  lv.GetLevel(),
		ServerURL: *opt.server,
		NewerThan: binq.Version,
	}
	err = install.Run(opts)
	switch {
	case err == nil:
		// OK
	case errors.As(err, &install.ErrVersionNotNewerThanThreshold):
		fmt.Fprintf(cmd.errs, "No need to upgrade\n")
		return exitOK
	default:
		fmt.Fprint(cmd.errs, logDest.String())
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}

	tmpBinq := filepath.Join(tmpdir, "binq")
	// Check installation and print version
	if err = cmd.runBinqVersion(tmpBinq); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to run \"%s version\". %v", tmpBinq, err)
		return exitNG
	}
	if err = os.Rename(tmpBinq, cmd.progPath); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to install binq. %v\n", err)
		return exitNG
	}

	fmt.Fprintf(cmd.errs, "%s is upgraded\n", cmd.progPath)
	return exitOK
}

func (cmd *selfUpgradeCmd) runBinqVersion(bin string) (err error) {
	exe := exec.Command(bin, "version")
	exe.Stdout = cmd.outs
	exe.Stderr = cmd.errs
	return exe.Run()
}
