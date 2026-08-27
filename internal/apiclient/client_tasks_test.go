package apiclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestGetTaskSummary(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/projects/project-1/tasks/summary" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("status") != "todo,in_progress" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"data":{"allTasks":{"total":3,"completed":1,"incomplete":2,"inProgress":1},"managementTasks":{"total":2,"completed":1,"incomplete":1,"inProgress":1},"progress":75,"dueSoon":1,"overdue":1,"noDueDate":0}}`)),
			Request:    r,
		}, nil
	})

	client := NewClient("https://taskaio.example", "token", time.Second)
	client.HTTPClient.Transport = transport
	summary, err := client.GetTaskSummary(context.Background(), "project-1", url.Values{"status": {"todo,in_progress"}})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Progress != 75 || summary.AllTasks.Total != 3 || summary.DueSoon != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
