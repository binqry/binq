package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/gookit/color"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/progrhyme/binq/internal/logs"
	"github.com/progrhyme/binq/schema/item"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

type reviseCmd struct {
	*commonCmd
	option *reviseOpts
}

type reviseOpts struct {
	urlFormat, replacements, extensions, checksums *string
	delete, latest, noLatest, yes                  *bool
	*commonOpts
}

func newReviseCmd(common *commonCmd) (self *reviseCmd) {
	self = &reviseCmd{commonCmd: common}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &reviseOpts{
		urlFormat:    fs.StringP("url", "u", "", "# JSON parameter for \"url-format\""),
		replacements: fs.StringP("replace", "r", "", "# JSON parameter for \"replacements\""),
		extensions:   fs.StringP("ext", "e", "", "# JSON parameter for \"extensions\""),
		checksums:    fs.StringP("sum", "s", "", "# JSON parameter for \"checksums\""),
		yes:          fs.BoolP("yes", "y", false, "# Update JSON file without confirmation"),
		delete:       fs.Bool("delete", false, "# Delete version"),
		latest:       fs.Bool("latest", false, "# Add or Update as Latest Version"),
		noLatest:     fs.Bool("no-latest", false, "# Add or Update as Not Latest Version"),
		commonOpts:   newCommonOpts(fs),
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
  <<.prog>> <<.name>> path/to/item.json VERSION \
    [-s|--sum CHECKSUMS] [-u|--url URL_FORMAT] [-r|--replace REPLACEMENTS] [-e|--ext EXTENSIONS] \
    [--latest] [--no-latest] [-y|--yes]

  # Delete Version
  <<.prog>> <<.name>> path/to/item.json VERSION [--delete] [-y|--yes]

Examples:
  # Add v0.1.1 if not exist
  <<.prog>> <<.name>> foo.json 0.1.1

  # Delete v0.1.0-dev if exists
  <<.prog>> <<.name>> foo.json 0.1.0-dev --delete

  # Add or Update v0.2.0 with version specific parameters
  <<.prog>> <<.name>> foo.json 0.2.0 \
    -s "foo-win.zip:${sha256_win},foo-mac.zip:${sha256_mac}" --latest

Parameters:
- CHECKSUMS

  Format: "<File1>:<Checksum1>[:<Algorithm1>],..."

  SHA-256 is the default algorithm. To use CRC, specify like this: '-s "foo.zip:1093117945:crc"'.
  Other algorithm is not supported for now.

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

	if *opt.debug {
		logs.SetLevel(logs.Debug)
	} else if *opt.verbose {
		logs.SetLevel(logs.Info)
	}

	file, version := args[0], args[1]
	orig, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Can't read file: %s. %v\n", file, err)
		return exitNG
	}

	obj, err := item.DecodeItemJSON(orig)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to decode Item JSON. %v\n", err)
		return exitNG
	}
	logs.Debugf("Decoded JSON: %s", obj)

	if *opt.delete {
		if deleted := obj.DeleteRevision(version); deleted == false {
			fmt.Fprintf(cmd.errs, "Error! Version does not exist: %s\n", version)
			return exitNG
		}
		logs.Debugf("Version %s deleted. After Item: %s", version, obj)
		return cmd.writeRevisedItem(obj, file, orig)
	}

	var replacements, extensions map[string]string
	if *opt.replacements != "" {
		replacements = parseArgToStrMap(*opt.replacements)
	}
	if *opt.extensions != "" {
		extensions = parseArgToStrMap(*opt.extensions)
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
	}

	obj.AddOrUpdateRevision(rev, mode)
	logs.Debugf("Version %s updated. After Item: %s", version, obj)

	return cmd.writeRevisedItem(obj, file, orig)
}

func (cmd *reviseCmd) writeRevisedItem(obj *item.Item, src string, orig []byte) (exit int) {
	revised, err := obj.Print(true)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to print Item JSON. %v\n", err)
		return exitNG
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(strings.TrimRight(string(orig), "\r\n")),
		B:        difflib.SplitLines(string(revised)),
		FromFile: src,
		ToFile:   "Revised",
		Context:  3,
	}
	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Can't get diff. %v\n", err)
		return exitNG
	}
	if text == "" {
		fmt.Fprintln(cmd.errs, "No change")
	} else {
		if !*cmd.option.yes {
			if terminal.IsTerminal(1) {
				fmt.Fprintln(cmd.outs, colorizeDiff(text))
			} else {
				fmt.Fprintln(cmd.outs, text)
			}
		}
		if terminal.IsTerminal(0) && !*cmd.option.yes {
			fmt.Fprintf(cmd.errs, "Overwrite %s? (Y/n) ", src)
			stdin := bufio.NewScanner(os.Stdin)
			stdin.Scan()
			ans := stdin.Text()
			if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "N") {
				fmt.Fprintln(cmd.errs, "Canceled")
				return exitOK
			}
		}

		file, err := os.OpenFile(src, os.O_WRONLY|os.O_TRUNC, 0664)
		if err != nil {
			fmt.Fprintf(cmd.errs, "Error! Can't open file: %s\n", src)
			return exitNG
		}
		defer file.Close()
		if _, err = file.Write(revised); err != nil {
			fmt.Fprintf(cmd.errs, "Error! Can't write file: %s\n", src)
			return exitNG
		}
		fmt.Fprintf(cmd.outs, "Updated %s\n", src)
	}

	return exitOK
}

func colorizeDiff(diff string) (colored string) {
	lines := strings.Split(diff, "\n")
	for i, s := range lines {
		switch {
		case strings.HasPrefix(s, "---"):
			lines[i] = color.Danger.Render(s)
		case strings.HasPrefix(s, "+++"):
			lines[i] = color.Success.Render(s)
		case strings.HasPrefix(s, "-"):
			lines[i] = color.Red.Render(s)
		case strings.HasPrefix(s, "+"):
			lines[i] = color.Green.Render(s)
		case strings.HasPrefix(s, "@@"):
			lines[i] = color.Note.Render(s)
		default:
			// Nothing to do
		}
	}
	return strings.Join(lines, "\n")
}
