package dlx

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

type Runner struct {
	Origin, DestDir, tmpdir, download, extracted string
}

var defaultRunner Runner

func Run(src, dir string, verbose bool) (err error) {
	defaultRunner.Origin = src
	defaultRunner.DestDir = dir
	return defaultRunner.Run(verbose)
}

func (r *Runner) Run(verbose bool) (err error) {
	if err = r.fetch(verbose); err != nil {
		return err
	}
	defer os.RemoveAll(r.tmpdir)
	// TODO: detect filetype
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
	// TODO: implement
	return err
}

func (r *Runner) locate(verbose bool) (err error) {
	dest := filepath.Join(r.DestDir, filepath.Base(r.download))
	if _err := os.Rename(r.download, dest); _err != nil {
		return errorwf(_err, "Failed to locate file: %s", dest)
	}
	return err
}
