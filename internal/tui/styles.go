package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Tab bar styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Padding(0, 1)

	tabBarStyle = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

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

	// Filter prompt
	filterPromptStyle = lipgloss.NewStyle().
				Bold(true)

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
