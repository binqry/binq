package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	"github.com/progrhyme/dlx"
	"github.com/progrhyme/dlx/internal/logs"
)

type Runner struct {
	Origin     string
	DestDir    string
	DestFile   string
	Logger     logs.Logging
	tmpdir     string
	download   string
	extractDir string
	extracted  bool
}

var defaultRunner Runner

func Run(src, dir, file string, outs io.Writer, lv logs.Level) (err error) {
	defaultRunner.Origin = src
	defaultRunner.DestDir = dir
	defaultRunner.DestFile = file
	defaultRunner.Logger = logs.New(outs, lv, 0)
	return defaultRunner.Run()
}

func (r *Runner) Run() (err error) {
	if err = r.fetch(); err != nil {
		return err
	}
	defer os.RemoveAll(r.tmpdir)
	if err = r.extract(); err != nil {
		return err
	}
	if err = r.locate(); err != nil {
		return err
	}
	return nil
}

func (r *Runner) fetch() (err error) {
	req, _err := http.NewRequest(http.MethodGet, r.Origin, nil)
	if _err != nil {
		return errorwf(_err, "Failed to create HTTP request")
	}
	req.Header.Set("User-Agent", fmt.Sprintf("dlx/%s", dlx.Version))
	r.Logger.Printf("GET %s", r.Origin)
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
	r.Logger.Debugf("Saved file %s", r.download)

	return nil
}

func (r *Runner) extract() (err error) {
	r.extracted = false
	uai, _err := archiver.ByExtension(r.download)
	if _err != nil {
		r.Logger.Noticef("Unarchiver can't be determined. %s", _err)
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
	r.Logger.Infof("Extracts archive %s", r.download)
	if _err = unarchiver.Unarchive(r.download, r.extractDir); _err != nil {
		return errorwf(_err, "Failed to unarchive: %s", r.download)
	}

	r.extracted = true
	return nil
}

func (r *Runner) locate() (err error) {
	if !r.extracted {
		var dest string
		if r.DestFile == "" {
			dest = filepath.Join(r.DestDir, filepath.Base(r.download))
		} else {
			dest = filepath.Join(r.DestDir, r.DestFile)
		}
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
		r.Logger.Printf("Installed %s", dest)
		return nil
	}

	// Walk through the extracted directory to find and locate executable files
	installed := []string{}
	err = filepath.Walk(r.extractDir, func(path string, info os.FileInfo, problem error) error {
		r.Logger.Debugf("Walking in archive: %s", path)
		if problem != nil || info.IsDir() {
			return problem
		}
		if isExecutable(info.Mode()) {
			dest := filepath.Join(r.DestDir, info.Name())
			if _err := os.Rename(path, dest); _err != nil {
				return errorwf(_err, "Failed to locate file: %s", dest)
			}
			r.Logger.Printf("Installed %s", dest)
			installed = append(installed, dest)
		}
		return nil
	})
	switch len(installed) {
	case 0:
		r.Logger.Warnf("Archive has no executables. None is installed")
	case 1:
		if r.DestFile != "" && r.DestFile != filepath.Base(installed[0]) {
			dest := filepath.Join(r.DestDir, r.DestFile)
			if _err := os.Rename(installed[0], dest); _err != nil {
				return errorwf(_err, "Failed to locate file: %s", dest)
			}
			r.Logger.Printf("Moved %s to %s", installed[0], dest)
		}
	default:
		// Do nothing
	}

	return err
}

// TODO: Support Windows
func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
