package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grantlucas/loom/internal/datasource"
)

// ListView displays a table of all issues with sorting and navigation.
type ListView struct {
	table  table.Model
	issues []datasource.Issue
}

// NewListView creates a new ListView with default settings.
func NewListView() *ListView {
	columns := []table.Column{
		{Title: "ID", Width: 14},
		{Title: "P", Width: 3},
		{Title: "Type", Width: 8},
		{Title: "Status", Width: 12},
		{Title: "Assignee", Width: 14},
		{Title: "Title", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	return &ListView{
		table: t,
	}
}

// SetIssues updates the table with the given issues.
func (v *ListView) SetIssues(issues []datasource.Issue) {
	v.issues = issues
	rows := make([]table.Row, len(issues))
	for i, issue := range issues {
		rows[i] = table.Row{
			issue.ID,
			fmt.Sprintf("P%d", issue.Priority),
			issue.IssueType,
			issue.Status,
			issue.Assignee,
			issue.Title,
		}
	}
	v.table.SetRows(rows)
}

// Update handles input messages.
func (v *ListView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return cmd
}

// View renders the issue list table.
func (v *ListView) View() string {
	count := len(v.issues)
	label := "issues"
	if count == 1 {
		label = "issue"
	}
	status := statusBarStyle.Render(fmt.Sprintf("%d %s", count, label))
	return v.table.View() + "\n" + status
}
