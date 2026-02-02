package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information - set via ldflags at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Displays the version, build time, and git commit of vibe.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vibe version %s\n", Version)
		if GitCommit != "unknown" {
			fmt.Printf("  commit: %s\n", GitCommit)
		}
		if BuildTime != "unknown" {
			fmt.Printf("  built:  %s\n", BuildTime)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
