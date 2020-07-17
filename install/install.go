// Package install implements installation functionality of binq.
package install

import (
	"errors"
	"regexp"
	"runtime"
)

type Mode int

const (
	ModeDLOnly Mode = 1 << iota
	ModeExtract
	ModeExecutable
	ModeDefault = ModeExtract | ModeExecutable
)

var ErrVersionNotNewerThanThreshold = errors.New("Item version is not newer than given threshold")

var (
	isWindows = runtime.GOOS == "windows"
	winRegExe = regexp.MustCompile(`^[\w\-]+\.exe$`)
)
