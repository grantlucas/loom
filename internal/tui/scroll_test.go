package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
)

// newTestViewport creates a viewport with enough content lines for offset testing.
func newTestViewport(width, height, contentLines int) viewport.Model {
	vp := viewport.New(width, height)
	lines := make([]string, contentLines)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	return vp
}

func TestEnsureLineVisible_ScrollsDownWhenCursorBelowViewport(t *testing.T) {
	vp := newTestViewport(80, 5, 20)
	vp.SetYOffset(0)

	ensureLineVisible(&vp, 7)

	if vp.YOffset != 3 {
		t.Errorf("expected YOffset 3, got %d", vp.YOffset)
	}
}

func TestEnsureLineVisible_ScrollsUpWhenCursorAboveViewport(t *testing.T) {
	vp := newTestViewport(80, 5, 20)
	vp.SetYOffset(5)

	ensureLineVisible(&vp, 2)

	if vp.YOffset != 2 {
		t.Errorf("expected YOffset 2, got %d", vp.YOffset)
	}
}

func TestEnsureLineVisible_NoChangeWhenCursorVisible(t *testing.T) {
	vp := newTestViewport(80, 5, 20)
	vp.SetYOffset(2)

	ensureLineVisible(&vp, 4)

	if vp.YOffset != 2 {
		t.Errorf("expected YOffset 2 (unchanged), got %d", vp.YOffset)
	}
}

// --- cursorToViewportMiddle ---

func TestCursorToViewportMiddle_SetsCursorToMiddle(t *testing.T) {
	vp := newTestViewport(80, 10, 30)
	vp.SetYOffset(10) // visible lines 10-19
	cursor := 0
	// contentOffset=2 (e.g. stats+blank), so mid = 10 + 5 - 2 = 13
	cursorToViewportMiddle(&vp, &cursor, 30, 2)
	if cursor != 13 {
		t.Errorf("expected cursor 13, got %d", cursor)
	}
}

func TestCursorToViewportMiddle_ClampsToZero(t *testing.T) {
	vp := newTestViewport(80, 10, 30)
	vp.SetYOffset(0)
	cursor := 5
	// contentOffset=8, so mid = 0 + 5 - 8 = -3 → clamped to 0
	cursorToViewportMiddle(&vp, &cursor, 30, 8)
	if cursor != 0 {
		t.Errorf("expected cursor 0, got %d", cursor)
	}
}

func TestCursorToViewportMiddle_ClampsToLastItem(t *testing.T) {
	vp := newTestViewport(80, 10, 30)
	vp.SetYOffset(25) // near the end
	cursor := 0
	// contentOffset=2, so mid = 25 + 5 - 2 = 28; itemCount=5 → clamped to 4
	cursorToViewportMiddle(&vp, &cursor, 5, 2)
	if cursor != 4 {
		t.Errorf("expected cursor 4, got %d", cursor)
	}
}

func TestCursorToViewportMiddle_ZeroItems(t *testing.T) {
	vp := newTestViewport(80, 10, 30)
	vp.SetYOffset(5)
	cursor := 0
	cursorToViewportMiddle(&vp, &cursor, 0, 2)
	if cursor != 0 {
		t.Errorf("expected cursor 0 with zero items, got %d", cursor)
	}
}
