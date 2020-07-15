// Package install implements installation functionality of binq.
package install

import "errors"

type Mode int

const (
	ModeDLOnly Mode = 1 << iota
	ModeExtract
	ModeExecutable
	ModeDefault = ModeExtract | ModeExecutable
)

var ErrVersionNotNewerThanThreshold = errors.New("Item version is not newer than given threshold")
