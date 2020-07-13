package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

type modifyCmd struct {
	*indexCmd
	option *modifyOpts
}

type modifyOpts struct {
	newName, path *string
	*confirmOpts
}

func (cmd *modifyCmd) getConfirmOpts() confirmFlavor {
	return cmd.option
}

func newModifyCmd(common *commonCmd) (self *modifyCmd) {
	self = &modifyCmd{indexCmd: &indexCmd{
		confirmCmd: &confirmCmd{
			commonCmd: common,
		},
	}}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &modifyOpts{
		newName:     fs.StringP("name", "n", "", "# New Name for the Item"),
		path:        fs.StringP("path", "p", "", "# New Path for the Item"),
		confirmOpts: newIndexOpts(fs),
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *modifyCmd) usage() {
	const help = `Summary:
  Modify the indice properties of an Item in Local {{.prog}} Index Dataset.

Usage:
  {{.prog}} {{.name}} pato/to/root[/index.json] NAME \
    [-n|--name NEW_NAME] [-p|--path PATH] [-y|--yes] [GENERAL_OPTIONS]

Example:
  {{.prog}} {{.name}} index-root-dir foo -n bar -p example.com/bar[/index.json]

The command above modify "foo" entry in Index, with new name and path.
If you want to update the content of an Item, use "{{.prog}} register" command.

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})

	cmd.flags.PrintDefaults()
}

func (cmd *modifyCmd) run(args []string) (exit int) {
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
		fmt.Fprintln(cmd.errs, "Error! Both PATH_OF_INDEX and NAME_OF_ITEM must be specified")
		cmd.usage()
		return exitNG
	}
	setLogLevelByOption(opt)

	fileIndex, err := resolveIndexPathByArg(args[0])
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %s\n", err)
		return exitNG
	}
	idx, err := decodeIndex(cmd, fileIndex)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %s\n", err)
		return exitNG
	}

	name := args[1]
	indice := idx.Find(name)
	if indice == nil {
		fmt.Fprintf(cmd.errs, "Item not found in index. Name: %s, Index: %s\n", name, fileIndex)
		return exitNG
	}
	lv.Noticef("Target indice: %s", indice)

	newName, newPathItem := *opt.newName, *opt.path

	var modified bool
	var oldPathItem string
	if newName != "" && newName != name {
		indice.Name = newName
		modified = true
	}
	if newPathItem != "" && newPathItem != indice.Path {
		oldPathItem = indice.Path
		indice.Path = newPathItem
		modified = true
	}

	if !modified {
		fmt.Fprintf(cmd.errs, "No change\n")
		return exitOK
	}

	if !idx.Swap(name, indice) {
		// Unexpected
		fmt.Fprintf(cmd.errs, "Error! Failed to update Index. Name: %s\n", name)
		return exitNG
	}

	err = writeNewIndex(cmd, idx, fileIndex)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}

	if oldPathItem == "" {
		return exitOK
	}

	oldPathItem = filepath.Join(filepath.Dir(fileIndex), oldPathItem)
	newPathItem = filepath.Join(filepath.Dir(fileIndex), newPathItem)
	newDir := filepath.Dir(newPathItem)
	if err = os.MkdirAll(newDir, 0755); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Can't make directory: %s. %v\n", newDir, err)
		return exitNG
	}
	if err = os.Rename(oldPathItem, newPathItem); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to move file: %s => %s. %v\n", oldPathItem, newPathItem, err)
		return exitNG
	}
	fmt.Fprintf(cmd.outs, "Moved Item JSON: %s => %s\n", oldPathItem, newPathItem)

	return exitOK
}
