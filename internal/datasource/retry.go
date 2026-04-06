package datasource

import (
	"errors"
	"time"
)

// RetryExecutor wraps an Executor and retries on ErrDatabaseLocked
// with exponential backoff.
type RetryExecutor struct {
	inner      Executor
	maxRetries int
	baseDelay  time.Duration
	sleep      func(time.Duration)
}

// NewRetryExecutor creates a RetryExecutor with sensible defaults:
// 5 retries with 100ms base delay (100, 200, 400, 800, 1600ms).
func NewRetryExecutor(inner Executor) *RetryExecutor {
	return &RetryExecutor{
		inner:      inner,
		maxRetries: 5,
		baseDelay:  100 * time.Millisecond,
		sleep:      time.Sleep,
	}
}

// Execute runs the inner executor, retrying with exponential backoff
// when ErrDatabaseLocked is encountered.
func (r *RetryExecutor) Execute(args ...string) ([]byte, error) {
	out, err := r.inner.Execute(args...)
	if err == nil {
		return out, nil
	}
	if !errors.Is(err, ErrDatabaseLocked) {
		return nil, err
	}
	delay := r.baseDelay
	for attempt := 0; attempt < r.maxRetries; attempt++ {
		r.sleep(delay)
		out, err = r.inner.Execute(args...)
		if err == nil {
			return out, nil
		}
		if !errors.Is(err, ErrDatabaseLocked) {
			return nil, err
		}
		delay *= 2
	}
	return nil, err
}
