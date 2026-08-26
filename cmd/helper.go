package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/taskaio/taskaio-cli/internal/apiclient"
	"github.com/taskaio/taskaio-cli/internal/config"
)

func getLoadedConfig(cmd *cobra.Command) (*config.Config, error) {
	configPath, _ := cmd.Flags().GetString("config")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	outputFormat, _ := cmd.Flags().GetString("output")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	cfg, err := config.LoadConfig(configPath, baseURL, token, outputFormat, timeout)
	if err != nil {
		return nil, &apiclient.ConfigError{Err: err}
	}
	return cfg, nil
}

func getClient(cmd *cobra.Command) (*apiclient.Client, *config.Config, error) {
	cfg, err := getLoadedConfig(cmd)
	if err != nil {
		return nil, nil, err
	}

	client := apiclient.NewClient(cfg.BaseURL, cfg.Token, cfg.Timeout)
	return client, cfg, nil
}

func requireAuth(cfg *config.Config) error {
	if cfg.Token == "" {
		return &apiclient.APIError{
			StatusCode: 401,
			Code:       "UNAUTHORIZED",
			Message:    "Authentication token is missing. Please run 'taskaio auth login' or set TASKAIO_TOKEN.",
		}
	}
	return nil
}

func validateInputMode(cmd *cobra.Command, inputSrc string, bodyFlags ...string) error {
	if inputSrc == "" {
		return nil
	}
	for _, name := range bodyFlags {
		if cmd.Flags().Changed(name) {
			return &apiclient.UsageError{Err: fmt.Errorf("--input cannot be combined with --%s", name)}
		}
	}
	return nil
}

func confirmDeletion(force bool, prompt string) (bool, error) {
	if force {
		return true, nil
	}

	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func exitWithError(err error) {
	if err == nil {
		os.Exit(apiclient.ExitCodeSuccess)
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(apiclient.GetExitCodeForError(err))
}
