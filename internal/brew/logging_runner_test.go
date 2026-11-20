package brew

import (
	"context"
	"errors"
	"testing"
)

// streamingRunner is a mock whose ExecuteStream yields no lines and reports
// err once on its error channel, as a failed brew command does.
func streamingRunner(err error) *MockRunner {
	return &MockRunner{
		ExecuteStreamFn: func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
			lines := make(chan string)
			close(lines)
			errs := make(chan error, 1)
			if err != nil {
				errs <- err
			}
			close(errs)
			return lines, errs
		},
	}
}

// A stream error reaches both the caller and OnExec. The old code had the
// logger and the caller both read one error channel, so one of them saw nil
// and a failed install could look like a success. Repeated, so the race
// between the two readers shows up often under -race.
func TestLoggingRunnerStreamErrorReachesCallerAndLog(t *testing.T) {
	want := errors.New("brew install failed")
	for i := 0; i < 200; i++ {
		var logged error
		r := NewLoggingRunner(streamingRunner(want), nil, func(args []string, err error) {
			logged = err
		})
		_, errs := r.ExecuteStream(context.Background(), "install", "wget")
		if got := <-errs; !errors.Is(got, want) {
			t.Fatalf("iteration %d: caller got %v, want %v", i, got, want)
		}
		if !errors.Is(logged, want) {
			t.Fatalf("iteration %d: OnExec got %v, want %v", i, logged, want)
		}
	}
}

// A successful stream reaches the caller as nil, and OnExec sees nil.
func TestLoggingRunnerStreamSuccessIsNilForBoth(t *testing.T) {
	for i := 0; i < 200; i++ {
		var logged error
		logCalls := 0
		r := NewLoggingRunner(streamingRunner(nil), nil, func(args []string, err error) {
			logged = err
			logCalls++
		})
		lines, errs := r.ExecuteStream(context.Background(), "upgrade", "wget")
		for range lines {
		}
		if got := <-errs; got != nil {
			t.Fatalf("iteration %d: caller got %v, want nil", i, got)
		}
		if logged != nil || logCalls != 1 {
			t.Fatalf("iteration %d: OnExec got %v (%d calls), want nil once", i, logged, logCalls)
		}
	}
}

// With no OnExec the stream is the inner one, unchanged: the caller gets the
// inner error channel itself.
func TestLoggingRunnerStreamWithoutOnExecPassesInnerThrough(t *testing.T) {
	want := errors.New("brew upgrade failed")
	inner := streamingRunner(want)
	r := NewLoggingRunner(inner, nil, nil)
	_, errs := r.ExecuteStream(context.Background(), "upgrade", "wget")
	if got := <-errs; !errors.Is(got, want) {
		t.Fatalf("caller got %v, want %v", got, want)
	}
}
