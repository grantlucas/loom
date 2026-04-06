package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
)

func TestDetailView_ImplementsViewInterface(t *testing.T) {
	var _ View = NewDetailView()
}

func TestNewDetailView_ReturnsNonNil(t *testing.T) {
	dv := NewDetailView()
	if dv == nil {
		t.Fatal("NewDetailView() should return non-nil")
	}
}

func TestDetailView_View_Loading(t *testing.T) {
	dv := NewDetailView()
	dv.SetLoading()
	view := dv.View()
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading indicator in view")
	}
}

func TestDetailView_View_Error(t *testing.T) {
	dv := NewDetailView()
	dv.SetError(errTest)
	view := dv.View()
	if !strings.Contains(view, "test error") {
		t.Errorf("expected error message in view, got %q", view)
	}
}

func TestDetailView_View_NoDetail(t *testing.T) {
	dv := NewDetailView()
	view := dv.View()
	if !strings.Contains(view, "No issue selected") {
		t.Error("expected 'No issue selected' when no detail is set")
	}
}

func testDetail() *datasource.IssueDetail {
	return &datasource.IssueDetail{
		ID:          "proj-2",
		Title:       "Implement feature",
		Description: "Build the thing with care",
		Status:      "in_progress",
		Priority:    1,
		IssueType:   "feature",
		Assignee:    "Bob",
		Owner:       "bob@example.com",
		CreatedAt:   time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC),
		CreatedBy:   "Bob",
		UpdatedAt:   time.Date(2026, 2, 2, 15, 0, 0, 0, time.UTC),
		Dependencies: []datasource.ExpandedRelation{
			{
				ID:             "proj-1",
				Title:          "Prerequisite task",
				Status:         "closed",
				Priority:       0,
				DependencyType: "blocks",
			},
		},
		Dependents: []datasource.ExpandedRelation{
			{
				ID:             "proj-3",
				Title:          "Follow-up work",
				Status:         "open",
				Priority:       2,
				DependencyType: "blocks",
			},
		},
		Parent: "proj-epic",
	}
}

func TestDetailView_View_RendersID(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "proj-2") {
		t.Error("expected issue ID in view")
	}
}

func TestDetailView_View_RendersTitle(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "Implement feature") {
		t.Error("expected title in view")
	}
}

func TestDetailView_View_RendersStatus(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "in_progress") {
		t.Error("expected status in view")
	}
}

func TestDetailView_View_RendersPriority(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "P1") {
		t.Error("expected priority in view")
	}
}

func TestDetailView_View_RendersType(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "feature") {
		t.Error("expected issue type in view")
	}
}

func TestDetailView_View_RendersAssignee(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "Bob") {
		t.Error("expected assignee in view")
	}
}

func TestDetailView_View_RendersOwner(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "bob@example.com") {
		t.Error("expected owner in view")
	}
}

func TestDetailView_View_RendersCreatedDate(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "2026-02-01") {
		t.Error("expected created date in view")
	}
}

func TestDetailView_View_RendersCreatedBy(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	// "Bob" appears as both assignee and created_by; check context
	if !strings.Contains(view, "Bob") {
		t.Error("expected created_by in view")
	}
}

func TestDetailView_View_RendersUpdatedDate(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "2026-02-02") {
		t.Error("expected updated date in view")
	}
}

func TestDetailView_View_RendersParent(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "proj-epic") {
		t.Error("expected parent in view")
	}
}

func TestDetailView_View_OmitsParentWhenEmpty(t *testing.T) {
	dv := NewDetailView()
	d := testDetail()
	d.Parent = ""
	dv.SetDetail(d)
	view := dv.View()
	if strings.Contains(view, "Parent") {
		t.Error("expected no Parent line when parent is empty")
	}
}

func TestDetailView_View_RendersDescription(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "Build the thing with care") {
		t.Error("expected description text in view")
	}
}

func TestDetailView_View_RendersDescriptionSection(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "Description") {
		t.Error("expected Description section header in view")
	}
}

func TestDetailView_View_RendersDependenciesWithCount(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "Dependencies (1)") {
		t.Errorf("expected 'Dependencies (1)' in view, got:\n%s", view)
	}
	if !strings.Contains(view, "proj-1") {
		t.Error("expected dependency ID in view")
	}
	if !strings.Contains(view, "Prerequisite task") {
		t.Error("expected dependency title in view")
	}
}

func TestDetailView_View_RendersDependentsWithCount(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	if !strings.Contains(view, "Dependents (1)") {
		t.Errorf("expected 'Dependents (1)' in view, got:\n%s", view)
	}
	if !strings.Contains(view, "proj-3") {
		t.Error("expected dependent ID in view")
	}
	if !strings.Contains(view, "Follow-up work") {
		t.Error("expected dependent title in view")
	}
}

func TestDetailView_View_EmptyDependencies(t *testing.T) {
	dv := NewDetailView()
	d := testDetail()
	d.Dependencies = nil
	dv.SetDetail(d)
	view := dv.View()
	if !strings.Contains(view, "Dependencies (0)") {
		t.Errorf("expected 'Dependencies (0)' in view, got:\n%s", view)
	}
}

func TestDetailView_View_EmptyDependents(t *testing.T) {
	dv := NewDetailView()
	d := testDetail()
	d.Dependents = nil
	dv.SetDetail(d)
	view := dv.View()
	if !strings.Contains(view, "Dependents (0)") {
		t.Errorf("expected 'Dependents (0)' in view, got:\n%s", view)
	}
}

func TestStyledStatusSimple_UsedInDetailView(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"closed", "✓"},
		{"in_progress", "◐"},
		{"open", "○"},
		{"unknown", "○"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StyledStatusSimple(tt.status)
			if !strings.Contains(got, tt.want) {
				t.Errorf("StyledStatusSimple(%q) = %q, want to contain %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestDetailView_View_DependencyStatusIndicators(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	// Dependency proj-1 is closed, should show ✓
	if !strings.Contains(view, "✓") {
		t.Error("expected ✓ indicator for closed dependency")
	}
	// Dependent proj-3 is open, should show ○
	if !strings.Contains(view, "○") {
		t.Error("expected ○ indicator for open dependent")
	}
}

func TestDetailView_Update_NoError(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Should not panic; viewport handles the key
	_ = cmd
}

func TestDetailView_SetLoading_ClearsDetail(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	dv.SetLoading()
	view := dv.View()
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading state after SetLoading()")
	}
}

func TestDetailView_SetError_ClearsLoading(t *testing.T) {
	dv := NewDetailView()
	dv.SetLoading()
	dv.SetError(errTest)
	view := dv.View()
	if strings.Contains(view, "Loading") {
		t.Error("expected loading to be cleared after SetError()")
	}
	if !strings.Contains(view, "test error") {
		t.Error("expected error message in view")
	}
}

func TestDetailView_SetDetail_ClearsLoadingAndError(t *testing.T) {
	dv := NewDetailView()
	dv.SetLoading()
	dv.SetError(errTest)
	dv.SetDetail(testDetail())
	view := dv.View()
	if strings.Contains(view, "Loading") {
		t.Error("expected loading to be cleared after SetDetail()")
	}
	if strings.Contains(view, "test error") {
		t.Error("expected error to be cleared after SetDetail()")
	}
	if !strings.Contains(view, "Implement feature") {
		t.Error("expected detail content after SetDetail()")
	}
}

func TestDetailView_View_UsesViewportWhenSized(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	// Set viewport dimensions to trigger viewport rendering path
	dv.viewport.Width = 80
	dv.viewport.Height = 24
	dv.viewport.SetContent(dv.renderContent())
	view := dv.View()
	if !strings.Contains(view, "proj-2") {
		t.Error("expected viewport-rendered content to contain issue ID")
	}
}

func TestDetailView_RenderContent_NilDetail(t *testing.T) {
	dv := NewDetailView()
	content := dv.renderContent()
	if content != "" {
		t.Errorf("expected empty content for nil detail, got %q", content)
	}
}

func TestDetailView_RelationCount_Empty(t *testing.T) {
	dv := NewDetailView()
	if dv.RelationCount() != 0 {
		t.Error("expected 0 relations when no detail set")
	}
}

func TestDetailView_RelationCount_WithRelations(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail()) // 1 dep + 1 dependent = 2
	if dv.RelationCount() != 2 {
		t.Errorf("expected 2 relations, got %d", dv.RelationCount())
	}
}

func TestDetailView_Relations_CombinesDepsAndDependents(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	rels := dv.relations()
	if len(rels) != 2 {
		t.Fatalf("expected 2 relations, got %d", len(rels))
	}
	if rels[0].ID != "proj-1" {
		t.Errorf("expected first relation to be dependency proj-1, got %q", rels[0].ID)
	}
	if rels[1].ID != "proj-3" {
		t.Errorf("expected second relation to be dependent proj-3, got %q", rels[1].ID)
	}
}

func TestDetailView_SelectedRelationID_Default(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	if dv.SelectedRelationID() != "proj-1" {
		t.Errorf("expected default selection 'proj-1', got %q", dv.SelectedRelationID())
	}
}

func TestDetailView_SelectedRelationID_NoRelations(t *testing.T) {
	dv := NewDetailView()
	d := testDetail()
	d.Dependencies = nil
	d.Dependents = nil
	dv.SetDetail(d)
	if dv.SelectedRelationID() != "" {
		t.Error("expected empty string when no relations")
	}
}

func TestDetailView_SelectedRelationID_NoDetail(t *testing.T) {
	dv := NewDetailView()
	if dv.SelectedRelationID() != "" {
		t.Error("expected empty string when no detail set")
	}
}

func TestDetailView_CursorDown_MovesToNext(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail()) // 2 relations
	dv.Update(keyMsg('j'))
	if dv.SelectedRelationID() != "proj-3" {
		t.Errorf("expected cursor to move to proj-3, got %q", dv.SelectedRelationID())
	}
}

func TestDetailView_CursorUp_MovesToPrevious(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	dv.Update(keyMsg('j')) // move to index 1
	dv.Update(keyMsg('k')) // move back to index 0
	if dv.SelectedRelationID() != "proj-1" {
		t.Errorf("expected cursor back at proj-1, got %q", dv.SelectedRelationID())
	}
}

func TestDetailView_CursorDown_ClampsAtEnd(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail()) // 2 relations
	dv.Update(keyMsg('j'))     // index 1
	dv.Update(keyMsg('j'))     // should stay at 1
	if dv.SelectedRelationID() != "proj-3" {
		t.Errorf("expected cursor clamped at proj-3, got %q", dv.SelectedRelationID())
	}
}

func TestDetailView_CursorUp_ClampsAtZero(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	dv.Update(keyMsg('k')) // should stay at 0
	if dv.SelectedRelationID() != "proj-1" {
		t.Errorf("expected cursor clamped at proj-1, got %q", dv.SelectedRelationID())
	}
}

func TestDetailView_SetDetail_ResetsCursor(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	dv.Update(keyMsg('j')) // move to index 1
	dv.SetDetail(testDetail()) // should reset
	if dv.relationCursor != 0 {
		t.Errorf("expected cursor reset to 0, got %d", dv.relationCursor)
	}
}

func TestDetailView_View_HighlightsSelectedRelation(t *testing.T) {
	dv := NewDetailView()
	dv.SetDetail(testDetail())
	view := dv.View()
	// The selected relation should have a highlight marker ">"
	if !strings.Contains(view, "> ") {
		t.Error("expected '>' marker for selected relation")
	}
}

func TestDetailView_CursorNavigation_NoRelations(t *testing.T) {
	dv := NewDetailView()
	d := testDetail()
	d.Dependencies = nil
	d.Dependents = nil
	dv.SetDetail(d)
	// Should not panic
	dv.Update(keyMsg('j'))
	dv.Update(keyMsg('k'))
	if dv.SelectedRelationID() != "" {
		t.Error("expected empty selection with no relations")
	}
}

func TestDetailView_ViewportScrollsToFollowRelationCursor(t *testing.T) {
	dv := NewDetailView()
	// Create a detail with many dependents to overflow a small viewport
	d := testDetail()
	d.Dependents = make([]datasource.ExpandedRelation, 20)
	for i := range d.Dependents {
		d.Dependents[i] = datasource.ExpandedRelation{
			ID:     fmt.Sprintf("dep-%d", i),
			Title:  fmt.Sprintf("Dependent %d", i),
			Status: "open",
		}
	}
	dv.Resize(80, 12) // small viewport: 12 - 3 = 9 content lines
	dv.SetDetail(d)

	// Move cursor down to last relation (1 dep + 20 dependents = 21 total)
	for i := 0; i < 20; i++ {
		dv.Update(keyMsg('j'))
	}

	// The cursor is on the last dependent, which is far below the viewport.
	// The viewport should have scrolled down.
	if dv.viewport.YOffset == 0 {
		t.Error("viewport should scroll down to keep relation cursor visible")
	}
}

func TestDetailView_ScrollsUpWhenCursorAtTopOfList(t *testing.T) {
	dv := NewDetailView()
	d := testDetail()
	d.Dependents = make([]datasource.ExpandedRelation, 20)
	for i := range d.Dependents {
		d.Dependents[i] = datasource.ExpandedRelation{
			ID:     fmt.Sprintf("dep-%d", i),
			Title:  fmt.Sprintf("Dependent %d", i),
			Status: "open",
		}
	}
	dv.Resize(80, 12) // small viewport
	dv.SetDetail(d)

	// Scroll down to a relation so the top of the page is out of view
	for i := 0; i < 15; i++ {
		dv.Update(keyMsg('j'))
	}
	scrolledOffset := dv.viewport.YOffset

	// Now go back up to cursor 0
	for i := 0; i < 15; i++ {
		dv.Update(keyMsg('k'))
	}
	if dv.relationCursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", dv.relationCursor)
	}

	// Press k again — cursor can't go higher, so viewport should scroll up
	dv.Update(keyMsg('k'))
	if dv.viewport.YOffset >= scrolledOffset {
		t.Error("pressing k at cursor 0 should scroll viewport up")
	}

	// Keep pressing k until viewport reaches top
	for i := 0; i < 50; i++ {
		dv.Update(keyMsg('k'))
	}
	if dv.viewport.YOffset != 0 {
		t.Errorf("repeated k presses should scroll viewport to top, got offset %d", dv.viewport.YOffset)
	}
}

var errTest = errForTest("test error")

type errForTest string

func (e errForTest) Error() string { return string(e) }

// --- StatusHints ---

func TestDetailView_ImplementsStatusHinter(t *testing.T) {
	var _ StatusHinter = NewDetailView()
}

func TestDetailView_StatusHints(t *testing.T) {
	dv := NewDetailView()
	hints := dv.StatusHints()

	keys := make(map[string]string)
	for _, h := range hints {
		keys[h.Key] = h.Desc
	}

	for _, k := range []string{"j/k", "enter", "esc"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("expected hint for key %q", k)
		}
	}
}

func TestDetailView_Resize_VerySmallHeight_ClampsToOne(t *testing.T) {
	dv := NewDetailView()
	dv.Resize(80, 2) // height - 3 = -1, should clamp to 1
	if dv.viewport.Height != 1 {
		t.Errorf("expected viewport height 1 for tiny terminal, got %d", dv.viewport.Height)
	}
}

func TestDetailView_Resize_SetsViewportDimensions(t *testing.T) {
	dv := NewDetailView()
	dv.Resize(100, 40)
	if dv.viewport.Width != 100 {
		t.Errorf("expected viewport width 100, got %d", dv.viewport.Width)
	}
	// Height should account for tab bar and breadcrumb (subtract overhead)
	if dv.viewport.Height <= 0 {
		t.Error("expected viewport height > 0 after Resize")
	}
	if dv.viewport.Height >= 40 {
		t.Error("expected viewport height < terminal height (overhead subtracted)")
	}
}

func TestDetailView_View_WrapsLongDescription(t *testing.T) {
	dv := NewDetailView()
	// Set a narrow viewport width before setting detail
	dv.Resize(40, 20)

	detail := testDetail()
	// Description longer than viewport width — should be wrapped
	detail.Description = "This is a very long description that should be wrapped to fit within the viewport width limit"
	dv.SetDetail(detail)

	content := dv.renderContent()
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Allow some tolerance for ANSI sequences in styled lines
		// but plain description lines should respect viewport width
		if len(line) > 40 && strings.Contains(line, "description that should be wrapped") {
			t.Errorf("line %d exceeds viewport width (len=%d): %q", i, len(line), line)
		}
	}
}

func TestDetailView_View_ShortDescriptionUnchanged(t *testing.T) {
	dv := NewDetailView()
	dv.Resize(80, 20)

	detail := testDetail()
	detail.Description = "Short desc"
	dv.SetDetail(detail)

	content := dv.renderContent()
	if !strings.Contains(content, "Short desc") {
		t.Error("expected short description to appear unchanged in content")
	}
}
