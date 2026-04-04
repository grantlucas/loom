package tui

import (
	"strings"
	"testing"

	"github.com/grantlucas/loom/internal/datasource"
)

func TestListView_RendersColumnHeaders(t *testing.T) {
	lv := NewListView()
	view := lv.View()

	headers := []string{"ID", "P", "Type", "Status", "Assignee", "Title"}
	for _, h := range headers {
		if !strings.Contains(view, h) {
			t.Errorf("expected view to contain column header %q", h)
		}
	}
}

func TestListView_ShowsIssueCount(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "First"},
		{ID: "a-2", Title: "Second"},
		{ID: "a-3", Title: "Third"},
	})

	view := lv.View()
	if !strings.Contains(view, "3 issues") {
		t.Errorf("expected view to contain '3 issues', got:\n%s", view)
	}
}

func TestListView_ShowsIssueCountSingular(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Only one"},
	})

	view := lv.View()
	if !strings.Contains(view, "1 issue") {
		t.Errorf("expected view to contain '1 issue', got:\n%s", view)
	}
	if strings.Contains(view, "1 issues") {
		t.Error("should use singular 'issue' not 'issues'")
	}
}

func TestListView_RendersIssueData(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{
			ID:        "loom-1",
			Priority:  1,
			IssueType: "task",
			Status:    "open",
			Assignee:  "alice",
			Title:     "Fix the widget",
		},
	})

	view := lv.View()
	for _, want := range []string{"loom-1", "P1", "task", "open", "alice", "Fix the widget"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected view to contain %q", want)
		}
	}
}
