package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/progrhyme/binq/internal/erron"
	"github.com/progrhyme/binq/internal/logs"
	"github.com/progrhyme/binq/schema"
	"github.com/progrhyme/binq/schema/item"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

type registerCmd struct {
	prevRawIndex []byte
	*commonCmd
	option *registerOpts
}

type registerOpts struct {
	name, path *string
	yes        *bool
	*commonOpts
}

func newRegisterCmd(common *commonCmd) (self *registerCmd) {
	self = &registerCmd{commonCmd: common}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &registerOpts{
		name:       fs.StringP("name", "n", "", "# Identical name for Item in Index"),
		path:       fs.StringP("path", "p", "", "# Path for Item in Index"),
		yes:        fs.BoolP("yes", "y", false, "# Update Index data without confirmation"),
		commonOpts: newCommonOpts(fs),
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *registerCmd) usage() {
	const help = `Summary:
  Register or Update Item content on {{.prog}} Index.

Usage:
  {{.prog}} {{.name}} pato/to/root[/index.json] path/to/item.json \
    [-n|--name NAME] [-p|--path PATH] [-y|--yes] [GENERAL_OPTIONS]

Example:
  {{.prog}} {{.name}} index-root-dir foo.json -n foo -p example.com/foo[/index.json]

The command above registers foo.json as "foo", copying the JSON file into
"index-root-dir/example.com/foo/index.json".
You can specify different path for Item like "example.com/foo.json".
But you can't choose different name from "index.json" for Index JSON.

With this command, Index JSON file will be created when it does not exist.
When existing name in Index is specified, the JSON file should be replaced with new file.

If you want to modify name or path in Index without altering its content, use "{{.prog}} modify"
command.

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})

	cmd.flags.PrintDefaults()
}

func (cmd *registerCmd) run(args []string) (exit int) {
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
		fmt.Fprintln(cmd.errs, "Error! Both PATH_OF_INDEX and PATH_OF_ITEM must be specified")
		cmd.usage()
		return exitNG
	}

	if *opt.debug {
		logs.SetLevel(logs.Debug)
	} else if *opt.verbose {
		logs.SetLevel(logs.Info)
	}

	fileIndex, err := parseArgIndex(args[0])
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %s\n", err)
		return exitNG
	}
	idx, err := cmd.decodeOrGenerateIndex(fileIndex)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %s\n", err)
		return exitNG
	}

	fileItem := args[1]
	rawItem, err := ioutil.ReadFile(fileItem)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Can't read item file: %s. %v\n", fileItem, err)
		return exitNG
	}

	obj, err := item.DecodeItemJSON(rawItem)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to decode Item JSON: %s. %v\n", fileItem, err)
		return exitNG
	}

	name, pathItem := *opt.name, *opt.path
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(fileItem), filepath.Ext(fileItem))
	}
	indice := idx.Find(name)
	if pathItem == "" && indice == nil {
		pathItem = selectPathForItem(obj, fileItem)
	}

	var oldPathItem string
	if indice == nil {
		indice = &schema.IndiceItem{Name: name, Path: pathItem}
		idx.Add(indice)
	} else if pathItem != "" {
		if pathItem != indice.Path {
			oldPathItem = indice.Path
			indice.Path = pathItem
			if !idx.Swap(name, indice) {
				// Unexpected
				fmt.Fprintf(cmd.errs, "Error! Failed to update Index. Name: %s\n", name)
				return exitNG
			}
		}
	} else {
		pathItem = indice.Path
	}

	err = cmd.writeNewIndex(idx, fileIndex)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}

	destPathItem := filepath.Join(filepath.Dir(fileIndex), pathItem)
	if err = copyFile(fileItem, destPathItem); err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}
	fmt.Fprintf(cmd.outs, "Copied Item JSON: %s => %s\n", fileItem, destPathItem)

	if oldPathItem != "" {
		oldPathItem = filepath.Join(filepath.Dir(fileIndex), oldPathItem)
		err = removeFile(oldPathItem)
		switch err {
		case nil:
			fmt.Fprintf(cmd.outs, "Deleted old Item JSON: %s\n", oldPathItem)
		case errFileNotFound:
			logs.Warnf("Can't remove file: %s. Not Found", oldPathItem)
		default:
			fmt.Fprintf(cmd.errs, "Error! %v\n", err)
			return exitNG
		}
	}

	return exitOK
}

func parseArgIndex(arg string) (pathIndex string, err error) {
	pathIndex = arg
	if strings.HasSuffix(pathIndex, ".json") {
		if filepath.Base(pathIndex) != "index.json" {
			err = fmt.Errorf("INDEX JSON filename must be \"index.json\". Given: %s", pathIndex)
			return pathIndex, err
		}
	} else {
		pathIndex = filepath.Join(pathIndex, "index.json")
	}
	return pathIndex, err
}

func (cmd *registerCmd) decodeOrGenerateIndex(file string) (idx *schema.Index, err error) {
	if _, _err := os.Stat(file); os.IsNotExist(_err) {
		logs.Noticef("Index file doesn't exist; will be created")
		return schema.NewIndex(), nil
	}

	// Decode Index JSON
	raw, _err := ioutil.ReadFile(file)
	if _err != nil {
		err = erron.Errorwf(_err, "Error! Can't read item file: %s", file)
		return idx, err
	}

	idx, _err = schema.DecodeIndexJSON(raw)
	if _err != nil {
		err = erron.Errorwf(_err, "Error! Can't decode Index JSON: %s", file)
		return idx, err
	}

	cmd.prevRawIndex = raw
	return idx, nil
}

func selectPathForItem(obj *item.Item, fileItem string) (pathItem string) {
	uf, err := url.Parse(obj.Meta.URLFormat)
	if err != nil {
		logs.Warnf("Failed to parse url-format of item: %s. %v", obj.Meta.URLFormat, err)
		return fileItem
	}

	// Ex) github.com/user/repo
	paths := strings.Split(strings.TrimPrefix(uf.Path, "/"), "/")
	if len(paths) >= 3 {
		return path.Join(uf.Host, paths[0], paths[1], "index.json")
	}
	return path.Join(append(append([]string{uf.Host}, paths...), "index.json")...)
}

func (cmd *registerCmd) writeNewIndex(idx *schema.Index, fileIndex string) (err error) {
	newRawIndex, _err := idx.Print(true)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to encode new Index")
	}

	fromFile := "<Null>"
	if len(cmd.prevRawIndex) > 0 {
		fromFile = fileIndex
	}

	diff, err := getDiff(diffArgs{
		textA: strings.TrimRight(string(cmd.prevRawIndex), "\r\n"),
		textB: string(newRawIndex),
		fileA: fromFile,
		fileB: fileIndex,
	})
	if err != nil {
		return err
	}
	if diff == "" {
		fmt.Fprintln(cmd.errs, "Index has no change")
		return nil
	}
	if !*cmd.option.yes {
		fprintDiff(cmd.outs, diff)
	}
	if terminal.IsTerminal(0) && !*cmd.option.yes {
		fmt.Fprintf(cmd.errs, "Write %s. Okay? (Y/n) ", fileIndex)
		stdin := bufio.NewScanner(os.Stdin)
		stdin.Scan()
		ans := stdin.Text()
		if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "N") {
			fmt.Fprintln(cmd.errs, "Canceled")
			return errCanceled
		}
	}

	return writeFile(fileIndex, newRawIndex, func() {
		fmt.Fprintf(cmd.outs, "Saved %s\n", fileIndex)
	})
}
