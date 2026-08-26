package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the taskaio CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("taskaio-cli version %s (commit: %s, built at: %s)\n", Version, Commit, Date)
	},
}

func init() {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("taskaio-cli version {{.Version}}\n")
}
