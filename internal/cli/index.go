package cli

import (
	"fmt"
	"text/template"

	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

const (
	outFmtText = "text"
	outFmtJSON = "json"
)

type indexCmd struct {
	*clientCmd
	option *indexOpts
}

type indexOpts struct {
	outfmt *string
	*clientOpts
}

func (cmd *indexCmd) getClientOpts() clientFlavor {
	return cmd.option
}

func newIndexCmd(common *commonCmd) (self *indexCmd) {
	self = &indexCmd{clientCmd: &clientCmd{commonCmd: common}}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &indexOpts{
		outfmt:     fs.StringP("output", "o", "", "# Output format (text,json)"),
		clientOpts: newClientOpts(fs),
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *indexCmd) usage() {
	const help = `Summary:
  List items on <<.prog>> index server.

Usage:
  <<.prog>> <<.name>> [-s|--server SERVER] [-o|--output FORMAT] [GENERAL_OPTIONS]

Options:
`

	t := template.Must(template.New("usage").Delims("<<", ">>").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})
	cmd.flags.PrintDefaults()
}

func (cmd *indexCmd) run(args []string) (exit int) {
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

	clt, err := getClient(cmd)
	if err != nil {
		return exitNG
	}
	index, err := clt.GetIndex()
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Can't get index data. Server: %s, Error: %v\n", cmd.server, err)
		return exitNG
	}

	switch *opt.outfmt {
	case outFmtJSON:
		b, err := index.ToJSON(true)
		if err != nil {
			fmt.Fprintf(cmd.errs, "Error! Failed to output index data. %v\n", err)
			return exitNG
		}
		fmt.Fprintf(cmd.outs, "%s\n", b)
	case outFmtText, "":
		fmt.Fprint(cmd.outs, index.ToText())
	default:
		lv.Noticef("Unknown output format: %s", *opt.outfmt)
		fmt.Fprint(cmd.outs, index.ToText())
	}

	return exitOK
}
