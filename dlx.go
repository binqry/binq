package dlx

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/mholt/archiver/v3"
)

type Runner struct {
	Origin     string
	DestDir    string
	Output     io.Writer
	tmpdir     string
	download   string
	extractDir string
	extracted  bool
}

var defaultRunner Runner

func Run(src, dir string, outs io.Writer, verbose bool) (err error) {
	defaultRunner.Origin = src
	defaultRunner.DestDir = dir
	defaultRunner.Output = outs
	return defaultRunner.Run(verbose)
}

func (r *Runner) Run(verbose bool) (err error) {
	if err = r.fetch(verbose); err != nil {
		return err
	}
	defer os.RemoveAll(r.tmpdir)
	if err = r.extract(verbose); err != nil {
		return err
	}
	if err = r.locate(verbose); err != nil {
		return err
	}
	return nil
}

func (r *Runner) fetch(verbose bool) (err error) {
	req, _err := http.NewRequest(http.MethodGet, r.Origin, nil)
	if _err != nil {
		return errorwf(_err, "Failed to create HTTP request")
	}
	req.Header.Set("User-Agent", fmt.Sprintf("dlx/%s", Version))
	res, _err := http.DefaultClient.Do(req)
	if _err != nil {
		return errorwf(_err, "Failed to execute HTTP request")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, r.Origin)
	}
	r.tmpdir, _err = ioutil.TempDir(os.TempDir(), "dlx.*")
	if _err != nil {
		return errorwf(_err, "Failed to create tempdir")
	}
	defer func() {
		if _err != nil {
			os.RemoveAll(r.tmpdir)
		}
	}()

	base := path.Base(r.Origin)
	r.download = filepath.Join(r.tmpdir, base)
	dl, _err := os.Create(r.download)
	if _err != nil {
		return errorwf(_err, "Failed to open file: %s", r.download)
	}
	defer dl.Close()
	_, _err = io.Copy(dl, res.Body)
	if _err != nil {
		return errorwf(_err, "Failed to read HTTP response")
	}

	return nil
}

func (r *Runner) extract(verbose bool) (err error) {
	r.extracted = false
	uai, _err := archiver.ByExtension(r.download)
	if _err != nil {
		fmt.Fprintf(r.Output, "[Notice] Unarchiver can't be determined. %s\n", _err)
		return nil
	}

	unarchiver, ok := uai.(archiver.Unarchiver)
	if !ok {
		err = fmt.Errorf(
			"Failed to determine Unarchiver. Probably file format is wrong. File: %s, Type: %T",
			r.download, uai)
		return err
	}

	r.extractDir = filepath.Join(r.tmpdir, "ext")
	os.Mkdir(r.extractDir, 0755)
	if _err = unarchiver.Unarchive(r.download, r.extractDir); _err != nil {
		return errorwf(_err, "Failed to unarchive: %s", r.download)
	}

	r.extracted = true
	return nil
}

func (r *Runner) locate(verbose bool) (err error) {
	if !r.extracted {
		dest := filepath.Join(r.DestDir, filepath.Base(r.download))
		if _err := os.Rename(r.download, dest); _err != nil {
			return errorwf(_err, "Failed to locate file: %s", dest)
		}
		fi, _err := os.Stat(dest)
		if _err != nil {
			return errorwf(_err, "Failed to get file info: %s", dest)
		}
		// Assume downloaded binary is executable
		if _err = os.Chmod(dest, fi.Mode()|0111); _err != nil {
			return errorwf(_err, "Failed to change file mode: %s", dest)
		}
		return nil
	}

	// Walk through the extracted directory to find and locate executable files
	return filepath.Walk(r.extractDir, func(path string, info os.FileInfo, problem error) error {
		if problem != nil || info.IsDir() {
			return problem
		}
		if isExecutable(info.Mode()) {
			dest := filepath.Join(r.DestDir, info.Name())
			if _err := os.Rename(path, dest); _err != nil {
				return errorwf(_err, "Failed to locate file: %s", dest)
			}
		}
		return nil
	})
}

// TODO: Support Windows
func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
