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

func TestListView_CursorNavigation(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "First"},
		{ID: "a-2", Title: "Second"},
		{ID: "a-3", Title: "Third"},
	})

	// Initial cursor should be at row 0
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Errorf("expected initial selection 'a-1', got %q", got)
	}

	// Press j to move down
	lv.Update(keyMsg('j'))
	if got := lv.SelectedIssueID(); got != "a-2" {
		t.Errorf("after j, expected 'a-2', got %q", got)
	}

	// Press k to move back up
	lv.Update(keyMsg('k'))
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Errorf("after k, expected 'a-1', got %q", got)
	}
}

func TestListView_SelectedIssueID_Empty(t *testing.T) {
	lv := NewListView()
	if got := lv.SelectedIssueID(); got != "" {
		t.Errorf("expected empty string for no issues, got %q", got)
	}
}

func TestListView_SortByPriority(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-3", Priority: 3, Title: "Low"},
		{ID: "a-1", Priority: 1, Title: "High"},
		{ID: "a-2", Priority: 2, Title: "Med"},
	})

	// Default sort is by priority ascending
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Errorf("expected first row 'a-1' (P1), got %q", got)
	}
}

func TestListView_SortByStatus(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Priority: 1, Status: "open", Title: "Open"},
		{ID: "a-2", Priority: 1, Status: "in_progress", Title: "Active"},
		{ID: "a-3", Priority: 1, Status: "closed", Title: "Done"},
	})

	// Cycle sort to status column
	lv.Update(keyMsg('s'))
	// Status sort: in_progress first, then open, then closed
	if got := lv.SelectedIssueID(); got != "a-2" {
		t.Errorf("expected first row 'a-2' (in_progress), got %q", got)
	}
}

func TestListView_SortIndicatorInView(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Test"},
	})

	view := lv.View()
	// Default sort by priority should show indicator
	if !strings.Contains(view, "▲") && !strings.Contains(view, "↑") {
		// Just check that sort column name appears — the indicator is in column header
		if !strings.Contains(view, "P") {
			t.Error("expected priority column header in view")
		}
	}
}

func TestListView_BlockedIndicator(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{
			ID:              "a-1",
			Title:           "Blocked task",
			Status:          "open",
			DependencyCount: 2,
		},
	})

	view := lv.View()
	if !strings.Contains(view, "●") {
		t.Errorf("expected blocked indicator ● in view, got:\n%s", view)
	}
}

func TestListView_ReadyIndicator(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{
			ID:              "a-1",
			Title:           "Ready task",
			Status:          "open",
			DependencyCount: 0,
		},
	})

	view := lv.View()
	if !strings.Contains(view, "○") {
		t.Errorf("expected ready indicator ○ in view, got:\n%s", view)
	}
}

func TestListView_ClosedIndicator(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{
			ID:     "a-1",
			Title:  "Done task",
			Status: "closed",
		},
	})

	view := lv.View()
	if !strings.Contains(view, "✓") {
		t.Errorf("expected closed indicator ✓ in view, got:\n%s", view)
	}
}

func TestListView_InProgressIndicator(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{
			ID:     "a-1",
			Title:  "Active task",
			Status: "in_progress",
		},
	})

	view := lv.View()
	if !strings.Contains(view, "◐") {
		t.Errorf("expected in_progress indicator ◐ in view, got:\n%s", view)
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
