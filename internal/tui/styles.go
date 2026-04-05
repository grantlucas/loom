package tui

import "github.com/charmbracelet/lipgloss"

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
)
