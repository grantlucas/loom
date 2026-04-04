package datasource

import (
	"fmt"
	"testing"
	"time"
)

// countingExecutor tracks how many times Execute is called.
type countingExecutor struct {
	callCount int
	output    []byte
	err       error
}

func (e *countingExecutor) Execute(args ...string) ([]byte, error) {
	e.callCount++
	return e.output, e.err
}

const listJSON = `[{"id":"proj-1","title":"Test","status":"open","priority":1,"issue_type":"task","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]`

const detailJSON = `[{"id":"proj-2","title":"Detail","status":"open","priority":1,"issue_type":"task","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","dependencies":[],"dependents":[]}]`

func TestCacheListIssuesDelegatesToClientOnMiss(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	issues, err := cache.ListIssues()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].ID != "proj-1" {
		t.Errorf("ID = %q, want %q", issues[0].ID, "proj-1")
	}
	if exec.callCount != 1 {
		t.Errorf("callCount = %d, want 1", exec.callCount)
	}
}

func TestCacheListIssuesReturnsCachedOnHit(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	// First call — miss
	_, _ = cache.ListIssues()
	// Second call — hit
	issues, err := cache.ListIssues()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != "proj-1" {
		t.Errorf("unexpected issues on cache hit: %+v", issues)
	}
	if exec.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (second call should be cached)", exec.callCount)
	}
}

func TestCacheListIssuesExpiresAfterTTL(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	fakeNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return fakeNow }

	// First call — miss
	_, _ = cache.ListIssues()
	// Advance past TTL
	fakeNow = fakeNow.Add(6 * time.Minute)
	// Second call — expired, should re-fetch
	_, err := cache.ListIssues()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (entry should have expired)", exec.callCount)
	}
}

func TestCacheGetIssueDelegatesToClientOnMiss(t *testing.T) {
	exec := &countingExecutor{output: []byte(detailJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	detail, err := cache.GetIssue("proj-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.ID != "proj-2" {
		t.Errorf("ID = %q, want %q", detail.ID, "proj-2")
	}
	if exec.callCount != 1 {
		t.Errorf("callCount = %d, want 1", exec.callCount)
	}
}

func TestCacheGetIssueReturnsCachedOnHit(t *testing.T) {
	exec := &countingExecutor{output: []byte(detailJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	_, _ = cache.GetIssue("proj-2")
	detail, err := cache.GetIssue("proj-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.ID != "proj-2" {
		t.Errorf("ID = %q, want %q", detail.ID, "proj-2")
	}
	if exec.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (second call should be cached)", exec.callCount)
	}
}

func TestCacheGetIssueDifferentIDsAreSeparate(t *testing.T) {
	exec := &countingExecutor{output: []byte(detailJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	_, _ = cache.GetIssue("proj-2")
	_, _ = cache.GetIssue("proj-3")
	if exec.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (different IDs should be separate cache entries)", exec.callCount)
	}
}

func TestCacheListReadyDelegatesToClientOnMiss(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	issues, err := cache.ListReady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != "proj-1" {
		t.Errorf("unexpected issues: %+v", issues)
	}
	if exec.callCount != 1 {
		t.Errorf("callCount = %d, want 1", exec.callCount)
	}
}

func TestCacheListReadyReturnsCachedOnHit(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	_, _ = cache.ListReady()
	_, err := cache.ListReady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (second call should be cached)", exec.callCount)
	}
}

func TestCacheErrorsAreNotCached(t *testing.T) {
	exec := &countingExecutor{err: fmt.Errorf("command failed")}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	_, err := cache.ListIssues()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Second call should retry (error was not cached)
	_, _ = cache.ListIssues()
	if exec.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (errors should not be cached)", exec.callCount)
	}
}

func TestCacheGetIssueErrorNotCached(t *testing.T) {
	exec := &countingExecutor{err: fmt.Errorf("show failed")}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	_, err := cache.GetIssue("proj-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	_, _ = cache.GetIssue("proj-1")
	if exec.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (errors should not be cached)", exec.callCount)
	}
}

func TestCacheListReadyErrorNotCached(t *testing.T) {
	exec := &countingExecutor{err: fmt.Errorf("ready failed")}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	_, err := cache.ListReady()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	_, _ = cache.ListReady()
	if exec.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (errors should not be cached)", exec.callCount)
	}
}

func TestCacheInvalidateClearsAllEntries(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	// Populate cache
	_, _ = cache.ListIssues()
	_, _ = cache.ListReady()
	if exec.callCount != 2 {
		t.Fatalf("setup: callCount = %d, want 2", exec.callCount)
	}

	// Invalidate
	cache.Invalidate()

	// Both should re-fetch
	_, _ = cache.ListIssues()
	_, _ = cache.ListReady()
	if exec.callCount != 4 {
		t.Errorf("callCount = %d, want 4 (both should re-fetch after invalidation)", exec.callCount)
	}
}

func TestCacheGetIssueExpiresAfterTTL(t *testing.T) {
	exec := &countingExecutor{output: []byte(detailJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	fakeNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return fakeNow }

	_, _ = cache.GetIssue("proj-2")
	fakeNow = fakeNow.Add(6 * time.Minute)
	_, err := cache.GetIssue("proj-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (entry should have expired)", exec.callCount)
	}
}

func TestCacheListReadyExpiresAfterTTL(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	fakeNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return fakeNow }

	_, _ = cache.ListReady()
	fakeNow = fakeNow.Add(6 * time.Minute)
	_, err := cache.ListReady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (entry should have expired)", exec.callCount)
	}
}

// TestDataSourceInterfaceCompliance verifies that both Client and Cache
// satisfy the DataSource interface at compile time.
func TestDataSourceInterfaceCompliance(t *testing.T) {
	var _ DataSource = (*Client)(nil)
	var _ DataSource = (*Cache)(nil)
}

func TestCacheNotExpiredBeforeTTL(t *testing.T) {
	exec := &countingExecutor{output: []byte(listJSON)}
	client := NewClient(exec)
	cache := NewCache(client, 5*time.Minute)

	fakeNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return fakeNow }

	_, _ = cache.ListIssues()
	// Advance but stay within TTL
	fakeNow = fakeNow.Add(4 * time.Minute)
	_, _ = cache.ListIssues()
	if exec.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (entry should still be valid)", exec.callCount)
	}
}
