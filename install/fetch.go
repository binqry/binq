package install

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/binqry/binq/client"
	"github.com/binqry/binq/internal/erron"
	"github.com/binqry/binq/schema/item"
)

func (r *Runner) fetch() (err error) {
	if r.sourceURL == "" {
		return fmt.Errorf("Can't fetch because sourceURL is not set. Source: %s", r.Source)
	}
	req, err := client.NewHttpGetRequest(r.sourceURL, map[string]string{})
	if err != nil {
		return err
	}
	r.Logger.Printf("GET %s", r.sourceURL)
	res, _err := r.httpClient.Do(req)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to execute HTTP request")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, r.Source)
	}
	r.tmpdir, _err = ioutil.TempDir(os.TempDir(), "binq.*")
	if _err != nil {
		return erron.Errorwf(_err, "Failed to create tempdir")
	}
	defer func() {
		if err != nil {
			os.RemoveAll(r.tmpdir)
		}
	}()

	url, _err := url.Parse(r.sourceURL)
	if _err != nil {
		// Unexpected case
		return erron.Errorwf(_err, "Failed to parse source URL: %v", r.sourceURL)
	}
	base := path.Base(url.Path)
	r.download = filepath.Join(r.tmpdir, base)
	dl, _err := os.Create(r.download)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to open file: %s", r.download)
	}
	defer dl.Close()

	if r.sourceItem != nil {
		if cs := r.sourceItem.GetChecksum(base); cs != nil {
			return r.downloadWithChecksum(cs, res.Body, dl)
		}
		r.Logger.Noticef("Checksum is not provided. Skip verification")
	}

	// Download without checksum
	_, _err = io.Copy(dl, res.Body)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to read HTTP response")
	}
	r.Logger.Debugf("Saved file %s", r.download)

	return nil
}

func (r *Runner) downloadWithChecksum(cs *item.ItemChecksum, content io.ReadCloser, destFile *os.File) (err error) {
	sum, hasher, _ := cs.GetSumAndHasher()
	tee := io.TeeReader(content, hasher)
	_, _err := io.Copy(destFile, tee)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to read HTTP response")
	}
	r.Logger.Debugf("Saved file %s", r.download)

	cksum := hex.EncodeToString(hasher.Sum(nil))
	r.Logger.Debugf("Sum: %s", cksum)
	if sum != cksum {
		r.Logger.Errorf("Checksum differs! Maybe corrupt? Want: %s, Got: %s", sum, cksum)
	} else {
		r.Logger.Infof("Checksum is OK")
	}
	return nil
}
