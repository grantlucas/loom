package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/grantlucas/loom/internal/datasource"
)

func TestActiveTabStyle_IsBold(t *testing.T) {
	if !activeTabStyle.GetBold() {
		t.Error("active tab style should be bold")
	}
}

func TestInactiveTabStyle_IsNotBold(t *testing.T) {
	if inactiveTabStyle.GetBold() {
		t.Error("inactive tab style should not be bold")
	}
}

func TestActiveTabStyle_HasBorderWithoutBottom(t *testing.T) {
	if !activeTabStyle.GetBorderTop() {
		t.Error("active tab style should have a top border")
	}
	if !activeTabStyle.GetBorderLeft() {
		t.Error("active tab style should have a left border")
	}
	if !activeTabStyle.GetBorderRight() {
		t.Error("active tab style should have a right border")
	}
	if activeTabStyle.GetBorderBottom() {
		t.Error("active tab style should NOT have a bottom border")
	}
}

func TestInactiveTabStyle_HasBorderWithoutBottom(t *testing.T) {
	if !inactiveTabStyle.GetBorderTop() {
		t.Error("inactive tab style should have a top border")
	}
	if inactiveTabStyle.GetBorderBottom() {
		t.Error("inactive tab style should NOT have a bottom border")
	}
}

func TestActiveTabStyle_HasPadding(t *testing.T) {
	_, right, _, left := activeTabStyle.GetPadding()
	if left == 0 && right == 0 {
		t.Error("active tab style should have horizontal padding")
	}
}

func TestInactiveTabStyle_HasPadding(t *testing.T) {
	_, right, _, left := inactiveTabStyle.GetPadding()
	if left == 0 && right == 0 {
		t.Error("inactive tab style should have horizontal padding")
	}
}

func TestDetailTitleStyle_IsBold(t *testing.T) {
	if !detailTitleStyle.GetBold() {
		t.Error("detail title style should be bold")
	}
}

func TestDetailSectionStyle_IsBold(t *testing.T) {
	if !detailSectionStyle.GetBold() {
		t.Error("detail section style should be bold")
	}
}

func TestDetailLabelStyle_IsDimmed(t *testing.T) {
	if !detailLabelStyle.GetFaint() {
		t.Error("detail label style should be faint/dimmed")
	}
}

func TestRelationSelectedStyle_IsReverse(t *testing.T) {
	if !relationSelectedStyle.GetReverse() {
		t.Error("relation selected style should be reverse")
	}
}

func TestBreadcrumbStyle_HasForegroundColor(t *testing.T) {
	fg := breadcrumbStyle.GetForeground()
	if fg == (lipgloss.NoColor{}) {
		t.Error("breadcrumb style should have a foreground color")
	}
}

func TestDashboardBarStyle_HasForegroundColor(t *testing.T) {
	fg := dashboardBarStyle.GetForeground()
	if fg == (lipgloss.NoColor{}) {
		t.Error("dashboard bar style should have a foreground color")
	}
}

func TestPriorityStyle_ReturnsDistinctColorsPerLevel(t *testing.T) {
	expected := map[int]lipgloss.Color{
		0: lipgloss.Color("196"), // red
		1: lipgloss.Color("208"), // orange
		2: lipgloss.Color("226"), // yellow
		3: lipgloss.Color("33"),  // blue
		4: lipgloss.Color("243"), // gray
	}
	for pri, wantColor := range expected {
		style := PriorityStyle(pri)
		got := style.GetForeground()
		if got != wantColor {
			t.Errorf("PriorityStyle(%d): got foreground %v, want %v", pri, got, wantColor)
		}
	}
}

func TestStyledPriority_ContainsPriorityLabel(t *testing.T) {
	for pri := 0; pri <= 4; pri++ {
		result := StyledPriority(pri)
		want := fmt.Sprintf("P%d", pri)
		if !strings.Contains(result, want) {
			t.Errorf("StyledPriority(%d): result %q does not contain %q", pri, result, want)
		}
	}
}

func TestStatusStyle_ReturnsDistinctColorsPerStatus(t *testing.T) {
	expected := map[string]lipgloss.Color{
		"closed":      lipgloss.Color("34"),  // green
		"in_progress": lipgloss.Color("226"), // yellow
		"open":        lipgloss.Color("252"), // white
	}
	for status, wantColor := range expected {
		style := StatusStyle(status)
		got := style.GetForeground()
		if got != wantColor {
			t.Errorf("StatusStyle(%q): got foreground %v, want %v", status, got, wantColor)
		}
	}
}

func TestStatusStyle_UnknownStatusDefaultsToOpen(t *testing.T) {
	style := StatusStyle("unknown")
	got := style.GetForeground()
	if got != lipgloss.Color("252") {
		t.Errorf("StatusStyle(unknown): got foreground %v, want white (252)", got)
	}
}

func TestStyledStatus_ReturnsCorrectIndicators(t *testing.T) {
	tests := []struct {
		status    string
		depCount  int
		wantIcon  string
	}{
		{"closed", 0, "✓"},
		{"in_progress", 0, "◐"},
		{"open", 0, "○"},
		{"open", 2, "●"},
	}
	for _, tt := range tests {
		issue := datasource.Issue{Status: tt.status, DependencyCount: tt.depCount}
		result := StyledStatus(issue)
		if !strings.Contains(result, tt.wantIcon) {
			t.Errorf("StyledStatus(%s, deps=%d): result %q does not contain %q",
				tt.status, tt.depCount, result, tt.wantIcon)
		}
	}
}

func TestStyledStatusSimple_ReturnsCorrectIndicators(t *testing.T) {
	tests := []struct {
		status   string
		wantIcon string
	}{
		{"closed", "✓"},
		{"in_progress", "◐"},
		{"open", "○"},
	}
	for _, tt := range tests {
		result := StyledStatusSimple(tt.status)
		if !strings.Contains(result, tt.wantIcon) {
			t.Errorf("StyledStatusSimple(%q): result %q does not contain %q",
				tt.status, result, tt.wantIcon)
		}
	}
}

func TestPriorityStyle_UnknownPriorityDefaultsToGray(t *testing.T) {
	style := PriorityStyle(99)
	got := style.GetForeground()
	if got != lipgloss.Color("243") {
		t.Errorf("PriorityStyle(99): got foreground %v, want gray (243)", got)
	}
}

func TestPlainPriority(t *testing.T) {
	for pri := 0; pri <= 4; pri++ {
		got := PlainPriority(pri)
		want := fmt.Sprintf("P%d", pri)
		if got != want {
			t.Errorf("PlainPriority(%d) = %q, want %q", pri, got, want)
		}
	}
}

func TestPlainPriority_ContainsNoANSI(t *testing.T) {
	got := PlainPriority(1)
	if strings.Contains(got, "\x1b") {
		t.Errorf("PlainPriority should not contain ANSI escape codes, got %q", got)
	}
}

func TestPlainStatus(t *testing.T) {
	tests := []struct {
		status   string
		depCount int
		want     string
	}{
		{"closed", 0, "✓"},
		{"in_progress", 0, "◐"},
		{"open", 0, "○"},
		{"open", 2, "●"},
	}
	for _, tt := range tests {
		issue := datasource.Issue{Status: tt.status, DependencyCount: tt.depCount}
		got := PlainStatus(issue)
		if got != tt.want {
			t.Errorf("PlainStatus(%s, deps=%d) = %q, want %q", tt.status, tt.depCount, got, tt.want)
		}
	}
}

func TestPlainStatus_ContainsNoANSI(t *testing.T) {
	issue := datasource.Issue{Status: "in_progress"}
	got := PlainStatus(issue)
	if strings.Contains(got, "\x1b") {
		t.Errorf("PlainStatus should not contain ANSI escape codes, got %q", got)
	}
}

func TestRenderSectionHeader_ContainsTitle(t *testing.T) {
	result := renderSectionHeader("Status", 40)
	if !strings.Contains(result, "Status") {
		t.Errorf("renderSectionHeader should contain title, got %q", result)
	}
}

func TestRenderSectionHeader_ContainsDashes(t *testing.T) {
	result := renderSectionHeader("Status", 40)
	if !strings.Contains(result, "──") {
		t.Errorf("renderSectionHeader should contain dash characters, got %q", result)
	}
}

func TestInfoLineStyle_HasForegroundColor(t *testing.T) {
	fg := infoLineStyle.GetForeground()
	if fg == (lipgloss.NoColor{}) {
		t.Error("info line style should have a foreground color")
	}
}

func TestStatusBarContainerStyle_HasTopBorder(t *testing.T) {
	if !statusBarContainerStyle.GetBorderTop() {
		t.Error("status bar container style should have a top border")
	}
}

func TestTabGapStyle_HasForegroundColor(t *testing.T) {
	fg := tabGapStyle.GetForeground()
	if fg == (lipgloss.NoColor{}) {
		t.Error("tab gap style should have a foreground color")
	}
}
