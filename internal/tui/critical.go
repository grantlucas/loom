package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
	"github.com/grantlucas/loom/internal/graph"
)

type criticalSortMode int

const (
	critSortByLength   criticalSortMode = iota
	critSortByPriority
)

// CriticalPathView displays longest blocking chains to completion.
type CriticalPathView struct {
	chains   []graph.Chain
	issues   map[string]datasource.Issue
	cursor   int
	sortMode criticalSortMode
	width    int
	viewport viewport.Model
	sortKey  key.Binding
	priKey   key.Binding
	upKey    key.Binding
	downKey  key.Binding
}

// NewCriticalPathView creates a new CriticalPathView.
func NewCriticalPathView() *CriticalPathView {
	return &CriticalPathView{
		issues: make(map[string]datasource.Issue),
		sortKey: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "sort by length"),
		),
		priKey: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "sort by priority"),
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

// SetIssues builds the dependency DAG and computes critical paths.
func (cv *CriticalPathView) SetIssues(issues []datasource.Issue) {
	cv.issues = make(map[string]datasource.Issue, len(issues))
	priorities := make(map[string]int, len(issues))
	g := graph.NewDAG()

	for _, issue := range issues {
		cv.issues[issue.ID] = issue
		priorities[issue.ID] = issue.Priority
		g.AddNode(issue.ID)
	}

	for _, issue := range issues {
		for _, dep := range issue.Dependencies {
			// dep.DependsOnID blocks dep.IssueID
			// Edge direction: blocker -> blocked
			g.AddEdge(dep.DependsOnID, dep.IssueID)
		}
	}

	cv.chains = graph.CriticalPaths(g, priorities)
	cv.sortChains()
	cv.cursor = 0
	cv.syncViewport()
}

// SelectedNodeID returns the issue ID under the cursor.
func (cv *CriticalPathView) SelectedNodeID() string {
	if len(cv.chains) == 0 {
		return ""
	}
	idx := 0
	for _, chain := range cv.chains {
		for _, node := range chain.Nodes {
			if idx == cv.cursor {
				return node
			}
			idx++
		}
	}
	return ""
}

// Resize adapts the critical path layout to the given terminal dimensions.
func (cv *CriticalPathView) Resize(width, height int) {
	cv.width = width
	contentHeight := height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}
	cv.viewport.Width = width
	cv.viewport.Height = contentHeight
	cv.syncViewport()
}

func (cv *CriticalPathView) titleMaxWidth() int {
	if cv.width <= 0 {
		return 40
	}
	// Subtract space for indent (2), indicator (3), ID (15), priority (4)
	maxW := cv.width - 24
	if maxW < 10 {
		maxW = 10
	}
	return maxW
}

// Update handles key messages for cursor navigation and sort toggling.
func (cv *CriticalPathView) Update(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		cv.viewport, cmd = cv.viewport.Update(msg)
		return cmd
	}

	switch {
	case key.Matches(km, cv.downKey):
		total := cv.totalNodes()
		if cv.cursor < total-1 {
			cv.cursor++
		}
		cv.syncViewport()
		return nil
	case key.Matches(km, cv.upKey):
		if cv.cursor > 0 {
			cv.cursor--
		}
		cv.syncViewport()
		return nil
	case key.Matches(km, cv.priKey):
		cv.sortMode = critSortByPriority
		cv.sortChains()
		cv.cursor = 0
		cv.syncViewport()
		return nil
	case key.Matches(km, cv.sortKey):
		cv.sortMode = critSortByLength
		cv.sortChains()
		cv.cursor = 0
		cv.syncViewport()
		return nil
	}

	var cmd tea.Cmd
	cv.viewport, cmd = cv.viewport.Update(msg)
	return cmd
}

// View renders the critical path display.
func (cv *CriticalPathView) View() string {
	if len(cv.chains) == 0 && len(cv.issues) == 0 {
		return "  No data loaded"
	}
	if len(cv.chains) == 0 {
		return "  No blocking chains found"
	}
	if cv.viewport.Width == 0 && cv.viewport.Height == 0 {
		return cv.renderContent()
	}
	return cv.viewport.View()
}

// renderContent builds the full critical path content string.
func (cv *CriticalPathView) renderContent() string {
	var b strings.Builder
	cv.renderSummary(&b)
	b.WriteString("\n")

	nodeIdx := 0
	for i, chain := range cv.chains {
		b.WriteString(detailSectionStyle.Render(
			fmt.Sprintf("── Chain %d (depth: %d, P%d) ──", i+1, chain.Length(), chain.MaxPriority),
		))
		b.WriteString("\n")
		for _, nodeID := range chain.Nodes {
			line := cv.renderNode(nodeID)
			if nodeIdx == cv.cursor {
				line = relationSelectedStyle.Render(line)
			}
			b.WriteString(line)
			b.WriteString("\n")
			nodeIdx++
		}
		b.WriteString("\n")
	}

	return b.String()
}

// syncViewport updates viewport content and scrolls to keep cursor visible.
func (cv *CriticalPathView) syncViewport() {
	cv.viewport.SetContent(cv.renderContent())
	ensureLineVisible(&cv.viewport, cv.cursorLine())
}

// cursorLine computes the output line number for the current cursor position.
func (cv *CriticalPathView) cursorLine() int {
	// Summary: 2 lines (summary text + blank)
	line := 2
	nodeIdx := 0
	for _, chain := range cv.chains {
		line++ // chain header
		for range chain.Nodes {
			if nodeIdx == cv.cursor {
				return line
			}
			line++
			nodeIdx++
		}
		line++ // blank line after chain
	}
	return line
}

func (cv *CriticalPathView) renderSummary(b *strings.Builder) {
	maxDepth := 0
	p0Goals := 0
	for _, chain := range cv.chains {
		if chain.Length() > maxDepth {
			maxDepth = chain.Length()
		}
		// Check if the sink (last node) is P0
		sinkID := chain.Nodes[len(chain.Nodes)-1]
		if issue, ok := cv.issues[sinkID]; ok && issue.Priority == 0 {
			p0Goals++
		}
	}
	chainWord := "chains"
	if len(cv.chains) == 1 {
		chainWord = "chain"
	}
	b.WriteString(fmt.Sprintf("  %d %s, max depth: %d, P0 goals: %d\n",
		len(cv.chains), chainWord, maxDepth, p0Goals))
}

func (cv *CriticalPathView) renderNode(id string) string {
	issue, ok := cv.issues[id]
	if !ok {
		return fmt.Sprintf("  %s  %-14s", "?", id)
	}
	indicator := StyledStatusSimple(issue.Status)
	title := issue.Title
	maxW := cv.titleMaxWidth()
	if len(title) > maxW {
		title = title[:maxW-3] + "..."
	}
	return fmt.Sprintf("  %s  %-14s %s  %s", indicator, issue.ID, StyledPriority(issue.Priority), title)
}

func (cv *CriticalPathView) totalNodes() int {
	total := 0
	for _, chain := range cv.chains {
		total += chain.Length()
	}
	return total
}

func (cv *CriticalPathView) sortChains() {
	switch cv.sortMode {
	case critSortByPriority:
		sort.Slice(cv.chains, func(i, j int) bool {
			if cv.chains[i].MaxPriority != cv.chains[j].MaxPriority {
				return cv.chains[i].MaxPriority < cv.chains[j].MaxPriority
			}
			return cv.chains[i].Length() > cv.chains[j].Length()
		})
	default: // critSortByLength
		sort.Slice(cv.chains, func(i, j int) bool {
			if cv.chains[i].Length() != cv.chains[j].Length() {
				return cv.chains[i].Length() > cv.chains[j].Length()
			}
			return cv.chains[i].MaxPriority < cv.chains[j].MaxPriority
		})
	}
}
