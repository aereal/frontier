package main

import (
	"context"
	"os"

	"github.com/aereal/frontier/internal/cli"
)

func main() {
	if err := cli.New(os.Stdin, os.Stdout, os.Stderr).Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
