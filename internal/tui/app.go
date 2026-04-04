package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Tab represents a navigable view tab.
type Tab int

const (
	TabDashboard Tab = iota
	TabIssues
	TabDetail
	TabTree
	TabCriticalPath
)

var tabNames = [...]string{
	TabDashboard:    "Dashboard",
	TabIssues:       "Issues",
	TabDetail:       "Detail",
	TabTree:         "Tree",
	TabCriticalPath: "Critical Path",
}

var allTabs = []Tab{TabDashboard, TabIssues, TabDetail, TabTree, TabCriticalPath}

// String returns the display name for a tab.
func (t Tab) String() string {
	if int(t) < len(tabNames) {
		return tabNames[t]
	}
	return "Unknown"
}

// App is the root Bubble Tea model for Loom.
type App struct {
	activeTab Tab
}

// NewApp creates a new App with default settings.
func NewApp() App {
	return App{
		activeTab: TabDashboard,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			a.activeTab = TabDashboard
		case "i":
			a.activeTab = TabIssues
		case "t":
			a.activeTab = TabTree
		case "c":
			a.activeTab = TabCriticalPath
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a App) View() string {
	var b strings.Builder
	b.WriteString(a.renderTabBar())
	b.WriteString("\n")
	return b.String()
}

func (a App) renderTabBar() string {
	var tabs []string
	for _, tab := range allTabs {
		if tab == a.activeTab {
			tabs = append(tabs, activeTabStyle.Render(tab.String()))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(tab.String()))
		}
	}
	return tabBarStyle.Render(strings.Join(tabs, ""))
}
