package cli

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/progrhyme/binq/schema/item"
	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

type createCmd struct {
	*commonCmd
	option *createOpts
}

type createOpts struct {
	version, replacements, extensions, renameFiles, file *string
	*commonOpts
}

func newCreateCmd(common *commonCmd) (self *createCmd) {
	self = &createCmd{commonCmd: common}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &createOpts{
		version:      fs.StringP("version", "v", "", "# JSON parameter for \"version\""),
		file:         fs.StringP("file", "f", "", "# Output File name"),
		replacements: fs.StringP("replace", "r", "", "# JSON parameter for \"replacements\""),
		extensions:   fs.StringP("ext", "e", "", "# JSON parameter for \"extensions\""),
		renameFiles:  fs.StringP("rename", "R", "", "# JSON parameter for \"rename-files\""),
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
  <<.prog>> <<.name>> URL_FORMAT [-v|--version VERSION] [-f|--file OUTPUT_FILE] \
    [-r|--replace REPLACEMENTS] [-e|--ext EXTENSIONS] [-R|--rename RENAME_FILES] \
    [GENERAL_OPTIONS]

Examples:
  <<.prog>> <<.name>> "https://github.com/rust-lang/mdBook/releases/download/v{{.Version}}/mdbook-v{{.Version}}-{{.Arch}}-{{.OS}}{{.Ext}}" \
    -v 0.4.0 -r amd64:x86_64,darwin:apple-darwin,linux:unknown-linux-gnu,windows:pc-windows-msvc \
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
    },
    "versions": [
      {
        "version": "0.4.0"
      }
    ]
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
	setLogLevelByOption(opt)

	var urlFormat string
	var replacements, extensions, renameFiles map[string]string
	urlFormat = args[0]
	if *opt.replacements != "" {
		replacements = parseArgToStrMap(*opt.replacements, "replacement")
	}
	if *opt.extensions != "" {
		extensions = parseArgToStrMap(*opt.extensions, "extension")
	}
	if *opt.renameFiles != "" {
		renameFiles = parseArgToStrMap(*opt.renameFiles, "rename-files")
	}

	rev := &item.ItemRevision{
		URLFormat:    urlFormat,
		Version:      *opt.version,
		Replacements: replacements,
		Extension:    extensions,
		RenameFiles:  renameFiles,
	}

	gen, err := item.GenerateItemJSON(rev, true)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to generate Item JSON. %v\n", err)
		return exitNG
	}

	if *opt.file != "" {
		writeFile(*opt.file, gen, func() {
			fmt.Fprintf(cmd.errs, "Written %s\n", *opt.file)
		})
	} else {
		fmt.Fprintln(cmd.outs, string(gen))
	}

	return exitOK
}

func parseArgToStrMap(arg, kind string) (m map[string]string) {
	m = make(map[string]string)
	for _, kv := range strings.Split(arg, ",") {
		params := strings.Split(kv, ":")
		switch len(params) {
		case 2:
			m[params[0]] = params[1]
		default:
			lv.Warnf("Wrong argement for %s: %s", kind, kv)
		}
	}
	return m
}
