package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
)

// Compile-time check: DashboardView must implement View.
var _ View = (*DashboardView)(nil)

func TestNewDashboardView_ReturnsNonNil(t *testing.T) {
	dv := NewDashboardView()
	if dv == nil {
		t.Fatal("NewDashboardView should return a non-nil pointer")
	}
}

func TestDashboardView_Update_ReturnsNilCmd(t *testing.T) {
	dv := NewDashboardView()
	cmd := dv.Update(tea.KeyMsg{})
	if cmd != nil {
		t.Error("DashboardView.Update should return nil")
	}
}

func TestDashboardView_EmptyState(t *testing.T) {
	dv := NewDashboardView()
	out := dv.View()
	if !strings.Contains(out, "No issues found") {
		t.Errorf("empty dashboard should show 'No issues found', got: %s", out)
	}
}

func TestDashboardView_SetIssues_StoresIssues(t *testing.T) {
	dv := NewDashboardView()
	issues := []datasource.Issue{
		{ID: "a", Status: "open"},
		{ID: "b", Status: "closed"},
	}
	dv.SetIssues(issues)
	if len(dv.issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(dv.issues))
	}
}

func TestDashboardView_SetReady_StoresIssues(t *testing.T) {
	dv := NewDashboardView()
	ready := []datasource.Issue{{ID: "r1"}}
	dv.SetReady(ready)
	if len(dv.ready) != 1 {
		t.Errorf("expected 1 ready issue, got %d", len(dv.ready))
	}
}

// --- Status counts ---

func TestDashboardView_StatusCounts(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "1", Status: "open"},
		{ID: "2", Status: "open"},
		{ID: "3", Status: "in_progress"},
		{ID: "4", Status: "closed"},
		{ID: "5", Status: "closed"},
		{ID: "6", Status: "closed"},
	})
	out := dv.View()
	if !strings.Contains(out, "Open: 2") {
		t.Errorf("should show Open: 2, got:\n%s", out)
	}
	if !strings.Contains(out, "In Progress: 1") {
		t.Errorf("should show In Progress: 1, got:\n%s", out)
	}
	if !strings.Contains(out, "Closed: 3") {
		t.Errorf("should show Closed: 3, got:\n%s", out)
	}
}

func TestDashboardView_StatusSection_Header(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{{ID: "1", Status: "open"}})
	out := dv.View()
	if !strings.Contains(out, "Status") {
		t.Error("should contain Status section header")
	}
}

// --- Priority distribution ---

func TestDashboardView_PriorityDistribution(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "1", Priority: 0},
		{ID: "2", Priority: 0},
		{ID: "3", Priority: 1},
		{ID: "4", Priority: 2},
	})
	out := dv.View()
	if !strings.Contains(out, "Priority") {
		t.Error("should contain Priority section header")
	}
	if !strings.Contains(out, "P0") {
		t.Error("should contain P0 label")
	}
	if !strings.Contains(out, "P1") {
		t.Error("should contain P1 label")
	}
	if !strings.Contains(out, "█") {
		t.Error("should contain bar chart characters")
	}
}

func TestDashboardView_PriorityDistribution_SkipsZeroCounts(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "1", Priority: 1},
	})
	out := dv.View()
	if strings.Contains(out, "P0") {
		t.Error("should not show P0 when count is zero")
	}
	if !strings.Contains(out, "P1") {
		t.Error("should show P1 when count is non-zero")
	}
}

func TestDashboardView_PriorityBarScaling(t *testing.T) {
	dv := NewDashboardView()
	issues := make([]datasource.Issue, 0)
	for i := 0; i < 4; i++ {
		issues = append(issues, datasource.Issue{ID: "a", Priority: 0})
	}
	for i := 0; i < 2; i++ {
		issues = append(issues, datasource.Issue{ID: "b", Priority: 1})
	}
	dv.SetIssues(issues)
	out := dv.View()
	// P0 has 4, P1 has 2; P0 bar should be longer
	p0Line := ""
	p1Line := ""
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "P0") {
			p0Line = line
		}
		if strings.Contains(line, "P1") {
			p1Line = line
		}
	}
	p0Bars := strings.Count(p0Line, "█")
	p1Bars := strings.Count(p1Line, "█")
	if p0Bars <= p1Bars {
		t.Errorf("P0 bar (%d) should be longer than P1 bar (%d)", p0Bars, p1Bars)
	}
}

func TestDashboardView_PriorityBarMinWidth(t *testing.T) {
	dv := NewDashboardView()
	// Create 31 issues at P0 and 1 at P1
	// P1: 1*30/31 = 0 → should be clamped to 1
	issues := make([]datasource.Issue, 0, 32)
	for i := 0; i < 31; i++ {
		issues = append(issues, datasource.Issue{ID: "a", Priority: 0})
	}
	issues = append(issues, datasource.Issue{ID: "b", Priority: 1})
	dv.SetIssues(issues)
	out := dv.View()
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "P1") {
			if !strings.Contains(line, "█") {
				t.Error("P1 bar should have at least one block even at minimum width")
			}
			break
		}
	}
}

// --- Ready queue ---

func TestDashboardView_ReadyQueue(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{{ID: "x"}})
	dv.SetReady([]datasource.Issue{
		{ID: "r-1", Priority: 1, Title: "First ready"},
		{ID: "r-2", Priority: 2, Title: "Second ready"},
	})
	out := dv.View()
	if !strings.Contains(out, "Ready Queue") {
		t.Error("should contain Ready Queue section header")
	}
	if !strings.Contains(out, "r-1") {
		t.Error("should contain ready issue ID r-1")
	}
	if !strings.Contains(out, "First ready") {
		t.Error("should contain ready issue title")
	}
}

func TestDashboardView_ReadyQueue_MaxFive(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{{ID: "x"}})
	ready := make([]datasource.Issue, 7)
	for i := range ready {
		ready[i] = datasource.Issue{ID: "r-" + string(rune('a'+i)), Title: "Ready"}
	}
	dv.SetReady(ready)
	out := dv.View()
	// Should only show 5 items
	count := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Ready") && strings.Contains(line, "r-") {
			count++
		}
	}
	if count > 5 {
		t.Errorf("ready queue should show at most 5 items, got %d", count)
	}
}

func TestDashboardView_ReadyQueue_Empty(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{{ID: "x"}})
	out := dv.View()
	if !strings.Contains(out, "Ready Queue") {
		t.Error("should show Ready Queue section even when empty")
	}
	if !strings.Contains(out, "None") {
		t.Error("should show 'None' when ready queue is empty")
	}
}

// --- Blocked issues ---

func TestDashboardView_BlockedIssues(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "blocker-1", Status: "open"},
		{
			ID:     "blocked-1",
			Status: "open",
			Dependencies: []datasource.RawDependency{
				{IssueID: "blocked-1", DependsOnID: "blocker-1"},
			},
		},
	})
	out := dv.View()
	if !strings.Contains(out, "Blocked") {
		t.Error("should contain Blocked section header")
	}
	if !strings.Contains(out, "blocked-1") {
		t.Error("should show blocked issue ID")
	}
	if !strings.Contains(out, "blocker-1") {
		t.Error("should show what's blocking the issue")
	}
}

func TestDashboardView_BlockedIssues_ClosedDepsNotBlocked(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "dep-1", Status: "closed"},
		{
			ID:     "issue-1",
			Status: "open",
			Dependencies: []datasource.RawDependency{
				{IssueID: "issue-1", DependsOnID: "dep-1"},
			},
		},
	})
	out := dv.View()
	// issue-1 depends on dep-1 which is closed, so not blocked
	lines := strings.Split(out, "\n")
	blockedSection := false
	for _, line := range lines {
		if strings.Contains(line, "Blocked") {
			blockedSection = true
		}
		if blockedSection && strings.Contains(line, "issue-1") {
			t.Error("issue-1 should not appear as blocked since its dep is closed")
		}
	}
}

func TestDashboardView_BlockedIssues_ClosedIssueNotBlocked(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "dep-1", Status: "open"},
		{
			ID:     "issue-1",
			Status: "closed",
			Dependencies: []datasource.RawDependency{
				{IssueID: "issue-1", DependsOnID: "dep-1"},
			},
		},
	})
	out := dv.View()
	lines := strings.Split(out, "\n")
	blockedSection := false
	for _, line := range lines {
		if strings.Contains(line, "Blocked") {
			blockedSection = true
		}
		if blockedSection && strings.Contains(line, "issue-1") {
			t.Error("closed issue should not appear as blocked")
		}
	}
}

func TestDashboardView_BlockedIssues_Empty(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "1", Status: "open"},
	})
	out := dv.View()
	if !strings.Contains(out, "Blocked") {
		t.Error("should show Blocked section even when empty")
	}
	if !strings.Contains(out, "None") {
		t.Error("should show 'None' when no blocked issues")
	}
}

// --- Longest chain ---

func TestDashboardView_LongestChain_Linear(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open"},
		{ID: "b", Status: "open", Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
		{ID: "c", Status: "open", Dependencies: []datasource.RawDependency{
			{IssueID: "c", DependsOnID: "b"},
		}},
	})
	out := dv.View()
	if !strings.Contains(out, "Longest chain: 3") {
		t.Errorf("should show longest chain of 3, got:\n%s", out)
	}
}

func TestDashboardView_LongestChain_NoDeps(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open"},
		{ID: "b", Status: "open"},
	})
	out := dv.View()
	if !strings.Contains(out, "Longest chain: 1") {
		t.Errorf("should show longest chain of 1, got:\n%s", out)
	}
}

func TestDashboardView_LongestChain_Cycle(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Dependencies: []datasource.RawDependency{
			{IssueID: "a", DependsOnID: "b"},
		}},
		{ID: "b", Status: "open", Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	// Should not panic or infinite loop
	out := dv.View()
	if !strings.Contains(out, "Longest chain:") {
		t.Error("should still render longest chain stat even with cycles")
	}
}

// --- Total count ---

func TestDashboardView_TotalCount(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "1", Status: "open"},
		{ID: "2", Status: "closed"},
		{ID: "3", Status: "in_progress"},
	})
	out := dv.View()
	if !strings.Contains(out, "Total: 3") {
		t.Errorf("should show total count of 3, got:\n%s", out)
	}
}

// --- Section headers ---

func TestDashboardView_AllSectionHeaders(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{{ID: "1", Status: "open"}})
	out := dv.View()
	for _, header := range []string{"Status", "Priority", "Ready Queue", "Blocked", "Stats"} {
		if !strings.Contains(out, header) {
			t.Errorf("should contain section header %q", header)
		}
	}
}

func TestDashboardView_UsesStyledPriorityInBarChart(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "1", Status: "open", Priority: 0},
	})
	out := dv.View()
	// StyledPriority returns the same text as plain in non-TTY, but the function should be called
	// Verify the output contains "P0" (from StyledPriority)
	if !strings.Contains(out, "P0") {
		t.Error("dashboard should display P0 in priority distribution")
	}
}

func TestDashboardView_UsesStyledPriorityInReadyQueue(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "ready-1", Status: "open", Priority: 2, Title: "Test"},
	})
	dv.SetReady([]datasource.Issue{
		{ID: "ready-1", Status: "open", Priority: 2, Title: "Test"},
	})
	out := dv.View()
	if !strings.Contains(out, "P2") {
		t.Error("ready queue should display P2")
	}
}

// --- StatusHints ---

func TestDashboardView_ImplementsStatusHinter(t *testing.T) {
	var _ StatusHinter = NewDashboardView()
}

func TestDashboardView_StatusHints_Empty(t *testing.T) {
	dv := NewDashboardView()
	hints := dv.StatusHints()

	if len(hints) != 0 {
		t.Errorf("expected no view-specific hints, got %d", len(hints))
	}
}

func TestDashboardView_Resize_VeryNarrow_ClampsBarWidth(t *testing.T) {
	dv := NewDashboardView()
	dv.Resize(10, 30)
	if dv.barMaxWidth < 10 {
		t.Errorf("expected barMaxWidth >= 10 for narrow terminal, got %d", dv.barMaxWidth)
	}
}

func TestDashboardView_Resize_VeryWide_ClampsBarWidth(t *testing.T) {
	dv := NewDashboardView()
	dv.Resize(300, 30)
	if dv.barMaxWidth > 80 {
		t.Errorf("expected barMaxWidth <= 80 for very wide terminal, got %d", dv.barMaxWidth)
	}
}

func TestDashboardView_Resize_BarWidthScalesWithTerminal(t *testing.T) {
	dv := NewDashboardView()
	dv.SetIssues([]datasource.Issue{
		{ID: "1", Status: "open", Priority: 0},
		{ID: "2", Status: "open", Priority: 0},
	})

	dv.Resize(80, 30)
	out80 := dv.View()

	dv.Resize(160, 30)
	out160 := dv.View()

	// Count the bar characters (█) in each output
	bars80 := strings.Count(out80, "█")
	bars160 := strings.Count(out160, "█")

	if bars160 <= bars80 {
		t.Errorf("expected wider terminal to produce wider bars: got %d chars at 160w vs %d at 80w",
			bars160, bars80)
	}
}
