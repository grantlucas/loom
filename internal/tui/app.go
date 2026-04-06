package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/grantlucas/loom/internal/datasource"
)

// Tab represents a navigable view tab.
type Tab int

const (
	TabDashboard Tab = iota
	TabIssues
	TabDetail
	TabTree
)

var tabNames = [...]string{
	TabDashboard: "Dashboard",
	TabIssues:    "Issues",
	TabDetail:    "Detail",
	TabTree:      "Tree",
}

var allTabs = []Tab{TabDashboard, TabIssues, TabTree}

// String returns the display name for a tab.
func (t Tab) String() string {
	if int(t) < len(tabNames) {
		return tabNames[t]
	}
	return "Unknown"
}

var tabShortcuts = [...]string{
	TabDashboard: "d",
	TabIssues:    "i",
	TabDetail:    "",
	TabTree:      "t",
}

// Shortcut returns the keyboard shortcut key for a tab, or empty string if none.
func (t Tab) Shortcut() string {
	if int(t) < len(tabShortcuts) {
		return tabShortcuts[t]
	}
	return ""
}

// View is the interface that each tab's view must implement.
type View interface {
	Update(msg tea.Msg) tea.Cmd
	View() string
	Resize(width, height int)
}

// InputCapturer is optionally implemented by views that capture keyboard input
// (e.g. a text filter). When active, global shortcuts should be suppressed.
type InputCapturer interface {
	IsCapturingInput() bool
}

// Jumper is optionally implemented by views that support Vim-style gg/G navigation.
type Jumper interface {
	JumpToTop()
	JumpToBottom()
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
	history  []string
	gPending bool
	width      int
	height     int
	loading    bool
	refreshing bool
	spinner    spinner.Model
}

// NewApp creates a new App wired to the given DataSource.
func NewApp(ds datasource.DataSource, interval time.Duration, watch bool) App {
	views := map[Tab]View{
		TabDashboard: NewDashboardView(),
		TabIssues:    NewListView(),
		TabDetail:    NewDetailView(),
		TabTree:      NewTreeView(),
	}
	s := spinner.New()
	s.Spinner = spinner.Dot

	return App{
		activeTab: TabDashboard,
		views:     views,
		keys:      DefaultKeyMap(),
		ds:        ds,
		interval:  interval,
		watchMode: watch,
		loading: true,
		spinner:   s,
	}
}

func (a App) Init() tea.Cmd {
	cmds := []tea.Cmd{a.fetchIssues(), a.fetchReady(), a.spinner.Tick}
	if a.watchMode {
		cmds = append(cmds, a.scheduleTick())
	}
	return tea.Batch(cmds...)
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
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case IssuesLoadedMsg:
		a.loading = false
		a.refreshing = false
		a.err = nil
		if lv, ok := a.views[TabIssues].(*ListView); ok {
			lv.SetIssues(msg.Issues)
		}
		if dv, ok := a.views[TabDashboard].(*DashboardView); ok {
			dv.SetIssues(msg.Issues)
		}
		if tv, ok := a.views[TabTree].(*TreeView); ok {
			tv.SetIssues(msg.Issues)
		}
		return a, nil

	case ReadyLoadedMsg:
		if dv, ok := a.views[TabDashboard].(*DashboardView); ok {
			dv.SetReady(msg.Issues)
		}
		return a, nil

	case ErrMsg:
		a.loading = false
		a.refreshing = false
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

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentHeight := msg.Height - 3 // reserve 3 lines for status bar (border + hints + info)
		for _, v := range a.views {
			v.Resize(msg.Width, contentHeight)
		}
		return a, nil

	case tea.KeyMsg:
		// If the active view is capturing input, delegate to it directly.
		if v, ok := a.views[a.activeTab]; ok {
			if ic, ok := v.(InputCapturer); ok && ic.IsCapturingInput() {
				cmd := v.Update(msg)
				return a, cmd
			}
		}

		// Handle G (shift+g) — jump to bottom
		if msg.String() == "G" {
			a.gPending = false
			if j, ok := a.views[a.activeTab].(Jumper); ok {
				j.JumpToBottom()
			}
			return a, nil
		}

		// Handle g prefix for gg navigation
		if msg.String() == "g" {
			if a.gPending {
				a.gPending = false
				if j, ok := a.views[a.activeTab].(Jumper); ok {
					j.JumpToTop()
				}
				return a, nil
			}
			a.gPending = true
			return a, nil
		}

		// Any other key cancels gPending and processes normally
		if a.gPending {
			a.gPending = false
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
		case key.Matches(msg, a.keys.Enter) && a.activeTab == TabTree:
			if tv, ok := a.views[TabTree].(*TreeView); ok {
				id := tv.SelectedNodeID()
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
			a.refreshing = true
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
	} else if a.loading {
		b.WriteString("  " + a.spinner.View() + " Loading...")
	} else if a.err != nil {
		b.WriteString(errStyle.Render("  " + friendlyError(a.err)))
	} else {
		if a.activeTab == TabDetail {
			b.WriteString(a.renderBreadcrumb())
			b.WriteString("\n")
		}
		if v, ok := a.views[a.activeTab]; ok {
			b.WriteString(v.View())
		}
	}

	// Status bar — pinned to bottom
	hints := a.globalHints()
	if v, ok := a.views[a.activeTab]; ok {
		if sh, ok := v.(StatusHinter); ok {
			hints = append(sh.StatusHints(), hints...)
		}
	}
	hintsLine := renderStatusBar(hints, a.width)

	// Build info line (below hints)
	var info string
	if a.loading {
		info = a.spinner.View() + " Loading..."
	} else if a.refreshing {
		info = a.spinner.View() + " Refreshing..."
	} else if v, ok := a.views[a.activeTab]; ok {
		if si, ok := v.(StatusInfoer); ok {
			info = si.StatusInfo()
		}
	}
	infoLine := renderInfoLine(info, a.width)

	statusBar := statusBarContainerStyle.Width(a.width).Render(
		hintsLine + "\n" + infoLine,
	)

	content := b.String()
	if a.height <= 0 {
		return content + "\n" + statusBar
	}

	// Place content at top, then pin status bar on the last line
	contentHeight := a.height - 3 // reserve 3 lines for status bar (border + hints + info)
	placed := lipgloss.PlaceVertical(contentHeight, lipgloss.Top, content)
	return placed + "\n" + statusBar
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
	type entry struct{ key, desc string }

	renderSection := func(title string, entries []entry) string {
		var sb strings.Builder
		sb.WriteString(renderSectionHeader(title, a.width))
		sb.WriteString("\n")
		for _, e := range entries {
			sb.WriteString(fmt.Sprintf("  %-10s %s\n", e.key, e.desc))
		}
		return sb.String()
	}

	var b strings.Builder

	// Navigation section
	b.WriteString(renderSection("Navigation", []entry{
		{"d", "Dashboard"},
		{"i", "Issues"},
		{"t", "Tree"},
		{"enter", "Open detail"},
		{"esc", "Back"},
		{"gg", "Jump to top"},
		{"G", "Jump to bottom"},
	}))
	b.WriteString("\n")

	// General section
	b.WriteString(renderSection("General", []entry{
		{"/", "filter (issues view)"},
		{"r", "Refresh data"},
		{"w", "Toggle watch mode"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}))

	// View-specific section
	var viewEntries []entry
	switch a.activeTab {
	case TabIssues:
		viewEntries = []entry{
			{"s", "Cycle sort column"},
			{"S", "Reverse sort direction"},
			{"/", "Filter issues"},
			{"c", "Toggle closed issues"},
		}
	case TabTree:
		viewEntries = []entry{
			{"j/k", "Move cursor"},
			{"e", "expand node"},
			{"c", "Collapse node"},
		}
	case TabDetail:
		viewEntries = []entry{
			{"j/k", "Navigate relations"},
		}
	}

	if len(viewEntries) > 0 {
		b.WriteString("\n")
		b.WriteString(renderSection(a.activeTab.String(), viewEntries))
	}

	return b.String()
}

func (a App) renderTabBar() string {
	var rendered []string
	var widths []int
	var activeIdx = -1

	for i, tab := range allTabs {
		label := tab.String()
		if s := tab.Shortcut(); s != "" {
			label += " (" + s + ")"
		}
		if tab == a.activeTab {
			activeIdx = i
			rendered = append(rendered, activeTabStyle.Render(label))
		} else {
			rendered = append(rendered, inactiveTabStyle.Render(label))
		}
		widths = append(widths, lipgloss.Width(rendered[len(rendered)-1]))
	}
	if a.watchMode {
		rendered = append(rendered, watchIndicatorStyle.Render("WATCH"))
		widths = append(widths, lipgloss.Width(rendered[len(rendered)-1]))
	}

	joined := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	// Build a bottom line: ─ under inactive tabs, spaces under active tab
	var bottom strings.Builder
	for i, w := range widths {
		if i == activeIdx {
			bottom.WriteString(strings.Repeat(" ", w))
		} else {
			bottom.WriteString(strings.Repeat("─", w))
		}
	}
	// Fill remaining width to terminal edge
	joinedWidth := lipgloss.Width(joined)
	if remaining := a.width - joinedWidth; remaining > 0 {
		bottom.WriteString(strings.Repeat("─", remaining))
	}

	return joined + "\n" + tabGapStyle.Render(bottom.String())
}

func (a App) globalHints() []StatusHint {
	if a.showHelp {
		return []StatusHint{
			{Key: "?", Desc: "close help"},
			{Key: "q", Desc: "quit"},
		}
	}
	return []StatusHint{
		{Key: "r", Desc: "refresh"},
		{Key: "gg/G", Desc: "top/bottom"},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	}
}

func friendlyError(err error) string {
	switch {
	case errors.Is(err, datasource.ErrBdNotFound):
		return "bd not found in PATH. Install beads: https://github.com/grantlucas/beads"
	case errors.Is(err, datasource.ErrProjectNotInitialized):
		return "No beads project found. Run 'bd init' to initialize."
	case errors.Is(err, datasource.ErrMalformedResponse):
		return "Unexpected response from bd. Check bd version."
	case errors.Is(err, datasource.ErrDatabaseLocked):
		return "Database locked by another process. Retries exhausted. Close other bd commands."
	default:
		return "Error: " + err.Error()
	}
}
