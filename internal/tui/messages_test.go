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
