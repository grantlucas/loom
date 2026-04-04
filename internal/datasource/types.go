package datasource

import "time"

// Issue represents an issue from bd list --json output.
type Issue struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Status          string          `json:"status"`
	Priority        int             `json:"priority"`
	IssueType       string          `json:"issue_type"`
	Assignee        string          `json:"assignee"`
	Owner           string          `json:"owner"`
	CreatedAt       time.Time       `json:"created_at"`
	CreatedBy       string          `json:"created_by"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Dependencies    []RawDependency `json:"dependencies"`
	DependencyCount int             `json:"dependency_count"`
	DependentCount  int             `json:"dependent_count"`
	CommentCount    int             `json:"comment_count"`
	Parent          string          `json:"parent"`
}

// RawDependency represents a dependency link from bd list --json output.
type RawDependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	Metadata    string `json:"metadata"`
}
