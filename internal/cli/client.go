package cli

import (
	"fmt"
	"net/url"

	"github.com/binqry/binq"
	"github.com/binqry/binq/client"
	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

type clientRunner interface {
	runner
	getClientOpts() clientFlavor
	getServer() string
	setServer(string)
}

type clientFlavor interface {
	flavor
	getServer() *string
}

type clientCmd struct {
	server string
	*commonCmd
	option *clientOpts
}

type clientOpts struct {
	server *string
	*commonOpts
}

func (cmd *clientCmd) getClientOpts() clientFlavor {
	return cmd.option
}

func (cmd *clientCmd) getServer() (svr string) {
	return cmd.server
}

func (cmd *clientCmd) setServer(svr string) {
	cmd.server = svr
}

func (opt *clientOpts) getServer() (svr *string) {
	return opt.server
}

func newClientOpts(fs *pflag.FlagSet) *clientOpts {
	return &clientOpts{
		server: fs.StringP("server", "s", "", "# Index Server URL"),
		commonOpts: &commonOpts{
			help:  fs.BoolP("help", "h", false, "# Show help"),
			logLv: fs.StringP("log-level", "L", "", "# Log level (debug,info,notice,warn,error)"),
		},
	}
}

func getClient(cmd clientRunner) (clt *client.Client, err error) {
	server := *cmd.getClientOpts().getServer()
	if server == "" {
		server = binq.DefaultBinqServer
	}
	cmd.setServer(server)

	level := logLevelByOption(cmd.getClientOpts())
	logger := lv.New(cmd.getErrs(), level, 0)
	svrURL, err := url.Parse(server)
	if err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! URL parse failed. %v\n", err)
		return nil, err
	}
	lv.Debugf("Server URL: %s", svrURL)

	clt = client.NewClient(svrURL, logger)
	return clt, nil
}
