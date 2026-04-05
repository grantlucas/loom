package tui

import (
	"errors"
	"testing"
	"time"

	"github.com/grantlucas/loom/internal/datasource"
)

func TestIssuesLoadedMsg_CarriesIssues(t *testing.T) {
	issues := []datasource.Issue{{ID: "test-1", Title: "Test"}}
	msg := IssuesLoadedMsg{Issues: issues}
	if len(msg.Issues) != 1 || msg.Issues[0].ID != "test-1" {
		t.Error("IssuesLoadedMsg should carry issues")
	}
}

func TestErrMsg_CarriesError(t *testing.T) {
	err := errors.New("fetch failed")
	msg := ErrMsg{Err: err}
	if msg.Err.Error() != "fetch failed" {
		t.Error("ErrMsg should carry the error")
	}
}

func TestTickMsg_IsTimeAlias(t *testing.T) {
	now := time.Now()
	msg := TickMsg(now)
	if time.Time(msg) != now {
		t.Error("TickMsg should be convertible to time.Time")
	}
}

func TestIssueDetailLoadedMsg_CarriesDetail(t *testing.T) {
	detail := &datasource.IssueDetail{ID: "proj-1", Title: "Test Issue"}
	msg := IssueDetailLoadedMsg{Detail: detail}
	if msg.Detail.ID != "proj-1" {
		t.Error("IssueDetailLoadedMsg should carry the detail")
	}
	if msg.Detail.Title != "Test Issue" {
		t.Error("IssueDetailLoadedMsg should carry the detail title")
	}
}

func TestIssueDetailErrMsg_CarriesError(t *testing.T) {
	err := errors.New("detail fetch failed")
	msg := IssueDetailErrMsg{Err: err}
	if msg.Err.Error() != "detail fetch failed" {
		t.Error("IssueDetailErrMsg should carry the error")
	}
}

func TestReadyLoadedMsg_CarriesIssues(t *testing.T) {
	issues := []datasource.Issue{
		{ID: "ready-1", Title: "Ready One"},
		{ID: "ready-2", Title: "Ready Two"},
	}
	msg := ReadyLoadedMsg{Issues: issues}
	if len(msg.Issues) != 2 {
		t.Errorf("ReadyLoadedMsg should carry 2 issues, got %d", len(msg.Issues))
	}
	if msg.Issues[0].ID != "ready-1" {
		t.Error("ReadyLoadedMsg should preserve issue order")
	}
}
