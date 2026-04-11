package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
)

const defaultBarMaxWidth = 30

// DashboardView shows project health at a glance.
type DashboardView struct {
	issues      []datasource.Issue
	ready       []datasource.Issue
	barMaxWidth int
	width       int
}

// NewDashboardView creates a new DashboardView.
func NewDashboardView() *DashboardView {
	return &DashboardView{barMaxWidth: defaultBarMaxWidth}
}

// SetIssues updates the full issue list for computing stats.
func (d *DashboardView) SetIssues(issues []datasource.Issue) {
	d.issues = issues
}

// SetReady updates the ready-queue issues.
func (d *DashboardView) SetReady(issues []datasource.Issue) {
	d.ready = issues
}

// Resize adapts the dashboard layout to the given terminal dimensions.
func (d *DashboardView) Resize(width, height int) {
	// Scale bar width: use ~40% of terminal width, clamped to reasonable range
	barWidth := width * 2 / 5
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 80 {
		barWidth = 80
	}
	d.barMaxWidth = barWidth
	d.width = width
}

// StatusHints returns contextual key hints for the status bar.
func (d *DashboardView) StatusHints() []StatusHint {
	return nil
}

// StatusInfo returns contextual info for the secondary status line.
func (d *DashboardView) StatusInfo() string {
	n := len(d.issues)
	if n == 0 {
		return ""
	}
	label := "issues"
	if n == 1 {
		label = "issue"
	}
	return fmt.Sprintf("%d %s", n, label)
}

// Update handles messages. The dashboard has no interactive elements.
func (d *DashboardView) Update(_ tea.Msg) tea.Cmd {
	return nil
}

// View renders the dashboard.
func (d *DashboardView) View() string {
	if len(d.issues) == 0 {
		return "  No issues found"
	}

	var b strings.Builder
	d.renderStatus(&b)
	d.renderInProgress(&b)
	d.renderPriority(&b)
	d.renderReadyQueue(&b)
	d.renderBlocked(&b)
	return b.String()
}

func (d *DashboardView) renderStatus(b *strings.Builder) {
	open, inProgress, closed := d.statusCounts()
	b.WriteString(renderSectionHeader("Status", d.width))
	b.WriteString("\n")
	fmt.Fprintf(b, "  Open: %d   In Progress: %d   Closed: %d\n", open, inProgress, closed)
	b.WriteString("\n")
}

func (d *DashboardView) statusCounts() (open, inProgress, closed int) {
	for _, issue := range d.issues {
		switch issue.Status {
		case "open":
			open++
		case "in_progress":
			inProgress++
		case "closed":
			closed++
		}
	}
	return
}

func (d *DashboardView) inProgressIssues() []datasource.Issue {
	var result []datasource.Issue
	for _, issue := range d.issues {
		if issue.Status == "in_progress" {
			result = append(result, issue)
		}
	}
	return result
}

func (d *DashboardView) renderInProgress(b *strings.Builder) {
	b.WriteString(renderSectionHeader("In Progress", d.width))
	b.WriteString("\n")
	active := d.inProgressIssues()
	if len(active) == 0 {
		b.WriteString("  None\n")
		b.WriteString("\n")
		return
	}
	limit := 5
	if len(active) < limit {
		limit = len(active)
	}
	for _, issue := range active[:limit] {
		assignee := issue.Assignee
		if assignee == "" {
			assignee = "unassigned"
		}
		fmt.Fprintf(b, "  %-14s %s  %-12s %s\n", issue.ID, StyledPriority(issue.Priority), assignee, issue.Title)
	}
	b.WriteString("\n")
}

func (d *DashboardView) renderPriority(b *strings.Builder) {
	dist := d.priorityDistribution()
	b.WriteString(renderSectionHeader("Priority", d.width))
	b.WriteString("\n")

	// Find max count for scaling
	maxCount := 0
	for _, count := range dist {
		if count > maxCount {
			maxCount = count
		}
	}

	// Sort priorities
	priorities := make([]int, 0, len(dist))
	for p := range dist {
		priorities = append(priorities, p)
	}
	sort.Ints(priorities)

	for _, p := range priorities {
		count := dist[p]
		barLen := count
		if maxCount > 0 {
			barLen = (count * d.barMaxWidth) / maxCount
		}
		if barLen < 1 {
			barLen = 1
		}
		bar := dashboardBarStyle.Render(strings.Repeat("█", barLen))
		fmt.Fprintf(b, "  %s %s %d\n", StyledPriority(p), bar, count)
	}
	b.WriteString("\n")
}

func (d *DashboardView) priorityDistribution() map[int]int {
	dist := make(map[int]int)
	for _, issue := range d.issues {
		dist[issue.Priority]++
	}
	return dist
}

func (d *DashboardView) renderReadyQueue(b *strings.Builder) {
	b.WriteString(renderSectionHeader("Ready Queue", d.width))
	b.WriteString("\n")
	if len(d.ready) == 0 {
		b.WriteString("  None\n")
		b.WriteString("\n")
		return
	}
	limit := 5
	if len(d.ready) < limit {
		limit = len(d.ready)
	}
	for _, issue := range d.ready[:limit] {
		fmt.Fprintf(b, "  %-14s %s  %s\n", issue.ID, StyledPriority(issue.Priority), issue.Title)
	}
	b.WriteString("\n")
}

type blockedInfo struct {
	issue     datasource.Issue
	blockedBy []string
}

func (d *DashboardView) renderBlocked(b *strings.Builder) {
	b.WriteString(renderSectionHeader("Blocked", d.width))
	b.WriteString("\n")
	blocked := d.blockedIssues()
	if len(blocked) == 0 {
		b.WriteString("  None\n")
		b.WriteString("\n")
		return
	}
	for _, bi := range blocked {
		fmt.Fprintf(b, "  %-14s Waiting for: %s\n", bi.issue.ID, strings.Join(bi.blockedBy, ", "))
	}
	b.WriteString("\n")
}

func (d *DashboardView) blockedIssues() []blockedInfo {
	statusMap := make(map[string]string, len(d.issues))
	for _, issue := range d.issues {
		statusMap[issue.ID] = issue.Status
	}

	var result []blockedInfo
	for _, issue := range d.issues {
		if issue.Status == "closed" {
			continue
		}
		var blockers []string
		for _, dep := range issue.Dependencies {
			if statusMap[dep.DependsOnID] != "closed" {
				blockers = append(blockers, dep.DependsOnID)
			}
		}
		if len(blockers) > 0 {
			result = append(result, blockedInfo{issue: issue, blockedBy: blockers})
		}
	}
	return result
}

