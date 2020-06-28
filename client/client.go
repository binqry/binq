package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/progrhyme/dlx"
	"github.com/progrhyme/dlx/internal/erron"
	"github.com/progrhyme/dlx/internal/logs"
)

type Runner struct {
	Source     string
	DestDir    string
	DestFile   string
	Logger     logs.Logging
	ServerURL  *url.URL
	httpClient *http.Client
	sourceURL  string
	tmpdir     string
	download   string
	extractDir string
	extracted  bool
}

type runOpts struct {
	source   string
	destDir  string
	destFile string
	outs     io.Writer
	level    logs.Level
	server   string
}

var defaultRunner Runner

func Run(opt runOpts) (err error) {
	defaultRunner.Source = opt.source
	defaultRunner.DestDir = opt.destDir
	defaultRunner.DestFile = opt.destFile
	defaultRunner.Logger = logs.New(opt.outs, opt.level, 0)
	defaultRunner.httpClient = newHttpClient(DefaultHTTPTimeout)

	var urlStr string
	if opt.server != "" {
		urlStr = opt.server
	} else if server := os.Getenv(dlx.EnvKeyServer); server != "" {
		urlStr = server
	}
	if urlStr != "" {
		uri, _err := url.Parse(urlStr)
		if _err != nil {
			return erron.Errorwf(_err, "Failed to parse server URL: %s", urlStr)
		}
		defaultRunner.ServerURL = uri
	}

	return defaultRunner.Run()
}

func (r *Runner) Run() (err error) {
	if err = r.prefetch(); err != nil {
		return err
	}
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

// prefetch query metadata for item info to fetch
func (r *Runner) prefetch() (err error) {
	if strings.HasPrefix(r.Source, "http") {
		r.sourceURL = r.Source
		return nil
	}
	if r.ServerURL == nil {
		return fmt.Errorf("No server is configured. Can't deal with source: %s", r.Source)
	}
	// Copy r.ServerURL
	urlItem, _err := url.Parse(r.ServerURL.String())
	if _err != nil {
		// Unexpected case
		return erron.Errorwf(_err, "Failed to parse server URL: %v", r.ServerURL)
	}
	urlItem.Path = path.Join(urlItem.Path, r.Source)

	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	req, err := newHttpGetRequest(urlItem.String(), headers)
	if err != nil {
		return err
	}
	r.Logger.Infof("GET %s", urlItem.String())
	// TODO: change timeout
	res, _err := r.httpClient.Do(req)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to execute HTTP request")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, r.Source)
	}

	// TODO: decode response

	return err
}

func (r *Runner) fetch() (err error) {
	if r.sourceURL == "" {
		return fmt.Errorf("Can't fetch because sourceURL is not set. Runner: %+v", r)
	}
	req, err := newHttpGetRequest(r.sourceURL, map[string]string{})
	if err != nil {
		return err
	}
	r.Logger.Printf("GET %s", r.Source)
	res, _err := r.httpClient.Do(req)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to execute HTTP request")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, r.Source)
	}
	r.tmpdir, _err = ioutil.TempDir(os.TempDir(), "dlx.*")
	if _err != nil {
		return erron.Errorwf(_err, "Failed to create tempdir")
	}
	defer func() {
		if _err != nil {
			os.RemoveAll(r.tmpdir)
		}
	}()

	base := path.Base(r.Source)
	r.download = filepath.Join(r.tmpdir, base)
	dl, _err := os.Create(r.download)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to open file: %s", r.download)
	}
	defer dl.Close()
	_, _err = io.Copy(dl, res.Body)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to read HTTP response")
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
		return erron.Errorwf(_err, "Failed to unarchive: %s", r.download)
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
			return erron.Errorwf(_err, "Failed to locate file: %s", dest)
		}
		fi, _err := os.Stat(dest)
		if _err != nil {
			return erron.Errorwf(_err, "Failed to get file info: %s", dest)
		}
		// Assume downloaded binary is executable
		if _err = os.Chmod(dest, fi.Mode()|0111); _err != nil {
			return erron.Errorwf(_err, "Failed to change file mode: %s", dest)
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
				return erron.Errorwf(_err, "Failed to locate file: %s", dest)
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
				return erron.Errorwf(_err, "Failed to locate file: %s", dest)
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
