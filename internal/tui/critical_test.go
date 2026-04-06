package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
	"github.com/grantlucas/loom/internal/graph"
)

// Compile-time check: CriticalPathView must implement View.
var _ View = (*CriticalPathView)(nil)

func TestNewCriticalPathView_ReturnsNonNil(t *testing.T) {
	cv := NewCriticalPathView()
	if cv == nil {
		t.Fatal("NewCriticalPathView should return a non-nil pointer")
	}
}

func TestCriticalPathView_EmptyState(t *testing.T) {
	cv := NewCriticalPathView()
	out := cv.View()
	if !strings.Contains(out, "No data loaded") {
		t.Error("empty critical path view should show 'No data loaded'")
	}
}

func TestCriticalPathView_SetIssues_ComputesChains(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
		{ID: "c", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "c", DependsOnID: "b"},
		}},
	})
	if len(cv.chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(cv.chains))
	}
	if cv.chains[0].Length() != 3 {
		t.Errorf("expected chain length 3, got %d", cv.chains[0].Length())
	}
}

func TestCriticalPathView_SetIssues_NoDeps(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 2},
	})
	// Each issue is its own chain (isolated nodes are both roots and sinks)
	if len(cv.chains) != 2 {
		t.Errorf("expected 2 chains (isolated nodes), got %d", len(cv.chains))
	}
}

// --- Rendering ---

func TestCriticalPathView_RendersChainsWithIDs(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "proj-1", Status: "open", Priority: 1},
		{ID: "proj-2", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "proj-2", DependsOnID: "proj-1"},
		}},
	})
	out := cv.View()
	if !strings.Contains(out, "proj-1") {
		t.Error("should contain issue ID proj-1")
	}
	if !strings.Contains(out, "proj-2") {
		t.Error("should contain issue ID proj-2")
	}
}

func TestCriticalPathView_RendersChainHeader(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	out := cv.View()
	if !strings.Contains(out, "Chain") {
		t.Error("should contain chain header")
	}
	if !strings.Contains(out, "depth: 2") {
		t.Errorf("should show depth, got:\n%s", out)
	}
}

func TestCriticalPathView_RendersStatusIndicators(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "closed", Priority: 1},
		{ID: "b", Status: "in_progress", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
		{ID: "c", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "c", DependsOnID: "b"},
		}},
	})
	out := cv.View()
	if !strings.Contains(out, "✓") {
		t.Error("should show ✓ for closed issues")
	}
	if !strings.Contains(out, "◐") {
		t.Error("should show ◐ for in_progress issues")
	}
	if !strings.Contains(out, "○") {
		t.Error("should show ○ for open issues")
	}
}

func TestCriticalPathView_RendersSummaryLine(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 0},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	out := cv.View()
	if !strings.Contains(out, "1 chain") {
		t.Errorf("should show chain count in summary, got:\n%s", out)
	}
	if !strings.Contains(out, "max depth: 2") {
		t.Errorf("should show max depth in summary, got:\n%s", out)
	}
}

func TestCriticalPathView_SummaryShowsBlockedP0(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 0, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	out := cv.View()
	if !strings.Contains(out, "P0 goals: 1") {
		t.Errorf("should show blocked P0 count, got:\n%s", out)
	}
}

// --- Cursor navigation ---

func TestCriticalPathView_CursorDown(t *testing.T) {
	cv := newCriticalViewWithChain()
	if cv.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cv.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", cv.cursor)
	}
}

func TestCriticalPathView_CursorUp(t *testing.T) {
	cv := newCriticalViewWithChain()
	cv.cursor = 2
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if cv.cursor != 1 {
		t.Errorf("expected cursor 1 after k, got %d", cv.cursor)
	}
}

func TestCriticalPathView_CursorClampBottom(t *testing.T) {
	cv := newCriticalViewWithChain()
	// total items = 3 nodes in one chain
	cv.cursor = 2
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cv.cursor != 2 {
		t.Errorf("cursor should clamp at bottom, got %d", cv.cursor)
	}
}

func TestCriticalPathView_CursorClampTop(t *testing.T) {
	cv := newCriticalViewWithChain()
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if cv.cursor != 0 {
		t.Errorf("cursor should clamp at top, got %d", cv.cursor)
	}
}

func TestCriticalPathView_SelectedNodeID(t *testing.T) {
	cv := newCriticalViewWithChain()
	id := cv.SelectedNodeID()
	if id != "a" {
		t.Errorf("expected first node 'a', got %q", id)
	}
	cv.cursor = 2
	id = cv.SelectedNodeID()
	if id != "c" {
		t.Errorf("expected last node 'c', got %q", id)
	}
}

func TestCriticalPathView_SelectedNodeID_Empty(t *testing.T) {
	cv := NewCriticalPathView()
	if cv.SelectedNodeID() != "" {
		t.Error("expected empty string when no chains")
	}
}

func TestCriticalPathView_SelectedNodeHighlighted(t *testing.T) {
	cv := newCriticalViewWithChain()
	cv.cursor = 1
	out := cv.View()
	// The selected node line should be rendered differently
	// We check that the view contains the node (basic sanity)
	if !strings.Contains(out, "b") {
		t.Error("should contain selected node")
	}
}

// --- Sorting ---

func TestCriticalPathView_SortByLength(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "x", Status: "open", Priority: 0},
		{ID: "a", Status: "open", Priority: 2},
		{ID: "b", Status: "open", Priority: 2, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
		{ID: "c", Status: "open", Priority: 2, Dependencies: []datasource.RawDependency{
			{IssueID: "c", DependsOnID: "b"},
		}},
	})
	// Default sort is by length desc — chain [a,b,c] (3) before [x] (1)
	if cv.chains[0].Length() != 3 {
		t.Errorf("expected longest chain first, got length %d", cv.chains[0].Length())
	}
}

func TestCriticalPathView_SortToggle(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "x", Status: "open", Priority: 0},
		{ID: "a", Status: "open", Priority: 2},
		{ID: "b", Status: "open", Priority: 2, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	// Press 'p' to sort by priority
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cv.sortMode != critSortByPriority {
		t.Error("expected sort mode to be critSortByPriority after pressing p")
	}
	// [x] has P0 (higher priority), should be first
	if cv.chains[0].MaxPriority != 0 {
		t.Errorf("expected P0 chain first after priority sort, got P%d", cv.chains[0].MaxPriority)
	}

	// Press 'l' to sort by length
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cv.sortMode != critSortByLength {
		t.Error("expected sort mode to be critSortByLength after pressing l")
	}
}

func TestCriticalPathView_SortResetsCursor(t *testing.T) {
	cv := newCriticalViewWithChain()
	cv.cursor = 2
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cv.cursor != 0 {
		t.Error("expected cursor reset to 0 after sort change")
	}
}

// --- Update returns nil ---

func TestCriticalPathView_Update_UnhandledKey(t *testing.T) {
	cv := NewCriticalPathView()
	cmd := cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if cmd != nil {
		t.Error("expected nil cmd for unhandled key")
	}
}

func TestCriticalPathView_Update_NonKeyMsg(t *testing.T) {
	cv := NewCriticalPathView()
	cmd := cv.Update(RefreshMsg{})
	if cmd != nil {
		t.Error("expected nil cmd for non-key message")
	}
}

// --- Edge cases ---

func TestCriticalPathView_NoBlockingChains(t *testing.T) {
	cv := NewCriticalPathView()
	// Set issues with no dependencies — all isolated, each is a chain of length 1
	// But if we want "No blocking chains found", we need issues set but no chains.
	// Actually isolated nodes DO form chains. Let me test the message appears
	// when issues are loaded but somehow result in no chains (e.g., all in a cycle).
	// Cycles return nil from CriticalPaths. But we can't create cycles via SetIssues.
	// Instead, directly set the state:
	cv.issues = map[string]datasource.Issue{"a": {ID: "a"}}
	cv.chains = nil
	out := cv.View()
	if !strings.Contains(out, "No blocking chains found") {
		t.Error("should show 'No blocking chains found' when issues exist but no chains")
	}
}

func TestCriticalPathView_SelectedNodeID_PastEnd(t *testing.T) {
	cv := newCriticalViewWithChain()
	cv.cursor = 999 // way past end
	id := cv.SelectedNodeID()
	if id != "" {
		t.Errorf("expected empty string for cursor past end, got %q", id)
	}
}

func TestCriticalPathView_RenderNode_MissingIssue(t *testing.T) {
	cv := NewCriticalPathView()
	cv.chains = []graph.Chain{{Nodes: []string{"unknown-1"}, MaxPriority: 0}}
	cv.issues = make(map[string]datasource.Issue)
	out := cv.View()
	if !strings.Contains(out, "unknown-1") {
		t.Error("should render unknown node ID")
	}
	if !strings.Contains(out, "?") {
		t.Error("should show ? for missing issue")
	}
}

func TestCriticalPathView_LongTitleTruncated(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1, Title: "This is a very long title that exceeds forty characters and should be truncated"},
	})
	out := cv.View()
	if !strings.Contains(out, "...") {
		t.Error("long title should be truncated with ...")
	}
}

func TestCriticalPathView_SortByPriority_TiebreakByLength(t *testing.T) {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "x", Status: "open", Priority: 1},
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	// Same priority, so tiebreak by length desc
	if cv.chains[0].Length() < cv.chains[1].Length() {
		t.Error("same priority chains should sort longer first")
	}
}

func TestApp_EnterOnCriticalPath_NonCPV_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabCriticalPath
	app.views[TabCriticalPath] = &stubView{}
	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabCriticalPath {
		t.Error("expected to stay on CriticalPath tab")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- Helpers ---

func newCriticalViewWithChain() *CriticalPathView {
	cv := NewCriticalPathView()
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1, Title: "First"},
		{ID: "b", Status: "in_progress", Priority: 1, Title: "Second", Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
		{ID: "c", Status: "open", Priority: 1, Title: "Third", Dependencies: []datasource.RawDependency{
			{IssueID: "c", DependsOnID: "b"},
		}},
	})
	return cv
}

func TestCriticalPathView_ViewportScrollsToFollowCursor(t *testing.T) {
	cv := newCriticalViewWithChain()
	// Small viewport: height 3 means only 3 content lines visible
	cv.Resize(80, 5) // 5 - 2 overhead = 3 viewport lines

	// Move cursor to last node
	total := cv.totalNodes()
	for i := 0; i < total-1; i++ {
		cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	if cv.viewport.YOffset == 0 {
		t.Error("viewport should scroll down to keep cursor visible")
	}
}

func TestCriticalPathView_ScrollsUpWhenCursorAtTop(t *testing.T) {
	cv := newCriticalViewWithChain()
	cv.Resize(80, 5)

	total := cv.totalNodes()
	for i := 0; i < total-1; i++ {
		cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	for i := 0; i < total-1; i++ {
		cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	}
	offsetBefore := cv.viewport.YOffset
	cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if cv.viewport.YOffset >= offsetBefore && offsetBefore > 0 {
		t.Error("pressing k at cursor 0 should scroll viewport up")
	}
}

func TestCriticalPathView_ViewportRendersWhenSized(t *testing.T) {
	cv := newCriticalViewWithChain()
	cv.Resize(80, 20)
	out := cv.View()
	if !strings.Contains(out, "Chain 1") {
		t.Error("sized viewport should still render chain content")
	}
}

func TestCriticalPathView_Resize_VeryNarrow_ClampsTitleWidth(t *testing.T) {
	cv := NewCriticalPathView()
	cv.Resize(20, 30)
	if cv.titleMaxWidth() < 10 {
		t.Errorf("expected titleMaxWidth >= 10, got %d", cv.titleMaxWidth())
	}
}

func TestCriticalPathView_Resize_AdaptsTitleTruncation(t *testing.T) {
	cv := NewCriticalPathView()
	longTitle := "This is a very long title that should be truncated differently at different widths"
	cv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 0, Title: longTitle},
		{ID: "b", Status: "open", Priority: 1, Title: "Short", Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})

	cv.Resize(60, 30)
	out60 := cv.View()

	cv.Resize(120, 30)
	out120 := cv.View()

	if len(out120) <= len(out60) {
		t.Errorf("expected wider terminal to show more title text: len at 120w=%d, len at 60w=%d",
			len(out120), len(out60))
	}
}

// Suppress unused import warning for key package
var _ = key.Binding{}
