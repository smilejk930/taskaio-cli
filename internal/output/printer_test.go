package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/taskaio/taskaio-cli/internal/apiclient"
)

func TestPrintTaskSummaryTable(t *testing.T) {
	summary := &apiclient.TaskSummary{
		AllTasks:        apiclient.TaskCounts{Total: 3, Completed: 1, Incomplete: 2, InProgress: 1},
		ManagementTasks: apiclient.TaskCounts{Total: 2, Completed: 1, Incomplete: 1, InProgress: 1},
		Progress:        75,
		DueSoon:         1,
		Overdue:         1,
	}
	var buffer bytes.Buffer
	if err := PrintTaskSummary(&buffer, "table", summary); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"MANAGEMENT", "PROGRESS", "DUE SOON", "75%"} {
		if !strings.Contains(buffer.String(), expected) {
			t.Fatalf("output missing %q: %s", expected, buffer.String())
		}
	}
}
