package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
	"github.com/grantlucas/loom/internal/graph"
)

// Compile-time check: FocusView must implement View.
var _ View = (*FocusView)(nil)

func TestNewFocusView_ReturnsNonNil(t *testing.T) {
	fv := NewFocusView()
	if fv == nil {
		t.Fatal("NewFocusView should return a non-nil pointer")
	}
}

func TestFocusView_EmptyState(t *testing.T) {
	fv := NewFocusView()
	out := fv.View()
	if !strings.Contains(out, "No data loaded") {
		t.Error("empty focus view should show 'No data loaded'")
	}
}

func TestFocusView_NoReadyIssues(t *testing.T) {
	fv := NewFocusView()
	fv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
	})
	fv.SetReady(nil)
	out := fv.View()
	if !strings.Contains(out, "No ready issues") {
		t.Errorf("should show 'No ready issues' when ready is empty, got:\n%s", out)
	}
}

// --- Populated view ---

func TestFocusView_ShowsRankedReadyIssues(t *testing.T) {
	fv := newFocusViewPopulated()
	out := fv.View()
	// Should show both ready issues
	if !strings.Contains(out, "bd-a1") {
		t.Error("should contain ready issue bd-a1")
	}
	if !strings.Contains(out, "bd-c3") {
		t.Error("should contain ready issue bd-c3")
	}
	// bd-a1 unblocks 2 (P0 + P2), bd-c3 unblocks 1 (P1)
	// bd-a1 priority sum = (4-0)+(4-2) = 6, bd-c3 = (4-1) = 3
	// So bd-a1 should appear first (rank 1)
	idx1 := strings.Index(out, "bd-a1")
	idx2 := strings.Index(out, "bd-c3")
	if idx1 > idx2 {
		t.Error("bd-a1 (higher impact) should appear before bd-c3")
	}
}

func TestFocusView_ShowsSummaryLine(t *testing.T) {
	fv := newFocusViewPopulated()
	out := fv.View()
	if !strings.Contains(out, "2 ready issues") {
		t.Errorf("should show ready count in summary, got:\n%s", out)
	}
	if !strings.Contains(out, "3 total blocked") {
		t.Errorf("should show total blocked count, got:\n%s", out)
	}
}

func TestFocusView_ShowsLeafLabel(t *testing.T) {
	fv := NewFocusView()
	fv.SetIssues([]datasource.Issue{
		{ID: "leaf-1", Status: "open", Priority: 3, Title: "Leaf issue"},
	})
	fv.SetReady([]datasource.Issue{
		{ID: "leaf-1"},
	})
	out := fv.View()
	if !strings.Contains(out, "leaf") {
		t.Errorf("should show 'leaf' for issues that unblock nothing, got:\n%s", out)
	}
}

func TestFocusView_ExpandedShowsDownstream(t *testing.T) {
	fv := newFocusViewPopulated()
	out := fv.View()
	// In expanded mode, should show downstream issue IDs
	if !strings.Contains(out, "bd-b2") {
		t.Error("expanded view should show downstream issue bd-b2")
	}
	if !strings.Contains(out, "bd-d4") {
		t.Error("expanded view should show downstream issue bd-d4")
	}
}

func TestFocusView_ExpandedShowsTreeChars(t *testing.T) {
	fv := newFocusViewPopulated()
	out := fv.View()
	if !strings.Contains(out, "└→") && !strings.Contains(out, "├→") {
		t.Error("expanded view should show tree connectors")
	}
}

func TestFocusView_CollapsedHidesDownstream(t *testing.T) {
	fv := newFocusViewPopulated()
	// Toggle to collapsed
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if fv.expanded {
		t.Fatal("expected collapsed after 'e' press")
	}
	out := fv.View()
	// Should still show ready issues but not downstream tree
	if !strings.Contains(out, "bd-a1") {
		t.Error("collapsed view should still show ready issue")
	}
	if strings.Contains(out, "├→") || strings.Contains(out, "└→") {
		t.Error("collapsed view should not show tree connectors")
	}
}

// --- Cursor navigation ---

func TestFocusView_CursorDown(t *testing.T) {
	fv := newFocusViewPopulated()
	if fv.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if fv.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", fv.cursor)
	}
}

func TestFocusView_CursorUp(t *testing.T) {
	fv := newFocusViewPopulated()
	fv.cursor = 2
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if fv.cursor != 1 {
		t.Errorf("expected cursor 1 after k, got %d", fv.cursor)
	}
}

func TestFocusView_CursorClampTop(t *testing.T) {
	fv := newFocusViewPopulated()
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if fv.cursor != 0 {
		t.Errorf("cursor should clamp at top, got %d", fv.cursor)
	}
}

func TestFocusView_CursorClampBottom(t *testing.T) {
	fv := newFocusViewPopulated()
	total := fv.totalLines()
	fv.cursor = total - 1
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if fv.cursor != total-1 {
		t.Errorf("cursor should clamp at bottom, got %d", fv.cursor)
	}
}

func TestFocusView_SelectedNodeID_ReadyItem(t *testing.T) {
	fv := newFocusViewPopulated()
	id := fv.SelectedNodeID()
	// First item should be the highest-impact ready issue
	if id == "" {
		t.Error("expected non-empty selection")
	}
}

func TestFocusView_SelectedNodeID_Downstream(t *testing.T) {
	fv := newFocusViewPopulated()
	// Move cursor to first downstream item (cursor 1)
	fv.cursor = 1
	id := fv.SelectedNodeID()
	if id == "" {
		t.Error("expected non-empty selection on downstream item")
	}
	// Should be a downstream ID, not the same as cursor 0
	if id == fv.items[0].NodeID {
		t.Error("downstream selection should differ from parent")
	}
}

func TestFocusView_SelectedNodeID_Empty(t *testing.T) {
	fv := NewFocusView()
	if fv.SelectedNodeID() != "" {
		t.Error("expected empty string when no items")
	}
}

func TestFocusView_SelectedNodeID_Collapsed(t *testing.T) {
	fv := newFocusViewPopulated()
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}) // collapse
	fv.cursor = 1
	id := fv.SelectedNodeID()
	if id == "" {
		t.Error("expected non-empty selection in collapsed mode")
	}
	// In collapsed mode cursor 1 should select the second ready item
	if id != fv.items[1].NodeID {
		t.Errorf("expected second item %q, got %q", fv.items[1].NodeID, id)
	}
}

// --- Sort modes ---

func TestFocusView_SortCycles(t *testing.T) {
	fv := newFocusViewPopulated()
	if fv.sortMode != focusSortByImpact {
		t.Fatal("initial sort should be impact")
	}
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if fv.sortMode != focusSortByPriority {
		t.Error("first cycle should go to priority")
	}
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if fv.sortMode != focusSortByDepth {
		t.Error("second cycle should go to depth")
	}
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if fv.sortMode != focusSortByUnblock {
		t.Error("third cycle should go to unblock")
	}
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if fv.sortMode != focusSortByImpact {
		t.Error("fourth cycle should wrap to impact")
	}
}

func TestFocusView_SortResetsCursor(t *testing.T) {
	fv := newFocusViewPopulated()
	fv.cursor = 3
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if fv.cursor != 0 {
		t.Error("sort should reset cursor to 0")
	}
}

func TestFocusView_ExpandToggleResetsCursor(t *testing.T) {
	fv := newFocusViewPopulated()
	fv.cursor = 3
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if fv.cursor != 0 {
		t.Error("expand toggle should reset cursor to 0")
	}
}

func TestFocusView_SortByPriority(t *testing.T) {
	fv := NewFocusView()
	fv.SetIssues([]datasource.Issue{
		{ID: "lo", Status: "open", Priority: 3, Title: "Low pri"},
		{ID: "hi", Status: "open", Priority: 0, Title: "High pri"},
	})
	fv.SetReady([]datasource.Issue{{ID: "lo"}, {ID: "hi"}})
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}) // switch to priority sort
	if fv.items[0].NodeID != "hi" {
		t.Errorf("priority sort should put P0 first, got %q", fv.items[0].NodeID)
	}
}

func TestFocusView_ShowsSortModeInSummary(t *testing.T) {
	fv := newFocusViewPopulated()
	out := fv.View()
	if !strings.Contains(out, "sort: impact") {
		t.Errorf("should show current sort mode, got:\n%s", out)
	}
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	out = fv.View()
	if !strings.Contains(out, "sort: priority") {
		t.Errorf("should show priority sort mode, got:\n%s", out)
	}
}

// --- Update edge cases ---

func TestFocusView_Update_NonKeyMsg(t *testing.T) {
	fv := NewFocusView()
	cmd := fv.Update(RefreshMsg{})
	if cmd != nil {
		t.Error("expected nil cmd for non-key message")
	}
}

func TestFocusView_Update_UnhandledKey(t *testing.T) {
	fv := NewFocusView()
	cmd := fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if cmd != nil {
		t.Error("expected nil cmd for unhandled key")
	}
}

func TestFocusView_LongTitleTruncated(t *testing.T) {
	fv := NewFocusView()
	fv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1, Title: "This is a very long title that exceeds forty characters and should be truncated"},
	})
	fv.SetReady([]datasource.Issue{{ID: "a"}})
	out := fv.View()
	if !strings.Contains(out, "...") {
		t.Error("long title should be truncated with ...")
	}
}

func TestFocusView_SelectedNodeID_PastEnd(t *testing.T) {
	fv := newFocusViewPopulated()
	fv.cursor = 999
	id := fv.SelectedNodeID()
	if id != "" {
		t.Errorf("expected empty string for cursor past end, got %q", id)
	}
}

func TestFocusView_SelectedNodeID_CollapsedPastEnd(t *testing.T) {
	fv := newFocusViewPopulated()
	fv.expanded = false
	fv.cursor = 999
	if fv.SelectedNodeID() != "" {
		t.Error("expected empty string for cursor past end in collapsed mode")
	}
}

func TestFocusView_RenderItemLine_MissingIssue(t *testing.T) {
	fv := NewFocusView()
	fv.SetIssues([]datasource.Issue{})
	// Manually set items with a node not in the issue map
	fv.readyIDs = []string{"ghost"}
	fv.items = []graph.Impact{{NodeID: "ghost"}}
	out := fv.View()
	if !strings.Contains(out, "ghost") {
		t.Error("should render unknown node ID")
	}
}

// --- Helper ---

func newFocusViewPopulated() *FocusView {
	fv := NewFocusView()
	issues := []datasource.Issue{
		{ID: "bd-a1", Status: "open", Priority: 1, Title: "Fix auth", IssueType: "task"},
		{ID: "bd-b2", Status: "open", Priority: 0, Title: "Deploy hotfix", IssueType: "task", Dependencies: []datasource.RawDependency{
			{IssueID: "bd-b2", DependsOnID: "bd-a1"},
		}},
		{ID: "bd-e5", Status: "open", Priority: 2, Title: "Release v2", IssueType: "task", Dependencies: []datasource.RawDependency{
			{IssueID: "bd-e5", DependsOnID: "bd-a1"},
		}},
		{ID: "bd-c3", Status: "open", Priority: 2, Title: "Add retry", IssueType: "feature"},
		{ID: "bd-d4", Status: "open", Priority: 1, Title: "Integration tests", IssueType: "task", Dependencies: []datasource.RawDependency{
			{IssueID: "bd-d4", DependsOnID: "bd-c3"},
		}},
	}
	fv.SetIssues(issues)
	fv.SetReady([]datasource.Issue{
		{ID: "bd-a1"},
		{ID: "bd-c3"},
	})
	return fv
}
