package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderStatusBar_FormatsHints(t *testing.T) {
	hints := []StatusHint{
		{Key: "s", Desc: "sort"},
		{Key: "/", Desc: "filter"},
		{Key: "?", Desc: "help"},
	}
	result := renderStatusBar(hints, 80)

	// Should contain all keys and descriptions
	for _, h := range hints {
		if !strings.Contains(result, h.Key) {
			t.Errorf("expected result to contain key %q", h.Key)
		}
		if !strings.Contains(result, h.Desc) {
			t.Errorf("expected result to contain desc %q", h.Desc)
		}
	}

	// Should contain separator between hints
	if !strings.Contains(result, "·") {
		t.Error("expected result to contain · separator")
	}
}

func TestRenderStatusBar_EmptyHints(t *testing.T) {
	result := renderStatusBar(nil, 80)
	if result != "" {
		t.Errorf("expected empty string for nil hints, got %q", result)
	}

	result = renderStatusBar([]StatusHint{}, 80)
	if result != "" {
		t.Errorf("expected empty string for empty hints, got %q", result)
	}
}

func TestRenderStatusBar_SingleHint(t *testing.T) {
	hints := []StatusHint{{Key: "q", Desc: "quit"}}
	result := renderStatusBar(hints, 80)

	if !strings.Contains(result, "q") {
		t.Error("expected result to contain key 'q'")
	}
	if !strings.Contains(result, "quit") {
		t.Error("expected result to contain desc 'quit'")
	}
	// Single hint should have no separator
	if strings.Contains(result, "·") {
		t.Error("single hint should not contain separator")
	}
}

func TestRenderStatusBar_TruncatesAtWidth(t *testing.T) {
	hints := []StatusHint{
		{Key: "s", Desc: "sort"},
		{Key: "/", Desc: "filter"},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
		{Key: "r", Desc: "refresh"},
		{Key: "w", Desc: "watch"},
	}
	// Use a very narrow width
	result := renderStatusBar(hints, 20)
	visibleWidth := lipgloss.Width(result)
	if visibleWidth > 20 {
		t.Errorf("rendered width %d exceeds max width 20", visibleWidth)
	}
}

func TestRenderStatusBar_ZeroWidth(t *testing.T) {
	hints := []StatusHint{{Key: "q", Desc: "quit"}}
	result := renderStatusBar(hints, 0)
	if result != "" {
		t.Errorf("expected empty string for zero width, got %q", result)
	}
}

func TestRenderInfoLine_FormatsText(t *testing.T) {
	result := renderInfoLine("3 issues", 80)
	if !strings.Contains(result, "3 issues") {
		t.Errorf("expected result to contain '3 issues', got %q", result)
	}
}

func TestRenderInfoLine_EmptyString(t *testing.T) {
	result := renderInfoLine("", 80)
	if result != "" {
		t.Errorf("expected empty string for empty info, got %q", result)
	}
}

func TestRenderInfoLine_ZeroWidth(t *testing.T) {
	result := renderInfoLine("3 issues", 0)
	if result != "" {
		t.Errorf("expected empty string for zero width, got %q", result)
	}
}

func TestRenderInfoLine_TruncatesAtWidth(t *testing.T) {
	result := renderInfoLine("this is a very long info line that exceeds the width", 20)
	visibleWidth := lipgloss.Width(result)
	if visibleWidth > 20 {
		t.Errorf("rendered width %d exceeds max width 20", visibleWidth)
	}
}
