package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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
	table       table.Model
	issues      []datasource.Issue
	filtered    []datasource.Issue
	sortCol     sortColumn
	sortKey     key.Binding
	filterMode  bool
	filterText  string
	filterInput textinput.Model
	width       int
	height      int
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

	fi := textinput.New()
	fi.Placeholder = "status:open priority:1 text..."
	fi.CharLimit = 100

	return &ListView{
		table:       t,
		sortCol:     sortByPriority,
		sortKey:     key.NewBinding(key.WithKeys("s")),
		filterInput: fi,
	}
}

// SetIssues updates the table with the given issues.
func (v *ListView) SetIssues(issues []datasource.Issue) {
	v.issues = make([]datasource.Issue, len(issues))
	copy(v.issues, issues)
	if v.filterText != "" {
		v.applyFilter()
	} else {
		v.filtered = nil
		v.sortAndRefresh()
	}
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
	issues := v.displayIssues()
	rows := make([]table.Row, len(issues))
	for i, issue := range issues {
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
		if v.filterMode {
			switch msg.Type {
			case tea.KeyEnter:
				v.filterMode = false
				v.filterInput.Blur()
				v.filterText = v.filterInput.Value()
				v.applyFilter()
				return nil
			case tea.KeyEscape:
				v.filterMode = false
				v.filterInput.Blur()
				v.filterText = ""
				v.filterInput.Reset()
				v.applyFilter()
				return nil
			default:
				var cmd tea.Cmd
				v.filterInput, cmd = v.filterInput.Update(msg)
				return cmd
			}
		}

		if key.Matches(msg, v.sortKey) {
			v.sortCol = (v.sortCol + 1) % (sortByTitle + 1)
			v.sortAndRefresh()
			return nil
		}
		if msg.String() == "/" {
			v.filterMode = true
			v.filterInput.Focus()
			return nil
		}
	}
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return cmd
}

func (v *ListView) applyFilter() {
	if v.filterText == "" {
		v.filtered = nil
		v.sortAndRefresh()
		return
	}
	v.filtered = nil
	for _, issue := range v.issues {
		if matchesFilter(issue, v.filterText) {
			v.filtered = append(v.filtered, issue)
		}
	}
	v.sortAndRefresh()
}

type parsedFilter struct {
	fields   map[string]string
	freetext string
}

func parseFilter(text string) parsedFilter {
	pf := parsedFilter{fields: make(map[string]string)}
	var freeWords []string
	for _, token := range strings.Fields(text) {
		if idx := strings.Index(token, ":"); idx > 0 && idx < len(token)-1 {
			pf.fields[strings.ToLower(token[:idx])] = strings.ToLower(token[idx+1:])
		} else {
			freeWords = append(freeWords, token)
		}
	}
	pf.freetext = strings.ToLower(strings.Join(freeWords, " "))
	return pf
}

func matchesFilter(issue datasource.Issue, filterText string) bool {
	pf := parseFilter(filterText)

	if pf.freetext != "" {
		titleLower := strings.ToLower(issue.Title)
		idLower := strings.ToLower(issue.ID)
		if !strings.Contains(titleLower, pf.freetext) && !strings.Contains(idLower, pf.freetext) {
			return false
		}
	}

	for field, value := range pf.fields {
		switch field {
		case "status":
			if strings.ToLower(issue.Status) != value {
				return false
			}
		case "priority":
			if fmt.Sprintf("%d", issue.Priority) != value {
				return false
			}
		case "type":
			if strings.ToLower(issue.IssueType) != value {
				return false
			}
		case "assignee":
			if strings.ToLower(issue.Assignee) != value {
				return false
			}
		}
	}

	return true
}

// View renders the issue list table.
func (v *ListView) View() string {
	var status string
	if v.filterMode {
		return v.table.View() + "\n" + filterPromptStyle.Render("Filter: ") + v.filterInput.View()
	}
	if v.filterText != "" {
		displayed := len(v.displayIssues())
		total := len(v.issues)
		label := "issues"
		if displayed == 1 {
			label = "issue"
		}
		status = statusBarStyle.Render(fmt.Sprintf("%d of %d %s", displayed, total, label))
	} else {
		count := len(v.issues)
		label := "issues"
		if count == 1 {
			label = "issue"
		}
		status = statusBarStyle.Render(fmt.Sprintf("%d %s", count, label))
	}
	return v.table.View() + "\n" + status
}

// Resize adapts the list layout to the given terminal dimensions.
func (v *ListView) Resize(width, height int) {
	v.width = width
	v.height = height
}

func (v *ListView) displayIssues() []datasource.Issue {
	if v.filtered != nil {
		return v.filtered
	}
	return v.issues
}
