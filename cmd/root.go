package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "taskaio",
	Short:        "taskAIO CLI - Command line interface for taskAIO project and task management",
	Long:         "taskAIO CLI allows you to manage projects, tasks, schedules, and view members from your terminal.",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "Config file path (default is $XDG_CONFIG_HOME/taskaio/config.yaml or ~/.config/taskaio/config.yaml)")
	rootCmd.PersistentFlags().String("base-url", "", "taskAIO server base URL (default: http://localhost:3000)")
	rootCmd.PersistentFlags().String("token", "", "Personal access token for authentication")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format (json, table; default: config or json)")
	rootCmd.PersistentFlags().Duration("timeout", 0, "HTTP request timeout duration (e.g. 30s)")

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(schedulesCmd)
	rootCmd.AddCommand(versionCmd)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
