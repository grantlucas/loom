package tui

import (
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
