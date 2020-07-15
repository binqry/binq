package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/binqry/binq/internal/erron"
	"github.com/binqry/binq/schema"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

// For index manipulator commands: "register", "modify", "deregister"
type indiceRunner interface {
	confirmRunner
	getPrevRawIndex() []byte
	setPrevRawIndex([]byte)
}

type indiceFlavor interface {
	getYes() *bool
}

type indiceCmd struct {
	prevRawIndex []byte
	*confirmCmd
}

func (cmd *indiceCmd) getPrevRawIndex() (b []byte) {
	return cmd.prevRawIndex
}

func (cmd *indiceCmd) setPrevRawIndex(b []byte) {
	cmd.prevRawIndex = b
}

func newIndiceOpts(fs *pflag.FlagSet) (opt *confirmOpts) {
	return &confirmOpts{
		yes:        fs.BoolP("yes", "y", false, "# Update Index data without confirmation"),
		commonOpts: newCommonOpts(fs),
	}
}

func resolveIndexPathByArg(arg string) (pathIndex string, err error) {
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

func decodeIndex(cmd indiceRunner, file string) (idx *schema.Index, err error) {
	if _, _err := os.Stat(file); os.IsNotExist(_err) {
		err = fmt.Errorf("Index file not found: %s", file)
		return idx, err
	}

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

	cmd.setPrevRawIndex(raw)
	return idx, nil
}

func writeNewIndex(cmd indiceRunner, idx *schema.Index, fileIndex string) (err error) {
	newRawIndex, _err := idx.ToJSON(true)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to encode new Index")
	}

	fromFile := "<Null>"
	if len(cmd.getPrevRawIndex()) > 0 {
		fromFile = fileIndex
	}

	diff, err := getDiff(diffArgs{
		textA: strings.TrimRight(string(cmd.getPrevRawIndex()), "\r\n"),
		textB: string(newRawIndex),
		fileA: fromFile,
		fileB: fileIndex,
	})
	if err != nil {
		return err
	}
	if diff == "" {
		fmt.Fprintln(cmd.getErrs(), "Index has no change")
		return nil
	}

	yes := *(cmd.getConfirmOpts().getYes())
	if !yes {
		fprintDiff(cmd.getOuts(), diff)
	}
	if terminal.IsTerminal(0) && !yes {
		fmt.Fprintf(cmd.getErrs(), "Write %s. Okay? (Y/n) ", fileIndex)
		stdin := bufio.NewScanner(os.Stdin)
		stdin.Scan()
		ans := stdin.Text()
		if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "N") {
			fmt.Fprintln(cmd.getErrs(), "Canceled")
			return errCanceled
		}
	}

	return writeFile(fileIndex, newRawIndex, func() {
		fmt.Fprintf(cmd.getOuts(), "Saved %s\n", fileIndex)
	})
}
