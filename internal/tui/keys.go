package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all global key bindings for the application.
type KeyMap struct {
	Dashboard    key.Binding
	Issues       key.Binding
	Tree         key.Binding
	CriticalPath key.Binding
	Refresh      key.Binding
	Watch        key.Binding
	Help         key.Binding
	Quit         key.Binding
	Enter        key.Binding
	Back         key.Binding
	Goto         key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Dashboard: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "dashboard"),
		),
		Issues: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "issues"),
		),
		Tree: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "tree"),
		),
		CriticalPath: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "critical path"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Watch: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle watch"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open detail"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Goto: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "goto issue"),
		),
	}
}
