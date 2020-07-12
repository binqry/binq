package cli

import (
	"fmt"
	"text/template"

	"github.com/progrhyme/binq/schema/item"
	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

type reviseCmd struct {
	*confirmCmd
	option *reviseOpts
}

type reviseOpts struct {
	version, urlFormat, replacements, extensions, renameFiles, checksums *string
	delete, latest, noLatest                                             *bool
	*confirmOpts
}

func (cmd *reviseCmd) getConfirmOpts() confirmFlavor {
	return cmd.option
}

func newReviseCmd(common *commonCmd) (self *reviseCmd) {
	self = &reviseCmd{confirmCmd: &confirmCmd{commonCmd: common}}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &reviseOpts{
		version:      fs.StringP("version", "v", "", "# JSON parameter for \"version\""),
		urlFormat:    fs.StringP("url", "u", "", "# JSON parameter for \"url-format\""),
		replacements: fs.StringP("replace", "r", "", "# JSON parameter for \"replacements\""),
		extensions:   fs.StringP("ext", "e", "", "# JSON parameter for \"extensions\""),
		renameFiles:  fs.StringP("rename", "R", "", "# JSON parameter for \"rename-files\""),
		checksums:    fs.StringP("sum", "s", "", "# JSON parameter for \"checksums\""),
		delete:       fs.Bool("delete", false, "# Delete version"),
		latest:       fs.Bool("latest", false, "# Add or Update as Latest Version"),
		noLatest:     fs.Bool("no-latest", false, "# Add or Update as Not Latest Version"),
		confirmOpts: &confirmOpts{
			yes:        fs.BoolP("yes", "y", false, "# Update JSON file without confirmation"),
			commonOpts: newCommonOpts(fs),
		},
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *reviseCmd) usage() {
	const help = `Summary:
  Revise a version in Item JSON for <<.prog>>.

Usage:
  # Add or Update Version
  <<.prog>> <<.name>> path/to/item.json [-v|--version] VERSION \
    [-s|--sum CHECKSUMS] [-u|--url URL_FORMAT] [-r|--replace REPLACEMENTS] [-e|--ext EXTENSIONS] \
    [-R|--rename RENAME_FILES] [--latest] [--no-latest] [-y|--yes] [GENERAL_OPTIONS]

  # Delete Version
  <<.prog>> <<.name>> path/to/item.json VERSION --delete [-y|--yes] [GENERAL_OPTIONS]

Examples:
  # Add v0.1.1 if not exist
  <<.prog>> <<.name>> foo.json -v 0.1.1

  # Delete v0.1.0-dev if exists
  <<.prog>> <<.name>> foo.json 0.1.0-dev --delete

  # Add or Update v0.2.0 with version specific parameters
  <<.prog>> <<.name>> foo.json 0.2.0 \
    -s "foo-win.zip:${sha256_win},foo-mac.zip:${sha256_mac}" --latest

Parameters:
- CHECKSUMS

  Format: "<File1>:<Checksum1>[:<Algorithm1>],..."

  SHA-256 is the default algorithm. To use CRC or MD5, specify suffix ":crc" or ":md5" like this:
  '-s "foo.zip:5993c24b:crc"'.
  Other algorithm is not supported for now.

Limitation:
  It is not expected to specify two or more types of checksums per file.

Options:
`

	t := template.Must(template.New("usage").Delims("<<", ">>").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})

	cmd.flags.PrintDefaults()
}

func (cmd *reviseCmd) run(args []string) (exit int) {
	if err := cmd.flags.Parse(args); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Parsing arguments failed. %s\n", err)
		return exitNG
	}

	opt := cmd.option
	if *opt.help {
		cmd.usage()
		return exitOK
	}
	if len(args) <= 1 {
		fmt.Fprintln(cmd.errs, "Error! Both JSON file and VERSION must be specified")
		cmd.usage()
		return exitNG
	}
	setLogLevelByOption(opt)

	file := args[0]
	var version string
	if *opt.version != "" {
		version = *opt.version
	} else {
		version = args[1]
	}
	orig, obj, err := readAndDecodeItemJSONFile(file)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}
	lv.Debugf("Decoded JSON: %s", obj)

	if *opt.delete {
		if deleted := obj.DeleteRevision(version); deleted == false {
			fmt.Fprintf(cmd.errs, "Error! Version does not exist: %s\n", version)
			return exitNG
		}
		lv.Debugf("Version %s deleted. After Item: %s", version, obj)
		return updateItemJSON(cmd, obj, file, orig)
	}

	var replacements, extensions, renameFiles map[string]string
	if *opt.replacements != "" {
		replacements = parseArgToStrMap(*opt.replacements, "replacement")
	}
	if *opt.extensions != "" {
		extensions = parseArgToStrMap(*opt.extensions, "extension")
	}
	if *opt.renameFiles != "" {
		renameFiles = parseArgToStrMap(*opt.renameFiles, "rename-files")
	}

	mode := item.ReviseModeNatural
	if *opt.latest {
		mode = item.ReviseModeLatest
	} else if *opt.noLatest {
		mode = item.ReviseModeOld
	}

	rev := &item.ItemRevision{
		Version:      version,
		Checksums:    item.NewItemChecksums(*opt.checksums),
		URLFormat:    *opt.urlFormat,
		Replacements: replacements,
		Extension:    extensions,
		RenameFiles:  renameFiles,
	}

	obj.AddOrUpdateRevision(rev, mode)
	lv.Debugf("Version %s updated. After Item: %s", version, obj)

	return updateItemJSON(cmd, obj, file, orig)
}
