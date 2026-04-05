package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
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
	keys      KeyMap
	ds        datasource.DataSource
	interval  time.Duration
	err       error
}

// NewApp creates a new App wired to the given DataSource.
func NewApp(ds datasource.DataSource, interval time.Duration, watch bool) App {
	views := map[Tab]View{
		TabIssues: NewListView(),
	}
	return App{
		activeTab: TabDashboard,
		views:     views,
		keys:      DefaultKeyMap(),
		ds:        ds,
		interval:  interval,
		watchMode: watch,
	}
}

func (a App) Init() tea.Cmd {
	return a.fetchIssues()
}

func (a App) fetchIssues() tea.Cmd {
	return func() tea.Msg {
		issues, err := a.ds.ListIssues()
		if err != nil {
			return ErrMsg{Err: err}
		}
		return IssuesLoadedMsg{Issues: issues}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Dashboard):
			a.activeTab = TabDashboard
			return a, nil
		case key.Matches(msg, a.keys.Issues):
			a.activeTab = TabIssues
			return a, nil
		case key.Matches(msg, a.keys.Tree):
			a.activeTab = TabTree
			return a, nil
		case key.Matches(msg, a.keys.CriticalPath):
			a.activeTab = TabCriticalPath
			return a, nil
		case key.Matches(msg, a.keys.Refresh):
			return a, func() tea.Msg { return RefreshMsg{} }
		case key.Matches(msg, a.keys.Watch):
			a.watchMode = !a.watchMode
			return a, nil
		case key.Matches(msg, a.keys.Help):
			a.showHelp = !a.showHelp
			return a, nil
		case key.Matches(msg, a.keys.Quit):
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
