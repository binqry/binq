package main

import (
	"os"

	"github.com/progrhyme/dlx"
)

func main() {
	os.Exit(dlx.NewCLI(os.Stdout, os.Stderr).Run(os.Args))
}
