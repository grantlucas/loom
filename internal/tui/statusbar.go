package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusHint is a key-description pair for the status bar.
type StatusHint struct {
	Key  string
	Desc string
}

// StatusHinter is optionally implemented by views that provide contextual hints.
type StatusHinter interface {
	StatusHints() []StatusHint
}

// renderStatusBar renders a single-line status bar from the given hints.
// Format: "key desc · key desc · key desc", truncated to fit width.
func renderStatusBar(hints []StatusHint, width int) string {
	if len(hints) == 0 || width <= 0 {
		return ""
	}

	sep := hintSepStyle.Render(" · ")
	sepWidth := lipgloss.Width(sep)

	var parts []string
	totalWidth := 0

	for i, h := range hints {
		part := hintKeyStyle.Render(h.Key) + " " + hintDescStyle.Render(h.Desc)
		partWidth := lipgloss.Width(part)

		addedWidth := partWidth
		if i > 0 {
			addedWidth += sepWidth
		}

		if totalWidth+addedWidth > width {
			break
		}

		parts = append(parts, part)
		totalWidth += addedWidth
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, sep)
}
