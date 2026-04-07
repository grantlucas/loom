package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
)

// Compile-time check: TreeView must implement View.
var _ View = (*TreeView)(nil)

func TestNewTreeView_ReturnsNonNil(t *testing.T) {
	tv := NewTreeView()
	if tv == nil {
		t.Fatal("NewTreeView should return a non-nil pointer")
	}
}

func TestTreeView_EmptyState(t *testing.T) {
	tv := NewTreeView()
	out := tv.View()
	if !strings.Contains(out, "No data loaded") {
		t.Error("empty tree view should show 'No data loaded'")
	}
}

func TestTreeView_SetIssues_BuildsForest(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	if len(tv.flatNodes) == 0 {
		t.Error("expected flat nodes to be populated")
	}
}

// --- Forest mode rendering ---

func TestTreeView_ForestMode_RendersRoots(t *testing.T) {
	tv := newTreeViewWithIssues()
	out := tv.View()
	if !strings.Contains(out, "root-1") {
		t.Error("should contain root issue ID")
	}
}

func TestTreeView_ForestMode_RendersChildren(t *testing.T) {
	tv := newTreeViewWithIssues()
	out := tv.View()
	if !strings.Contains(out, "child-1") {
		t.Error("should contain child issue ID")
	}
}

func TestTreeView_ForestMode_RendersTreeChars(t *testing.T) {
	tv := newTreeViewWithIssues()
	out := tv.View()
	// Should contain tree drawing characters
	if !strings.Contains(out, "├") && !strings.Contains(out, "└") {
		t.Error("should contain tree drawing characters (├ or └)")
	}
}

func TestTreeView_StatusIndicators(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "closed", Priority: 1},
		{ID: "b", Status: "in_progress", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
		{ID: "c", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "c", DependsOnID: "b"},
		}},
	})
	out := tv.View()
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

func TestTreeView_MultipleRoots(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{
		{ID: "root-a", Status: "open", Priority: 1},
		{ID: "root-b", Status: "open", Priority: 2},
	})
	out := tv.View()
	if !strings.Contains(out, "root-a") || !strings.Contains(out, "root-b") {
		t.Error("should render multiple roots in forest mode")
	}
}

// --- Rooted mode ---

func TestTreeView_SetRoot_ShowsSubtree(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.SetRoot("root-1")
	out := tv.View()
	if !strings.Contains(out, "root-1") {
		t.Error("should show root issue in rooted mode")
	}
	if !strings.Contains(out, "child-1") {
		t.Error("should show children of root")
	}
}

func TestTreeView_SetRoot_HidesOtherRoots(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{
		{ID: "root-a", Status: "open", Priority: 1},
		{ID: "child-a", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "child-a", DependsOnID: "root-a"},
		}},
		{ID: "root-b", Status: "open", Priority: 2},
	})
	tv.SetRoot("root-a")
	out := tv.View()
	if strings.Contains(out, "root-b") {
		t.Error("should not show other roots in rooted mode")
	}
}

func TestTreeView_ClearRoot_ReturnsToForest(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.SetRoot("root-1")
	tv.ClearRoot()
	out := tv.View()
	// Should be back in forest mode showing all roots
	if tv.rootID != "" {
		t.Error("rootID should be empty after ClearRoot")
	}
	if !strings.Contains(out, "root-1") {
		t.Error("should show all roots after ClearRoot")
	}
}

func TestTreeView_SetRoot_ResetsCursor(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.cursor = 2
	tv.SetRoot("root-1")
	if tv.cursor != 0 {
		t.Error("cursor should reset to 0 after SetRoot")
	}
}

// --- Expand/Collapse ---

func TestTreeView_CollapseKey_HidesChildren(t *testing.T) {
	tv := newTreeViewWithIssues()
	// Cursor on root-1 (first node), press 'c' to collapse
	initialCount := len(tv.flatNodes)
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if len(tv.flatNodes) >= initialCount {
		t.Error("collapsing root should reduce visible nodes")
	}
}

func TestTreeView_ExpandKey_ShowsChildren(t *testing.T) {
	tv := newTreeViewWithIssues()
	// Collapse first
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	collapsed := len(tv.flatNodes)
	// Expand
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if len(tv.flatNodes) <= collapsed {
		t.Error("expanding should show more nodes")
	}
}

func TestTreeView_CollapseLeaf_IsNoop(t *testing.T) {
	tv := newTreeViewWithIssues()
	// Move cursor to a leaf node
	tv.cursor = len(tv.flatNodes) - 1
	before := len(tv.flatNodes)
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if len(tv.flatNodes) != before {
		t.Error("collapsing a leaf should be a no-op")
	}
}

// --- Cursor navigation ---

func TestTreeView_CursorDown(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tv.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", tv.cursor)
	}
}

func TestTreeView_CursorUp(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.cursor = 1
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tv.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", tv.cursor)
	}
}

func TestTreeView_CursorClampBottom(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.cursor = len(tv.flatNodes) - 1
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tv.cursor != len(tv.flatNodes)-1 {
		t.Error("cursor should clamp at bottom")
	}
}

func TestTreeView_CursorClampTop(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tv.cursor != 0 {
		t.Error("cursor should clamp at top")
	}
}

// --- JumpToTop / JumpToBottom ---

func TestTreeView_JumpToTop_MovesCursorToZero(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.cursor = len(tv.flatNodes) - 1 // move to bottom
	tv.JumpToTop()
	if tv.cursor != 0 {
		t.Errorf("expected cursor 0 after JumpToTop, got %d", tv.cursor)
	}
}

func TestTreeView_JumpToBottom_MovesCursorToLast(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.JumpToBottom()
	want := len(tv.flatNodes) - 1
	if tv.cursor != want {
		t.Errorf("expected cursor %d after JumpToBottom, got %d", want, tv.cursor)
	}
}

func TestTreeView_JumpToTop_EmptyNodes_IsNoop(t *testing.T) {
	tv := NewTreeView()
	tv.JumpToTop() // should not panic
	if tv.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", tv.cursor)
	}
}

func TestTreeView_JumpToBottom_EmptyNodes_IsNoop(t *testing.T) {
	tv := NewTreeView()
	tv.JumpToBottom() // should not panic
	if tv.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", tv.cursor)
	}
}

func TestTreeView_ImplementsJumper(t *testing.T) {
	var _ Jumper = NewTreeView()
}

// --- SelectedNodeID ---

func TestTreeView_SelectedNodeID(t *testing.T) {
	tv := newTreeViewWithIssues()
	id := tv.SelectedNodeID()
	if id != "root-1" {
		t.Errorf("expected root-1, got %q", id)
	}
}

func TestTreeView_SelectedNodeID_Empty(t *testing.T) {
	tv := NewTreeView()
	if tv.SelectedNodeID() != "" {
		t.Error("expected empty string when no nodes")
	}
}

// --- Stats ---

func TestTreeView_RendersStats(t *testing.T) {
	tv := newTreeViewWithIssues()
	out := tv.View()
	if !strings.Contains(out, "nodes") {
		t.Error("should show node count in stats")
	}
	if !strings.Contains(out, "roots") {
		t.Error("should show root count in stats")
	}
}

// --- Update edge cases ---

func TestTreeView_Update_UnhandledKey(t *testing.T) {
	tv := NewTreeView()
	cmd := tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if cmd != nil {
		t.Error("expected nil cmd for unhandled key")
	}
}

func TestTreeView_Update_NonKeyMsg(t *testing.T) {
	tv := NewTreeView()
	cmd := tv.Update(RefreshMsg{})
	if cmd != nil {
		t.Error("expected nil cmd for non-key message")
	}
}

// --- Edge cases ---

func TestTreeView_NoTreeNodes(t *testing.T) {
	tv := NewTreeView()
	tv.issues = map[string]datasource.Issue{"a": {ID: "a"}}
	// dag is nil, so flatNodes stays empty
	out := tv.View()
	if !strings.Contains(out, "No tree nodes") {
		t.Error("should show 'No tree nodes' when dag is nil but issues exist")
	}
}

func TestTreeView_RenderNode_MissingIssue(t *testing.T) {
	tv := NewTreeView()
	// issues has one entry so we pass "No data loaded", but flatNodes references a different ID
	tv.issues = map[string]datasource.Issue{"other": {ID: "other"}}
	tv.flatNodes = []flatNode{{id: "missing-1", prefix: "  "}}
	out := tv.View()
	if !strings.Contains(out, "missing-1") {
		t.Error("should render missing node ID")
	}
	if !strings.Contains(out, "?") {
		t.Error("should show ? for missing issue")
	}
}

func TestTreeView_LongTitleTruncated(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1, Title: "This is a very long title that exceeds forty characters and should be truncated"},
	})
	out := tv.View()
	if !strings.Contains(out, "...") {
		t.Error("long title should be truncated with ...")
	}
}

func TestTreeView_CollapsedIndicator(t *testing.T) {
	tv := newTreeViewWithIssues()
	// Collapse root-1
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	out := tv.View()
	if !strings.Contains(out, "+") {
		t.Error("collapsed node should show + indicator")
	}
}

func TestTreeView_RebuildWithNilDAG(t *testing.T) {
	tv := NewTreeView()
	tv.dag = nil
	tv.rebuild()
	if len(tv.flatNodes) != 0 {
		t.Error("rebuild with nil dag should produce no nodes")
	}
}

// --- Helpers ---

func newTreeViewWithIssues() *TreeView {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{
		{ID: "root-1", Status: "open", Priority: 1, Title: "Root One"},
		{ID: "child-1", Status: "open", Priority: 1, Title: "Child One", Dependencies: []datasource.RawDependency{
			{IssueID: "child-1", DependsOnID: "root-1"},
		}},
		{ID: "child-2", Status: "closed", Priority: 2, Title: "Child Two", Dependencies: []datasource.RawDependency{
			{IssueID: "child-2", DependsOnID: "root-1"},
		}},
		{ID: "grandchild-1", Status: "in_progress", Priority: 1, Title: "Grandchild", Dependencies: []datasource.RawDependency{
			{IssueID: "grandchild-1", DependsOnID: "child-1"},
		}},
	})
	return tv
}

func TestTreeView_ViewportScrollsToFollowCursor(t *testing.T) {
	tv := newTreeViewWithIssues()
	// Set a small viewport: height 3 means only 3 lines visible
	tv.Resize(80, 5) // 5 - 2 overhead = 3 viewport lines

	// Move cursor to last node (index 3: root, child-1, grandchild, child-2)
	for i := 0; i < len(tv.flatNodes)-1; i++ {
		tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	// The cursor is on line 2+3=5 in content (stats + blank + 4 nodes).
	// With viewport height 3, offset must be > 0 to show cursor.
	if tv.viewport.YOffset == 0 {
		t.Error("viewport should scroll down to keep cursor visible")
	}
}

func TestTreeView_ScrollsUpWhenCursorAtTop(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.Resize(80, 5) // small viewport

	// Scroll down
	for i := 0; i < len(tv.flatNodes)-1; i++ {
		tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	// Go back to top
	for i := 0; i < len(tv.flatNodes)-1; i++ {
		tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	}
	// Press k again at cursor 0 — should scroll viewport up
	offsetBefore := tv.viewport.YOffset
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tv.viewport.YOffset >= offsetBefore && offsetBefore > 0 {
		t.Error("pressing k at cursor 0 should scroll viewport up")
	}
}

func TestTreeView_ViewportRendersWhenSized(t *testing.T) {
	tv := newTreeViewWithIssues()
	tv.Resize(80, 20)
	out := tv.View()
	if !strings.Contains(out, "root-1") {
		t.Error("sized viewport should still render content with root-1")
	}
}

// --- StatusHints ---

func TestTreeView_ImplementsStatusHinter(t *testing.T) {
	var _ StatusHinter = NewTreeView()
}

func TestTreeView_StatusHints(t *testing.T) {
	tv := NewTreeView()
	hints := tv.StatusHints()

	keys := make(map[string]string)
	for _, h := range hints {
		keys[h.Key] = h.Desc
	}

	for _, k := range []string{"j/k", "e", "c", "enter"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("expected hint for key %q", k)
		}
	}
}

func TestTreeView_Resize_VeryNarrow_ClampsTitleWidth(t *testing.T) {
	tv := NewTreeView()
	tv.Resize(20, 30)
	if tv.titleMaxWidth() < 10 {
		t.Errorf("expected titleMaxWidth >= 10, got %d", tv.titleMaxWidth())
	}
}

func TestTreeView_Resize_AdaptsTitleTruncation(t *testing.T) {
	tv := NewTreeView()
	longTitle := "This is a very long title that should be truncated differently at different widths"
	tv.SetIssues([]datasource.Issue{
		{ID: "root-1", Status: "open", Priority: 1, Title: longTitle},
	})

	tv.Resize(60, 30)
	out60 := tv.View()

	tv.Resize(120, 30)
	out120 := tv.View()

	// Wider terminal should show more of the title
	if len(out120) <= len(out60) {
		t.Errorf("expected wider terminal to show more title text: len at 120w=%d, len at 60w=%d",
			len(out120), len(out60))
	}
}

func TestTreeView_ImplementsStatusInfoer(t *testing.T) {
	var _ StatusInfoer = NewTreeView()
}

func TestTreeView_StatusInfo_ShowsNodeCount(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{
		{ID: "a-1"},
		{ID: "a-2"},
		{ID: "a-3"},
	})
	info := tv.StatusInfo()
	if !strings.Contains(info, "3 nodes") {
		t.Errorf("expected '3 nodes', got: %q", info)
	}
}

func TestTreeView_StatusInfo_NoIssues(t *testing.T) {
	tv := NewTreeView()
	info := tv.StatusInfo()
	if info != "" {
		t.Errorf("expected empty StatusInfo with no issues, got: %q", info)
	}
}

func TestTreeView_StatusInfo_RootIDSet(t *testing.T) {
	tv := NewTreeView()
	tv.issues = map[string]datasource.Issue{"a-1": {ID: "a-1"}}
	tv.rootID = "a-1"
	info := tv.StatusInfo()
	if info != "1 nodes, 1 roots" {
		t.Errorf("expected '1 nodes, 1 roots', got: %q", info)
	}
}

func TestTreeView_Resize_HeightZero_ClampsToOne(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{{ID: "a-1"}})
	tv.Resize(80, 0)
	if tv.viewport.Height != 1 {
		t.Errorf("expected viewport height 1 for zero height, got: %d", tv.viewport.Height)
	}
}

func TestTreeView_Resize_HeightOne_ClampsToOne(t *testing.T) {
	tv := NewTreeView()
	tv.SetIssues([]datasource.Issue{{ID: "a-1"}})
	tv.Resize(80, 1)
	if tv.viewport.Height != 1 {
		t.Errorf("expected viewport height 1 for height=1, got: %d", tv.viewport.Height)
	}
}

// --- Ctrl-D / Ctrl-U page scroll cursor tracking ---

func newLargeTreeView() *TreeView {
	tv := NewTreeView()
	issues := make([]datasource.Issue, 30)
	for i := range issues {
		issues[i] = datasource.Issue{
			ID:       fmt.Sprintf("node-%02d", i),
			Status:   "open",
			Priority: 1,
			Title:    fmt.Sprintf("Node %d", i),
		}
	}
	tv.SetIssues(issues)
	tv.Resize(80, 14) // 14 - 2 overhead = 12 viewport lines
	return tv
}

func TestTreeView_CtrlD_MovesCursorToMiddle(t *testing.T) {
	tv := newLargeTreeView()
	// Cursor starts at 0, send Ctrl-D (half page down)
	tv.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if tv.cursor == 0 {
		t.Error("expected cursor to move from 0 after Ctrl-D")
	}
	// Cursor should be near middle of visible area
	mid := tv.viewport.YOffset + tv.viewport.Height/2 - 2
	if tv.cursor != mid {
		t.Errorf("expected cursor at %d (viewport middle), got %d", mid, tv.cursor)
	}
}

func TestTreeView_CtrlU_MovesCursorToMiddle(t *testing.T) {
	tv := newLargeTreeView()
	// First scroll down
	tv.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	tv.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	// Now scroll back up
	tv.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	mid := tv.viewport.YOffset + tv.viewport.Height/2 - 2
	if tv.cursor != mid {
		t.Errorf("expected cursor at %d (viewport middle), got %d", mid, tv.cursor)
	}
}

func TestTreeView_CtrlD_CursorClampedAtEnd(t *testing.T) {
	tv := newLargeTreeView()
	// Scroll down many times past the end
	for i := 0; i < 20; i++ {
		tv.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	}
	if tv.cursor >= len(tv.flatNodes) {
		t.Errorf("cursor %d should be < item count %d", tv.cursor, len(tv.flatNodes))
	}
}

func TestTreeView_JkAtEdge_DoesNotJumpCursor(t *testing.T) {
	tv := newLargeTreeView()
	// Move cursor to last item
	tv.JumpToBottom()
	lastCursor := tv.cursor
	// Press j at bottom — cursor should stay, viewport may scroll by 1
	tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tv.cursor != lastCursor {
		t.Errorf("j at bottom should keep cursor at %d, got %d", lastCursor, tv.cursor)
	}
}
