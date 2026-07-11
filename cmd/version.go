package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the muxic version string, injected at build time by goreleaser
// via ldflags (-X muxic/cmd.Version={{.Tag}}). The default "dev" value is used
// for developer builds (`go build` without ldflags).
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the muxic version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
