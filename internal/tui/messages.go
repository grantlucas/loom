package tui

import (
	"time"

	"github.com/grantlucas/loom/internal/datasource"
)

// IssuesLoadedMsg carries fetched issues into the update loop.
type IssuesLoadedMsg struct {
	Issues []datasource.Issue
}

// ErrMsg carries a data-fetch error into the update loop.
type ErrMsg struct {
	Err error
}

// TickMsg signals a watch-mode tick.
type TickMsg time.Time

// RefreshMsg signals that data should be re-fetched from bd.
type RefreshMsg struct{}

// IssueDetailLoadedMsg carries a fetched issue detail into the update loop.
type IssueDetailLoadedMsg struct {
	Detail *datasource.IssueDetail
}

// IssueDetailErrMsg carries an issue-detail fetch error into the update loop.
type IssueDetailErrMsg struct {
	Err error
}
