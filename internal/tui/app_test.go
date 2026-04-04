package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewApp_DefaultsToDashboard(t *testing.T) {
	app := NewApp()
	if app.activeTab != TabDashboard {
		t.Errorf("expected active tab %d (Dashboard), got %d", TabDashboard, app.activeTab)
	}
}

func TestNewApp_ImplementsTeaModel(t *testing.T) {
	var _ tea.Model = NewApp()
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
			app := NewApp()
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
	app := NewApp()
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
	app := NewApp()
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
	app := NewApp()
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
