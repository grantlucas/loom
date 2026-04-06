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

// StatusInfoer is optionally implemented by views that provide contextual info
// for the secondary status line (e.g. issue counts).
type StatusInfoer interface {
	StatusInfo() string
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

// renderInfoLine renders a single-line info bar from the given text.
// Returns empty string if info is empty or width is zero.
func renderInfoLine(info string, width int) string {
	if info == "" || width <= 0 {
		return ""
	}
	rendered := infoLineStyle.Render(info)
	if lipgloss.Width(rendered) > width {
		// Truncate to fit
		rendered = infoLineStyle.Width(width).MaxWidth(width).Render(info)
	}
	return rendered
}
