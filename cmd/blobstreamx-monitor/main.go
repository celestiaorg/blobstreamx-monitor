package main

import (
	"context"
	"github.com/celestiaorg/blobstreamx-monitor/cmd/blobstreamx-monitor/root"
	"os"
)

func main() {
	rootCmd := root.Cmd()
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}
