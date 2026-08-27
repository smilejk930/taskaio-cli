package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/taskaio/taskaio-cli/internal/apiclient"
	"github.com/taskaio/taskaio-cli/internal/input"
	"github.com/taskaio/taskaio-cli/internal/output"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects and view members",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		search, _ := cmd.Flags().GetString("search")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		fetchAll, _ := cmd.Flags().GetBool("all")

		var allProjects []apiclient.Project
		var meta apiclient.Meta
		currentCursor := cursor

		for {
			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}
			if limit > 0 {
				params.Set("limit", strconv.Itoa(limit))
			}
			if currentCursor != "" {
				params.Set("cursor", currentCursor)
			}

			resp, err := client.ListProjects(context.Background(), params)
			if err != nil {
				exitWithError(err)
			}

			allProjects = append(allProjects, resp.Data...)
			meta = resp.Meta

			if !fetchAll || !resp.Meta.HasMore || resp.Meta.NextCursor == nil || *resp.Meta.NextCursor == "" {
				break
			}
			currentCursor = *resp.Meta.NextCursor
		}
		if fetchAll {
			meta.HasMore = false
			meta.NextCursor = nil
			if meta.Total == nil {
				total := len(allProjects)
				meta.Total = &total
			}
		}

		if err := output.PrintProjectList(os.Stdout, cfg.Output, &apiclient.ListResponse[apiclient.Project]{Data: allProjects, Meta: meta}); err != nil {
			exitWithError(err)
		}
	},
}

var projectsGetCmd = &cobra.Command{
	Use:   "get <projectId>",
	Short: "Get project details by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		projectID := args[0]
		project, err := client.GetProject(context.Background(), projectID)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintProject(os.Stdout, cfg.Output, project); err != nil {
			exitWithError(err)
		}
	},
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		inputSrc, _ := cmd.Flags().GetString("input")
		name, _ := cmd.Flags().GetString("name")
		desc, _ := cmd.Flags().GetString("description")

		var payload apiclient.CreateProjectInput
		if err := validateInputMode(cmd, inputSrc, "name", "description"); err != nil {
			exitWithError(err)
		}

		if inputSrc != "" {
			if err := input.ReadInput(inputSrc, &payload); err != nil {
				exitWithError(err)
			}
		}

		if name != "" {
			payload.Name = name
		}
		if desc != "" {
			payload.Description = &desc
		}

		if payload.Name == "" {
			exitWithError(fmt.Errorf("project name is required (use --name or --input)"))
		}

		created, err := client.CreateProject(context.Background(), payload)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintProject(os.Stdout, cfg.Output, created); err != nil {
			exitWithError(err)
		}
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update <projectId>",
	Short: "Update project information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		projectID := args[0]
		inputSrc, _ := cmd.Flags().GetString("input")
		name, _ := cmd.Flags().GetString("name")
		desc, _ := cmd.Flags().GetString("description")

		var payload apiclient.UpdateProjectInput
		if err := validateInputMode(cmd, inputSrc, "name", "description"); err != nil {
			exitWithError(err)
		}

		if inputSrc != "" {
			if err := input.ReadInput(inputSrc, &payload); err != nil {
				exitWithError(err)
			}
		}

		if cmd.Flags().Changed("name") {
			payload.Name = &name
		}
		if cmd.Flags().Changed("description") {
			payload.Description = &desc
		}

		updated, err := client.UpdateProject(context.Background(), projectID, payload)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintProject(os.Stdout, cfg.Output, updated); err != nil {
			exitWithError(err)
		}
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:   "delete <projectId>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		projectID := args[0]
		force, _ := cmd.Flags().GetBool("yes")

		confirmed, err := confirmDeletion(force, fmt.Sprintf("Are you sure you want to delete project %s?", projectID))
		if err != nil {
			exitWithError(err)
		}
		if !confirmed {
			output.PrintMessage(os.Stderr, "Aborted.")
			return
		}

		if err := client.DeleteProject(context.Background(), projectID); err != nil {
			exitWithError(err)
		}

		output.PrintMessage(os.Stderr, fmt.Sprintf("Project %s deleted successfully.", projectID))
	},
}

var projectsMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage project members",
}

var projectsMembersListCmd = &cobra.Command{
	Use:   "list <projectId>",
	Short: "List project members",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		projectID := args[0]
		search, _ := cmd.Flags().GetString("search")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		fetchAll, _ := cmd.Flags().GetBool("all")

		var allMembers []apiclient.ProjectMember
		var meta apiclient.Meta
		currentCursor := cursor

		for {
			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}
			if limit > 0 {
				params.Set("limit", strconv.Itoa(limit))
			}
			if currentCursor != "" {
				params.Set("cursor", currentCursor)
			}

			resp, err := client.ListProjectMembers(context.Background(), projectID, params)
			if err != nil {
				exitWithError(err)
			}

			allMembers = append(allMembers, resp.Data...)
			meta = resp.Meta

			if !fetchAll || !resp.Meta.HasMore || resp.Meta.NextCursor == nil || *resp.Meta.NextCursor == "" {
				break
			}
			currentCursor = *resp.Meta.NextCursor
		}
		if fetchAll {
			meta.HasMore = false
			meta.NextCursor = nil
			if meta.Total == nil {
				total := len(allMembers)
				meta.Total = &total
			}
		}

		if err := output.PrintMemberList(os.Stdout, cfg.Output, &apiclient.ListResponse[apiclient.ProjectMember]{Data: allMembers, Meta: meta}); err != nil {
			exitWithError(err)
		}
	},
}

func init() {
	projectsListCmd.Flags().String("search", "", "Search query for project name or description")
	projectsListCmd.Flags().Int("limit", 50, "Number of items per page")
	projectsListCmd.Flags().String("cursor", "", "Cursor for pagination")
	projectsListCmd.Flags().Bool("all", false, "Fetch all pages")

	projectsCreateCmd.Flags().String("name", "", "Project name")
	projectsCreateCmd.Flags().String("description", "", "Project description")
	projectsCreateCmd.Flags().String("input", "", "JSON input file path or '-' for stdin")

	projectsUpdateCmd.Flags().String("name", "", "Project name")
	projectsUpdateCmd.Flags().String("description", "", "Project description")
	projectsUpdateCmd.Flags().String("input", "", "JSON input file path or '-' for stdin")

	projectsDeleteCmd.Flags().BoolP("yes", "y", false, "Skip deletion confirmation prompt")

	projectsMembersListCmd.Flags().String("search", "", "Search query for member display name or email")
	projectsMembersListCmd.Flags().Int("limit", 50, "Number of items per page")
	projectsMembersListCmd.Flags().String("cursor", "", "Cursor for pagination")
	projectsMembersListCmd.Flags().Bool("all", false, "Fetch all pages")

	projectsMembersCmd.AddCommand(projectsMembersListCmd)

	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
	projectsCmd.AddCommand(projectsCreateCmd)
	projectsCmd.AddCommand(projectsUpdateCmd)
	projectsCmd.AddCommand(projectsDeleteCmd)
	projectsCmd.AddCommand(projectsMembersCmd)
}
