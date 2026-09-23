package cmd_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func cliBinary(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "taskaio")
	build := exec.Command("go", "build", "-o", path, "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	return path
}

func runCLI(t *testing.T, binary, url, stdin string, args ...string) string {
	t.Helper()
	command := exec.Command(binary, append([]string{"--config", os.DevNull, "--token", "test-token", "--base-url", url}, args...)...)
	command.Stdin = strings.NewReader(stdin)
	command.Env = append(os.Environ(), "TASKAIO_CONFIG="+os.DevNull, "TASKAIO_BASE_URL="+url, "TASKAIO_TOKEN=test-token")
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func TestCLIInputPreservesValuesAndNull(t *testing.T) {
	binary := cliBinary(t)
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	runCLI(t, binary, server.URL, `{"title":"task","status":"review","priority":"urgent"}`, "tasks", "create", "--project", "p", "--input", "-")
	runCLI(t, binary, server.URL, `{"name":"holiday","startDate":"2026-09-23","endDate":"2026-09-23","type":"public_holiday"}`, "schedules", "create", "--input", "-")
	runCLI(t, binary, server.URL, `{"assigneeId":null,"parentId":null,"startDate":null,"endDate":null}`, "tasks", "update", "t", "--input", "-")
	runCLI(t, binary, server.URL, `{"description":null}`, "projects", "update", "p", "--input", "-")
	runCLI(t, binary, server.URL, `{"memberId":null,"note":null}`, "schedules", "update", "s", "--input", "-")
	if bodies[0]["status"] != "review" || bodies[0]["priority"] != "urgent" {
		t.Errorf("task input overwritten: %#v", bodies[0])
	}
	if bodies[1]["type"] != "public_holiday" {
		t.Errorf("schedule input overwritten: %#v", bodies[1])
	}
	for _, key := range []string{"assigneeId", "parentId", "startDate", "endDate"} {
		value, ok := bodies[2][key]
		if !ok || value != nil {
			t.Errorf("missing explicit null for %s: %#v", key, bodies[2])
		}
	}
	for i, keys := range [][]string{{"description"}, {"memberId", "note"}} {
		for _, key := range keys {
			value, ok := bodies[i+3][key]
			if !ok || value != nil {
				t.Errorf("missing explicit null for %s: %#v", key, bodies[i+3])
			}
		}
	}
}

func TestCLIEmptyListAndValidationDetails(t *testing.T) {
	binary := cliBinary(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[],"meta":{"hasMore":false}}`))
			return
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":{"code":"UNPROCESSABLE_ENTITY","message":"Validation failed","details":[{"path":["title"],"message":"Required"}]}}`))
	}))
	defer server.Close()

	for _, args := range [][]string{{"projects", "list"}, {"projects", "members", "list", "p"}, {"tasks", "list", "--project", "p"}, {"schedules", "list"}} {
		out := runCLI(t, binary, server.URL, "", args...)
		if !strings.Contains(out, `"data": []`) {
			t.Errorf("empty list is not an array for %v: %s", args, out)
		}
	}
	command := exec.Command(binary, "--config", os.DevNull, "--token", "test-token", "--base-url", server.URL, "tasks", "create", "--project", "p", "--title", "x")
	outBytes, _ := command.CombinedOutput()
	if !strings.Contains(string(outBytes), "title") || !strings.Contains(string(outBytes), "Required") {
		t.Errorf("validation details missing: %s", outBytes)
	}
}
