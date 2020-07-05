package cli

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/progrhyme/binq/internal/logs"
	"github.com/spf13/pflag"
)

type deregisterCmd struct {
	*indexCmd
}

func newDeregisterCmd(common *commonCmd) (self *deregisterCmd) {
	self = &deregisterCmd{indexCmd: &indexCmd{commonCmd: common}}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = newIndexOpts(fs)
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *deregisterCmd) usage() {
	const help = `Summary:
  Deregister an Item from Local {{.prog}} Index Dataset.

Usage:
  {{.prog}} {{.name}} pato/to/root[/index.json] NAME [-y|--yes] [GENERAL_OPTIONS]

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})

	cmd.flags.PrintDefaults()
}

func (cmd *deregisterCmd) run(args []string) (exit int) {
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
	logs.Noticef("Target indice: %s", indice)

	if !idx.Remove(name) {
		// Unexpected
		fmt.Fprintf(cmd.errs, "Error! Failed to update Index. Name: %s\n", name)
		return exitNG
	}

	err = writeNewIndex(cmd, idx, fileIndex)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}

	var exists bool
	pathItem := indice.Path
	for _, i := range idx.Items {
		if i.Path == pathItem {
			logs.Noticef("Item \"%s\" still refers to \"%s\"", i.Name, i.Path)
			exists = true
		}
	}

	if !exists {
		pathItem = filepath.Join(filepath.Dir(fileIndex), pathItem)
		err = removeFile(pathItem)
		switch err {
		case nil:
			fmt.Fprintf(cmd.outs, "Deleted Item JSON: %s\n", pathItem)
		case errFileNotFound:
			logs.Warnf("Can't remove file: %s. Not Found", pathItem)
		default:
			fmt.Fprintf(cmd.errs, "Error! %v\n", err)
			return exitNG
		}
	}

	return exitOK
}
