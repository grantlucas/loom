package tui

import "github.com/charmbracelet/bubbles/viewport"

// ensureLineVisible adjusts the viewport offset so that the given line is visible.
func ensureLineVisible(vp *viewport.Model, line int) {
	if line < vp.YOffset {
		vp.SetYOffset(line)
	}
	if line >= vp.YOffset+vp.Height {
		vp.SetYOffset(line - vp.Height + 1)
	}
}
