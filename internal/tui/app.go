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

// RefreshMsg signals that data should be re-fetched from bd.
type RefreshMsg struct{}

// View is the interface that each tab's view must implement.
type View interface {
	Update(msg tea.Msg) tea.Cmd
	View() string
}

// App is the root Bubble Tea model for Loom.
type App struct {
	activeTab Tab
	showHelp  bool
	watchMode bool
	views     map[Tab]View
}

// NewApp creates a new App with default settings.
func NewApp() App {
	return App{
		activeTab: TabDashboard,
		views:     make(map[Tab]View),
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
			return a, nil
		case "i":
			a.activeTab = TabIssues
			return a, nil
		case "t":
			a.activeTab = TabTree
			return a, nil
		case "c":
			a.activeTab = TabCriticalPath
			return a, nil
		case "r":
			return a, func() tea.Msg { return RefreshMsg{} }
		case "w":
			a.watchMode = !a.watchMode
			return a, nil
		case "?":
			a.showHelp = !a.showHelp
			return a, nil
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	}

	// Delegate to active view
	if v, ok := a.views[a.activeTab]; ok {
		cmd := v.Update(msg)
		return a, cmd
	}
	return a, nil
}

func (a App) View() string {
	var b strings.Builder
	b.WriteString(a.renderTabBar())
	b.WriteString("\n")
	if a.showHelp {
		b.WriteString(a.renderHelp())
		b.WriteString("\n")
	} else if v, ok := a.views[a.activeTab]; ok {
		b.WriteString(v.View())
	}
	return b.String()
}

func (a App) renderHelp() string {
	help := []struct{ key, desc string }{
		{"d", "Dashboard"},
		{"i", "Issues"},
		{"t", "Tree"},
		{"c", "Critical Path"},
		{"r", "Refresh"},
		{"w", "Toggle watch mode"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}
	var lines []string
	for _, h := range help {
		lines = append(lines, "  "+h.key+"  "+h.desc)
	}
	return strings.Join(lines, "\n")
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
