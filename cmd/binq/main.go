package main

import (
	"os"

	"github.com/binqry/binq/internal/cli"
)

func main() {
	os.Exit(cli.NewCLI(os.Stdout, os.Stderr).Run(os.Args))
}
