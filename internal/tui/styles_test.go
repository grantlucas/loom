package tui

import "testing"

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
