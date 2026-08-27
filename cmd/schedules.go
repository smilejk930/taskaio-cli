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

var schedulesCmd = &cobra.Command{
	Use:   "schedules",
	Short: "Manage schedules",
}

var schedulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List schedules",
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		search, _ := cmd.Flags().GetString("search")
		sType, _ := cmd.Flags().GetString("type")
		user, _ := cmd.Flags().GetString("user")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		fetchAll, _ := cmd.Flags().GetBool("all")

		var allSchedules []apiclient.Schedule
		var meta apiclient.Meta
		currentCursor := cursor

		for {
			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}
			if sType != "" {
				params.Set("type", sType)
			}
			if user != "" {
				params.Set("userId", user)
			}
			if from != "" {
				params.Set("from", from)
			}
			if to != "" {
				params.Set("to", to)
			}
			if limit > 0 {
				params.Set("limit", strconv.Itoa(limit))
			}
			if currentCursor != "" {
				params.Set("cursor", currentCursor)
			}

			resp, err := client.ListSchedules(context.Background(), params)
			if err != nil {
				exitWithError(err)
			}

			allSchedules = append(allSchedules, resp.Data...)
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
				total := len(allSchedules)
				meta.Total = &total
			}
		}

		if err := output.PrintScheduleList(os.Stdout, cfg.Output, &apiclient.ListResponse[apiclient.Schedule]{Data: allSchedules, Meta: meta}); err != nil {
			exitWithError(err)
		}
	},
}

var schedulesGetCmd = &cobra.Command{
	Use:   "get <scheduleId>",
	Short: "Get schedule details by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		scheduleID := args[0]
		schedule, err := client.GetSchedule(context.Background(), scheduleID)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintSchedule(os.Stdout, cfg.Output, schedule); err != nil {
			exitWithError(err)
		}
	},
}

var schedulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new schedule",
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
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		sType, _ := cmd.Flags().GetString("type")
		user, _ := cmd.Flags().GetString("user")
		note, _ := cmd.Flags().GetString("note")

		var payload apiclient.CreateScheduleInput
		if err := validateInputMode(cmd, inputSrc, "name", "start-date", "end-date", "type", "user", "note"); err != nil {
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
		if startDate != "" {
			payload.StartDate = startDate
		}
		if endDate != "" {
			payload.EndDate = endDate
		}
		if sType != "" {
			payload.Type = sType
		}
		if user != "" {
			payload.MemberID = &user
		}
		if note != "" {
			payload.Note = &note
		}

		if payload.Name == "" || payload.StartDate == "" || payload.EndDate == "" || payload.Type == "" {
			exitWithError(fmt.Errorf("name, start-date, end-date, and type are required (use flags or --input)"))
		}

		created, err := client.CreateSchedule(context.Background(), payload)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintSchedule(os.Stdout, cfg.Output, created); err != nil {
			exitWithError(err)
		}
	},
}

var schedulesUpdateCmd = &cobra.Command{
	Use:   "update <scheduleId>",
	Short: "Update schedule information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		scheduleID := args[0]
		inputSrc, _ := cmd.Flags().GetString("input")
		name, _ := cmd.Flags().GetString("name")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		sType, _ := cmd.Flags().GetString("type")
		user, _ := cmd.Flags().GetString("user")
		note, _ := cmd.Flags().GetString("note")

		var payload apiclient.UpdateScheduleInput
		if err := validateInputMode(cmd, inputSrc, "name", "start-date", "end-date", "type", "user", "note"); err != nil {
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
		if cmd.Flags().Changed("start-date") {
			payload.StartDate = &startDate
		}
		if cmd.Flags().Changed("end-date") {
			payload.EndDate = &endDate
		}
		if cmd.Flags().Changed("type") {
			payload.Type = &sType
		}
		if cmd.Flags().Changed("user") {
			payload.MemberID = &user
		}
		if cmd.Flags().Changed("note") {
			payload.Note = &note
		}

		updated, err := client.UpdateSchedule(context.Background(), scheduleID, payload)
		if err != nil {
			exitWithError(err)
		}

		if err := output.PrintSchedule(os.Stdout, cfg.Output, updated); err != nil {
			exitWithError(err)
		}
	},
}

var schedulesDeleteCmd = &cobra.Command{
	Use:   "delete <scheduleId>",
	Short: "Delete a schedule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, cfg, err := getClient(cmd)
		if err != nil {
			exitWithError(err)
		}
		if err := requireAuth(cfg); err != nil {
			exitWithError(err)
		}

		scheduleID := args[0]
		force, _ := cmd.Flags().GetBool("yes")

		confirmed, err := confirmDeletion(force, fmt.Sprintf("Are you sure you want to delete schedule %s?", scheduleID))
		if err != nil {
			exitWithError(err)
		}
		if !confirmed {
			output.PrintMessage(os.Stderr, "Aborted.")
			return
		}

		if err := client.DeleteSchedule(context.Background(), scheduleID); err != nil {
			exitWithError(err)
		}

		output.PrintMessage(os.Stderr, fmt.Sprintf("Schedule %s deleted successfully.", scheduleID))
	},
}

func init() {
	schedulesListCmd.Flags().String("search", "", "Search query for schedule name or note")
	schedulesListCmd.Flags().String("type", "", "Filter by schedule type (member_leave, business_trip, public_holiday, workshop, supervision, other)")
	schedulesListCmd.Flags().String("user", "", "Filter by user ID (memberId)")
	schedulesListCmd.Flags().String("from", "", "Start date filter (YYYY-MM-DD)")
	schedulesListCmd.Flags().String("to", "", "End date filter (YYYY-MM-DD)")
	schedulesListCmd.Flags().Int("limit", 50, "Number of items per page")
	schedulesListCmd.Flags().String("cursor", "", "Cursor for pagination")
	schedulesListCmd.Flags().Bool("all", false, "Fetch all pages")

	schedulesCreateCmd.Flags().String("name", "", "Schedule name")
	schedulesCreateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	schedulesCreateCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	schedulesCreateCmd.Flags().String("type", "member_leave", "Schedule type")
	schedulesCreateCmd.Flags().String("user", "", "Member user ID (optional)")
	schedulesCreateCmd.Flags().String("note", "", "Schedule note or description")
	schedulesCreateCmd.Flags().String("input", "", "JSON input file path or '-' for stdin")

	schedulesUpdateCmd.Flags().String("name", "", "Schedule name")
	schedulesUpdateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	schedulesUpdateCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	schedulesUpdateCmd.Flags().String("type", "", "Schedule type")
	schedulesUpdateCmd.Flags().String("user", "", "Member user ID")
	schedulesUpdateCmd.Flags().String("note", "", "Schedule note or description")
	schedulesUpdateCmd.Flags().String("input", "", "JSON input file path or '-' for stdin")

	schedulesDeleteCmd.Flags().BoolP("yes", "y", false, "Skip deletion confirmation prompt")

	schedulesCmd.AddCommand(schedulesListCmd)
	schedulesCmd.AddCommand(schedulesGetCmd)
	schedulesCmd.AddCommand(schedulesCreateCmd)
	schedulesCmd.AddCommand(schedulesUpdateCmd)
	schedulesCmd.AddCommand(schedulesDeleteCmd)
}
