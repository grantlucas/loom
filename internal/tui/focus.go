package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
	"github.com/grantlucas/loom/internal/graph"
)

type focusSortMode int

const (
	focusSortByImpact focusSortMode = iota
	focusSortByPriority
	focusSortByDepth
	focusSortByUnblock
)

// FocusView displays ready issues ranked by downstream impact.
type FocusView struct {
	items    []graph.Impact
	issues   map[string]datasource.Issue
	readyIDs []string
	dag      *graph.DAG
	cursor   int
	sortMode focusSortMode
	expanded bool
	width    int

	sortKey key.Binding
	expKey  key.Binding
	upKey   key.Binding
	downKey key.Binding
}

// NewFocusView creates a new FocusView.
func NewFocusView() *FocusView {
	return &FocusView{
		issues:   make(map[string]datasource.Issue),
		expanded: true,
		sortKey: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "cycle sort mode"),
		),
		expKey: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "expand/collapse"),
		),
		upKey: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k", "up"),
		),
		downKey: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j", "down"),
		),
	}
}

// SetIssues builds the dependency DAG from all issues.
func (fv *FocusView) SetIssues(issues []datasource.Issue) {
	fv.issues = make(map[string]datasource.Issue, len(issues))
	fv.dag = graph.NewDAG()

	for _, issue := range issues {
		fv.issues[issue.ID] = issue
		fv.dag.AddNode(issue.ID)
	}
	for _, issue := range issues {
		for _, dep := range issue.Dependencies {
			fv.dag.AddEdge(dep.DependsOnID, dep.IssueID)
		}
	}

	fv.rebuild()
}

// SetReady sets the ready issue IDs and recomputes impact.
func (fv *FocusView) SetReady(ready []datasource.Issue) {
	fv.readyIDs = make([]string, len(ready))
	for i, r := range ready {
		fv.readyIDs[i] = r.ID
	}
	fv.rebuild()
}

func (fv *FocusView) rebuild() {
	if fv.dag == nil || len(fv.readyIDs) == 0 {
		fv.items = nil
		fv.cursor = 0
		return
	}

	priorities := make(map[string]int, len(fv.issues))
	for id, issue := range fv.issues {
		priorities[id] = issue.Priority
	}

	fv.items = graph.DownstreamImpact(fv.dag, fv.readyIDs, priorities)
	fv.sortItems()
	fv.cursor = 0
}

// SelectedNodeID returns the issue ID under the cursor.
func (fv *FocusView) SelectedNodeID() string {
	if len(fv.items) == 0 {
		return ""
	}
	if fv.expanded {
		idx := 0
		for _, item := range fv.items {
			if idx == fv.cursor {
				return item.NodeID
			}
			idx++
			for _, ds := range item.Downstream {
				if idx == fv.cursor {
					return ds
				}
				idx++
			}
		}
		return ""
	}
	if fv.cursor < len(fv.items) {
		return fv.items[fv.cursor].NodeID
	}
	return ""
}

// Resize adapts the focus layout to the given terminal dimensions.
func (fv *FocusView) Resize(width, height int) {
	fv.width = width
}

func (fv *FocusView) titleMaxWidth() int {
	if fv.width <= 0 {
		return 40
	}
	// Subtract space for rank (4), ID (15), priority bracket (5), type (~8), spacing
	maxW := fv.width - 35
	if maxW < 10 {
		maxW = 10
	}
	return maxW
}

func (fv *FocusView) downstreamTitleMaxWidth() int {
	if fv.width <= 0 {
		return 35
	}
	// Subtract space for prefix (8), ID (15), priority bracket (5), spacing
	maxW := fv.width - 30
	if maxW < 10 {
		maxW = 10
	}
	return maxW
}

// Update handles key messages.
func (fv *FocusView) Update(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(km, fv.downKey):
		total := fv.totalLines()
		if fv.cursor < total-1 {
			fv.cursor++
		}
	case key.Matches(km, fv.upKey):
		if fv.cursor > 0 {
			fv.cursor--
		}
	case key.Matches(km, fv.sortKey):
		fv.sortMode = (fv.sortMode + 1) % 4
		fv.sortItems()
		fv.cursor = 0
	case key.Matches(km, fv.expKey):
		fv.expanded = !fv.expanded
		fv.cursor = 0
	}
	return nil
}

// View renders the focus display.
func (fv *FocusView) View() string {
	if len(fv.items) == 0 && len(fv.issues) == 0 {
		return "  No data loaded"
	}
	if len(fv.items) == 0 {
		return "  No ready issues"
	}

	var b strings.Builder
	fv.renderSummary(&b)
	b.WriteString("\n")

	lineIdx := 0
	for i, item := range fv.items {
		line := fv.renderItemLine(i, item)
		if lineIdx == fv.cursor {
			line = relationSelectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
		lineIdx++

		if fv.expanded {
			for j, ds := range item.Downstream {
				isLast := j == len(item.Downstream)-1
				line := fv.renderDownstreamLine(ds, isLast)
				if lineIdx == fv.cursor {
					line = relationSelectedStyle.Render(line)
				}
				b.WriteString(line)
				b.WriteString("\n")
				lineIdx++
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (fv *FocusView) renderSummary(b *strings.Builder) {
	totalBlocked := 0
	for _, item := range fv.items {
		totalBlocked += item.UnblockCount
	}
	issueWord := "issues"
	if len(fv.items) == 1 {
		issueWord = "issue"
	}
	sortName := [...]string{"impact", "priority", "chain depth", "unblock count"}
	b.WriteString(fmt.Sprintf("  %d ready %s, %d total blocked  [sort: %s]\n",
		len(fv.items), issueWord, totalBlocked, sortName[fv.sortMode]))
}

func (fv *FocusView) renderItemLine(rank int, item graph.Impact) string {
	issue, ok := fv.issues[item.NodeID]
	if !ok {
		return fmt.Sprintf("  %d. %s  (unknown)", rank+1, item.NodeID)
	}
	title := issue.Title
	maxW := fv.titleMaxWidth()
	if len(title) > maxW {
		title = title[:maxW-3] + "..."
	}
	if item.UnblockCount == 0 {
		return fmt.Sprintf("  %d. %-14s [P%d] %s  %s\n     Impact: leaf — unblocks nothing",
			rank+1, issue.ID, issue.Priority, title, issue.IssueType)
	}
	return fmt.Sprintf("  %d. %-14s [P%d] %s  %s\n     Impact: unblocks %d issues  Σpri: %d  depth: %d",
		rank+1, issue.ID, issue.Priority, title, issue.IssueType,
		item.UnblockCount, item.PrioritySum, item.MaxDepth)
}

func (fv *FocusView) renderDownstreamLine(id string, isLast bool) string {
	prefix := "     ├→ "
	if isLast {
		prefix = "     └→ "
	}
	issue, ok := fv.issues[id]
	if !ok {
		return prefix + id
	}
	title := issue.Title
	maxW := fv.downstreamTitleMaxWidth()
	if len(title) > maxW {
		title = title[:maxW-3] + "..."
	}
	return fmt.Sprintf("%s%-14s [P%d] %s", prefix, issue.ID, issue.Priority, title)
}

func (fv *FocusView) totalLines() int {
	if !fv.expanded {
		return len(fv.items)
	}
	total := 0
	for _, item := range fv.items {
		total++ // the item itself
		total += len(item.Downstream)
	}
	return total
}

func (fv *FocusView) sortItems() {
	switch fv.sortMode {
	case focusSortByPriority:
		sort.Slice(fv.items, func(i, j int) bool {
			if fv.items[i].OwnPriority != fv.items[j].OwnPriority {
				return fv.items[i].OwnPriority < fv.items[j].OwnPriority
			}
			return fv.items[i].PrioritySum > fv.items[j].PrioritySum
		})
	case focusSortByDepth:
		sort.Slice(fv.items, func(i, j int) bool {
			if fv.items[i].MaxDepth != fv.items[j].MaxDepth {
				return fv.items[i].MaxDepth > fv.items[j].MaxDepth
			}
			return fv.items[i].PrioritySum > fv.items[j].PrioritySum
		})
	case focusSortByUnblock:
		sort.Slice(fv.items, func(i, j int) bool {
			if fv.items[i].UnblockCount != fv.items[j].UnblockCount {
				return fv.items[i].UnblockCount > fv.items[j].UnblockCount
			}
			return fv.items[i].PrioritySum > fv.items[j].PrioritySum
		})
	default: // focusSortByImpact — already sorted by DownstreamImpact
	}
}
