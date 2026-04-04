package tui

import tea "github.com/charmbracelet/bubbletea"

// Tab represents a navigable view tab.
type Tab int

const (
	TabDashboard Tab = iota
	TabIssues
	TabDetail
	TabTree
	TabCriticalPath
)

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
	return a, nil
}

func (a App) View() string {
	return ""
}
