package root

import (
	"github.com/celestiaorg/blobstreamx-monitor/cmd/blobstreamx-monitor/version"
	"github.com/celestiaorg/blobstreamx-monitor/cmd/blobstreamx-monitor/watch"
	"github.com/spf13/cobra"
)

// Cmd creates a new root command for the Blobstreamx-monitor CLI. It is called once in the
// main function.
func Cmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "blobstreamx-monitor",
		Short:        "The BlobstreamX monitor CLI",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(
		version.Cmd,
		watch.Command(),
	)

	rootCmd.SetHelpCommand(&cobra.Command{})

	return rootCmd
}
