package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/taskaio/taskaio-cli/internal/config"
	"github.com/taskaio/taskaio-cli/internal/output"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")
		force, _ := cmd.Flags().GetBool("force")

		if configPath == "" {
			var err error
			configPath, err = config.DefaultConfigPath()
			if err != nil {
				exitWithError(err)
			}
		}

		if _, err := os.Stat(configPath); err == nil && !force {
			output.PrintMessage(os.Stderr, fmt.Sprintf("Config file already exists at %s (use --force to overwrite)", configPath))
			return
		}

		cfg := &config.Config{
			BaseURL: config.DefaultBaseURL,
			Output:  config.DefaultOutput,
			Timeout: config.DefaultTimeout,
		}

		if err := config.SaveConfig(cfg, configPath); err != nil {
			exitWithError(err)
		}

		output.PrintMessage(os.Stderr, fmt.Sprintf("Initialized configuration file at %s with 0600 permissions", configPath))
	},
}

func init() {
	configInitCmd.Flags().Bool("force", false, "Overwrite existing config file if present")
	configCmd.AddCommand(configInitCmd)
}
