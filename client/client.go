package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/binqry/binq"
	"github.com/binqry/binq/internal/erron"
	"github.com/binqry/binq/schema/item"
	"github.com/mholt/archiver/v3"
	"github.com/progrhyme/go-lv"
)

type Mode int

const (
	ModeDLOnly Mode = 1 << iota
	ModeExtract
	ModeExecutable
	ModeDefault = ModeExtract | ModeExecutable
)

type Runner struct {
	Mode       Mode
	Source     string
	DestDir    string
	DestFile   string
	Logger     lv.Loggable
	ServerURL  *url.URL
	httpClient *http.Client
	sourceURL  string
	sourceItem *item.ItemRevision
	os         string
	arch       string
	tmpdir     string
	download   string
	extractDir string
	extracted  bool
}

type RunOption struct {
	Mode      Mode
	Source    string
	DestDir   string
	DestFile  string
	Output    io.Writer
	LogLevel  lv.Level
	ServerURL string
}

var defaultRunner Runner

func Run(opt RunOption) (err error) {
	defaultRunner = Runner{
		Source:     opt.Source,
		DestDir:    opt.DestDir,
		DestFile:   opt.DestFile,
		Logger:     lv.New(opt.Output, opt.LogLevel, 0),
		httpClient: NewHttpClient(DefaultHTTPTimeout),
		os:         runtime.GOOS,
		arch:       runtime.GOARCH,
	}
	if opt.Mode == 0 {
		defaultRunner.Mode = ModeDefault
	} else {
		defaultRunner.Mode = opt.Mode
	}

	var urlStr string
	if opt.ServerURL != "" {
		urlStr = opt.ServerURL
	} else if server := os.Getenv(binq.EnvKeyServer); server != "" {
		urlStr = server
	} else {
		urlStr = binq.DefaultBinqServer
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
	if _err := r.prefetch(); _err != nil {
		return erron.Errorwf(_err, "Can't fetch item data. Target: %s, Server: %s", r.Source, r.ServerURL)
	}
	if err = r.fetch(); err != nil {
		return err
	}
	defer os.RemoveAll(r.tmpdir)
	if r.Mode&ModeExtract != 0 {
		if err = r.extract(); err != nil {
			return err
		}
	}
	if err = r.locate(); err != nil {
		return err
	}
	return nil
}

func (r *Runner) extract() (err error) {
	r.extracted = false
	uai, _err := archiver.ByExtension(r.download)
	if _err != nil {
		r.Logger.Debugf("Unarchiver can't be determined. %s", _err)
		r.Logger.Noticef("Skip extraction")
		return nil
	}

	unarchiver, ok := uai.(archiver.Unarchiver)
	if !ok {
		err = fmt.Errorf(
			"Failed to determine Unarchiver. Probably file format is wrong. File: %s, Type: %T",
			r.download, uai)
		return err
	}

	base := strings.TrimSuffix(filepath.Base(r.download), filepath.Ext(r.download))
	r.extractDir = filepath.Join(r.tmpdir, base)
	os.Mkdir(r.extractDir, 0755)
	r.Logger.Printf("Extracts archive %s", r.download)
	if _err = unarchiver.Unarchive(r.download, r.extractDir); _err != nil {
		return erron.Errorwf(_err, "Failed to unarchive: %s", r.download)
	}

	r.extracted = true
	return nil
}

func (r *Runner) locate() (err error) {
	// !ModeExtract OR Unextractable file
	if !r.extracted {
		var dest string
		if r.DestFile == "" {
			destFile := r.renameFileBySchema(filepath.Base(r.download))
			if destFile != "" {
				dest = filepath.Join(r.DestDir, destFile)
			} else {
				dest = filepath.Join(r.DestDir, filepath.Base(r.download))
			}
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
		if r.Mode&ModeExecutable != 0 {
			// Assume downloaded binary is executable
			if _err = os.Chmod(dest, fi.Mode()|0111); _err != nil {
				return erron.Errorwf(_err, "Failed to change file mode: %s", dest)
			}
		}
		r.Logger.Printf("Installed %s", dest)
		return nil
	}

	// ModeExtract AND Succeed to Extract
	// Codes below satisfy these conditions

	// ModeExtract AND !ModeExecutable
	// Just locate to destination directory
	if r.Mode&ModeExecutable == 0 {
		dest := filepath.Join(r.DestDir, filepath.Base(r.extractDir))
		if _err := os.Rename(r.extractDir, dest); _err != nil {
			return erron.Errorwf(_err, "Failed to locate content: %s", dest)
		}
		r.Logger.Printf("Installed %s", dest)
		return nil
	}

	// ModeExtract AND ModeExecutable
	// Walk through the extracted directory to find and locate executable files
	installed := []string{}
	err = filepath.Walk(r.extractDir, func(path string, info os.FileInfo, problem error) error {
		r.Logger.Debugf("Walking in archive: %s", path)
		if problem != nil || info.IsDir() {
			return problem
		}
		if isExecutable(info.Mode()) {
			var dest string
			destFile := r.renameFileBySchema(info.Name())
			if destFile != "" {
				dest = filepath.Join(r.DestDir, destFile)
			} else {
				dest = filepath.Join(r.DestDir, info.Name())
			}
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

func (r *Runner) renameFileBySchema(orig string) (tobe string) {
	if r.sourceItem == nil {
		return ""
	}
	param := item.FormatParam{OS: r.os, Arch: r.arch}
	tobe = r.sourceItem.ConvertFileName(orig, param)
	if tobe != "" {
		r.Logger.Infof("Rename: %s => %s", orig, tobe)
	}
	return tobe
}

// TODO: Support Windows
func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
