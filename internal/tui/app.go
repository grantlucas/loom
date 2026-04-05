package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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
	history   []string
	gotoMode  bool
	gotoInput textinput.Model
}

// NewApp creates a new App wired to the given DataSource.
func NewApp(ds datasource.DataSource, interval time.Duration, watch bool) App {
	views := map[Tab]View{
		TabDashboard: NewDashboardView(),
		TabIssues:    NewListView(),
		TabDetail:    NewDetailView(),
	}
	ti := textinput.New()
	ti.Placeholder = "issue ID"
	ti.CharLimit = 30
	return App{
		activeTab: TabDashboard,
		views:     views,
		keys:      DefaultKeyMap(),
		ds:        ds,
		interval:  interval,
		watchMode: watch,
		gotoInput: ti,
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(a.fetchIssues(), a.fetchReady())
}

func tickMsg(t time.Time) tea.Msg {
	return TickMsg(t)
}

func (a App) scheduleTick() tea.Cmd {
	return tea.Tick(a.interval, tickMsg)
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

func (a App) fetchReady() tea.Cmd {
	return func() tea.Msg {
		issues, err := a.ds.ListReady()
		if err != nil {
			return ErrMsg{Err: err}
		}
		return ReadyLoadedMsg{Issues: issues}
	}
}

func (a App) fetchIssueDetail(id string) tea.Cmd {
	return func() tea.Msg {
		detail, err := a.ds.GetIssue(id)
		if err != nil {
			return IssueDetailErrMsg{Err: err}
		}
		return IssueDetailLoadedMsg{Detail: detail}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case IssuesLoadedMsg:
		a.err = nil
		if lv, ok := a.views[TabIssues].(*ListView); ok {
			lv.SetIssues(msg.Issues)
		}
		if dv, ok := a.views[TabDashboard].(*DashboardView); ok {
			dv.SetIssues(msg.Issues)
		}
		return a, nil

	case ReadyLoadedMsg:
		if dv, ok := a.views[TabDashboard].(*DashboardView); ok {
			dv.SetReady(msg.Issues)
		}
		return a, nil

	case ErrMsg:
		a.err = msg.Err
		return a, nil

	case IssueDetailLoadedMsg:
		if dv, ok := a.views[TabDetail].(*DetailView); ok {
			dv.SetDetail(msg.Detail)
		}
		return a, nil

	case IssueDetailErrMsg:
		if dv, ok := a.views[TabDetail].(*DetailView); ok {
			dv.SetError(msg.Err)
		}
		return a, nil

	case TickMsg:
		if a.watchMode {
			return a, tea.Batch(
				a.fetchIssues(),
				a.fetchReady(),
				a.scheduleTick(),
			)
		}
		return a, nil

	case tea.KeyMsg:
		if a.gotoMode {
			switch msg.Type {
			case tea.KeyEnter:
				id := strings.TrimSpace(a.gotoInput.Value())
				a.gotoMode = false
				a.gotoInput.Blur()
				if id == "" {
					return a, nil
				}
				if a.activeTab == TabDetail {
					if dv, ok := a.views[TabDetail].(*DetailView); ok && dv.detail != nil {
						a.history = append(a.history, dv.detail.ID)
					}
				} else {
					a.history = nil
				}
				a.activeTab = TabDetail
				if dv, ok := a.views[TabDetail].(*DetailView); ok {
					dv.SetLoading()
				}
				return a, a.fetchIssueDetail(id)
			case tea.KeyEscape:
				a.gotoMode = false
				a.gotoInput.Blur()
				return a, nil
			default:
				var cmd tea.Cmd
				a.gotoInput, cmd = a.gotoInput.Update(msg)
				return a, cmd
			}
		}

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
		case key.Matches(msg, a.keys.Enter) && a.activeTab == TabIssues:
			if lv, ok := a.views[TabIssues].(*ListView); ok {
				id := lv.SelectedIssueID()
				if id == "" {
					return a, nil
				}
				a.history = nil
				a.activeTab = TabDetail
				if dv, ok := a.views[TabDetail].(*DetailView); ok {
					dv.SetLoading()
				}
				return a, a.fetchIssueDetail(id)
			}
			return a, nil
		case key.Matches(msg, a.keys.Enter) && a.activeTab == TabDetail:
			if dv, ok := a.views[TabDetail].(*DetailView); ok {
				targetID := dv.SelectedRelationID()
				if targetID == "" {
					return a, nil
				}
				if dv.detail != nil {
					a.history = append(a.history, dv.detail.ID)
				}
				dv.SetLoading()
				return a, a.fetchIssueDetail(targetID)
			}
			return a, nil
		case key.Matches(msg, a.keys.Back) && a.activeTab == TabDetail:
			if len(a.history) > 0 {
				prevID := a.history[len(a.history)-1]
				a.history = a.history[:len(a.history)-1]
				if dv, ok := a.views[TabDetail].(*DetailView); ok {
					dv.SetLoading()
				}
				return a, a.fetchIssueDetail(prevID)
			}
			a.activeTab = TabIssues
			return a, nil
		case key.Matches(msg, a.keys.Refresh):
			if inv, ok := a.ds.(interface{ Invalidate() }); ok {
				inv.Invalidate()
			}
			return a, tea.Batch(a.fetchIssues(), a.fetchReady())
		case key.Matches(msg, a.keys.Watch):
			a.watchMode = !a.watchMode
			if a.watchMode {
				return a, a.scheduleTick()
			}
			return a, nil
		case key.Matches(msg, a.keys.Help):
			a.showHelp = !a.showHelp
			return a, nil
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Goto):
			a.gotoMode = true
			a.gotoInput.Reset()
			a.gotoInput.Focus()
			return a, nil
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
	} else if a.gotoMode {
		b.WriteString(gotoPromptStyle.Render("Go to: ") + a.gotoInput.View())
		b.WriteString("\n")
	} else {
		if a.activeTab == TabDetail {
			b.WriteString(a.renderBreadcrumb())
			b.WriteString("\n")
		}
		if v, ok := a.views[a.activeTab]; ok {
			b.WriteString(v.View())
		}
	}
	return b.String()
}

func (a App) renderBreadcrumb() string {
	parts := []string{"Issues"}
	parts = append(parts, a.history...)
	if dv, ok := a.views[TabDetail].(*DetailView); ok && dv.detail != nil {
		parts = append(parts, dv.detail.ID)
	}
	return breadcrumbStyle.Render(strings.Join(parts, " > "))
}

func (a App) renderHelp() string {
	help := []struct{ key, desc string }{
		{"d", "Dashboard"},
		{"i", "Issues"},
		{"t", "Tree"},
		{"c", "Critical Path"},
		{"enter", "Open detail"},
		{"esc", "Back"},
		{"g", "goto issue"},
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
