package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

func TestTabBarStyle_HasBottomBorder(t *testing.T) {
	if !tabBarStyle.GetBorderBottom() {
		t.Error("tab bar style should have a bottom border")
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

func TestGotoPromptStyle_IsBold(t *testing.T) {
	if !gotoPromptStyle.GetBold() {
		t.Error("goto prompt style should be bold")
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

func TestPriorityStyle_UnknownPriorityDefaultsToGray(t *testing.T) {
	style := PriorityStyle(99)
	got := style.GetForeground()
	if got != lipgloss.Color("243") {
		t.Errorf("PriorityStyle(99): got foreground %v, want gray (243)", got)
	}
}
