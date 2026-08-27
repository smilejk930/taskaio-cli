package cmd

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestNormalizeTaskFilterValues(t *testing.T) {
	got, err := normalizeTaskFilterValues("status", []string{"할일,진행중", "todo", "review"}, taskStatusAliases)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"todo", "in_progress", "review"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if _, err := normalizeTaskFilterValues("status", []string{"대기"}, taskStatusAliases); err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestTaskWeekRange(t *testing.T) {
	location := time.FixedZone("KST", 9*60*60)
	now := time.Date(2026, 8, 27, 18, 0, 0, 0, location)
	from, to, err := taskWeekRange([]string{"지난주", "next"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if from != "2026-08-17" || to != "2026-09-06" {
		t.Fatalf("got %s..%s", from, to)
	}
}

func TestBuildTaskFilterParams(t *testing.T) {
	previousNow := taskFilterNow
	taskFilterNow = func() time.Time {
		return time.Date(2026, 8, 27, 12, 0, 0, 0, time.FixedZone("KST", 9*60*60))
	}
	t.Cleanup(func() { taskFilterNow = previousNow })

	command := &cobra.Command{Use: "test"}
	addTaskFilterFlags(command, true)
	args := []string{
		"--status", "할일,진행중",
		"--priority", "urgent", "--priority", "높음",
		"--week", "이번주",
		"--management-only",
		"--due-soon",
	}
	if err := command.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	params, err := buildTaskFilterParams(command, true)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"status":   "todo,in_progress",
		"priority": "urgent,high",
		"from":     "2026-08-24",
		"to":       "2026-08-30",
		"parentId": "root",
		"due":      "soon",
		"asOf":     "2026-08-27",
	}
	for key, value := range want {
		if params.Get(key) != value {
			t.Errorf("%s: got %q, want %q", key, params.Get(key), value)
		}
	}
}

func TestBuildTaskFilterParamsRejectsConflicts(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	addTaskFilterFlags(command, true)
	if err := command.ParseFlags([]string{"--due-soon", "--overdue"}); err != nil {
		t.Fatal(err)
	}
	if _, err := buildTaskFilterParams(command, true); err == nil {
		t.Fatal("expected due filter conflict")
	}
}
