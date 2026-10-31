package gui

import (
	"context"
	"errors"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// negativeDuration matches a duration rendered with a minus sign, e.g. "-1s"
// or "-500ms". A timeout message must never carry one.
var negativeDuration = regexp.MustCompile(`-\d+(\.\d+)?(ns|µs|us|ms|s|m|h)`)

// TestBoundedReadInSpentParentBudgetIsNeverNegative covers the case where the
// parent's deadline has already passed: the remaining budget is negative, so
// the reported duration used to read "timed out after -1s". The budget must
// be clamped, and the read must not be started at all — the brew client must
// never be invoked with a dead context.
func TestBoundedReadInSpentParentBudgetIsNeverNegative(t *testing.T) {
	// A parent whose deadline is already in the past.
	parent, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	var called atomic.Bool
	read := func(ctx context.Context) (string, error) {
		called.Store(true)
		return "should not happen", nil
	}

	type result struct {
		value string
		err   error
	}
	done := make(chan result, 1)
	go func() {
		v, err := boundedReadIn(parent, 30*time.Second, "list", read)
		done <- result{value: v, err: err}
	}()

	var res result
	select {
	case res = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("boundedReadIn did not return promptly for a spent parent")
	}

	if res.err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	var toErr *brew.TimeoutError
	if !errors.As(res.err, &toErr) {
		t.Fatalf("expected *brew.TimeoutError, got %T: %v", res.err, res.err)
	}
	if toErr.Command != "list" {
		t.Errorf("timeout command = %q, want %q", toErr.Command, "list")
	}
	if toErr.Timeout < 0 {
		t.Errorf("timeout duration = %s, want a non-negative duration", toErr.Timeout)
	}
	if negativeDuration.MatchString(res.err.Error()) {
		t.Errorf("timeout message %q reports a negative duration", res.err.Error())
	}
	if called.Load() {
		t.Error("the read was invoked even though the parent's budget was already spent")
	}
}
