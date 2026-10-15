package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/taskaio/taskaio-cli/internal/apiclient"
)

// Run only against a disposable test server: this creates and deletes resources.
func TestCLILiveLifecycle(t *testing.T) {
	baseURL := os.Getenv("TASKAIO_E2E_BASE_URL")
	if baseURL == "" {
		t.Skip("set TASKAIO_E2E_BASE_URL and TASKAIO_E2E_TOKEN for a disposable test server")
	}
	token := os.Getenv("TASKAIO_E2E_TOKEN")
	if token == "" {
		t.Fatal("TASKAIO_E2E_TOKEN is required")
	}
	binary := cliBinary(t)
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	env := make([]string, 0)
	for _, item := range os.Environ() {
		if !strings.HasPrefix(item, "TASKAIO_") {
			env = append(env, item)
		}
	}
	env = append(env, "TASKAIO_BASE_URL="+baseURL, "TASKAIO_OUTPUT=json")
	run := func(stdin string, wantExit int, args ...string) []byte {
		t.Helper()
		command := exec.Command(binary, append([]string{"--config", configPath}, args...)...)
		command.Env = env
		command.Stdin = strings.NewReader(stdin)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		code := 0
		if err != nil {
			exit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("start CLI: %v", err)
			}
			code = exit.ExitCode()
		}
		if code != wantExit {
			t.Fatalf("CLI %v: exit %d, want %d; stderr: %s", args, code, wantExit, stderr.String())
		}
		if wantExit != 0 && stdout.Len() != 0 {
			t.Fatalf("CLI error contaminated stdout: %s", stdout.String())
		}
		return stdout.Bytes()
	}
	decode := func(data []byte, out any) {
		t.Helper()
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatalf("invalid JSON output: %v; %s", err, data)
		}
	}

	run("", 0, "config", "init")
	run(token, 0, "auth", "login", "--token-stdin")
	info, err := os.Stat(configPath)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("config permissions: %v, %v", info, err)
	}
	var auth apiclient.AuthMeResponse
	decode(run("", 0, "auth", "status"), &auth)
	if auth.User.ID == "" {
		t.Fatal("auth response has no user ID")
	}
	run("", 3, "auth", "status", "--token", "invalid-test-token")

	name := fmt.Sprintf("cli-e2e-%d", time.Now().UnixNano())
	var project apiclient.Project
	decode(run("", 0, "projects", "create", "--name", name, "--description", "test description"), &project)
	if project.ID == "" || project.Name != name {
		t.Fatalf("created project: %+v", project)
	}
	t.Cleanup(func() { run("", 0, "projects", "delete", project.ID, "--yes") })
	var secondProject apiclient.Project
	decode(run(fmt.Sprintf(`{"name":%q}`, name+"-second"), 0, "projects", "create", "--input", "-"), &secondProject)
	t.Cleanup(func() { run("", 0, "projects", "delete", secondProject.ID, "--yes") })
	var projects apiclient.ListResponse[apiclient.Project]
	decode(run("", 0, "projects", "list", "--search", name, "--limit", "1"), &projects)
	if len(projects.Data) != 1 || !projects.Meta.HasMore || projects.Meta.NextCursor == nil {
		t.Fatalf("first project page: %+v", projects)
	}
	decode(run("", 0, "projects", "list", "--search", name, "--limit", "1", "--cursor", *projects.Meta.NextCursor), &projects)
	if len(projects.Data) != 1 || projects.Meta.HasMore {
		t.Fatalf("second project page: %+v", projects)
	}
	decode(run("", 0, "projects", "list", "--search", name, "--limit", "1", "--all"), &projects)
	if len(projects.Data) != 2 || projects.Meta.HasMore {
		t.Fatalf("all project pages: %+v", projects)
	}
	run(`{"description":null}`, 0, "projects", "update", project.ID, "--input", "-")
	var updatedProject apiclient.Project
	decode(run("", 0, "projects", "get", project.ID), &updatedProject)
	if updatedProject.Description != nil {
		t.Fatal("project description was not cleared")
	}
	var members apiclient.ListResponse[apiclient.ProjectMember]
	decode(run("", 0, "projects", "members", "list", project.ID, "--all"), &members)
	if len(members.Data) != 1 || members.Data[0].UserID != auth.User.ID || members.Data[0].Role != "owner" {
		t.Fatalf("project owner: %+v", members)
	}

	var parent, child apiclient.Task
	decode(run("", 0, "tasks", "create", "--project", project.ID, "--title", name, "--assignee", auth.User.ID, "--start-date", "2026-10-12", "--end-date", "2026-10-18"), &parent)
	t.Cleanup(func() { run("", 0, "tasks", "delete", parent.ID, "--yes") })
	childInput := fmt.Sprintf(`{"title":"child","parentId":%q,"assigneeId":%q,"status":"review","priority":"urgent","startDate":"2026-10-12","endDate":"2026-10-14"}`, parent.ID, auth.User.ID)
	decode(run(childInput, 0, "tasks", "create", "--project", project.ID, "--input", "-"), &child)
	var tasks apiclient.ListResponse[apiclient.Task]
	decode(run("", 0, "tasks", "list", "--project", project.ID, "--all", "--limit", "1"), &tasks)
	if len(tasks.Data) != 2 || tasks.Meta.HasMore {
		t.Fatalf("all task pages: %+v", tasks)
	}
	for _, filter := range [][]string{
		{"--parent", parent.ID},
		{"--status", "리뷰", "--priority", "긴급"},
		{"--overdue", "--as-of", "2026-10-15"},
		{"--from", "2026-10-14", "--to", "2026-10-14", "--priority", "urgent"},
	} {
		decode(run("", 0, append([]string{"tasks", "list", "--project", project.ID}, filter...)...), &tasks)
		if len(tasks.Data) != 1 || tasks.Data[0].ID != child.ID {
			t.Fatalf("task filter %v: %+v", filter, tasks)
		}
	}
	decode(run("", 0, "tasks", "list", "--project", project.ID, "--management-only", "--due-soon", "--as-of", "2026-10-15"), &tasks)
	if len(tasks.Data) != 1 || tasks.Data[0].ID != parent.ID {
		t.Fatalf("management due-soon filter: %+v", tasks)
	}
	var summary apiclient.TaskSummary
	decode(run("", 0, "tasks", "summary", "--project", project.ID, "--as-of", "2026-10-15"), &summary)
	if summary.AllTasks.Total != 2 || summary.ManagementTasks.Total != 1 || summary.DueSoon != 1 || summary.Overdue != 1 {
		t.Fatalf("task summary: %+v", summary)
	}
	run(`{"parentId":null,"startDate":null,"endDate":null,"progress":100}`, 0, "tasks", "update", child.ID, "--input", "-")
	var updatedTask apiclient.Task
	decode(run("", 0, "tasks", "get", child.ID), &updatedTask)
	if updatedTask.ParentID != nil || updatedTask.StartDate != nil || updatedTask.EndDate != nil || updatedTask.Progress != 100 || updatedTask.Status != "done" {
		t.Fatalf("task update and nulls: %+v", updatedTask)
	}
	run("", 0, "tasks", "delete", child.ID, "--yes")
	run("", 5, "tasks", "get", child.ID)

	var schedule apiclient.Schedule
	decode(run("", 0, "schedules", "create", "--name", name, "--start-date", "2026-10-15", "--end-date", "2026-10-15", "--user", auth.User.ID, "--note", "clear me"), &schedule)
	t.Cleanup(func() { run("", 0, "schedules", "delete", schedule.ID, "--yes") })
	if schedule.Type != "member_leave" {
		t.Fatalf("schedule default type = %q", schedule.Type)
	}
	run(`{"note":null,"memberId":null,"type":"other"}`, 0, "schedules", "update", schedule.ID, "--input", "-")
	var updatedSchedule apiclient.Schedule
	decode(run("", 0, "schedules", "get", schedule.ID), &updatedSchedule)
	if updatedSchedule.Note != nil || updatedSchedule.MemberID != nil || updatedSchedule.Type != "other" {
		t.Fatalf("schedule update and nulls: %+v", updatedSchedule)
	}
	var schedules apiclient.ListResponse[apiclient.Schedule]
	decode(run("", 0, "schedules", "list", "--search", name, "--from", "2026-10-15", "--to", "2026-10-15", "--type", "other", "--all"), &schedules)
	if len(schedules.Data) != 1 || schedules.Data[0].ID != schedule.ID {
		t.Fatalf("schedule filters: %+v", schedules)
	}
	if table := run("", 0, "tasks", "summary", "--project", project.ID, "--output", "table"); !bytes.Contains(table, []byte("PROGRESS")) {
		t.Fatalf("table output: %s", table)
	}
	run(`{"status":"invalid"}`, 2, "tasks", "update", parent.ID, "--input", "-")
}
