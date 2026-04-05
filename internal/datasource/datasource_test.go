package datasource

import (
	"errors"
	"fmt"
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

func TestParseIssueDetailInvalidJSON(t *testing.T) {
	_, err := ParseIssueDetail([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseIssueDetailEmptyArray(t *testing.T) {
	_, err := ParseIssueDetail([]byte(`[]`))
	if err == nil {
		t.Fatal("expected error for empty array, got nil")
	}
}

// mockExecutor records calls and returns canned responses.
type mockExecutor struct {
	calls  [][]string
	output []byte
	err    error
}

func (m *mockExecutor) Execute(args ...string) ([]byte, error) {
	m.calls = append(m.calls, args)
	return m.output, m.err
}

func TestClientListIssues(t *testing.T) {
	mock := &mockExecutor{
		output: []byte(`[{"id":"proj-1","title":"Test","status":"open","priority":1,"issue_type":"task","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]`),
	}
	client := NewClient(mock)

	issues, err := client.ListIssues()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].ID != "proj-1" {
		t.Errorf("ID = %q, want %q", issues[0].ID, "proj-1")
	}

	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	args := mock.calls[0]
	if len(args) != 2 || args[0] != "list" || args[1] != "--json" {
		t.Errorf("args = %v, want [list --json]", args)
	}
}

func TestClientGetIssue(t *testing.T) {
	mock := &mockExecutor{
		output: []byte(`[{"id":"proj-2","title":"Detail","status":"open","priority":1,"issue_type":"task","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","dependencies":[],"dependents":[]}]`),
	}
	client := NewClient(mock)

	detail, err := client.GetIssue("proj-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.ID != "proj-2" {
		t.Errorf("ID = %q, want %q", detail.ID, "proj-2")
	}

	args := mock.calls[0]
	if len(args) != 3 || args[0] != "show" || args[1] != "proj-2" || args[2] != "--json" {
		t.Errorf("args = %v, want [show proj-2 --json]", args)
	}
}

func TestClientListReady(t *testing.T) {
	mock := &mockExecutor{
		output: []byte(`[{"id":"proj-3","title":"Ready","status":"open","priority":2,"issue_type":"task","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]`),
	}
	client := NewClient(mock)

	issues, err := client.ListReady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != "proj-3" {
		t.Errorf("unexpected issues: %+v", issues)
	}

	args := mock.calls[0]
	if len(args) != 2 || args[0] != "ready" || args[1] != "--json" {
		t.Errorf("args = %v, want [ready --json]", args)
	}
}

func TestClientExecutorError(t *testing.T) {
	mock := &mockExecutor{
		err: fmt.Errorf("command failed"),
	}
	client := NewClient(mock)

	_, err := client.ListIssues()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "command failed" {
		t.Errorf("error = %q, want %q", err.Error(), "command failed")
	}
}

func TestBdExecutorRunsCommand(t *testing.T) {
	// Use "echo" as a stand-in for bd to test the executor plumbing
	exec := &BdExecutor{BinPath: "echo"}
	out, err := exec.Execute("hello", "world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := string(out)
	if got != "hello world\n" {
		t.Errorf("output = %q, want %q", got, "hello world\n")
	}
}

func TestBdExecutorMissingBinary(t *testing.T) {
	exec := &BdExecutor{BinPath: "/nonexistent/binary"}
	_, err := exec.Execute("list")
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
}

func TestBdExecutorMissingBinary_ReturnsErrBdNotFound(t *testing.T) {
	exec := &BdExecutor{BinPath: "/nonexistent/binary"}
	_, err := exec.Execute("list")
	if !errors.Is(err, ErrBdNotFound) {
		t.Errorf("expected ErrBdNotFound, got %v", err)
	}
}

func TestBdExecutorMissingBinaryInPATH_ReturnsErrBdNotFound(t *testing.T) {
	exec := &BdExecutor{BinPath: "definitely-not-a-real-binary-xyz"}
	_, err := exec.Execute("list")
	if !errors.Is(err, ErrBdNotFound) {
		t.Errorf("expected ErrBdNotFound, got %v", err)
	}
}

func TestBdExecutorWorkDir(t *testing.T) {
	exec := &BdExecutor{BinPath: "pwd", WorkDir: "/tmp"}
	out, err := exec.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// /tmp may resolve to /private/tmp on macOS
	got := string(out)
	if got != "/tmp\n" && got != "/private/tmp\n" {
		t.Errorf("output = %q, want /tmp or /private/tmp", got)
	}
}

func TestBdExecutorDefaultBinPath(t *testing.T) {
	exec := &BdExecutor{}
	// This will fail if bd is not in PATH, but that's fine —
	// we just verify it tries to run "bd" by default
	_, err := exec.Execute("--version")
	// If bd is available, no error. If not, we just check it tried.
	if err != nil {
		t.Skipf("bd not in PATH, skipping: %v", err)
	}
}

func TestClientGetIssueExecutorError(t *testing.T) {
	mock := &mockExecutor{err: fmt.Errorf("show failed")}
	client := NewClient(mock)
	_, err := client.GetIssue("proj-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientListReadyExecutorError(t *testing.T) {
	mock := &mockExecutor{err: fmt.Errorf("ready failed")}
	client := NewClient(mock)
	_, err := client.ListReady()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBdExecutorExitError(t *testing.T) {
	// Run a command that exits with non-zero to trigger ExitError branch
	exec := &BdExecutor{BinPath: "sh"}
	_, err := exec.Execute("-c", "echo stderr >&2; exit 1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBdExecutorNoBeadsProject_ReturnsErrProjectNotInitialized(t *testing.T) {
	exec := &BdExecutor{BinPath: "sh"}
	_, err := exec.Execute("-c", "echo 'Error: no beads database found' >&2; exit 1")
	if !errors.Is(err, ErrProjectNotInitialized) {
		t.Errorf("expected ErrProjectNotInitialized, got %v", err)
	}
}

func TestBdExecutorExitError_ReturnsBdError(t *testing.T) {
	exec := &BdExecutor{BinPath: "sh"}
	_, err := exec.Execute("-c", "echo 'some other error' >&2; exit 1")
	var bdErr *BdError
	if !errors.As(err, &bdErr) {
		t.Fatalf("expected *BdError, got %T: %v", err, err)
	}
	if bdErr.Stderr != "some other error" {
		t.Errorf("Stderr = %q, want %q", bdErr.Stderr, "some other error")
	}
}

func TestParseIssueListInvalidJSON(t *testing.T) {
	_, err := ParseIssueList([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
