package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grantlucas/loom/internal/datasource"
)

// sortColumn identifies which column the list is sorted by.
type sortColumn int

const (
	sortByPriority sortColumn = iota
	sortByStatus
	sortByID
	sortByType
	sortByAssignee
	sortByTitle
)

// statusOrder defines sort priority for issue statuses.
var statusOrder = map[string]int{
	"in_progress": 0,
	"open":        1,
	"closed":      2,
}

// ListView displays a table of all issues with sorting and navigation.
type ListView struct {
	table   table.Model
	issues  []datasource.Issue
	sortCol sortColumn
	sortKey key.Binding
}

// NewListView creates a new ListView with default settings.
func NewListView() *ListView {
	columns := []table.Column{
		{Title: "ID", Width: 14},
		{Title: "P ▲", Width: 5},
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
		table:   t,
		sortCol: sortByPriority,
		sortKey: key.NewBinding(key.WithKeys("s")),
	}
}

// SetIssues updates the table with the given issues.
func (v *ListView) SetIssues(issues []datasource.Issue) {
	v.issues = make([]datasource.Issue, len(issues))
	copy(v.issues, issues)
	v.sortAndRefresh()
}

func (v *ListView) sortAndRefresh() {
	v.sortIssues()
	v.rebuildRows()
	v.updateColumnHeaders()
}

func (v *ListView) sortIssues() {
	sort.SliceStable(v.issues, func(i, j int) bool {
		a, b := v.issues[i], v.issues[j]
		switch v.sortCol {
		case sortByPriority:
			return a.Priority < b.Priority
		case sortByStatus:
			return statusOrder[a.Status] < statusOrder[b.Status]
		case sortByID:
			return a.ID < b.ID
		case sortByType:
			return a.IssueType < b.IssueType
		case sortByAssignee:
			return a.Assignee < b.Assignee
		default:
			return a.Title < b.Title
		}
	})
}

func statusIndicator(issue datasource.Issue) string {
	switch issue.Status {
	case "closed":
		return "✓"
	case "in_progress":
		return "◐"
	default:
		if issue.DependencyCount > 0 {
			return "●"
		}
		return "○"
	}
}

func (v *ListView) rebuildRows() {
	rows := make([]table.Row, len(v.issues))
	for i, issue := range v.issues {
		rows[i] = table.Row{
			issue.ID,
			fmt.Sprintf("P%d", issue.Priority),
			issue.IssueType,
			statusIndicator(issue) + " " + issue.Status,
			issue.Assignee,
			issue.Title,
		}
	}
	v.table.SetRows(rows)
}

var columnHeaders = [...]string{"ID", "P", "Type", "Status", "Assignee", "Title"}

func (v *ListView) updateColumnHeaders() {
	widths := []int{14, 5, 8, 12, 14, 40}
	cols := make([]table.Column, len(columnHeaders))
	for i, h := range columnHeaders {
		title := h
		if sortColumn(i) == v.sortCol {
			title = h + " ▲"
		}
		cols[i] = table.Column{Title: title, Width: widths[i]}
	}
	v.table.SetColumns(cols)
}

// SelectedIssueID returns the ID of the currently highlighted issue,
// or empty string if no issues are loaded.
func (v *ListView) SelectedIssueID() string {
	row := v.table.SelectedRow()
	if row == nil {
		return ""
	}
	return row[0]
}

// Update handles input messages.
func (v *ListView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, v.sortKey) {
			v.sortCol = (v.sortCol + 1) % (sortByTitle + 1)
			v.sortAndRefresh()
			return nil
		}
	}
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
