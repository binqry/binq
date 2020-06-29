package main

import (
	"os"

	"github.com/progrhyme/binq/internal/cli"
)

func main() {
	os.Exit(cli.NewCLI(os.Stdout, os.Stderr).Run(os.Args))
}
