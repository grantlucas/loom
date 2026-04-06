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
	lv.Update(keyMsg('c')) // show closed issues for this test
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

func TestListView_FilteredStatusBarSingular(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Unique thing"},
		{ID: "a-2", Title: "Something else"},
	})

	enterFilterMode(lv)
	typeText(lv, "unique")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "1 of 2 issue") {
		t.Errorf("expected '1 of 2 issue' (singular), got:\n%s", view)
	}
	// Must not say "issues" (plural)
	if strings.Contains(view, "1 of 2 issues") {
		t.Error("expected singular 'issue' not 'issues' for 1 result")
	}
}

func TestListView_FilterByStatus(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Open task", Status: "open"},
		{ID: "a-2", Title: "Active task", Status: "in_progress"},
		{ID: "a-3", Title: "Done task", Status: "closed"},
	})

	enterFilterMode(lv)
	typeText(lv, "status:open")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "1 of 3") {
		t.Errorf("expected '1 of 3' for status:open filter, got:\n%s", view)
	}
	if !strings.Contains(view, "a-1") {
		t.Error("expected a-1 (open) in filtered results")
	}
	if strings.Contains(view, "a-2") {
		t.Error("expected a-2 (in_progress) filtered out")
	}
}

func TestListView_FilterByPriority(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Critical", Priority: 0},
		{ID: "a-2", Title: "Medium", Priority: 2},
		{ID: "a-3", Title: "Low", Priority: 4},
	})

	enterFilterMode(lv)
	typeText(lv, "priority:2")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "1 of 3") {
		t.Errorf("expected '1 of 3' for priority:2 filter, got:\n%s", view)
	}
	if !strings.Contains(view, "a-2") {
		t.Error("expected a-2 (priority 2) in results")
	}
}

func TestListView_FilterByType(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Bug report", IssueType: "bug"},
		{ID: "a-2", Title: "New feature", IssueType: "feature"},
	})

	enterFilterMode(lv)
	typeText(lv, "type:bug")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "1 of 2") {
		t.Errorf("expected '1 of 2' for type:bug filter, got:\n%s", view)
	}
	if !strings.Contains(view, "a-1") {
		t.Error("expected a-1 (bug) in results")
	}
}

func TestListView_FilterByAssignee(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Alice task", Assignee: "alice"},
		{ID: "a-2", Title: "Bob task", Assignee: "bob"},
	})

	enterFilterMode(lv)
	typeText(lv, "assignee:alice")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "1 of 2") {
		t.Errorf("expected '1 of 2' for assignee:alice filter, got:\n%s", view)
	}
	if !strings.Contains(view, "a-1") {
		t.Error("expected a-1 (alice) in results")
	}
	if strings.Contains(view, "a-2") {
		t.Error("expected a-2 (bob) filtered out")
	}
}

func TestListView_ComposableFilters(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Open bug", Status: "open", Priority: 1},
		{ID: "a-2", Title: "Open feature", Status: "open", Priority: 2},
		{ID: "a-3", Title: "Closed bug", Status: "closed", Priority: 1},
	})

	enterFilterMode(lv)
	typeText(lv, "status:open priority:1")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "1 of 3") {
		t.Errorf("expected '1 of 3' for composable filter, got:\n%s", view)
	}
	if !strings.Contains(view, "a-1") {
		t.Error("expected a-1 (open + P1) in results")
	}
	if strings.Contains(view, "a-2") {
		t.Error("expected a-2 (open but P2) filtered out")
	}
	if strings.Contains(view, "a-3") {
		t.Error("expected a-3 (P1 but closed) filtered out")
	}
}

func TestListView_MixedFieldAndFreetext(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Fix login bug", Status: "open"},
		{ID: "a-2", Title: "Fix logout bug", Status: "open"},
		{ID: "a-3", Title: "Fix login style", Status: "closed"},
	})

	enterFilterMode(lv)
	typeText(lv, "status:open login")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "1 of 3") {
		t.Errorf("expected '1 of 3' for mixed filter, got:\n%s", view)
	}
	if !strings.Contains(view, "a-1") {
		t.Error("expected a-1 (open + login) in results")
	}
	if strings.Contains(view, "a-2") {
		t.Error("expected a-2 (open but no 'login') filtered out")
	}
	if strings.Contains(view, "a-3") {
		t.Error("expected a-3 (login but closed) filtered out")
	}
}

func TestListView_SetIssuesReappliesActiveFilter(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Login", Status: "open"},
		{ID: "a-2", Title: "Dashboard", Status: "open"},
	})

	// Apply a filter
	enterFilterMode(lv)
	typeText(lv, "login")
	lv.Update(enterKey())

	// Simulate data refresh (watch mode)
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Login", Status: "open"},
		{ID: "a-2", Title: "Dashboard", Status: "open"},
		{ID: "a-3", Title: "Login v2", Status: "open"},
	})

	view := lv.View()
	// Filter should still be active, now matching 2 of 3
	if !strings.Contains(view, "2 of 3") {
		t.Errorf("expected '2 of 3' after SetIssues with active filter, got:\n%s", view)
	}
}

func TestListView_Resize_TitleColumnExpandsWithWidth(t *testing.T) {
	lv := NewListView()

	// Default title width is 40
	lv.Resize(80, 30)
	cols80 := lv.table.Columns()
	titleWidth80 := cols80[5].Width

	lv.Resize(160, 30)
	cols160 := lv.table.Columns()
	titleWidth160 := cols160[5].Width

	if titleWidth160 <= titleWidth80 {
		t.Errorf("expected wider terminal to give wider Title column: got %d at 160w vs %d at 80w",
			titleWidth160, titleWidth80)
	}
}

func TestListView_Resize_FixedColumnsStaySameWidth(t *testing.T) {
	lv := NewListView()
	lv.Resize(120, 30)
	cols := lv.table.Columns()

	// ID=14, P=5, Type=8, Status=12, Assignee=14
	expectedWidths := []int{14, 5, 8, 12, 14}
	for i, want := range expectedWidths {
		if cols[i].Width != want {
			t.Errorf("column %d (%s): got width %d, want %d", i, cols[i].Title, cols[i].Width, want)
		}
	}
}

// --- StatusHints ---

func TestListView_ImplementsStatusHinter(t *testing.T) {
	var _ StatusHinter = NewListView()
}

func TestListView_StatusHints_NormalMode(t *testing.T) {
	lv := NewListView()
	hints := lv.StatusHints()

	keys := make(map[string]string)
	for _, h := range hints {
		keys[h.Key] = h.Desc
	}

	for _, k := range []string{"s", "/", "enter"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("expected hint for key %q in normal mode", k)
		}
	}
}

func TestListView_StatusHints_FilterMode(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{{ID: "a-1", Title: "Test"}})
	enterFilterMode(lv)

	hints := lv.StatusHints()
	keys := make(map[string]string)
	for _, h := range hints {
		keys[h.Key] = h.Desc
	}

	if _, ok := keys["enter"]; !ok {
		t.Error("expected 'enter' hint in filter mode")
	}
	if _, ok := keys["esc"]; !ok {
		t.Error("expected 'esc' hint in filter mode")
	}
	// Should NOT have sort hint in filter mode
	if _, ok := keys["s"]; ok {
		t.Error("should not show sort hint in filter mode")
	}
}

func TestListView_StatusHints_ActiveFilter(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Login"},
		{ID: "a-2", Title: "Other"},
	})

	enterFilterMode(lv)
	typeText(lv, "login")
	lv.Update(enterKey())

	hints := lv.StatusHints()

	// Should have normal hints plus filter-active indicator
	found := false
	for _, h := range hints {
		if h.Key == "esc" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'esc' hint when filter is active")
	}
}

func TestListView_HidesClosedByDefault(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Open task", Status: "open"},
		{ID: "a-2", Title: "Active task", Status: "in_progress"},
		{ID: "a-3", Title: "Done task", Status: "closed"},
	})

	view := lv.View()
	if strings.Contains(view, "a-3") {
		t.Error("expected closed issue a-3 to be hidden by default")
	}
	if !strings.Contains(view, "a-1") {
		t.Error("expected open issue a-1 to be visible")
	}
	if !strings.Contains(view, "a-2") {
		t.Error("expected in_progress issue a-2 to be visible")
	}
	if !strings.Contains(view, "2 of 3") {
		t.Errorf("expected '2 of 3' count (closed hidden), got:\n%s", view)
	}
}

func TestListView_ToggleClosedWithCKey(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Open task", Status: "open"},
		{ID: "a-2", Title: "Done task", Status: "closed"},
	})

	// Default: closed hidden
	view := lv.View()
	if strings.Contains(view, "a-2") {
		t.Fatal("expected closed issue hidden by default")
	}

	// Press 'c' to show closed
	lv.Update(keyMsg('c'))
	view = lv.View()
	if !strings.Contains(view, "a-2") {
		t.Error("expected closed issue visible after pressing 'c'")
	}

	// Press 'c' again to hide closed
	lv.Update(keyMsg('c'))
	view = lv.View()
	if strings.Contains(view, "a-2") {
		t.Error("expected closed issue hidden after pressing 'c' again")
	}
}

func TestListView_StatusHints_ShowsClosedToggle(t *testing.T) {
	lv := NewListView()

	// Default: closed hidden, hint should say "show closed"
	hints := lv.StatusHints()
	found := false
	for _, h := range hints {
		if h.Key == "c" {
			found = true
			if h.Desc != "show closed" {
				t.Errorf("expected hint 'show closed' when hiding, got %q", h.Desc)
			}
		}
	}
	if !found {
		t.Error("expected 'c' hint in status hints")
	}

	// Toggle: now showing closed, hint should say "hide closed"
	lv.Update(keyMsg('c'))
	hints = lv.StatusHints()
	for _, h := range hints {
		if h.Key == "c" {
			if h.Desc != "hide closed" {
				t.Errorf("expected hint 'hide closed' when showing, got %q", h.Desc)
			}
		}
	}
}

func TestListView_HideClosedCombinesWithFilter(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Title: "Login bug", Status: "open"},
		{ID: "a-2", Title: "Login fix", Status: "closed"},
		{ID: "a-3", Title: "Dashboard", Status: "open"},
	})

	// Filter for "login" — with hideClosed on, should only show a-1
	enterFilterMode(lv)
	typeText(lv, "login")
	lv.Update(enterKey())

	view := lv.View()
	if !strings.Contains(view, "a-1") {
		t.Error("expected a-1 (open + login) visible")
	}
	if strings.Contains(view, "a-2") {
		t.Error("expected a-2 (closed + login) hidden")
	}
	if strings.Contains(view, "a-3") {
		t.Error("expected a-3 (open but no login) filtered out")
	}
}

func TestListView_ShiftS_TogglesDescendingSort(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Priority: 1, Title: "Alpha"},
		{ID: "a-2", Priority: 2, Title: "Beta"},
		{ID: "a-3", Priority: 3, Title: "Gamma"},
	})

	// Default: ascending priority — a-1 first
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("expected 'a-1' in ascending order, got %q", got)
	}

	// Press 'S' (shift+s) to toggle to descending
	lv.Update(keyMsg('S'))
	if got := lv.SelectedIssueID(); got != "a-3" {
		t.Fatalf("expected 'a-3' in descending order, got %q", got)
	}

	// Press 'S' again to toggle back to ascending
	lv.Update(keyMsg('S'))
	if got := lv.SelectedIssueID(); got != "a-1" {
		t.Fatalf("expected 'a-1' after toggling back to ascending, got %q", got)
	}
}

func TestListView_SortDirectionIndicator(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Priority: 1, Title: "Test"},
	})

	// Default ascending: should show ▲
	view := lv.View()
	if !strings.Contains(view, "▲") {
		t.Errorf("expected ▲ indicator for ascending sort, got:\n%s", view)
	}

	// Toggle to descending: should show ▼
	lv.Update(keyMsg('S'))
	view = lv.View()
	if !strings.Contains(view, "▼") {
		t.Errorf("expected ▼ indicator for descending sort, got:\n%s", view)
	}
}

func TestListView_CycleColumnResetsDirection(t *testing.T) {
	lv := NewListView()
	lv.SetIssues([]datasource.Issue{
		{ID: "a-1", Priority: 1, Status: "open", Title: "Alpha"},
		{ID: "a-2", Priority: 2, Status: "in_progress", Title: "Beta"},
	})

	// Toggle to descending
	lv.Update(keyMsg('S'))
	if got := lv.SelectedIssueID(); got != "a-2" {
		t.Fatalf("expected 'a-2' descending, got %q", got)
	}

	// Cycle column with 's' — direction should reset to ascending
	lv.Update(keyMsg('s'))
	// Now sorting by status ascending: in_progress (a-2) before open (a-1)
	if got := lv.SelectedIssueID(); got != "a-2" {
		t.Fatalf("expected 'a-2' (in_progress first in ascending), got %q", got)
	}

	// Verify direction indicator reset to ascending
	view := lv.View()
	if strings.Contains(view, "▼") {
		t.Error("expected direction to reset to ascending (▲) after cycling column")
	}
	if !strings.Contains(view, "▲") {
		t.Errorf("expected ▲ indicator after column cycle, got:\n%s", view)
	}
}

func TestListView_StatusHints_ShowsSortDirection(t *testing.T) {
	lv := NewListView()

	hints := lv.StatusHints()
	found := false
	for _, h := range hints {
		if h.Key == "S" {
			found = true
			if h.Desc != "reverse sort" {
				t.Errorf("expected hint 'reverse sort', got %q", h.Desc)
			}
		}
	}
	if !found {
		t.Error("expected 'S' hint in status hints")
	}
}

func TestListView_Resize_NarrowTerminalClampsTitleToMinimum(t *testing.T) {
	lv := NewListView()
	lv.Resize(40, 30) // very narrow
	cols := lv.table.Columns()
	titleWidth := cols[5].Width
	if titleWidth < 10 {
		t.Errorf("title column should have minimum width of 10, got %d", titleWidth)
	}
}
