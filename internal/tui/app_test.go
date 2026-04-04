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
