package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestKeyMap_AllBindingsHaveHelp(t *testing.T) {
	km := DefaultKeyMap()

	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Dashboard", km.Dashboard},
		{"Issues", km.Issues},
		{"Tree", km.Tree},
		{"CriticalPath", km.CriticalPath},
		{"Refresh", km.Refresh},
		{"Watch", km.Watch},
		{"Help", km.Help},
		{"Quit", km.Quit},
		{"Enter", km.Enter},
		{"Back", km.Back},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			help := b.binding.Help()
			if help.Key == "" {
				t.Errorf("binding %q has empty help key", b.name)
			}
			if help.Desc == "" {
				t.Errorf("binding %q has empty help description", b.name)
			}
		})
	}
}

func TestKeyMap_BindingKeys(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name string
		bind key.Binding
		keys []string
	}{
		{"Dashboard", km.Dashboard, []string{"d"}},
		{"Issues", km.Issues, []string{"i"}},
		{"Tree", km.Tree, []string{"t"}},
		{"CriticalPath", km.CriticalPath, []string{"c"}},
		{"Refresh", km.Refresh, []string{"r"}},
		{"Watch", km.Watch, []string{"w"}},
		{"Help", km.Help, []string{"?"}},
		{"Quit", km.Quit, []string{"q", "ctrl+c"}},
		{"Enter", km.Enter, []string{"enter"}},
		{"Back", km.Back, []string{"esc"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !key.Matches(keyFromString(tt.keys[0]), tt.bind) {
				t.Errorf("expected binding %q to match key %q", tt.name, tt.keys[0])
			}
		})
	}
}

// keyFromString creates a tea.KeyMsg-like input for key.Matches testing.
func keyFromString(s string) keyMsgAdapter {
	return keyMsgAdapter{str: s}
}

type keyMsgAdapter struct{ str string }

func (k keyMsgAdapter) String() string { return k.str }
