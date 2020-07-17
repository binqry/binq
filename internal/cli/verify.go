package cli

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/binqry/binq/client/http"
	"github.com/binqry/binq/internal/erron"
	"github.com/binqry/binq/schema/item"
	"github.com/progrhyme/go-lv"
	"github.com/spf13/pflag"
)

const initialDummyChecksum = "--dummy--"

type verifyCmd struct {
	*confirmCmd
	option *verifyOpts
}

type verifyOpts struct {
	version, os, arch *string
	keep              *bool
	*confirmOpts
}

func (cmd *verifyCmd) getConfirmOpts() confirmFlavor {
	return cmd.option
}

func newVerifyCmd(common *commonCmd) (self *verifyCmd) {
	self = &verifyCmd{confirmCmd: &confirmCmd{commonCmd: common}}

	fs := pflag.NewFlagSet(self.name, pflag.ContinueOnError)
	fs.SetOutput(self.errs)
	self.option = &verifyOpts{
		version: fs.StringP("version", "v", "", "# JSON parameter for \"version\""),
		os:      fs.String("os", "", "# JSON parameter for \"{{.OS}}\""),
		arch:    fs.StringP("arch", "a", "", "# JSON parameter for \"{{.Arch}}\""),
		keep:    fs.Bool("keep", false, "# Delete version"),
		confirmOpts: &confirmOpts{
			yes:        fs.BoolP("yes", "y", false, "# Update JSON file without confirmation"),
			commonOpts: newCommonOpts(fs),
		},
	}
	fs.Usage = self.usage
	self.flags = fs

	return self
}

func (cmd *verifyCmd) usage() {
	const help = `Summary:
  Download a specified version in <<.prog>> Item Manifest JSON and Verify its checksum.
  And update the checksum when needed.

Usage:
  <<.prog>> <<.name>> path/to/item.json [-v|--version VERSION] [--os OS] [-a|--arch ARCH] \
    [-y|--yes] [--keep] [GENERAL_OPTIONS]

When VERSION argument is omitted, the latest version will be verified.

Parameters:
- OS ... windows, darwin, linux etc.
- ARCH ... 386, amd64, arm etc.

When OS or ARCH parameter is omitted, value from running environment will be complemented.

Options:
`

	t := template.Must(template.New("usage").Delims("<<", ">>").Parse(help))
	t.Execute(cmd.errs, map[string]string{"prog": cmd.prog, "name": cmd.name})

	cmd.flags.PrintDefaults()
}

func (cmd *verifyCmd) run(args []string) (exit int) {
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
		fmt.Fprintln(cmd.errs, "Error! Target is not specified")
		cmd.usage()
		return exitNG
	}
	setLogLevelByOption(opt)

	fileItem := args[0]
	orig, obj, err := readAndDecodeItemJSONFile(fileItem)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %v\n", err)
		return exitNG
	}
	lv.Debugf("Decoded JSON: %s", obj)

	rev, err := getItemRevisionByOpt(obj, opt)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! %s\n", err)
		return exitNG
	}

	urlStr, err := rev.GetURL(buildURLParamToVerify(opt))
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! URL build failed. %v\n", err)
		return exitNG
	}
	if urlStr == "" {
		fmt.Fprintf(cmd.errs, "Error! URL is undefined. Version: %s\n", rev.Version)
		return exitNG
	}
	fmt.Fprintf(cmd.outs, "GET %s\n", urlStr)

	res, err := http.Fetch(urlStr)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to execute HTTP request. %v\n", err)
		return exitNG
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Fprintf(cmd.errs, "Error! HTTP response is not OK. Code: %d\n", res.StatusCode)
		return exitNG
	}

	tmpdir, err := ioutil.TempDir(os.TempDir(), "binq-verify.*")
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to create tempdir. %v\n", err)
		return exitNG
	}
	defer func() {
		if !*opt.keep {
			os.RemoveAll(tmpdir)
			fmt.Fprintf(cmd.errs, "Cleaning up completed\n")
		}
	}()

	urlObj, err := url.Parse(urlStr)
	if err != nil {
		// Unexpected case
		fmt.Fprintf(cmd.errs, "Failed to parse URL: %s", urlStr)
		return exitNG
	}
	file := path.Base(urlObj.Path)
	dlPath := filepath.Join(tmpdir, file)
	dlFile, err := os.Create(dlPath)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Failed to open file: %s", dlPath)
		return exitNG
	}
	defer dlFile.Close()

	cs := rev.GetChecksum(file)
	if cs == nil {
		lv.Noticef("Checksum is not provided")
		cs = &item.ItemChecksum{
			File:   file,
			SHA256: initialDummyChecksum, // To be replaced on verification
		}
	}

	updated, err := cmd.downloadAndVerify(cs, res.Body, dlFile, dlPath)
	if err != nil {
		fmt.Fprintf(cmd.errs, "Error! Failed to verify: %v", err)
		return exitNG
	}

	if updated {
		obj.UpdateRevisionChecksum(rev.Version, cs)
		lv.Debugf("Item updated. After Item: %s", obj)
		return updateItemJSON(cmd, obj, fileItem, orig)
	}

	if *opt.keep {
		fmt.Fprintf(cmd.errs, "Downloaded item remains at %s\nRemove it manually\n", dlPath)
	}
	return exitOK
}

func getItemRevisionByOpt(obj *item.Item, opt *verifyOpts) (rev *item.ItemRevision, err error) {
	if *opt.version != "" {
		rev = obj.GetRevision(*opt.version)
		if rev == nil {
			err = fmt.Errorf("Can't get item revision. Version: %s", *opt.version)
			return nil, err
		}
	} else {
		rev = obj.GetLatest()
		if rev == nil {
			err = fmt.Errorf("Can't get latest item revision")
			return nil, err
		}
	}
	return rev, nil
}

func buildURLParamToVerify(opt *verifyOpts) (param item.FormatParam) {
	if *opt.os != "" {
		param.OS = *opt.os
	} else {
		param.OS = runtime.GOOS
	}
	if *opt.arch != "" {
		param.Arch = *opt.arch
	} else {
		param.Arch = runtime.GOARCH
	}
	return param
}

func (cmd *verifyCmd) downloadAndVerify(
	cs *item.ItemChecksum, content io.ReadCloser, destFile *os.File, destPath string,
) (different bool, err error) {
	sum, hasher, kind := cs.GetSumAndHasher()
	tee := io.TeeReader(content, hasher)
	_, _err := io.Copy(destFile, tee)
	if _err != nil {
		return different, erron.Errorwf(_err, "Failed to read HTTP response")
	}
	lv.Debugf("Saved file %s", destPath)

	cksum := hex.EncodeToString(hasher.Sum(nil))
	lv.Debugf("Sum: %s", cksum)
	if sum != cksum {
		if sum != initialDummyChecksum {
			fmt.Fprintf(cmd.errs, "Warning! Checksum differs. Expected: %s, Got: %s\n", sum, cksum)
		}
		cs.SetSum(cksum, kind)
		return true, nil
	}

	fmt.Fprintf(cmd.outs, "Checksum is OK\n")
	return false, nil
}
