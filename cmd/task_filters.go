package cmd

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var taskFilterNow = time.Now

var taskStatusAliases = map[string]string{
	"todo":        "todo",
	"in_progress": "in_progress",
	"review":      "review",
	"done":        "done",
	"할일":          "todo",
	"진행중":         "in_progress",
	"리뷰":          "review",
	"완료":          "done",
}

var taskPriorityAliases = map[string]string{
	"urgent": "urgent",
	"high":   "high",
	"medium": "medium",
	"low":    "low",
	"긴급":     "urgent",
	"높음":     "high",
	"보통":     "medium",
	"낮음":     "low",
}

var taskWeekAliases = map[string]int{
	"last": -1,
	"this": 0,
	"next": 1,
	"지난주":  -1,
	"이번주":  0,
	"다음주":  1,
}

func addTaskFilterFlags(command *cobra.Command, includeSpecial bool) {
	command.Flags().String("search", "", "Search query for task title or description")
	command.Flags().StringSlice("status", nil, "Filter by status; repeat or comma-separate (할일, 진행중, 리뷰, 완료 or todo, in_progress, review, done)")
	command.Flags().StringSlice("priority", nil, "Filter by priority; repeat or comma-separate (긴급, 높음, 보통, 낮음 or urgent, high, medium, low)")
	command.Flags().String("assignee", "", "Filter by assignee user ID")
	command.Flags().StringSlice("week", nil, "Filter by task period; repeat or comma-separate (지난주, 이번주, 다음주 or last, this, next)")
	command.Flags().String("from", "", "Period start date (YYYY-MM-DD)")
	command.Flags().String("to", "", "Period end date (YYYY-MM-DD)")
	command.Flags().String("as-of", "", "Reference date for due counts and due filters (YYYY-MM-DD)")
	if includeSpecial {
		command.Flags().String("parent", "", "Filter by parent task ID (use 'root' or 'null' for root tasks)")
		command.Flags().Bool("management-only", false, "Show only directly matching 1-depth management tasks")
		command.Flags().Bool("due-soon", false, "Show incomplete tasks due today through the next 3 days")
		command.Flags().Bool("overdue", false, "Show incomplete tasks whose due date has passed")
	}
}

func buildTaskFilterParams(command *cobra.Command, includeSpecial bool) (url.Values, error) {
	params := url.Values{}
	search, _ := command.Flags().GetString("search")
	statuses, _ := command.Flags().GetStringSlice("status")
	priorities, _ := command.Flags().GetStringSlice("priority")
	assignee, _ := command.Flags().GetString("assignee")
	weeks, _ := command.Flags().GetStringSlice("week")
	from, _ := command.Flags().GetString("from")
	to, _ := command.Flags().GetString("to")
	asOf, _ := command.Flags().GetString("as-of")
	if len(weeks) > 0 && (from != "" || to != "") {
		return nil, fmt.Errorf("--week cannot be combined with --from or --to")
	}
	for name, value := range map[string]string{"from": from, "to": to, "as-of": asOf} {
		if value == "" {
			continue
		}
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil || parsed.Format("2006-01-02") != value {
			return nil, fmt.Errorf("invalid --%s date: %q (expected YYYY-MM-DD)", name, value)
		}
	}
	if from != "" && to != "" && from > to {
		return nil, fmt.Errorf("--to must be on or after --from")
	}

	if search != "" {
		params.Set("search", search)
	}
	if assignee != "" {
		params.Set("assigneeId", assignee)
	}

	normalizedStatuses, err := normalizeTaskFilterValues("status", statuses, taskStatusAliases)
	if err != nil {
		return nil, err
	}
	if len(normalizedStatuses) > 0 {
		params.Set("status", strings.Join(normalizedStatuses, ","))
	}

	normalizedPriorities, err := normalizeTaskFilterValues("priority", priorities, taskPriorityAliases)
	if err != nil {
		return nil, err
	}
	if len(normalizedPriorities) > 0 {
		params.Set("priority", strings.Join(normalizedPriorities, ","))
	}

	if len(weeks) > 0 {
		weekFrom, weekTo, err := taskWeekRange(weeks, taskFilterNow())
		if err != nil {
			return nil, err
		}
		from, to = weekFrom, weekTo
	}
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	if asOf == "" {
		asOf = taskFilterNow().Format("2006-01-02")
	}

	if !includeSpecial {
		params.Set("asOf", asOf)
		return params, nil
	}

	parent, _ := command.Flags().GetString("parent")
	managementOnly, _ := command.Flags().GetBool("management-only")
	dueSoon, _ := command.Flags().GetBool("due-soon")
	overdue, _ := command.Flags().GetBool("overdue")
	if managementOnly && parent != "" {
		return nil, fmt.Errorf("--management-only cannot be used with --parent")
	}
	if dueSoon && overdue {
		return nil, fmt.Errorf("--due-soon cannot be used with --overdue")
	}
	if managementOnly {
		params.Set("parentId", "root")
	} else if parent != "" {
		params.Set("parentId", parent)
	}
	if dueSoon || overdue {
		if dueSoon {
			params.Set("due", "soon")
		} else {
			params.Set("due", "overdue")
		}
		params.Set("asOf", asOf)
	} else if command.Flags().Changed("as-of") {
		params.Set("asOf", asOf)
	}

	return params, nil
}

func normalizeTaskFilterValues(name string, values []string, aliases map[string]string) ([]string, error) {
	seen := make(map[string]bool)
	result := make([]string, 0, len(values))
	for _, raw := range values {
		for _, item := range strings.Split(raw, ",") {
			value := strings.TrimSpace(strings.ToLower(item))
			if value == "" {
				continue
			}
			normalized, ok := aliases[value]
			if !ok {
				return nil, fmt.Errorf("invalid %s filter: %s", name, item)
			}
			if !seen[normalized] {
				seen[normalized] = true
				result = append(result, normalized)
			}
		}
	}
	return result, nil
}

func taskWeekRange(values []string, now time.Time) (string, string, error) {
	offsets := make(map[int]bool)
	for _, raw := range values {
		for _, item := range strings.Split(raw, ",") {
			value := strings.TrimSpace(strings.ToLower(item))
			if value == "" {
				continue
			}
			offset, ok := taskWeekAliases[value]
			if !ok {
				return "", "", fmt.Errorf("invalid week filter: %s", item)
			}
			offsets[offset] = true
		}
	}
	if len(offsets) == 0 {
		return "", "", fmt.Errorf("week filter cannot be empty")
	}

	ordered := make([]int, 0, len(offsets))
	for offset := range offsets {
		ordered = append(ordered, offset)
	}
	sort.Ints(ordered)
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	thisMonday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -daysSinceMonday)
	from := thisMonday.AddDate(0, 0, ordered[0]*7)
	to := thisMonday.AddDate(0, 0, ordered[len(ordered)-1]*7+6)
	return from.Format("2006-01-02"), to.Format("2006-01-02"), nil
}
