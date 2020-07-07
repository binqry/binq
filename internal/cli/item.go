package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/progrhyme/binq/internal/erron"
	"github.com/progrhyme/binq/schema/item"
	"golang.org/x/crypto/ssh/terminal"
)

func readAndDecodeItemJSONFile(file string) (raw []byte, obj *item.Item, err error) {
	raw, _err := ioutil.ReadFile(file)
	if _err != nil {
		err = erron.Errorwf(_err, "Can't read item file: %s", file)
		return raw, obj, err
	}
	obj, _err = item.DecodeItemJSON(raw)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to decode Item JSON: %s", file)
		return raw, obj, err
	}
	return raw, obj, nil
}

func updateItemJSON(cmd confirmRunner, obj *item.Item, file string, prev []byte) (exit int) {
	updated, err := obj.Print(true)
	if err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! Failed to print Item JSON. %v\n", err)
		return exitNG
	}

	diff, err := getDiff(diffArgs{
		textA: strings.TrimRight(string(prev), "\r\n"),
		textB: string(updated),
		fileA: file,
		fileB: "Updated",
	})
	if err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! %v\n", err)
		return exitNG
	}
	if diff == "" {
		fmt.Fprintln(cmd.getErrs(), "No change")
		return exitOK
	}
	if !*cmd.getConfirmOpts().getYes() {
		fprintDiff(cmd.getOuts(), diff)
	}
	if terminal.IsTerminal(0) && !*cmd.getConfirmOpts().getYes() {
		fmt.Fprintf(cmd.getErrs(), "Overwrite %s? (Y/n) ", file)
		stdin := bufio.NewScanner(os.Stdin)
		stdin.Scan()
		ans := stdin.Text()
		if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "N") {
			fmt.Fprintln(cmd.getErrs(), "Canceled")
			return exitOK
		}
	}

	if err = writeFile(file, updated, func() {
		fmt.Fprintf(cmd.getOuts(), "Updated %s\n", file)
	}); err != nil {
		return exitNG
	}

	return exitOK
}