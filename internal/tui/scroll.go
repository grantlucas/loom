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

// cursorToViewportMiddle sets the cursor to the middle visible item after a
// multi-line viewport scroll. contentOffset is the number of non-item lines
// before the first item (e.g. 2 for tree stats+blank line).
func cursorToViewportMiddle(vp *viewport.Model, cursor *int, itemCount, contentOffset int) {
	if itemCount == 0 {
		*cursor = 0
		return
	}
	mid := vp.YOffset + vp.Height/2 - contentOffset
	if mid < 0 {
		mid = 0
	}
	if mid >= itemCount {
		mid = itemCount - 1
	}
	*cursor = mid
}
