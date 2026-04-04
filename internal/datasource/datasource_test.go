package datasource

import (
	"testing"
	"time"
)

func TestParseIssueList(t *testing.T) {
	input := []byte(`[
		{
			"id": "proj-1",
			"title": "First issue",
			"description": "Do something",
			"status": "open",
			"priority": 1,
			"issue_type": "task",
			"assignee": "Alice",
			"owner": "alice@example.com",
			"created_at": "2026-01-15T10:00:00Z",
			"created_by": "Alice",
			"updated_at": "2026-01-15T12:00:00Z",
			"dependencies": [
				{
					"issue_id": "proj-1",
					"depends_on_id": "proj-0",
					"type": "blocks",
					"created_at": "2026-01-15T10:00:00Z",
					"created_by": "Alice",
					"metadata": "{}"
				}
			],
			"dependency_count": 1,
			"dependent_count": 2,
			"comment_count": 3,
			"parent": "proj-0"
		}
	]`)

	issues, err := ParseIssueList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]
	if issue.ID != "proj-1" {
		t.Errorf("ID = %q, want %q", issue.ID, "proj-1")
	}
	if issue.Title != "First issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "First issue")
	}
	if issue.Description != "Do something" {
		t.Errorf("Description = %q, want %q", issue.Description, "Do something")
	}
	if issue.Status != "open" {
		t.Errorf("Status = %q, want %q", issue.Status, "open")
	}
	if issue.Priority != 1 {
		t.Errorf("Priority = %d, want %d", issue.Priority, 1)
	}
	if issue.IssueType != "task" {
		t.Errorf("IssueType = %q, want %q", issue.IssueType, "task")
	}
	if issue.Assignee != "Alice" {
		t.Errorf("Assignee = %q, want %q", issue.Assignee, "Alice")
	}
	if issue.Owner != "alice@example.com" {
		t.Errorf("Owner = %q, want %q", issue.Owner, "alice@example.com")
	}
	if issue.CreatedBy != "Alice" {
		t.Errorf("CreatedBy = %q, want %q", issue.CreatedBy, "Alice")
	}
	if issue.Parent != "proj-0" {
		t.Errorf("Parent = %q, want %q", issue.Parent, "proj-0")
	}

	wantCreated := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	if !issue.CreatedAt.Equal(wantCreated) {
		t.Errorf("CreatedAt = %v, want %v", issue.CreatedAt, wantCreated)
	}
	wantUpdated := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	if !issue.UpdatedAt.Equal(wantUpdated) {
		t.Errorf("UpdatedAt = %v, want %v", issue.UpdatedAt, wantUpdated)
	}

	if issue.DependencyCount != 1 {
		t.Errorf("DependencyCount = %d, want %d", issue.DependencyCount, 1)
	}
	if issue.DependentCount != 2 {
		t.Errorf("DependentCount = %d, want %d", issue.DependentCount, 2)
	}
	if issue.CommentCount != 3 {
		t.Errorf("CommentCount = %d, want %d", issue.CommentCount, 3)
	}

	if len(issue.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(issue.Dependencies))
	}
	dep := issue.Dependencies[0]
	if dep.IssueID != "proj-1" {
		t.Errorf("dep.IssueID = %q, want %q", dep.IssueID, "proj-1")
	}
	if dep.DependsOnID != "proj-0" {
		t.Errorf("dep.DependsOnID = %q, want %q", dep.DependsOnID, "proj-0")
	}
	if dep.Type != "blocks" {
		t.Errorf("dep.Type = %q, want %q", dep.Type, "blocks")
	}
}

func TestParseIssueListEmpty(t *testing.T) {
	issues, err := ParseIssueList([]byte(`[]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestParseIssueDetail(t *testing.T) {
	input := []byte(`[
		{
			"id": "proj-2",
			"title": "Implement feature",
			"description": "Build the thing",
			"status": "in_progress",
			"priority": 0,
			"issue_type": "feature",
			"assignee": "Bob",
			"owner": "bob@example.com",
			"created_at": "2026-02-01T09:00:00Z",
			"created_by": "Bob",
			"updated_at": "2026-02-02T15:00:00Z",
			"dependencies": [
				{
					"id": "proj-1",
					"title": "Prerequisite task",
					"description": "Must do first",
					"status": "closed",
					"priority": 1,
					"issue_type": "task",
					"assignee": "Alice",
					"owner": "alice@example.com",
					"created_at": "2026-01-15T10:00:00Z",
					"created_by": "Alice",
					"updated_at": "2026-01-20T10:00:00Z",
					"closed_at": "2026-01-20T10:00:00Z",
					"close_reason": "Done",
					"dependency_type": "blocks"
				}
			],
			"dependents": [
				{
					"id": "proj-3",
					"title": "Follow-up work",
					"description": "After feature",
					"status": "open",
					"priority": 2,
					"issue_type": "task",
					"owner": "carol@example.com",
					"created_at": "2026-02-01T09:00:00Z",
					"created_by": "Carol",
					"updated_at": "2026-02-01T09:00:00Z",
					"dependency_type": "blocks"
				}
			],
			"parent": "proj-epic"
		}
	]`)

	detail, err := ParseIssueDetail(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if detail.ID != "proj-2" {
		t.Errorf("ID = %q, want %q", detail.ID, "proj-2")
	}
	if detail.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", detail.Status, "in_progress")
	}
	if detail.Priority != 0 {
		t.Errorf("Priority = %d, want %d", detail.Priority, 0)
	}

	// Dependencies
	if len(detail.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(detail.Dependencies))
	}
	dep := detail.Dependencies[0]
	if dep.ID != "proj-1" {
		t.Errorf("dep.ID = %q, want %q", dep.ID, "proj-1")
	}
	if dep.Status != "closed" {
		t.Errorf("dep.Status = %q, want %q", dep.Status, "closed")
	}
	if dep.DependencyType != "blocks" {
		t.Errorf("dep.DependencyType = %q, want %q", dep.DependencyType, "blocks")
	}
	if dep.ClosedAt != "2026-01-20T10:00:00Z" {
		t.Errorf("dep.ClosedAt = %q, want %q", dep.ClosedAt, "2026-01-20T10:00:00Z")
	}
	if dep.CloseReason != "Done" {
		t.Errorf("dep.CloseReason = %q, want %q", dep.CloseReason, "Done")
	}

	// Dependents
	if len(detail.Dependents) != 1 {
		t.Fatalf("expected 1 dependent, got %d", len(detail.Dependents))
	}
	dependent := detail.Dependents[0]
	if dependent.ID != "proj-3" {
		t.Errorf("dependent.ID = %q, want %q", dependent.ID, "proj-3")
	}
	if dependent.DependencyType != "blocks" {
		t.Errorf("dependent.DependencyType = %q, want %q", dependent.DependencyType, "blocks")
	}
	if dependent.Assignee != "" {
		t.Errorf("dependent.Assignee = %q, want empty (omitted)", dependent.Assignee)
	}

	if detail.Parent != "proj-epic" {
		t.Errorf("Parent = %q, want %q", detail.Parent, "proj-epic")
	}
}

func TestParseIssueListInvalidJSON(t *testing.T) {
	_, err := ParseIssueList([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
