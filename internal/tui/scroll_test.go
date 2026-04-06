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
