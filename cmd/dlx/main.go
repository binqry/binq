package main

import (
	"os"

	"github.com/progrhyme/dlx/client"
)

func main() {
	os.Exit(client.NewCLI(os.Stdout, os.Stderr).Run(os.Args))
}
