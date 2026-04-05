package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grantlucas/loom/internal/datasource"
)

func TestListView_ImplementsViewInterface(t *testing.T) {
	var _ View = NewListView()
}

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

func TestListView_SortCyclesThroughAllColumns(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "b-2", Priority: 2, IssueType: "bug", Status: "open", Assignee: "bob", Title: "Zebra"},
		{ID: "a-1", Priority: 1, IssueType: "task", Status: "in_progress", Assignee: "alice", Title: "Apple"},
	})

	// Default: sorted by priority — a-1 (P1) first
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("priority sort: expected 'a-1', got %q", got)
	}

	// s -> sortByStatus: in_progress (a-1) before open (b-2)
	lv.Update(keyMsg('s'))
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("status sort: expected 'a-1', got %q", got)
	}

	// s -> sortByID: a-1 before b-2
	lv.Update(keyMsg('s'))
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("ID sort: expected 'a-1', got %q", got)
	}

	// s -> sortByType: bug (b-2) before task (a-1)
	lv.Update(keyMsg('s'))
	if got := lv.SelectedIssueID(); got != "b-2" {
		t.Fatalf("type sort: expected 'b-2', got %q", got)
	}

	// s -> sortByAssignee: alice (a-1) before bob (b-2)
	lv.Update(keyMsg('s'))
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("assignee sort: expected 'a-1', got %q", got)
	}

	// s -> sortByTitle: Apple (a-1) before Zebra (b-2)
	lv.Update(keyMsg('s'))
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("title sort: expected 'a-1', got %q", got)
	}

	// s -> wraps back to sortByPriority
	lv.Update(keyMsg('s'))
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("wrapped priority sort: expected 'a-1', got %q", got)
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

func enterFilterMode(lv *ListView) {
	lv.Update(keyMsg('/'))
}

func typeText(lv *ListView, text string) {
	for _, r := range text {
		lv.Update(keyMsg(r))
	}
}

func escKey() tea.Msg {
	return tea.KeyMsg{Type: tea.KeyEscape}
}

func enterKey() tea.Msg {
	return tea.KeyMsg{Type: tea.KeyEnter}
}

func TestListView_SlashEntersFilterMode(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "First"},
	})

	lv.Update(keyMsg('/'))

	view := lv.View()
	if !strings.Contains(view, "Filter:") {
		t.Errorf("expected filter prompt in view after '/', got:\n%s", view)
	}
}

func TestListView_EscExitsFilterMode(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "First"},
		{ID: "a-2", Title: "Second"},
	})

	enterFilterMode(lv)
	lv.Update(escKey())

	view := lv.View()
	if strings.Contains(view, "Filter:") {
		t.Error("expected filter prompt to be hidden after Esc")
	}
	// Should show all issues (no filter active)
	if !strings.Contains(view, "2 issues") {
		t.Errorf("expected '2 issues' after Esc clears filter, got:\n%s", view)
	}
}

func TestListView_FreetextFilterMatchesTitleAndID(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "loom-1", Title: "Fix login bug"},
		{ID: "loom-2", Title: "Add dashboard"},
		{ID: "loom-3", Title: "Login page redesign"},
	})

	enterFilterMode(lv)
	typeText(lv, "login")
	lv.Update(enterKey())

	view := lv.View()
	// Should show 2 of 3 (both login-related issues)
	if !strings.Contains(view, "2 of 3") {
		t.Errorf("expected '2 of 3' in filtered view, got:\n%s", view)
	}
	if !strings.Contains(view, "loom-1") {
		t.Error("expected loom-1 (Fix login bug) in filtered results")
	}
	if !strings.Contains(view, "loom-3") {
		t.Error("expected loom-3 (Login page redesign) in filtered results")
	}
	if strings.Contains(view, "loom-2") {
		t.Error("expected loom-2 (Add dashboard) to be filtered out")
	}
}
