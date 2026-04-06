package datasource

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryExecutorImplementsExecutor(t *testing.T) {
	var _ Executor = (*RetryExecutor)(nil)
}

func TestRetryExecutorPassesThroughOnSuccess(t *testing.T) {
	inner := &mockExecutor{output: []byte("ok")}
	retry := NewRetryExecutor(inner)

	out, err := retry.Execute("list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "ok" {
		t.Errorf("output = %q, want %q", string(out), "ok")
	}
	if len(inner.calls) != 1 {
		t.Errorf("calls = %d, want 1", len(inner.calls))
	}
}

// sleepRecorder captures sleep durations for test verification.
type sleepRecorder struct {
	durations []time.Duration
}

func (s *sleepRecorder) sleep(d time.Duration) {
	s.durations = append(s.durations, d)
}

// failThenSucceedExecutor fails with a given error N times, then succeeds.
type failThenSucceedExecutor struct {
	failErr    error
	failCount  int
	callCount  int
	output     []byte
}

func (e *failThenSucceedExecutor) Execute(args ...string) ([]byte, error) {
	e.callCount++
	if e.callCount <= e.failCount {
		return nil, e.failErr
	}
	return e.output, nil
}

func TestRetryExecutorRetriesOnLockAndSucceeds(t *testing.T) {
	lockErr := fmt.Errorf("%w: another process holds the exclusive lock", ErrDatabaseLocked)
	inner := &failThenSucceedExecutor{
		failErr:   lockErr,
		failCount: 2,
		output:    []byte("ok"),
	}
	sr := &sleepRecorder{}
	retry := NewRetryExecutor(inner)
	retry.sleep = sr.sleep

	out, err := retry.Execute("list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "ok" {
		t.Errorf("output = %q, want %q", string(out), "ok")
	}
	if inner.callCount != 3 {
		t.Errorf("callCount = %d, want 3 (1 initial + 2 retries)", inner.callCount)
	}
}

func TestRetryExecutorExhaustsRetriesAndReturnsError(t *testing.T) {
	lockErr := fmt.Errorf("%w: another process holds the exclusive lock", ErrDatabaseLocked)
	inner := &failThenSucceedExecutor{
		failErr:   lockErr,
		failCount: 100, // always fail
		output:    []byte("ok"),
	}
	sr := &sleepRecorder{}
	retry := NewRetryExecutor(inner)
	retry.sleep = sr.sleep

	_, err := retry.Execute("list", "--json")
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	if !errors.Is(err, ErrDatabaseLocked) {
		t.Errorf("expected ErrDatabaseLocked, got %v", err)
	}
	// 1 initial + 5 retries = 6 total calls
	if inner.callCount != 6 {
		t.Errorf("callCount = %d, want 6 (1 initial + 5 retries)", inner.callCount)
	}
}

func TestRetryExecutorDoesNotRetryNonLockErrors(t *testing.T) {
	inner := &mockExecutor{err: fmt.Errorf("some other error")}
	sr := &sleepRecorder{}
	retry := NewRetryExecutor(inner)
	retry.sleep = sr.sleep

	_, err := retry.Execute("list", "--json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(inner.calls) != 1 {
		t.Errorf("calls = %d, want 1 (should not retry non-lock errors)", len(inner.calls))
	}
	if len(sr.durations) != 0 {
		t.Errorf("sleep called %d times, want 0", len(sr.durations))
	}
}

func TestRetryExecutorStopsRetryingOnNonLockError(t *testing.T) {
	// First call: lock error (triggers retry), second call: different error (stops)
	inner := &sequenceExecutor{
		results: []execResult{
			{err: fmt.Errorf("%w: locked", ErrDatabaseLocked)},
			{err: fmt.Errorf("unexpected failure")},
		},
	}
	sr := &sleepRecorder{}
	retry := NewRetryExecutor(inner)
	retry.sleep = sr.sleep

	_, err := retry.Execute("list")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unexpected failure" {
		t.Errorf("error = %q, want %q", err.Error(), "unexpected failure")
	}
	if inner.callCount != 2 {
		t.Errorf("callCount = %d, want 2", inner.callCount)
	}
	if len(sr.durations) != 1 {
		t.Errorf("sleep called %d times, want 1", len(sr.durations))
	}
}

type execResult struct {
	output []byte
	err    error
}

type sequenceExecutor struct {
	results   []execResult
	callCount int
}

func (e *sequenceExecutor) Execute(args ...string) ([]byte, error) {
	idx := e.callCount
	e.callCount++
	if idx < len(e.results) {
		return e.results[idx].output, e.results[idx].err
	}
	return nil, fmt.Errorf("no more results")
}

func TestRetryExecutorUsesExponentialBackoff(t *testing.T) {
	lockErr := fmt.Errorf("%w: locked", ErrDatabaseLocked)
	inner := &failThenSucceedExecutor{
		failErr:   lockErr,
		failCount: 100, // always fail
	}
	sr := &sleepRecorder{}
	retry := NewRetryExecutor(inner)
	retry.sleep = sr.sleep

	_, _ = retry.Execute("list")

	want := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
		1600 * time.Millisecond,
	}
	if len(sr.durations) != len(want) {
		t.Fatalf("sleep called %d times, want %d", len(sr.durations), len(want))
	}
	for i, d := range sr.durations {
		if d != want[i] {
			t.Errorf("sleep[%d] = %v, want %v", i, d, want[i])
		}
	}
}
