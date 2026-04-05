package tui

import (
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

func TestRelationStatusIndicator(t *testing.T) {
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
			got := relationStatusIndicator(tt.status)
			if got != tt.want {
				t.Errorf("relationStatusIndicator(%q) = %q, want %q", tt.status, got, tt.want)
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

var errTest = errForTest("test error")

type errForTest string

func (e errForTest) Error() string { return string(e) }
