package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/grantlucas/loom/internal/datasource"
)

var (
	// Tab bar styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			BorderBottom(false).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(false).
				Padding(0, 1)

	// Gap fill line extending from tabs to terminal edge
	tabGapStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// Watch mode indicator
	watchIndicatorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("46")).
				Padding(0, 1)

	// Status bar at bottom
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	// Detail view styles
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))

	detailSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("243"))

	detailLabelStyle = lipgloss.NewStyle().
				Faint(true)

	// Relation selection highlight
	relationSelectedStyle = lipgloss.NewStyle().
				Reverse(true)

	// Breadcrumb trail
	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	// Goto prompt
	gotoPromptStyle = lipgloss.NewStyle().
			Bold(true)

	// Dashboard bar chart
	dashboardBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))

	// Error message
	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	// Filter prompt
	filterPromptStyle = lipgloss.NewStyle().
				Bold(true)

	// Status bar hint styles
	hintKeyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	hintDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	hintSepStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	// Status bar container with top border
	statusBarContainerStyle = lipgloss.NewStyle().
				BorderTop(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("236"))

	// Priority color map
	priorityColors = map[int]lipgloss.Color{
		0: lipgloss.Color("196"), // red
		1: lipgloss.Color("208"), // orange
		2: lipgloss.Color("226"), // yellow
		3: lipgloss.Color("33"),  // blue
		4: lipgloss.Color("243"), // gray
	}
)

// PriorityStyle returns a lipgloss style with the foreground color for the given priority level.
func PriorityStyle(priority int) lipgloss.Style {
	color, ok := priorityColors[priority]
	if !ok {
		color = priorityColors[4] // default to gray
	}
	return lipgloss.NewStyle().Foreground(color)
}

// StyledPriority returns a color-coded priority string like "P0", "P1", etc.
func StyledPriority(priority int) string {
	return PriorityStyle(priority).Render(fmt.Sprintf("P%d", priority))
}

// StatusStyle returns a lipgloss style with the foreground color for the given status.
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "closed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	case "in_progress":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	}
}

// StyledStatus returns a color-coded status indicator for an issue.
// Uses ● for open issues with dependencies, ○ for open, ◐ for in_progress, ✓ for closed.
func StyledStatus(issue datasource.Issue) string {
	var icon string
	switch issue.Status {
	case "closed":
		icon = "✓"
	case "in_progress":
		icon = "◐"
	default:
		if issue.DependencyCount > 0 {
			icon = "●"
		} else {
			icon = "○"
		}
	}
	return StatusStyle(issue.Status).Render(icon)
}

// PlainPriority returns a priority string like "P0", "P1", etc. without ANSI styling.
func PlainPriority(priority int) string {
	return fmt.Sprintf("P%d", priority)
}

// PlainStatus returns a status indicator for an issue without ANSI styling.
func PlainStatus(issue datasource.Issue) string {
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

// StyledStatusSimple returns a color-coded status indicator from just a status string.
// Uses ○ for open (no dependency info available), ◐ for in_progress, ✓ for closed.
func StyledStatusSimple(status string) string {
	var icon string
	switch status {
	case "closed":
		icon = "✓"
	case "in_progress":
		icon = "◐"
	default:
		icon = "○"
	}
	return StatusStyle(status).Render(icon)
}

// renderSectionHeader renders a section header like "── Title ─────..." that
// fills to the given width.
func renderSectionHeader(title string, width int) string {
	prefix := "── "
	suffix := " "
	remaining := width - len(prefix) - len(title) - len(suffix)
	if remaining < 3 {
		remaining = 3
	}
	line := prefix + title + suffix + strings.Repeat("─", remaining)
	return detailSectionStyle.Render(line)
}
