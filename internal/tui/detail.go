package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
)

// DetailView displays full detail for a single issue with a scrollable viewport.
type DetailView struct {
	detail   *datasource.IssueDetail
	viewport viewport.Model
	loading  bool
	err      error
}

// NewDetailView creates a new DetailView.
func NewDetailView() *DetailView {
	return &DetailView{}
}

// SetDetail stores the issue detail and rebuilds the viewport content.
func (v *DetailView) SetDetail(d *datasource.IssueDetail) {
	v.detail = d
	v.loading = false
	v.err = nil
	v.viewport.SetContent(v.renderContent())
}

// SetLoading puts the view into loading state.
func (v *DetailView) SetLoading() {
	v.loading = true
	v.detail = nil
	v.err = nil
}

// SetError puts the view into error state.
func (v *DetailView) SetError(err error) {
	v.err = err
	v.loading = false
}

// Update handles input messages, delegating to the viewport for scrolling.
func (v *DetailView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

// View renders the detail view.
func (v *DetailView) View() string {
	if v.loading {
		return "  Loading..."
	}
	if v.err != nil {
		return fmt.Sprintf("  Error: %s", v.err.Error())
	}
	if v.detail == nil {
		return "  No issue selected"
	}
	// When viewport has no dimensions (e.g. in tests), render content directly
	if v.viewport.Width == 0 && v.viewport.Height == 0 {
		return v.renderContent()
	}
	return v.viewport.View()
}

// renderContent builds the full content string from the issue detail.
func (v *DetailView) renderContent() string {
	d := v.detail
	if d == nil {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("%s: %s", d.ID, d.Title)))
	b.WriteString("\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	b.WriteString("\n\n")

	// Metadata
	b.WriteString(fmt.Sprintf("%s %s   %s P%d   %s %s\n",
		detailLabelStyle.Render("Status:"), d.Status,
		detailLabelStyle.Render("Priority:"), d.Priority,
		detailLabelStyle.Render("Type:"), d.IssueType,
	))
	b.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		detailLabelStyle.Render("Assignee:"), d.Assignee,
		detailLabelStyle.Render("Owner:"), d.Owner,
	))
	b.WriteString(fmt.Sprintf("%s %s by %s   %s %s\n",
		detailLabelStyle.Render("Created:"), d.CreatedAt.Format("2006-01-02"),
		d.CreatedBy,
		detailLabelStyle.Render("Updated:"), d.UpdatedAt.Format("2006-01-02"),
	))
	if d.Parent != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", detailLabelStyle.Render("Parent:"), d.Parent))
	}

	// Description
	b.WriteString("\n")
	b.WriteString(detailSectionStyle.Render("── Description ──────────────────────────────────"))
	b.WriteString("\n\n")
	b.WriteString(d.Description)
	b.WriteString("\n")

	// Dependencies
	b.WriteString("\n")
	b.WriteString(detailSectionStyle.Render(fmt.Sprintf("── Dependencies (%d) ─────────────────────────────", len(d.Dependencies))))
	b.WriteString("\n\n")
	if len(d.Dependencies) == 0 {
		b.WriteString("  None\n")
	} else {
		for _, dep := range d.Dependencies {
			b.WriteString(fmt.Sprintf("  %s %-14s  %-40s  %s\n",
				relationStatusIndicator(dep.Status),
				dep.ID,
				dep.Title,
				dep.Status,
			))
		}
	}

	// Dependents
	b.WriteString("\n")
	b.WriteString(detailSectionStyle.Render(fmt.Sprintf("── Dependents (%d) ───────────────────────────────", len(d.Dependents))))
	b.WriteString("\n\n")
	if len(d.Dependents) == 0 {
		b.WriteString("  None\n")
	} else {
		for _, dep := range d.Dependents {
			b.WriteString(fmt.Sprintf("  %s %-14s  %-40s  %s\n",
				relationStatusIndicator(dep.Status),
				dep.ID,
				dep.Title,
				dep.Status,
			))
		}
	}

	return b.String()
}

// relationStatusIndicator returns a status indicator character for an expanded relation.
func relationStatusIndicator(status string) string {
	switch status {
	case "closed":
		return "✓"
	case "in_progress":
		return "◐"
	default:
		return "○"
	}
}
