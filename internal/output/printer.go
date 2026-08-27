package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/taskaio/taskaio-cli/internal/apiclient"
)

// PrintJSON formats and prints data as indented JSON to the provided writer.
func PrintJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// PrintAuthMe prints auth status either in table or JSON.
func PrintAuthMe(w io.Writer, format string, data *apiclient.AuthMeResponse) error {
	if format != "table" {
		return PrintJSON(w, data)
	}

	headers := []string{"USER ID", "EMAIL", "ADMIN", "TOKEN ID", "TOKEN NAME"}
	adminStr := "false"
	if data.User.IsAdmin {
		adminStr = "true"
	}
	rows := [][]string{
		{
			data.User.ID,
			data.User.Email,
			adminStr,
			data.Token.ID,
			data.Token.Name,
		},
	}
	RenderTable(w, headers, rows)
	return nil
}

// PrintProjects prints project list either in table or JSON.
func PrintProjects(w io.Writer, format string, projects []apiclient.Project) error {
	if format != "table" {
		return PrintJSON(w, projects)
	}

	headers := []string{"ID", "NAME", "ROLE", "DESCRIPTION", "CREATED AT"}
	var rows [][]string
	for _, p := range projects {
		role := "-"
		if p.Role != nil {
			role = *p.Role
		}
		desc := "-"
		if p.Description != nil {
			desc = *p.Description
		}
		rows = append(rows, []string{
			p.ID,
			p.Name,
			role,
			desc,
			p.CreatedAt,
		})
	}
	RenderTable(w, headers, rows)
	return nil
}

// PrintProjectList preserves pagination metadata in JSON output.
func PrintProjectList(w io.Writer, format string, response *apiclient.ListResponse[apiclient.Project]) error {
	if format != "table" {
		return PrintJSON(w, response)
	}
	return PrintProjects(w, format, response.Data)
}

// PrintProject prints a single project either in table or JSON.
func PrintProject(w io.Writer, format string, p *apiclient.Project) error {
	if format != "table" {
		return PrintJSON(w, p)
	}
	return PrintProjects(w, format, []apiclient.Project{*p})
}

// PrintMembers prints project members.
func PrintMembers(w io.Writer, format string, members []apiclient.ProjectMember) error {
	if format != "table" {
		return PrintJSON(w, members)
	}

	headers := []string{"USER ID", "NAME / EMAIL", "ROLE"}
	var rows [][]string
	for _, m := range members {
		name := "-"
		if m.DisplayName != nil && *m.DisplayName != "" {
			name = *m.DisplayName
		} else if m.Email != nil {
			name = *m.Email
		}
		rows = append(rows, []string{
			m.UserID,
			name,
			m.Role,
		})
	}
	RenderTable(w, headers, rows)
	return nil
}

// PrintMemberList preserves pagination metadata in JSON output.
func PrintMemberList(w io.Writer, format string, response *apiclient.ListResponse[apiclient.ProjectMember]) error {
	if format != "table" {
		return PrintJSON(w, response)
	}
	return PrintMembers(w, format, response.Data)
}

// PrintTasks prints task list.
func PrintTasks(w io.Writer, format string, tasks []apiclient.Task) error {
	if format != "table" {
		return PrintJSON(w, tasks)
	}

	headers := []string{"ID", "TITLE", "STATUS", "PRIORITY", "PROGRESS", "ASSIGNEE", "START", "END"}
	var rows [][]string
	for _, t := range tasks {
		assignee := "-"
		if t.AssigneeID != nil {
			assignee = *t.AssigneeID
		}
		start := "-"
		if t.StartDate != nil {
			start = *t.StartDate
		}
		end := "-"
		if t.EndDate != nil {
			end = *t.EndDate
		}

		rows = append(rows, []string{
			t.ID,
			t.Title,
			t.Status,
			t.Priority,
			strconv.Itoa(t.Progress) + "%",
			assignee,
			start,
			end,
		})
	}
	RenderTable(w, headers, rows)
	return nil
}

// PrintTaskList preserves pagination metadata in JSON output.
func PrintTaskList(w io.Writer, format string, response *apiclient.ListResponse[apiclient.Task]) error {
	if format != "table" {
		return PrintJSON(w, response)
	}
	return PrintTasks(w, format, response.Data)
}

// PrintTask prints a single task.
func PrintTask(w io.Writer, format string, t *apiclient.Task) error {
	if format != "table" {
		return PrintJSON(w, t)
	}
	return PrintTasks(w, format, []apiclient.Task{*t})
}

// PrintScheduleList preserves pagination metadata in JSON output.
func PrintScheduleList(w io.Writer, format string, response *apiclient.ListResponse[apiclient.Schedule]) error {
	if format != "table" {
		return PrintJSON(w, response)
	}
	return PrintSchedules(w, format, response.Data)
}

// PrintSchedules prints schedule list.
func PrintSchedules(w io.Writer, format string, schedules []apiclient.Schedule) error {
	if format != "table" {
		return PrintJSON(w, schedules)
	}

	headers := []string{"ID", "NAME", "TYPE", "MEMBER ID", "START DATE", "END DATE"}
	var rows [][]string
	for _, s := range schedules {
		member := "-"
		if s.MemberID != nil {
			member = *s.MemberID
		}
		rows = append(rows, []string{
			s.ID,
			s.Name,
			s.Type,
			member,
			s.StartDate,
			s.EndDate,
		})
	}
	RenderTable(w, headers, rows)
	return nil
}

// PrintSchedule prints a single schedule.
func PrintSchedule(w io.Writer, format string, s *apiclient.Schedule) error {
	if format != "table" {
		return PrintJSON(w, s)
	}
	return PrintSchedules(w, format, []apiclient.Schedule{*s})
}

// PrintMessage prints a diagnostic message or success notice.
func PrintMessage(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}
