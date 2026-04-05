package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grantlucas/loom/internal/datasource"
)

// mockDataSource implements datasource.DataSource and tracks calls.
type mockDataSource struct {
	issues      []datasource.Issue
	err         error
	callCount   int
	invalidated bool
}

func (m *mockDataSource) ListIssues() ([]datasource.Issue, error) {
	m.callCount++
	return m.issues, m.err
}

func (m *mockDataSource) GetIssue(string) (*datasource.IssueDetail, error) {
	return nil, nil
}

func (m *mockDataSource) ListReady() ([]datasource.Issue, error) {
	return nil, nil
}

func (m *mockDataSource) Invalidate() {
	m.invalidated = true
}

func newTestApp() App {
	return NewApp(&mockDataSource{}, 5*time.Second, false)
}

func newTestAppWithDS(ds datasource.DataSource) App {
	return NewApp(ds, 5*time.Second, false)
}

func TestNewApp_DefaultsToDashboard(t *testing.T) {
	app := newTestApp()
	if app.activeTab != TabDashboard {
		t.Errorf("expected active tab %d (Dashboard), got %d", TabDashboard, app.activeTab)
	}
}

func TestNewApp_StoresDataSource(t *testing.T) {
	ds := &mockDataSource{}
	app := NewApp(ds, 5*time.Second, false)
	if app.ds != ds {
		t.Error("expected NewApp to store the provided DataSource")
	}
}

func TestNewApp_StoresInterval(t *testing.T) {
	app := NewApp(&mockDataSource{}, 10*time.Second, false)
	if app.interval != 10*time.Second {
		t.Errorf("expected interval 10s, got %v", app.interval)
	}
}

func TestNewApp_SetsWatchMode(t *testing.T) {
	app := NewApp(&mockDataSource{}, 5*time.Second, true)
	if !app.watchMode {
		t.Error("expected watch mode to be set when passed true")
	}
}

func TestNewApp_RegistersListView(t *testing.T) {
	app := newTestApp()
	v, ok := app.views[TabIssues]
	if !ok {
		t.Fatal("expected TabIssues view to be registered")
	}
	if _, ok := v.(*ListView); !ok {
		t.Errorf("expected *ListView, got %T", v)
	}
}

func TestNewApp_ImplementsTeaModel(t *testing.T) {
	var _ tea.Model = newTestApp()
}

func keyMsg(r rune) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestApp_TabSwitching(t *testing.T) {
	tests := []struct {
		key  rune
		want Tab
	}{
		{'d', TabDashboard},
		{'i', TabIssues},
		{'t', TabTree},
		{'c', TabCriticalPath},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			app := newTestApp()
			// Start on a different tab to confirm switching works
			app.activeTab = TabCriticalPath
			if tt.key == 'c' {
				app.activeTab = TabDashboard
			}

			model, _ := app.Update(keyMsg(tt.key))
			got := model.(App).activeTab
			if got != tt.want {
				t.Errorf("key %q: expected tab %d, got %d", tt.key, tt.want, got)
			}
		})
	}
}

func TestApp_QuitKey(t *testing.T) {
	app := newTestApp()
	_, cmd := app.Update(keyMsg('q'))
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	// Execute the command to verify it produces tea.QuitMsg
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestApp_ViewRendersTabBar(t *testing.T) {
	app := newTestApp()
	view := app.View()

	// All tab labels should appear in the output
	tabs := []string{"Dashboard", "Issues", "Detail", "Tree", "Critical Path"}
	for _, tab := range tabs {
		if !strings.Contains(view, tab) {
			t.Errorf("expected view to contain tab label %q", tab)
		}
	}
}

func TestApp_ViewHighlightsActiveTab(t *testing.T) {
	// When on Issues tab, the tab bar should render differently for the active tab
	app := newTestApp()
	app.activeTab = TabIssues
	view := app.View()

	// The active tab should be present
	if !strings.Contains(view, "Issues") {
		t.Error("expected view to contain 'Issues'")
	}
}

func TestApp_TabNames(t *testing.T) {
	tests := []struct {
		tab  Tab
		want string
	}{
		{TabDashboard, "Dashboard"},
		{TabIssues, "Issues"},
		{TabDetail, "Detail"},
		{TabTree, "Tree"},
		{TabCriticalPath, "Critical Path"},
	}

	for _, tt := range tests {
		got := tt.tab.String()
		if got != tt.want {
			t.Errorf("Tab(%d).String() = %q, want %q", tt.tab, got, tt.want)
		}
	}
}

func TestApp_HelpToggle(t *testing.T) {
	app := newTestApp()
	if app.showHelp {
		t.Fatal("help should be hidden by default")
	}

	// Toggle on
	model, _ := app.Update(keyMsg('?'))
	app = model.(App)
	if !app.showHelp {
		t.Error("expected help to be visible after pressing ?")
	}

	// Toggle off
	model, _ = app.Update(keyMsg('?'))
	app = model.(App)
	if app.showHelp {
		t.Error("expected help to be hidden after pressing ? again")
	}
}

func TestApp_HelpOverlayInView(t *testing.T) {
	app := newTestApp()
	app.showHelp = true
	view := app.View()

	// Help overlay should show key bindings
	for _, expected := range []string{"d", "i", "t", "c", "r", "q"} {
		if !strings.Contains(view, expected) {
			t.Errorf("help overlay should contain %q", expected)
		}
	}
}

func TestApp_RefreshKey(t *testing.T) {
	app := newTestApp()
	_, cmd := app.Update(keyMsg('r'))
	if cmd == nil {
		t.Fatal("expected refresh command from r key, got nil")
	}
	msg := cmd()
	if _, ok := msg.(RefreshMsg); !ok {
		t.Errorf("expected RefreshMsg, got %T", msg)
	}
}

func TestApp_WatchToggle(t *testing.T) {
	app := newTestApp()
	if app.watchMode {
		t.Fatal("watch mode should be off by default")
	}

	model, _ := app.Update(keyMsg('w'))
	app = model.(App)
	if !app.watchMode {
		t.Error("expected watch mode on after pressing w")
	}

	model, _ = app.Update(keyMsg('w'))
	app = model.(App)
	if app.watchMode {
		t.Error("expected watch mode off after pressing w again")
	}
}

// stubView is a minimal View implementation for testing dispatching.
type stubView struct {
	updateCalled bool
	lastMsg      tea.Msg
	content      string
}

func (v *stubView) Update(msg tea.Msg) tea.Cmd {
	v.updateCalled = true
	v.lastMsg = msg
	return nil
}

func (v *stubView) View() string {
	return v.content
}

func TestApp_ViewDelegatesToActiveView(t *testing.T) {
	app := newTestApp()
	stub := &stubView{content: "dashboard content here"}
	app.views[TabDashboard] = stub

	view := app.View()
	if !strings.Contains(view, "dashboard content here") {
		t.Error("expected View() to include active view's content")
	}
}

func TestApp_UpdateDelegatesToActiveView(t *testing.T) {
	app := newTestApp()
	stub := &stubView{}
	app.views[TabDashboard] = stub

	// Send a non-global key that should be forwarded to the view
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !stub.updateCalled {
		t.Error("expected Update to delegate to active view")
	}
}

func TestApp_ViewSwitchingChangesDelegate(t *testing.T) {
	app := newTestApp()
	dashStub := &stubView{content: "dash"}
	issueStub := &stubView{content: "issues"}
	app.views[TabDashboard] = dashStub
	app.views[TabIssues] = issueStub

	// Switch to issues
	model, _ := app.Update(keyMsg('i'))
	app = model.(App)
	// Re-attach views since App is value type
	app.views[TabDashboard] = dashStub
	app.views[TabIssues] = issueStub

	view := app.View()
	if !strings.Contains(view, "issues") {
		t.Error("expected issues view content after switching to Issues tab")
	}
}

func TestApp_Init_ReturnsNil(t *testing.T) {
	app := newTestApp()
	cmd := app.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestTab_String_OutOfRange(t *testing.T) {
	tab := Tab(99)
	if got := tab.String(); got != "Unknown" {
		t.Errorf("out-of-range Tab.String() = %q, want %q", got, "Unknown")
	}
}

func TestApp_Update_NonKeyMsg_NoView(t *testing.T) {
	// Non-key messages with no registered view should pass through
	app := newTestApp()
	model, cmd := app.Update(RefreshMsg{})
	if cmd != nil {
		t.Error("expected nil cmd for non-key message with no view")
	}
	got := model.(App).activeTab
	if got != TabDashboard {
		t.Error("active tab should be unchanged")
	}
}

func TestApp_Update_NonKeyMsg_WithView(t *testing.T) {
	// Non-key messages should be delegated to the active view
	app := newTestApp()
	stub := &stubView{}
	app.views[TabDashboard] = stub

	app.Update(RefreshMsg{})
	if !stub.updateCalled {
		t.Error("expected non-key message to be delegated to active view")
	}
}

func TestApp_Update_UnhandledKey_NoView(t *testing.T) {
	// Unhandled key with no view registered should return nil cmd
	app := newTestApp()
	_, cmd := app.Update(keyMsg('x'))
	if cmd != nil {
		t.Error("expected nil cmd for unhandled key with no view")
	}
}

func TestApp_Update_UnhandledKey_WithView(t *testing.T) {
	// Unhandled keys should be forwarded to the active view
	app := newTestApp()
	stub := &stubView{}
	app.views[TabDashboard] = stub

	app.Update(keyMsg('x'))
	if !stub.updateCalled {
		t.Error("expected unhandled key to be delegated to active view")
	}
}
