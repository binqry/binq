package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/progrhyme/binq/internal/erron"
)

func writeFile(file string, raw []byte, onSuccess func()) (err error) {
	dir := filepath.Dir(file)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Can't make directory: %s", dir)
	}

	fout, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("Can't open file: %s", file)
	}
	defer fout.Close()
	if _, err = fout.Write(append(raw, "\n"[0])); err != nil {
		return fmt.Errorf("Can't write file: %s", file)
	}
	onSuccess()
	return nil
}

func copyFile(src, dest string) (err error) {
	fin, _err := os.Open(src)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to open src file: %s", src)
	}
	defer fin.Close()

	dir := filepath.Dir(dest)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Can't make directory: %s", dir)
	}

	fout, _err := os.Create(dest)
	if _err != nil {
		return erron.Errorwf(_err, "Failed to open dest file: %s", dest)
	}
	defer fout.Close()

	if _, _err = io.Copy(fout, fin); _err != nil {
		return erron.Errorwf(_err, "Failed to copy src to dest: %s => %s", src, dest)
	}

	return nil
}

func removeFile(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return errFileNotFound
	}

	return os.Remove(file)
}
