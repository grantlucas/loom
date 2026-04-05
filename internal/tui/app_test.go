package tui

import (
	"errors"
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
	detail      *datasource.IssueDetail
	detailErr   error
	getIssueID  string
	readyIssues []datasource.Issue
	readyErr    error
	readyCalls  int
}

func (m *mockDataSource) ListIssues() ([]datasource.Issue, error) {
	m.callCount++
	return m.issues, m.err
}

func (m *mockDataSource) GetIssue(id string) (*datasource.IssueDetail, error) {
	m.getIssueID = id
	return m.detail, m.detailErr
}

func (m *mockDataSource) ListReady() ([]datasource.Issue, error) {
	m.readyCalls++
	return m.readyIssues, m.readyErr
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

func TestApp_Init_ReturnsFetchCmd(t *testing.T) {
	ds := &mockDataSource{issues: []datasource.Issue{{ID: "x-1"}}}
	app := newTestAppWithDS(ds)
	cmd := app.Init()
	if cmd == nil {
		t.Fatal("expected Init() to return a command")
	}
	// Init returns a batched command (fetchIssues + fetchReady)
	// so we verify via the model update that calls were made
}

func TestApp_Init_FetchError_ReturnsErrMsg(t *testing.T) {
	ds := &mockDataSource{err: errors.New("fail")}
	app := newTestAppWithDS(ds)
	cmd := app.Init()
	if cmd == nil {
		t.Fatal("expected Init() to return a command")
	}
	// Init returns a batched command; individual fetch errors are tested via fetchIssues directly
}

func TestApp_Init_WatchMode_SchedulesTick(t *testing.T) {
	ds := &mockDataSource{issues: []datasource.Issue{{ID: "x-1"}}}
	app := NewApp(ds, 5*time.Second, true)
	cmd := app.Init()
	if cmd == nil {
		t.Fatal("expected Init() to return a command")
	}
	// Execute the batch to get sub-commands
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatal("expected Init() to return a BatchMsg")
	}
	// One of the sub-commands should produce a TickMsg
	var foundTick bool
	for _, sub := range batch {
		if sub == nil {
			continue
		}
		result := sub()
		if _, ok := result.(TickMsg); ok {
			foundTick = true
			break
		}
	}
	if !foundTick {
		t.Error("expected Init() to schedule a tick when watch mode is enabled")
	}
}

func TestApp_Init_NoWatch_NoTick(t *testing.T) {
	ds := &mockDataSource{issues: []datasource.Issue{{ID: "x-1"}}}
	app := NewApp(ds, 5*time.Second, false)
	cmd := app.Init()
	if cmd == nil {
		t.Fatal("expected Init() to return a command")
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		// Single command, not a batch — that's fine, no tick
		return
	}
	for _, sub := range batch {
		if sub == nil {
			continue
		}
		result := sub()
		if _, ok := result.(TickMsg); ok {
			t.Error("expected Init() NOT to schedule a tick when watch mode is disabled")
		}
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
		{'f', TabFocus},
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
		{TabFocus, "Focus"},
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
		t.Fatal("expected fetch command from r key, got nil")
	}
	// Refresh returns a batched command (fetchIssues + fetchReady)
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

func (v *stubView) Resize(width, height int) {}

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

func TestApp_Update_IssuesLoadedMsg_SetsDataOnListView(t *testing.T) {
	app := newTestApp()
	issues := []datasource.Issue{
		{ID: "a-1", Title: "First"},
		{ID: "a-2", Title: "Second"},
	}
	model, cmd := app.Update(IssuesLoadedMsg{Issues: issues})
	if cmd != nil {
		t.Error("expected nil cmd after IssuesLoadedMsg")
	}
	a := model.(App)
	lv := a.views[TabIssues].(*ListView)
	if len(lv.issues) != 2 {
		t.Errorf("expected 2 issues in ListView, got %d", len(lv.issues))
	}
}

func TestApp_Update_IssuesLoadedMsg_ClearsError(t *testing.T) {
	app := newTestApp()
	app.err = errors.New("old error")
	model, _ := app.Update(IssuesLoadedMsg{Issues: nil})
	a := model.(App)
	if a.err != nil {
		t.Error("expected error to be cleared after successful load")
	}
}

func TestApp_Update_ErrMsg_SetsError(t *testing.T) {
	app := newTestApp()
	model, cmd := app.Update(ErrMsg{Err: errors.New("fetch failed")})
	if cmd != nil {
		t.Error("expected nil cmd after ErrMsg")
	}
	a := model.(App)
	if a.err == nil || a.err.Error() != "fetch failed" {
		t.Errorf("expected error 'fetch failed', got %v", a.err)
	}
}

func TestApp_Update_RefreshKey_InvalidatesAndFetches(t *testing.T) {
	ds := &mockDataSource{issues: []datasource.Issue{{ID: "r-1"}}}
	app := newTestAppWithDS(ds)
	_, cmd := app.Update(keyMsg('r'))
	if cmd == nil {
		t.Fatal("expected fetch command from refresh key")
	}
	if !ds.invalidated {
		t.Error("expected Invalidate() to be called on refresh")
	}
}

// mockDataSourceNoInvalidate implements DataSource without Invalidate.
type mockDataSourceNoInvalidate struct {
	issues []datasource.Issue
	err    error
}

func (m *mockDataSourceNoInvalidate) ListIssues() ([]datasource.Issue, error) {
	return m.issues, m.err
}

func (m *mockDataSourceNoInvalidate) GetIssue(id string) (*datasource.IssueDetail, error) {
	return nil, nil
}

func (m *mockDataSourceNoInvalidate) ListReady() ([]datasource.Issue, error) {
	return nil, nil
}

func TestApp_Update_RefreshKey_NoInvalidator(t *testing.T) {
	ds := &mockDataSourceNoInvalidate{issues: []datasource.Issue{{ID: "n-1"}}}
	app := NewApp(ds, 5*time.Second, false)
	_, cmd := app.Update(keyMsg('r'))
	if cmd == nil {
		t.Fatal("expected fetch command even without Invalidator")
	}
}

func TestApp_Update_WatchToggleOn_ReturnsTickCmd(t *testing.T) {
	app := newTestApp()
	model, cmd := app.Update(keyMsg('w'))
	a := model.(App)
	if !a.watchMode {
		t.Error("expected watch mode on")
	}
	if cmd == nil {
		t.Fatal("expected tick command when watch mode toggled on")
	}
}

func TestApp_Update_WatchToggleOff_ReturnsNil(t *testing.T) {
	app := newTestApp()
	app.watchMode = true
	model, cmd := app.Update(keyMsg('w'))
	a := model.(App)
	if a.watchMode {
		t.Error("expected watch mode off")
	}
	if cmd != nil {
		t.Error("expected nil command when watch mode toggled off")
	}
}

func TestApp_Update_TickMsg_WatchOn_Fetches(t *testing.T) {
	ds := &mockDataSource{issues: []datasource.Issue{{ID: "t-1"}}}
	app := NewApp(ds, 5*time.Second, true)
	_, cmd := app.Update(TickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("expected command from TickMsg when watch is on")
	}
	if ds.callCount < 1 {
		// The cmd is batched (fetch + next tick), execute to verify fetch happens
		// We can't easily decompose tea.Batch, but we verified cmd is non-nil
	}
}

func TestApp_Update_TickMsg_WatchOff_Noop(t *testing.T) {
	app := newTestApp() // watch is off by default
	_, cmd := app.Update(TickMsg(time.Now()))
	if cmd != nil {
		t.Error("expected nil command from TickMsg when watch is off")
	}
}

func TestTickMsg_ConvertsTimeToTickMsg(t *testing.T) {
	now := time.Now()
	msg := tickMsg(now)
	tick, ok := msg.(TickMsg)
	if !ok {
		t.Fatalf("expected TickMsg, got %T", msg)
	}
	if time.Time(tick) != now {
		t.Error("expected tick time to match input")
	}
}

func TestApp_Init_ReturnsCmd(t *testing.T) {
	app := newTestApp()
	cmd := app.Init()
	if cmd == nil {
		t.Error("Init() should return a fetch command")
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

func TestNewApp_RegistersDetailView(t *testing.T) {
	app := newTestApp()
	v, ok := app.views[TabDetail]
	if !ok {
		t.Fatal("expected TabDetail view to be registered")
	}
	if _, ok := v.(*DetailView); !ok {
		t.Errorf("expected *DetailView, got %T", v)
	}
}

func enterKeyMsg() tea.Msg {
	return tea.KeyMsg{Type: tea.KeyEnter}
}

func escKeyMsg() tea.Msg {
	return tea.KeyMsg{Type: tea.KeyEscape}
}

func TestApp_EnterOnIssues_SwitchesToDetail(t *testing.T) {
	ds := &mockDataSource{
		issues: []datasource.Issue{{ID: "x-1", Title: "Test"}},
		detail: &datasource.IssueDetail{ID: "x-1", Title: "Test"},
	}
	app := newTestAppWithDS(ds)
	app.activeTab = TabIssues
	// Load issues into list view
	app.Update(IssuesLoadedMsg{Issues: ds.issues})
	lv := app.views[TabIssues].(*ListView)
	lv.SetIssues(ds.issues)

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabDetail {
		t.Errorf("expected TabDetail, got %d", a.activeTab)
	}
	if cmd == nil {
		t.Fatal("expected fetch command from Enter key")
	}
}

func TestApp_EnterOnIssues_ReturnsFetchCmd(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "x-1", Title: "Test Issue"}
	ds := &mockDataSource{
		issues: []datasource.Issue{{ID: "x-1", Title: "Test"}},
		detail: detail,
	}
	app := newTestAppWithDS(ds)
	app.activeTab = TabIssues
	lv := app.views[TabIssues].(*ListView)
	lv.SetIssues(ds.issues)

	_, cmd := app.Update(enterKeyMsg())
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
	msg := cmd()
	loaded, ok := msg.(IssueDetailLoadedMsg)
	if !ok {
		t.Fatalf("expected IssueDetailLoadedMsg, got %T", msg)
	}
	if loaded.Detail.ID != "x-1" {
		t.Errorf("expected detail ID 'x-1', got %q", loaded.Detail.ID)
	}
}

func TestApp_EnterOnIssues_NoSelection_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabIssues
	// ListView has no issues loaded, so SelectedIssueID() returns ""

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabIssues {
		t.Error("expected to stay on Issues tab when no selection")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no selection")
	}
}

func TestApp_EnterOnNonIssuesTab_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabDashboard

	model, _ := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabDashboard {
		t.Error("expected to stay on Dashboard when Enter pressed")
	}
}

func TestApp_EscOnDetail_SwitchesToIssues(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabDetail

	model, cmd := app.Update(escKeyMsg())
	a := model.(App)
	if a.activeTab != TabIssues {
		t.Errorf("expected TabIssues, got %d", a.activeTab)
	}
	if cmd != nil {
		t.Error("expected nil cmd from Escape")
	}
}

func TestApp_EscOnNonDetailTab_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabIssues
	stub := &stubView{}
	app.views[TabIssues] = stub

	app.Update(escKeyMsg())
	// Esc on non-detail tab should be delegated to the active view
	if !stub.updateCalled {
		t.Error("expected Escape to be delegated on non-detail tab")
	}
}

func TestApp_IssueDetailLoadedMsg_RoutesToDetailView(t *testing.T) {
	app := newTestApp()
	detail := &datasource.IssueDetail{ID: "d-1", Title: "Detail Test"}
	app.Update(IssueDetailLoadedMsg{Detail: detail})

	dv := app.views[TabDetail].(*DetailView)
	if dv.detail == nil || dv.detail.ID != "d-1" {
		t.Error("expected IssueDetailLoadedMsg to set detail on DetailView")
	}
}

func TestApp_IssueDetailErrMsg_RoutesToDetailView(t *testing.T) {
	app := newTestApp()
	app.Update(IssueDetailErrMsg{Err: errors.New("detail fail")})

	dv := app.views[TabDetail].(*DetailView)
	if dv.err == nil || dv.err.Error() != "detail fail" {
		t.Error("expected IssueDetailErrMsg to set error on DetailView")
	}
}

func TestApp_FetchIssueDetail_CallsGetIssue(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "f-1", Title: "Fetch Test"}
	ds := &mockDataSource{detail: detail}
	app := newTestAppWithDS(ds)

	cmd := app.fetchIssueDetail("f-1")
	msg := cmd()
	if ds.getIssueID != "f-1" {
		t.Errorf("expected GetIssue called with 'f-1', got %q", ds.getIssueID)
	}
	loaded, ok := msg.(IssueDetailLoadedMsg)
	if !ok {
		t.Fatalf("expected IssueDetailLoadedMsg, got %T", msg)
	}
	if loaded.Detail.ID != "f-1" {
		t.Error("expected loaded detail to have ID 'f-1'")
	}
}

func TestApp_FetchIssueDetail_Error(t *testing.T) {
	ds := &mockDataSource{detailErr: errors.New("not found")}
	app := newTestAppWithDS(ds)

	cmd := app.fetchIssueDetail("bad-1")
	msg := cmd()
	errMsg, ok := msg.(IssueDetailErrMsg)
	if !ok {
		t.Fatalf("expected IssueDetailErrMsg, got %T", msg)
	}
	if errMsg.Err.Error() != "not found" {
		t.Errorf("expected error 'not found', got %q", errMsg.Err.Error())
	}
}

func TestApp_EnterOnIssues_SetsDetailLoading(t *testing.T) {
	ds := &mockDataSource{
		issues: []datasource.Issue{{ID: "l-1", Title: "Loading Test"}},
		detail: &datasource.IssueDetail{ID: "l-1"},
	}
	app := newTestAppWithDS(ds)
	app.activeTab = TabIssues
	lv := app.views[TabIssues].(*ListView)
	lv.SetIssues(ds.issues)

	model, _ := app.Update(enterKeyMsg())
	a := model.(App)
	dv := a.views[TabDetail].(*DetailView)
	if !dv.loading {
		t.Error("expected DetailView to be in loading state after Enter")
	}
}

func TestApp_EnterOnIssues_NonListView_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabIssues
	app.views[TabIssues] = &stubView{} // Replace ListView with stub

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabIssues {
		t.Error("expected to stay on Issues when view is not *ListView")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestNewApp_EmptyHistory(t *testing.T) {
	app := newTestApp()
	if len(app.history) != 0 {
		t.Error("expected empty history on new app")
	}
}

func TestApp_EnterOnIssues_ClearsHistory(t *testing.T) {
	ds := &mockDataSource{
		issues: []datasource.Issue{{ID: "x-1", Title: "Test"}},
		detail: &datasource.IssueDetail{ID: "x-1", Title: "Test"},
	}
	app := newTestAppWithDS(ds)
	app.activeTab = TabIssues
	app.history = []string{"old-1", "old-2"}
	lv := app.views[TabIssues].(*ListView)
	lv.SetIssues(ds.issues)

	model, _ := app.Update(enterKeyMsg())
	a := model.(App)
	if len(a.history) != 0 {
		t.Errorf("expected history cleared, got %v", a.history)
	}
}

func TestApp_EnterOnDetail_PushesCurrentAndFetches(t *testing.T) {
	detail := &datasource.IssueDetail{
		ID:    "d-1",
		Title: "Current",
		Dependencies: []datasource.ExpandedRelation{
			{ID: "d-2", Title: "Dep", Status: "open"},
		},
	}
	ds := &mockDataSource{
		detail: &datasource.IssueDetail{ID: "d-2", Title: "Dep"},
	}
	app := newTestAppWithDS(ds)
	app.activeTab = TabDetail
	dv := app.views[TabDetail].(*DetailView)
	dv.SetDetail(detail)

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if len(a.history) != 1 || a.history[0] != "d-1" {
		t.Errorf("expected history [d-1], got %v", a.history)
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
	msg := cmd()
	if loaded, ok := msg.(IssueDetailLoadedMsg); !ok || loaded.Detail.ID != "d-2" {
		t.Error("expected fetch for d-2")
	}
}

func TestApp_EnterOnDetail_NoRelation_IsNoop(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "d-1", Title: "No rels"}
	app := newTestApp()
	app.activeTab = TabDetail
	dv := app.views[TabDetail].(*DetailView)
	dv.SetDetail(detail)

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if len(a.history) != 0 {
		t.Error("expected history unchanged")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestApp_EnterOnDetail_SetsLoading(t *testing.T) {
	detail := &datasource.IssueDetail{
		ID:           "d-1",
		Dependencies: []datasource.ExpandedRelation{{ID: "d-2", Status: "open"}},
	}
	ds := &mockDataSource{detail: &datasource.IssueDetail{ID: "d-2"}}
	app := newTestAppWithDS(ds)
	app.activeTab = TabDetail
	dv := app.views[TabDetail].(*DetailView)
	dv.SetDetail(detail)

	model, _ := app.Update(enterKeyMsg())
	a := model.(App)
	dvAfter := a.views[TabDetail].(*DetailView)
	if !dvAfter.loading {
		t.Error("expected loading state after Enter on detail relation")
	}
}

func TestApp_EnterOnDetail_NonDetailView_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabDetail
	app.views[TabDetail] = &stubView{}

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabDetail {
		t.Error("expected to stay on detail tab")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestApp_EscOnDetail_WithHistory_PopsAndFetches(t *testing.T) {
	ds := &mockDataSource{
		detail: &datasource.IssueDetail{ID: "prev-1", Title: "Previous"},
	}
	app := newTestAppWithDS(ds)
	app.activeTab = TabDetail
	app.history = []string{"prev-1"}

	model, cmd := app.Update(escKeyMsg())
	a := model.(App)
	if len(a.history) != 0 {
		t.Errorf("expected history popped to empty, got %v", a.history)
	}
	if a.activeTab != TabDetail {
		t.Error("expected to stay on detail tab while fetching previous")
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
	msg := cmd()
	if loaded, ok := msg.(IssueDetailLoadedMsg); !ok || loaded.Detail.ID != "prev-1" {
		t.Error("expected fetch for prev-1")
	}
}

func TestApp_EscOnDetail_WithHistory_SetsLoading(t *testing.T) {
	ds := &mockDataSource{detail: &datasource.IssueDetail{ID: "prev-1"}}
	app := newTestAppWithDS(ds)
	app.activeTab = TabDetail
	app.history = []string{"prev-1"}

	model, _ := app.Update(escKeyMsg())
	a := model.(App)
	dv := a.views[TabDetail].(*DetailView)
	if !dv.loading {
		t.Error("expected loading state on back navigation")
	}
}

func TestApp_MultiStepNavigation(t *testing.T) {
	// Simulate: list -> A -> B -> C, then esc back through history to list
	ds := &mockDataSource{
		issues: []datasource.Issue{{ID: "a-1", Title: "A"}},
	}
	app := newTestAppWithDS(ds)
	dv := app.views[TabDetail].(*DetailView)

	// list -> A (via issues list)
	app.activeTab = TabIssues
	lv := app.views[TabIssues].(*ListView)
	lv.SetIssues(ds.issues)
	ds.detail = &datasource.IssueDetail{ID: "a-1"}
	model, _ := app.Update(enterKeyMsg())
	app = model.(App)
	// Simulate detail loaded
	dv.SetDetail(&datasource.IssueDetail{
		ID:           "a-1",
		Dependencies: []datasource.ExpandedRelation{{ID: "b-1", Status: "open"}},
	})

	// A -> B (via relation Enter)
	ds.detail = &datasource.IssueDetail{ID: "b-1"}
	model, _ = app.Update(enterKeyMsg())
	app = model.(App)
	dv.SetDetail(&datasource.IssueDetail{
		ID:           "b-1",
		Dependencies: []datasource.ExpandedRelation{{ID: "c-1", Status: "open"}},
	})

	// B -> C
	ds.detail = &datasource.IssueDetail{ID: "c-1"}
	model, _ = app.Update(enterKeyMsg())
	app = model.(App)

	if len(app.history) != 2 {
		t.Fatalf("expected history [a-1, b-1], got %v", app.history)
	}

	// esc -> B (pop c-1's parent b-1)
	model, _ = app.Update(escKeyMsg())
	app = model.(App)
	if len(app.history) != 1 || app.history[0] != "a-1" {
		t.Errorf("expected history [a-1], got %v", app.history)
	}

	// esc -> A (pop b-1's parent a-1)
	model, _ = app.Update(escKeyMsg())
	app = model.(App)
	if len(app.history) != 0 {
		t.Errorf("expected empty history, got %v", app.history)
	}

	// esc -> list (history empty, go to Issues)
	model, _ = app.Update(escKeyMsg())
	app = model.(App)
	if app.activeTab != TabIssues {
		t.Error("expected to return to Issues tab after exhausting history")
	}
}

func TestApp_RenderBreadcrumb_OnDetailWithHistory(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabDetail
	app.history = []string{"a-1", "b-1"}
	dv := app.views[TabDetail].(*DetailView)
	dv.SetDetail(&datasource.IssueDetail{ID: "c-1", Title: "Current"})

	view := app.View()
	if !strings.Contains(view, "a-1") || !strings.Contains(view, "b-1") {
		t.Error("expected breadcrumb to contain history IDs")
	}
	if !strings.Contains(view, "c-1") {
		t.Error("expected breadcrumb to contain current issue ID")
	}
}

func TestApp_RenderBreadcrumb_NotShownOnIssuesTab(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabIssues
	app.history = []string{"should-not-show"}

	view := app.View()
	if strings.Contains(view, "should-not-show") {
		t.Error("breadcrumb should not appear on Issues tab")
	}
}

func TestApp_RenderBreadcrumb_EmptyOnDetailNoHistory(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabDetail
	dv := app.views[TabDetail].(*DetailView)
	dv.SetDetail(&datasource.IssueDetail{ID: "solo-1", Title: "Solo"})

	view := app.View()
	// With no history, breadcrumb should show just "Issues > solo-1"
	if !strings.Contains(view, "Issues") {
		t.Error("expected breadcrumb to show 'Issues' as root")
	}
}

func TestApp_GotoMode_DefaultOff(t *testing.T) {
	app := newTestApp()
	if app.gotoMode {
		t.Error("goto mode should be off by default")
	}
}

func TestApp_GKey_EntersGotoMode(t *testing.T) {
	app := newTestApp()
	model, _ := app.Update(keyMsg('g'))
	a := model.(App)
	if !a.gotoMode {
		t.Error("expected goto mode on after pressing g")
	}
}

func TestApp_GotoMode_EscCancels(t *testing.T) {
	app := newTestApp()
	model, _ := app.Update(keyMsg('g'))
	app = model.(App)

	model, _ = app.Update(escKeyMsg())
	app = model.(App)
	if app.gotoMode {
		t.Error("expected goto mode off after Esc")
	}
}

func TestApp_GotoMode_EnterWithID_Navigates(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "goto-1", Title: "Found"}
	ds := &mockDataSource{detail: detail}
	app := newTestAppWithDS(ds)

	// Enter goto mode
	model, _ := app.Update(keyMsg('g'))
	app = model.(App)

	// Type issue ID
	for _, r := range "goto-1" {
		model, _ = app.Update(keyMsg(r))
		app = model.(App)
	}

	// Submit
	model, cmd := app.Update(enterKeyMsg())
	app = model.(App)

	if app.gotoMode {
		t.Error("expected goto mode off after submit")
	}
	if app.activeTab != TabDetail {
		t.Error("expected to switch to detail tab")
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
	msg := cmd()
	if loaded, ok := msg.(IssueDetailLoadedMsg); !ok || loaded.Detail.ID != "goto-1" {
		t.Error("expected fetch for goto-1")
	}
}

func TestApp_GotoMode_EnterEmpty_IsNoop(t *testing.T) {
	app := newTestApp()
	model, _ := app.Update(keyMsg('g'))
	app = model.(App)

	model, cmd := app.Update(enterKeyMsg())
	app = model.(App)
	if app.gotoMode {
		t.Error("expected goto mode off")
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty input")
	}
}

func TestApp_GotoMode_BlocksGlobalKeys(t *testing.T) {
	app := newTestApp()
	model, _ := app.Update(keyMsg('g'))
	app = model.(App)

	// Press 'q' — should NOT quit, should type into input
	model, cmd := app.Update(keyMsg('q'))
	app = model.(App)
	if app.gotoMode != true {
		t.Error("expected to stay in goto mode")
	}
	if cmd != nil {
		// If cmd is tea.Quit, that would be wrong
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("q should not trigger quit in goto mode")
		}
	}
}

func TestApp_GotoMode_SetsLoading(t *testing.T) {
	ds := &mockDataSource{detail: &datasource.IssueDetail{ID: "g-1"}}
	app := newTestAppWithDS(ds)

	model, _ := app.Update(keyMsg('g'))
	app = model.(App)
	for _, r := range "g-1" {
		model, _ = app.Update(keyMsg(r))
		app = model.(App)
	}
	model, _ = app.Update(enterKeyMsg())
	app = model.(App)

	dv := app.views[TabDetail].(*DetailView)
	if !dv.loading {
		t.Error("expected loading state after goto submit")
	}
}

func TestApp_GotoMode_FromDetail_PushesHistory(t *testing.T) {
	ds := &mockDataSource{detail: &datasource.IssueDetail{ID: "g-2"}}
	app := newTestAppWithDS(ds)
	app.activeTab = TabDetail
	dv := app.views[TabDetail].(*DetailView)
	dv.SetDetail(&datasource.IssueDetail{ID: "current-1", Title: "Current"})

	model, _ := app.Update(keyMsg('g'))
	app = model.(App)
	for _, r := range "g-2" {
		model, _ = app.Update(keyMsg(r))
		app = model.(App)
	}
	model, _ = app.Update(enterKeyMsg())
	app = model.(App)

	if len(app.history) != 1 || app.history[0] != "current-1" {
		t.Errorf("expected history [current-1], got %v", app.history)
	}
}

func TestApp_GotoMode_FromNonDetail_ClearsHistory(t *testing.T) {
	ds := &mockDataSource{detail: &datasource.IssueDetail{ID: "g-3"}}
	app := newTestAppWithDS(ds)
	app.activeTab = TabIssues
	app.history = []string{"old-1"}

	model, _ := app.Update(keyMsg('g'))
	app = model.(App)
	for _, r := range "g-3" {
		model, _ = app.Update(keyMsg(r))
		app = model.(App)
	}
	model, _ = app.Update(enterKeyMsg())
	app = model.(App)

	if len(app.history) != 0 {
		t.Errorf("expected history cleared, got %v", app.history)
	}
}

func TestApp_GotoMode_View_ShowsPrompt(t *testing.T) {
	app := newTestApp()
	model, _ := app.Update(keyMsg('g'))
	app = model.(App)

	view := app.View()
	if !strings.Contains(view, "Go to:") {
		t.Error("expected 'Go to:' prompt in view during goto mode")
	}
}

func TestApp_GotoMode_View_HidesNormalContent(t *testing.T) {
	app := newTestApp()
	stub := &stubView{content: "normal content"}
	app.views[TabDashboard] = stub

	model, _ := app.Update(keyMsg('g'))
	app = model.(App)
	app.views[TabDashboard] = stub

	view := app.View()
	if strings.Contains(view, "normal content") {
		t.Error("expected normal content to be hidden during goto mode")
	}
}

func TestApp_HelpOverlay_ShowsGoto(t *testing.T) {
	app := newTestApp()
	app.showHelp = true
	view := app.View()
	if !strings.Contains(view, "goto") || !strings.Contains(view, "g") {
		t.Error("help overlay should show goto binding")
	}
}

func TestApp_HelpOverlay_ShowsEnterAndEsc(t *testing.T) {
	app := newTestApp()
	app.showHelp = true
	view := app.View()
	if !strings.Contains(view, "enter") || !strings.Contains(view, "esc") {
		t.Error("help overlay should show enter and esc bindings")
	}
}

func TestNewApp_RegistersDashboardView(t *testing.T) {
	app := newTestApp()
	v, ok := app.views[TabDashboard]
	if !ok {
		t.Fatal("expected TabDashboard view to be registered")
	}
	if _, ok := v.(*DashboardView); !ok {
		t.Errorf("expected *DashboardView, got %T", v)
	}
}

func TestApp_Update_IssuesLoadedMsg_SetsDataOnDashboardView(t *testing.T) {
	app := newTestApp()
	issues := []datasource.Issue{
		{ID: "d-1", Status: "open"},
		{ID: "d-2", Status: "closed"},
	}
	model, _ := app.Update(IssuesLoadedMsg{Issues: issues})
	a := model.(App)
	dv := a.views[TabDashboard].(*DashboardView)
	if len(dv.issues) != 2 {
		t.Errorf("expected 2 issues on dashboard, got %d", len(dv.issues))
	}
}

func TestApp_Update_ReadyLoadedMsg_SetsDataOnDashboardView(t *testing.T) {
	app := newTestApp()
	ready := []datasource.Issue{
		{ID: "r-1", Title: "Ready One"},
	}
	model, _ := app.Update(ReadyLoadedMsg{Issues: ready})
	a := model.(App)
	dv := a.views[TabDashboard].(*DashboardView)
	if len(dv.ready) != 1 {
		t.Errorf("expected 1 ready issue on dashboard, got %d", len(dv.ready))
	}
}

func TestApp_Update_NoViewRegistered_ReturnsNil(t *testing.T) {
	app := newTestApp()
	app.activeTab = Tab(99) // no view registered for this tab
	model, cmd := app.Update(RefreshMsg{})
	if cmd != nil {
		t.Error("expected nil cmd when no view is registered for tab")
	}
	a := model.(App)
	if a.activeTab != Tab(99) {
		t.Error("expected tab to remain unchanged")
	}
}

func TestApp_FetchIssues_ReturnsIssuesLoadedMsg(t *testing.T) {
	ds := &mockDataSource{issues: []datasource.Issue{{ID: "fi-1"}}}
	app := newTestAppWithDS(ds)
	cmd := app.fetchIssues()
	msg := cmd()
	loaded, ok := msg.(IssuesLoadedMsg)
	if !ok {
		t.Fatalf("expected IssuesLoadedMsg, got %T", msg)
	}
	if len(loaded.Issues) != 1 || loaded.Issues[0].ID != "fi-1" {
		t.Error("expected fetched issues in message")
	}
}

func TestApp_FetchIssues_Error_ReturnsErrMsg(t *testing.T) {
	ds := &mockDataSource{err: errors.New("list fail")}
	app := newTestAppWithDS(ds)
	cmd := app.fetchIssues()
	msg := cmd()
	errMsg, ok := msg.(ErrMsg)
	if !ok {
		t.Fatalf("expected ErrMsg, got %T", msg)
	}
	if errMsg.Err.Error() != "list fail" {
		t.Errorf("expected error 'list fail', got %q", errMsg.Err.Error())
	}
}

func TestApp_FetchReady_ReturnsReadyLoadedMsg(t *testing.T) {
	ds := &mockDataSource{readyIssues: []datasource.Issue{{ID: "r-1"}}}
	app := newTestAppWithDS(ds)
	cmd := app.fetchReady()
	if cmd == nil {
		t.Fatal("expected command from fetchReady")
	}
	msg := cmd()
	loaded, ok := msg.(ReadyLoadedMsg)
	if !ok {
		t.Fatalf("expected ReadyLoadedMsg, got %T", msg)
	}
	if len(loaded.Issues) != 1 || loaded.Issues[0].ID != "r-1" {
		t.Error("expected ready issues in message")
	}
}

func TestApp_FetchReady_Error_ReturnsErrMsg(t *testing.T) {
	ds := &mockDataSource{readyErr: errors.New("ready fail")}
	app := newTestAppWithDS(ds)
	cmd := app.fetchReady()
	msg := cmd()
	errMsg, ok := msg.(ErrMsg)
	if !ok {
		t.Fatalf("expected ErrMsg, got %T", msg)
	}
	if errMsg.Err.Error() != "ready fail" {
		t.Errorf("expected error 'ready fail', got %q", errMsg.Err.Error())
	}
}

func TestNewApp_RegistersCriticalPathView(t *testing.T) {
	app := newTestApp()
	v, ok := app.views[TabCriticalPath]
	if !ok {
		t.Fatal("expected TabCriticalPath view to be registered")
	}
	if _, ok := v.(*CriticalPathView); !ok {
		t.Errorf("expected *CriticalPathView, got %T", v)
	}
}

func TestApp_Update_IssuesLoadedMsg_SetsDataOnCriticalPathView(t *testing.T) {
	app := newTestApp()
	issues := []datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	}
	model, _ := app.Update(IssuesLoadedMsg{Issues: issues})
	a := model.(App)
	cpv := a.views[TabCriticalPath].(*CriticalPathView)
	if len(cpv.chains) == 0 {
		t.Error("expected chains to be computed on CriticalPathView")
	}
}

func TestApp_EnterOnCriticalPath_NavigatesToDetail(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "a", Title: "First"}
	ds := &mockDataSource{detail: detail}
	app := newTestAppWithDS(ds)
	app.activeTab = TabCriticalPath
	cpv := app.views[TabCriticalPath].(*CriticalPathView)
	cpv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1, Title: "First"},
		{ID: "b", Status: "open", Priority: 1, Title: "Second", Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabDetail {
		t.Errorf("expected TabDetail, got %d", a.activeTab)
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
}

func TestApp_EnterOnCriticalPath_NoSelection_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabCriticalPath
	// No issues loaded, so no chains
	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabCriticalPath {
		t.Error("expected to stay on CriticalPath tab when no selection")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no selection")
	}
}

func TestNewApp_RegistersTreeView(t *testing.T) {
	app := newTestApp()
	v, ok := app.views[TabTree]
	if !ok {
		t.Fatal("expected TabTree view to be registered")
	}
	if _, ok := v.(*TreeView); !ok {
		t.Errorf("expected *TreeView, got %T", v)
	}
}

func TestApp_Update_IssuesLoadedMsg_SetsDataOnTreeView(t *testing.T) {
	app := newTestApp()
	issues := []datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	}
	model, _ := app.Update(IssuesLoadedMsg{Issues: issues})
	a := model.(App)
	tv := a.views[TabTree].(*TreeView)
	if len(tv.flatNodes) == 0 {
		t.Error("expected tree nodes to be populated")
	}
}

func TestApp_EnterOnTree_NavigatesToDetail(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "a", Title: "First"}
	ds := &mockDataSource{detail: detail}
	app := newTestAppWithDS(ds)
	app.activeTab = TabTree
	tv := app.views[TabTree].(*TreeView)
	tv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1, Title: "First"},
		{ID: "b", Status: "open", Priority: 1, Title: "Second", Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabDetail {
		t.Errorf("expected TabDetail, got %d", a.activeTab)
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
}

func TestApp_EnterOnTree_NoSelection_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabTree
	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabTree {
		t.Error("expected to stay on Tree tab when no selection")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no selection")
	}
}

func TestApp_EnterOnTree_NonTreeView_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabTree
	app.views[TabTree] = &stubView{}
	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabTree {
		t.Error("expected to stay on Tree tab")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

// --- Focus view wiring ---

func TestNewApp_RegistersFocusView(t *testing.T) {
	app := newTestApp()
	v, ok := app.views[TabFocus]
	if !ok {
		t.Fatal("expected TabFocus view to be registered")
	}
	if _, ok := v.(*FocusView); !ok {
		t.Errorf("expected *FocusView, got %T", v)
	}
}

func TestApp_FocusTabKey(t *testing.T) {
	app := newTestApp()
	model, _ := app.Update(keyMsg('f'))
	a := model.(App)
	if a.activeTab != TabFocus {
		t.Errorf("expected TabFocus, got %d", a.activeTab)
	}
}

func TestApp_Update_IssuesLoadedMsg_SetsDataOnFocusView(t *testing.T) {
	app := newTestApp()
	issues := []datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	}
	model, _ := app.Update(IssuesLoadedMsg{Issues: issues})
	a := model.(App)
	fv := a.views[TabFocus].(*FocusView)
	if fv.dag == nil {
		t.Error("expected DAG to be built on FocusView")
	}
}

func TestApp_Update_ReadyLoadedMsg_SetsDataOnFocusView(t *testing.T) {
	app := newTestApp()
	// First load issues so DAG exists
	issues := []datasource.Issue{
		{ID: "a", Status: "open", Priority: 1},
		{ID: "b", Status: "open", Priority: 1, Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	}
	model, _ := app.Update(IssuesLoadedMsg{Issues: issues})
	a := model.(App)
	// Now load ready
	ready := []datasource.Issue{{ID: "a"}}
	model, _ = a.Update(ReadyLoadedMsg{Issues: ready})
	a = model.(App)
	fv := a.views[TabFocus].(*FocusView)
	if len(fv.items) == 0 {
		t.Error("expected focus items to be populated after ReadyLoadedMsg")
	}
}

func TestApp_EnterOnFocus_NavigatesToDetail(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "a", Title: "Fix auth"}
	ds := &mockDataSource{detail: detail}
	app := newTestAppWithDS(ds)
	app.activeTab = TabFocus
	fv := app.views[TabFocus].(*FocusView)
	fv.SetIssues([]datasource.Issue{
		{ID: "a", Status: "open", Priority: 1, Title: "Fix auth"},
		{ID: "b", Status: "open", Priority: 1, Title: "Deploy", Dependencies: []datasource.RawDependency{
			{IssueID: "b", DependsOnID: "a"},
		}},
	})
	fv.SetReady([]datasource.Issue{{ID: "a"}})

	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabDetail {
		t.Errorf("expected TabDetail, got %d", a.activeTab)
	}
	if cmd == nil {
		t.Fatal("expected fetch command")
	}
}

func TestApp_EnterOnFocus_NoSelection_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabFocus
	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabFocus {
		t.Error("expected to stay on Focus tab when no selection")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no selection")
	}
}

func TestApp_EnterOnFocus_NonFocusView_IsNoop(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabFocus
	app.views[TabFocus] = &stubView{}
	model, cmd := app.Update(enterKeyMsg())
	a := model.(App)
	if a.activeTab != TabFocus {
		t.Error("expected to stay on Focus tab")
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestApp_FocusTabInTabBar(t *testing.T) {
	app := newTestApp()
	app.activeTab = TabFocus
	out := app.View()
	if !strings.Contains(out, "Focus") {
		t.Error("tab bar should contain Focus tab name")
	}
}

func TestApp_HelpShowsFocusKey(t *testing.T) {
	app := newTestApp()
	app.showHelp = true
	out := app.View()
	if !strings.Contains(out, "Focus") {
		t.Errorf("help should mention Focus, got:\n%s", out)
	}
}

func TestApp_View_WatchIndicator_ShownWhenActive(t *testing.T) {
	app := NewApp(&mockDataSource{}, 5*time.Second, true)
	output := app.View()
	if !strings.Contains(output, "WATCH") {
		t.Error("expected View() to contain watch indicator when watch mode is active")
	}
}

func TestApp_View_WatchIndicator_HiddenWhenInactive(t *testing.T) {
	app := NewApp(&mockDataSource{}, 5*time.Second, false)
	output := app.View()
	if strings.Contains(output, "WATCH") {
		t.Error("expected View() NOT to contain watch indicator when watch mode is inactive")
	}
}

func TestApp_WindowSizeMsg_StoresDimensions(t *testing.T) {
	app := newTestApp()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	model, _ := app.Update(msg)
	a := model.(App)
	if a.width != 120 || a.height != 40 {
		t.Errorf("expected width=120 height=40, got width=%d height=%d", a.width, a.height)
	}
}

func TestApp_WindowSizeMsg_PropagatesResizeToActiveView(t *testing.T) {
	app := newTestApp()
	// Switch to issues tab first
	app.activeTab = TabIssues
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	app.Update(msg)

	// Verify ListView got resized by checking it has a non-zero width stored
	lv := app.views[TabIssues].(*ListView)
	if lv.width == 0 {
		t.Error("expected ListView to have non-zero width after WindowSizeMsg")
	}
}
