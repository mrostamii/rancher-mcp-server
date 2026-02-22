package main

import (
	"os"

	"github.com/mrostamii/rancher-mcp-server/internal/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
