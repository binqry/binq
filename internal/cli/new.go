package cli

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/progrhyme/binq/internal/logs"
	"github.com/progrhyme/binq/schema/item"
	"github.com/spf13/pflag"
)

type createCmd struct {
	*commonCmd
	option *createOpts
}

type createOpts struct {
	replacements, extensions, file *string
	*commonOpts
}

func newCreateCmd(common *commonCmd) (self *createCmd) {
	self = &createCmd{commonCmd: common}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &createOpts{
		file:         fs.StringP("file", "f", "", "# Output File name"),
		replacements: fs.StringP("replace", "r", "", "# JSON parameter for \"replacements\""),
		extensions:   fs.StringP("ext", "e", "", "# JSON parameter for \"extensions\""),
		commonOpts:   newCommonOpts(fs),
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *createCmd) usage() {
	const help = `Summary:
  Generate a template Item JSON for <<.prog>>.

Usage:
  <<.prog>> <<.name>> URL_FORMAT [VERSION] [-f|--file OUTPUT_FILE] \
    [-r|--replace REPLACEMENTS] [-e|--ext EXTENSIONS]

Examples:
  <<.prog>> <<.name>> "https://github.com/rust-lang/mdBook/releases/download/v{{.Version}}/mdbook-v{{.Version}}-{{.Arch}}-{{.OS}}{{.Ext}}" \
    0.4.0 -r amd64:x86_64,darwin:apple-darwin,linux:unknown-linux-gnu,windows:pc-windows-msvc \
    -e default:.tar.gz,windows:.zip

The command above generates JSON like this:

  {
    "meta": {
      "url-format": "https://github.com/rust-lang/mdBook/releases/download/v{{.Version}}/mdbook-v{{.Version}}-{{.Arch}}-{{.OS}}{{.Ext}}",
      "replacements": {
        "amd64": "x86_64",
        "darwin": "apple-darwin",
        "linux": "unknown-linux-gnu",
        "windows": "pc-windows-msvc"
      },
      "extension": {
        "default": ".tar.gz",
        "windows": ".zip"
      }
    },
    "latest": {
      "version": "0.4.0"
    }
  }

This is a valid JSON with which <<.prog>> download and install the archive "mdbook".

Options:
`

	t := template.Must(template.New("usage").Delims("<<", ">>").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})

	cmd.flags.PrintDefaults()
}

func (cmd *createCmd) run(args []string) (exit int) {
	if err := cmd.flags.Parse(args); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Parsing arguments failed. %s\n", err)
		return exitNG
	}

	opt := cmd.option
	if *opt.help {
		cmd.usage()
		return exitOK
	}
	if len(args) == 0 {
		fmt.Fprintln(cmd.errs, "Error! URL Format is not specified")
		cmd.usage()
		return exitNG
	}

	var urlFormat, version string
	var replacements, extensions map[string]string
	urlFormat = args[0]
	if len(args) >= 2 {
		version = args[1]
	}
	if *opt.replacements != "" {
		replacements = parseArgToStrMap(*opt.replacements)
	}
	if *opt.extensions != "" {
		extensions = parseArgToStrMap(*opt.extensions)
	}

	if *opt.debug {
		logs.SetLevel(logs.Debug)
	} else if *opt.verbose {
		logs.SetLevel(logs.Info)
	}

	rev := &item.ItemRevision{
		URLFormat:    urlFormat,
		Version:      version,
		Replacements: replacements,
		Extension:    extensions,
	}

	gen, err := item.GenerateItemJSON(rev, true)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to generate Item JSON. %v\n", err)
		return exitNG
	}

	if *opt.file != "" {
		file, err := os.OpenFile(*opt.file, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Fprintf(cmd.errs, "Error! Can't open file: %s\n", *opt.file)
			fmt.Fprintln(cmd.outs, string(gen))
			return exitNG
		}
		fmt.Fprintln(file, string(gen))
		fmt.Fprintf(cmd.errs, "Written %s\n", *opt.file)
	} else {
		fmt.Fprintln(cmd.outs, string(gen))
	}

	return exitOK
}

func parseArgToStrMap(arg string) (m map[string]string) {
	m = make(map[string]string)
	for _, kv := range strings.Split(arg, ",") {
		params := strings.Split(kv, ":")
		switch len(params) {
		case 2:
			m[params[0]] = params[1]
		default:
			logs.Warnf("Wrong argement for replacement: %s", kv)
		}
	}
	return m
}
