package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/progrhyme/binq"
	"github.com/progrhyme/go-lv"
)

const (
	exitOK = iota
	exitNG
)

const defaultLogLevel lv.Level = lv.LNotice

var (
	errFileNotFound = errors.New("File not found")
	errCanceled     = errors.New("Canceled")
)

type CLI struct {
	OutStream, ErrStream io.Writer
}

func NewCLI(outs, errs io.Writer) *CLI {
	return &CLI{OutStream: outs, ErrStream: errs}
}

func (c *CLI) Run(args []string) (exit int) {
	prog := filepath.Base(args[0])

	lv.Configure(c.ErrStream, defaultLogLevel, 0)
	common := &commonCmd{outs: c.OutStream, errs: c.ErrStream, prog: prog, name: "install"}
	installer := newInstallCmd(common)

	if len(args) == 1 {
		installer.usage(false)
		return exitNG
	}

	switch args[1] {
	case "install":
		return installer.run(args[1:])
	case "new":
		creator := newCreateCmd(common)
		creator.name = "new"
		return creator.run(args[2:])
	case "revise":
		revisor := newReviseCmd(common)
		revisor.name = "revise"
		return revisor.run(args[2:])
	case "verify":
		verifier := newVerifyCmd(common)
		verifier.name = "verify"
		return verifier.run(args[2:])
	case "register":
		registrar := newRegisterCmd(common)
		registrar.name = "register"
		return registrar.run(args[2:])
	case "modify":
		modifier := newModifyCmd(common)
		modifier.name = "modify"
		return modifier.run(args[2:])
	case "deregister":
		deregistrar := newDeregisterCmd(common)
		deregistrar.name = "deregister"
		return deregistrar.run(args[2:])
	case "version":
		fmt.Fprintf(c.OutStream, "Version: %s\n", binq.Version)
		return exitOK
	}

	return installer.run(args[1:])
}
