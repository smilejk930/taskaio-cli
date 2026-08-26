package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/taskaio/taskaio-cli/internal/apiclient"
	"github.com/taskaio/taskaio-cli/internal/config"
	"github.com/taskaio/taskaio-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication and tokens",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check current authentication status and user information",
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}

		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		authMe, err := client.GetAuthMe(context.Background())
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintAuthMe(os.Stdout, cfg.Output, authMe); err != nil {
			exitWithError(err)
		}
	},
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a Personal Access Token",
	Run: func(cmd *cobra.Command, args []string) {
		_, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}

		tokenStdin, _ := cmd.Flags().GetBool("token-stdin")
		flagToken, _ := cmd.Flags().GetString("token")

		var token string
		if tokenStdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				exitWithError(fmt.Errorf("failed to read token from stdin: %w", err))
			}
			token = strings.TrimSpace(string(data))
		} else if flagToken != "" {
			token = flagToken
		} else {
			fmt.Fprint(os.Stderr, "Enter Personal Access Token: ")
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				exitWithError(fmt.Errorf("failed to read token: %w", err))
			}
			token = strings.TrimSpace(line)
		}

		if token == "" {
			exitWithError(&apiclient.APIError{
				StatusCode: 401,
				Code:       "UNAUTHORIZED",
				Message:    "Token cannot be empty.",
			})
		}

		// Verify token with server
		testClient := apiclient.NewClient(cfg.BaseURL, token, cfg.Timeout)
		authMe, err := testClient.GetAuthMe(context.Background())
		if err != nil {
			exitWithError(fmt.Errorf("authentication failed: %w", err))
		}

		// Save token to config
		cfg.Token = token
		if err := config.SaveConfig(cfg, cfg.ConfigPath); err != nil {
			exitWithError(fmt.Errorf("failed to save config: %w", err))
		}

		output.PrintMessage(os.Stderr, fmt.Sprintf("Successfully logged in as %s (User ID: %s)", authMe.User.Email, authMe.User.ID))
	},
}

func init() {
	authLoginCmd.Flags().Bool("token-stdin", false, "Read token from standard input")
	authLoginCmd.Flags().String("token", "", "Personal Access Token value")

	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLoginCmd)
}
