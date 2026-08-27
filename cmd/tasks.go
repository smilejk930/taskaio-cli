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

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
}

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks for a project",
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		projectID, _ := cmd.Flags().GetString("project")
		if projectID == "" {
			exitWithError(fmt.Errorf("project ID is required (--project)"))
		}

		search, _ := cmd.Flags().GetString("search")
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		assignee, _ := cmd.Flags().GetString("assignee")
		parent, _ := cmd.Flags().GetString("parent")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		fetchAll, _ := cmd.Flags().GetBool("all")

		var allTasks []apiclient.Task
		var meta apiclient.Meta
		currentCursor := cursor

		for {
			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}
			if status != "" {
				params.Set("status", status)
			}
			if priority != "" {
				params.Set("priority", priority)
			}
			if assignee != "" {
				params.Set("assigneeId", assignee)
			}
			if parent != "" {
				params.Set("parentId", parent)
			}
			if limit > 0 {
				params.Set("limit", strconv.Itoa(limit))
			}
			if currentCursor != "" {
				params.Set("cursor", currentCursor)
			}

			resp, err := client.ListTasks(context.Background(), projectID, params)
			if err != nil {
				exitWithError(err)
			}

			allTasks = append(allTasks, resp.Data...)
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
				total := len(allTasks)
				meta.Total = &total
			}
		}

		if err := output.PrintTaskList(os.Stdout, cfg.Output, &apiclient.ListResponse[apiclient.Task]{Data: allTasks, Meta: meta}); err != nil {
			exitWithError(err)
		}
	},
}

var tasksGetCmd = &cobra.Command{
	Use:   "get <taskId>",
	Short: "Get task details by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		taskID := args[0]
		task, err := client.GetTask(context.Background(), taskID)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintTask(os.Stdout, cfg.Output, task); err != nil {
			exitWithError(err)
		}
	},
}

var tasksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task in a project",
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		projectID, _ := cmd.Flags().GetString("project")
		inputSrc, _ := cmd.Flags().GetString("input")
		title, _ := cmd.Flags().GetString("title")
		desc, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		assignee, _ := cmd.Flags().GetString("assignee")
		parent, _ := cmd.Flags().GetString("parent")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		progress, _ := cmd.Flags().GetInt("progress")
		color, _ := cmd.Flags().GetString("color")

		var payload apiclient.CreateTaskInput
		if err := validateInputMode(cmd, inputSrc, "title", "description", "status", "priority", "assignee", "parent", "start-date", "end-date", "progress", "color"); err != nil {
			exitWithError(err)
		}

		if inputSrc != "" {
			if err := input.ReadInput(inputSrc, &payload); err != nil {
				exitWithError(err)
			}
		}

		if title != "" {
			payload.Title = title
		}
		if desc != "" {
			payload.Description = &desc
		}
		if status != "" {
			payload.Status = &status
		}
		if priority != "" {
			payload.Priority = &priority
		}
		if assignee != "" {
			payload.AssigneeID = &assignee
		}
		if parent != "" {
			payload.ParentID = &parent
		}
		if startDate != "" {
			payload.StartDate = &startDate
		}
		if endDate != "" {
			payload.EndDate = &endDate
		}
		if cmd.Flags().Changed("progress") {
			payload.Progress = &progress
		}
		if color != "" {
			payload.Color = &color
		}

		if projectID == "" {
			exitWithError(fmt.Errorf("project ID is required (--project)"))
		}
		if payload.Title == "" {
			exitWithError(fmt.Errorf("task title is required (use --title or --input)"))
		}

		created, err := client.CreateTask(context.Background(), projectID, payload)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintTask(os.Stdout, cfg.Output, created); err != nil {
			exitWithError(err)
		}
	},
}

var tasksUpdateCmd = &cobra.Command{
	Use:   "update <taskId>",
	Short: "Update task information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		taskID := args[0]
		inputSrc, _ := cmd.Flags().GetString("input")
		title, _ := cmd.Flags().GetString("title")
		desc, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		assignee, _ := cmd.Flags().GetString("assignee")
		parent, _ := cmd.Flags().GetString("parent")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		progress, _ := cmd.Flags().GetInt("progress")
		color, _ := cmd.Flags().GetString("color")

		var payload apiclient.UpdateTaskInput
		if err := validateInputMode(cmd, inputSrc, "title", "description", "status", "priority", "assignee", "parent", "start-date", "end-date", "progress", "color"); err != nil {
			exitWithError(err)
		}

		if inputSrc != "" {
			if err := input.ReadInput(inputSrc, &payload); err != nil {
				exitWithError(err)
			}
		}

		if cmd.Flags().Changed("title") {
			payload.Title = &title
		}
		if cmd.Flags().Changed("description") {
			payload.Description = &desc
		}
		if cmd.Flags().Changed("status") {
			payload.Status = &status
		}
		if cmd.Flags().Changed("priority") {
			payload.Priority = &priority
		}
		if cmd.Flags().Changed("assignee") {
			payload.AssigneeID = &assignee
		}
		if cmd.Flags().Changed("parent") {
			payload.ParentID = &parent
		}
		if cmd.Flags().Changed("start-date") {
			payload.StartDate = &startDate
		}
		if cmd.Flags().Changed("end-date") {
			payload.EndDate = &endDate
		}
		if cmd.Flags().Changed("progress") {
			payload.Progress = &progress
		}
		if cmd.Flags().Changed("color") {
			payload.Color = &color
		}

		updated, err := client.UpdateTask(context.Background(), taskID, payload)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintTask(os.Stdout, cfg.Output, updated); err != nil {
			exitWithError(err)
		}
	},
}

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete <taskId>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		taskID := args[0]
		force, _ := cmd.Flags().GetBool("yes")

		confirmed, err := confirmDeletion(force, fmt.Sprintf("Are you sure you want to delete task %s?", taskID))
		if err != nil {
			exitWithError(err)
		}
		if !confirmed {
			output.PrintMessage(os.Stderr, "Aborted.")
			return
		}

		if err := client.DeleteTask(context.Background(), taskID); err != nil {
			exitWithError(err)
		}

		output.PrintMessage(os.Stderr, fmt.Sprintf("Task %s deleted successfully.", taskID))
	},
}

func init() {
	tasksListCmd.Flags().String("project", "", "Project ID (required)")
	tasksListCmd.Flags().String("search", "", "Search query for task title or description")
	tasksListCmd.Flags().String("status", "", "Filter by task status (todo, in_progress, review, done)")
	tasksListCmd.Flags().String("priority", "", "Filter by task priority (low, medium, high, urgent)")
	tasksListCmd.Flags().String("assignee", "", "Filter by assignee user ID")
	tasksListCmd.Flags().String("parent", "", "Filter by parent task ID (use 'root' or 'null' for root tasks)")
	tasksListCmd.Flags().Int("limit", 50, "Number of items per page")
	tasksListCmd.Flags().String("cursor", "", "Cursor for pagination")
	tasksListCmd.Flags().Bool("all", false, "Fetch all pages")

	tasksCreateCmd.Flags().String("project", "", "Project ID (required)")
	tasksCreateCmd.Flags().String("title", "", "Task title")
	tasksCreateCmd.Flags().String("description", "", "Task description")
	tasksCreateCmd.Flags().String("status", "todo", "Task status")
	tasksCreateCmd.Flags().String("priority", "medium", "Task priority")
	tasksCreateCmd.Flags().String("assignee", "", "Assignee user ID")
	tasksCreateCmd.Flags().String("parent", "", "Parent task ID")
	tasksCreateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	tasksCreateCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	tasksCreateCmd.Flags().Int("progress", 0, "Progress percentage (0-100)")
	tasksCreateCmd.Flags().String("color", "", "Hex color code")
	tasksCreateCmd.Flags().String("input", "", "JSON input file path or '-' for stdin")

	tasksUpdateCmd.Flags().String("title", "", "Task title")
	tasksUpdateCmd.Flags().String("description", "", "Task description")
	tasksUpdateCmd.Flags().String("status", "", "Task status")
	tasksUpdateCmd.Flags().String("priority", "", "Task priority")
	tasksUpdateCmd.Flags().String("assignee", "", "Assignee user ID")
	tasksUpdateCmd.Flags().String("parent", "", "Parent task ID")
	tasksUpdateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	tasksUpdateCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	tasksUpdateCmd.Flags().Int("progress", 0, "Progress percentage (0-100)")
	tasksUpdateCmd.Flags().String("color", "", "Hex color code")
	tasksUpdateCmd.Flags().String("input", "", "JSON input file path or '-' for stdin")

	tasksDeleteCmd.Flags().BoolP("yes", "y", false, "Skip deletion confirmation prompt")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksGetCmd)
	tasksCmd.AddCommand(tasksCreateCmd)
	tasksCmd.AddCommand(tasksUpdateCmd)
	tasksCmd.AddCommand(tasksDeleteCmd)
}
