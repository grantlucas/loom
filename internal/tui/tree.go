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

// flatNode is a pre-rendered tree node with its display properties.
type flatNode struct {
	id       string
	prefix   string // tree drawing chars (├── └── │   etc.)
	depth    int
	expanded bool
	hasKids  bool
}

// TreeView displays an ASCII-rendered dependency tree.
type TreeView struct {
	dag       *graph.DAG
	issues    map[string]datasource.Issue
	flatNodes []flatNode
	collapsed map[string]bool
	cursor    int
	rootID    string // empty = forest mode
	width     int
	upKey     key.Binding
	downKey   key.Binding
	expandKey key.Binding
	collapKey key.Binding
}

// NewTreeView creates a new TreeView.
func NewTreeView() *TreeView {
	return &TreeView{
		issues:    make(map[string]datasource.Issue),
		collapsed: make(map[string]bool),
		upKey: key.NewBinding(
			key.WithKeys("k", "up"),
		),
		downKey: key.NewBinding(
			key.WithKeys("j", "down"),
		),
		expandKey: key.NewBinding(
			key.WithKeys("e"),
		),
		collapKey: key.NewBinding(
			key.WithKeys("c"),
		),
	}
}

// SetIssues builds the DAG and flattens the tree for rendering.
func (tv *TreeView) SetIssues(issues []datasource.Issue) {
	tv.issues = make(map[string]datasource.Issue, len(issues))
	tv.dag = graph.NewDAG()

	for _, issue := range issues {
		tv.issues[issue.ID] = issue
		tv.dag.AddNode(issue.ID)
	}

	for _, issue := range issues {
		for _, dep := range issue.Dependencies {
			tv.dag.AddEdge(dep.DependsOnID, dep.IssueID)
		}
	}

	tv.rebuild()
}

// SetRoot switches to rooted mode showing only the subtree of the given ID.
func (tv *TreeView) SetRoot(id string) {
	tv.rootID = id
	tv.cursor = 0
	tv.rebuild()
}

// ClearRoot switches back to forest mode.
func (tv *TreeView) ClearRoot() {
	tv.rootID = ""
	tv.cursor = 0
	tv.rebuild()
}

// SelectedNodeID returns the issue ID under the cursor.
func (tv *TreeView) SelectedNodeID() string {
	if tv.cursor < 0 || tv.cursor >= len(tv.flatNodes) {
		return ""
	}
	return tv.flatNodes[tv.cursor].id
}

// Resize adapts the tree layout to the given terminal dimensions.
func (tv *TreeView) Resize(width, height int) {
	tv.width = width
}

func (tv *TreeView) titleMaxWidth() int {
	if tv.width <= 0 {
		return 40
	}
	// Subtract space for prefix (~10), collapse char (2), indicator (2), ID (15), priority (4)
	maxW := tv.width - 33
	if maxW < 10 {
		maxW = 10
	}
	return maxW
}

// Update handles key messages.
func (tv *TreeView) Update(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(km, tv.downKey):
		if tv.cursor < len(tv.flatNodes)-1 {
			tv.cursor++
		}
	case key.Matches(km, tv.upKey):
		if tv.cursor > 0 {
			tv.cursor--
		}
	case key.Matches(km, tv.collapKey):
		if tv.cursor >= 0 && tv.cursor < len(tv.flatNodes) {
			node := tv.flatNodes[tv.cursor]
			if node.hasKids {
				tv.collapsed[node.id] = true
				tv.rebuild()
			}
		}
	case key.Matches(km, tv.expandKey):
		if tv.cursor >= 0 && tv.cursor < len(tv.flatNodes) {
			node := tv.flatNodes[tv.cursor]
			delete(tv.collapsed, node.id)
			tv.rebuild()
		}
	}
	return nil
}

// View renders the tree.
func (tv *TreeView) View() string {
	if len(tv.issues) == 0 {
		return "  No data loaded"
	}
	if len(tv.flatNodes) == 0 {
		return "  No tree nodes"
	}

	var b strings.Builder
	tv.renderStats(&b)
	b.WriteString("\n")

	for i, node := range tv.flatNodes {
		line := tv.renderLine(node)
		if i == tv.cursor {
			line = relationSelectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (tv *TreeView) renderStats(b *strings.Builder) {
	totalNodes := len(tv.issues)
	var roots int
	if tv.rootID != "" {
		roots = 1
	} else if tv.dag != nil {
		roots = len(tv.dag.Roots())
	}
	b.WriteString(fmt.Sprintf("  %d nodes, %d roots", totalNodes, roots))
	if tv.rootID != "" {
		b.WriteString(fmt.Sprintf(" (rooted: %s)", tv.rootID))
	}
	b.WriteString("\n")
}

func (tv *TreeView) renderLine(node flatNode) string {
	issue, ok := tv.issues[node.id]
	if !ok {
		return fmt.Sprintf("%s? %s", node.prefix, node.id)
	}
	indicator := treeStatusIndicator(issue)
	collapse := " "
	if node.hasKids {
		if tv.collapsed[node.id] {
			collapse = "+"
		} else {
			collapse = "-"
		}
	}
	title := issue.Title
	maxW := tv.titleMaxWidth()
	if len(title) > maxW {
		title = title[:maxW-3] + "..."
	}
	return fmt.Sprintf("%s%s %s %-14s P%d  %s", node.prefix, collapse, indicator, issue.ID, issue.Priority, title)
}

func treeStatusIndicator(issue datasource.Issue) string {
	switch issue.Status {
	case "closed":
		return "✓"
	case "in_progress":
		return "◐"
	default:
		return "○"
	}
}

// rebuild flattens the tree based on current state (root, collapsed).
func (tv *TreeView) rebuild() {
	tv.flatNodes = nil
	if tv.dag == nil {
		return
	}

	if tv.rootID != "" {
		tv.flattenSubtree(tv.rootID, "", true)
		return
	}

	// Forest mode: render all roots
	roots := tv.dag.Roots()
	sort.Strings(roots)
	for i, root := range roots {
		isLast := i == len(roots)-1
		tv.flattenSubtree(root, "", isLast)
	}
}

func (tv *TreeView) flattenSubtree(id, indent string, isLast bool) {
	children := tv.dag.Successors(id)
	sort.Strings(children)
	hasKids := len(children) > 0

	var prefix string
	if indent == "" && tv.rootID != "" {
		// Root node in rooted mode — no tree chars
		prefix = "  "
	} else if indent == "" {
		// Top-level root in forest mode
		if isLast {
			prefix = "  └── "
		} else {
			prefix = "  ├── "
		}
	} else {
		if isLast {
			prefix = indent + "└── "
		} else {
			prefix = indent + "├── "
		}
	}

	tv.flatNodes = append(tv.flatNodes, flatNode{
		id:       id,
		prefix:   prefix,
		depth:    len(indent) / 4, // approximate depth
		expanded: !tv.collapsed[id],
		hasKids:  hasKids,
	})

	if tv.collapsed[id] || !hasKids {
		return
	}

	var childIndent string
	if indent == "" && tv.rootID != "" {
		childIndent = "    "
	} else if indent == "" {
		if isLast {
			childIndent = "      "
		} else {
			childIndent = "  │   "
		}
	} else {
		if isLast {
			childIndent = indent + "    "
		} else {
			childIndent = indent + "│   "
		}
	}

	for i, child := range children {
		childIsLast := i == len(children)-1
		tv.flattenSubtree(child, childIndent, childIsLast)
	}
}
