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
	sortAsc     bool
	sortKey     key.Binding
	filterMode  bool
	filterText  string
	filterInput textinput.Model
	hideClosed  bool
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
		sortAsc:     true,
		sortKey:     key.NewBinding(key.WithKeys("s")),
		filterInput: fi,
		hideClosed:  true,
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
		if !v.sortAsc {
			i, j = j, i
		}
		a, b := v.issues[i], v.issues[j]
		switch v.sortCol {
		case sortByPriority:
			if a.Priority != b.Priority {
				return a.Priority < b.Priority
			}
			return statusOrder[a.Status] < statusOrder[b.Status]
		default: // sortByStatus
			sa, sb := statusOrder[a.Status], statusOrder[b.Status]
			if sa != sb {
				return sa < sb
			}
			return a.Priority < b.Priority
		}
	})
}

func (v *ListView) rebuildRows() {
	issues := v.displayIssues()
	rows := make([]table.Row, len(issues))
	for i, issue := range issues {
		rows[i] = table.Row{
			issue.ID,
			PlainPriority(issue.Priority),
			issue.IssueType,
			PlainStatus(issue) + " " + issue.Status,
			issue.Title,
		}
	}
	v.table.SetRows(rows)
}

var columnHeaders = [...]string{"ID", "P", "Type", "Status", "Title"}

// sortColumnIndex maps each sortColumn to its position in columnHeaders.
var sortColumnIndex = map[sortColumn]int{
	sortByPriority: 1, // "P"
	sortByStatus:   3, // "Status"
}

func (v *ListView) updateColumnHeaders() {
	widths := []int{14, 5, 8, 12, 40}
	cols := make([]table.Column, len(columnHeaders))
	for i, h := range columnHeaders {
		title := h
		if i == sortColumnIndex[v.sortCol] {
			if v.sortAsc {
				title = h + " ▲"
			} else {
				title = h + " ▼"
			}
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

// IsCapturingInput returns true when the filter input is focused.
func (v *ListView) IsCapturingInput() bool {
	return v.filterMode
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
			v.sortCol = (v.sortCol + 1) % (sortByStatus + 1)
			v.sortAsc = true
			v.sortAndRefresh()
			return nil
		}
		if msg.String() == "S" {
			v.sortAsc = !v.sortAsc
			v.sortAndRefresh()
			return nil
		}
		if msg.String() == "c" {
			v.hideClosed = !v.hideClosed
			v.rebuildRows()
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

// StatusHints returns contextual key hints for the status bar.
func (v *ListView) StatusHints() []StatusHint {
	if v.filterMode {
		return []StatusHint{
			{Key: "enter", Desc: "apply"},
			{Key: "esc", Desc: "cancel"},
		}
	}
	closedDesc := "show closed"
	if !v.hideClosed {
		closedDesc = "hide closed"
	}
	hints := []StatusHint{
		{Key: "s", Desc: "sort"},
		{Key: "S", Desc: "reverse sort"},
		{Key: "/", Desc: "filter"},
		{Key: "c", Desc: closedDesc},
		{Key: "enter", Desc: "open"},
	}
	if v.filterText != "" {
		hints = append(hints, StatusHint{Key: "esc", Desc: "clear filter"})
	}
	return hints
}

// View renders the issue list table.
func (v *ListView) View() string {
	var status string
	if v.filterMode {
		return v.table.View() + "\n" + filterPromptStyle.Render("Filter: ") + v.filterInput.View()
	}
	displayed := len(v.displayIssues())
	total := len(v.issues)
	if v.filterText != "" || v.hideClosed {
		label := "issues"
		if displayed == 1 {
			label = "issue"
		}
		status = statusBarStyle.Render(fmt.Sprintf("%d of %d %s", displayed, total, label))
	} else {
		label := "issues"
		if total == 1 {
			label = "issue"
		}
		status = statusBarStyle.Render(fmt.Sprintf("%d %s", total, label))
	}
	return v.table.View() + "\n" + status
}

// Fixed column widths for non-Title columns.
var fixedColumnWidths = [...]int{14, 5, 8, 12} // ID, P, Type, Status

const minTitleWidth = 10

// Resize adapts the list layout to the given terminal dimensions.
func (v *ListView) Resize(width, height int) {
	v.width = width
	v.height = height

	fixedTotal := 0
	for _, w := range fixedColumnWidths {
		fixedTotal += w
	}
	titleWidth := width - fixedTotal - 2 // 2 for table borders/padding
	if titleWidth < minTitleWidth {
		titleWidth = minTitleWidth
	}

	cols := v.table.Columns()
	if len(cols) == 5 {
		for i := 0; i < 4; i++ {
			cols[i].Width = fixedColumnWidths[i]
		}
		cols[4].Width = titleWidth
		v.table.SetColumns(cols)
	}
}

func (v *ListView) displayIssues() []datasource.Issue {
	issues := v.issues
	if v.filtered != nil {
		issues = v.filtered
	}
	if v.hideClosed {
		visible := make([]datasource.Issue, 0, len(issues))
		for _, issue := range issues {
			if issue.Status != "closed" {
				visible = append(visible, issue)
			}
		}
		return visible
	}
	return issues
}
