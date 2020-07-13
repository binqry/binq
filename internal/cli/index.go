package cli

import (
	"fmt"
	"net/url"
	"text/template"

	"github.com/binqry/binq"
	"github.com/binqry/binq/client"
	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

const (
	outFmtText = "text"
	outFmtJSON = "json"
)

type indexCmd struct {
	*commonCmd
	option *indexOpts
}

type indexOpts struct {
	server, outfmt *string
	*commonOpts
}

func newIndexCmd(common *commonCmd) (self *indexCmd) {
	self = &indexCmd{commonCmd: common}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &indexOpts{
		server:     fs.StringP("server", "s", "", "# Index Server URL"),
		outfmt:     fs.StringP("output", "o", "", "# Output format (text,json)"),
		commonOpts: newCommonOpts(fs),
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *indexCmd) usage() {
	const help = `Summary:
  List items on <<.prog>> index server.

Usage:
  <<.prog>> <<.name>> [-s|--server SERVER] [GENERAL_OPTIONS]

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
	server := *opt.server
	if server == "" {
		server = binq.DefaultBinqServer
	}
	level := lv.GetLevel()
	if *opt.logLv != "" {
		level = lv.WordToLevel(*opt.logLv)
	}
	logger := lv.New(cmd.errs, level, 0)
	svrURL, err := url.Parse(server)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! URL parse failed. %v\n", err)
		return exitNG
	}
	lv.Debugf("Server URL: %s", svrURL)

	clt := client.NewClient(svrURL, logger)
	index, err := clt.GetIndex()
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Can't get index data. Server: %s, Error: %v\n", server, err)
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
