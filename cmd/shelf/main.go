package main

import (
	"os"

	"github.com/kyaoi/gitshelf/internal/cli"
	"github.com/kyaoi/gitshelf/internal/versioninfo"
)

var version string

func main() {
	if err := cli.NewRootCommand(versioninfo.Resolve(version)).Execute(); err != nil {
		os.Exit(1)
	}
}
