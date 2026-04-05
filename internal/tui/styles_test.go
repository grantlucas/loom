package tui

import (
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
