package brew

import (
	"context"
	"errors"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// streamResult is the channel pair of a stream that yields no lines and
// reports err once on its error channel, as a failed brew command does.
func streamResult(err error) (<-chan string, <-chan error) {
	lines := make(chan string)
	close(lines)
	errs := make(chan error, 1)
	if err != nil {
		errs <- err
	}
	close(errs)
	return lines, errs
}

// streamingRunner is a mock whose ExecuteStream yields no lines and reports
// err once on its error channel.
func streamingRunner(err error) *brewtest.MockRunner {
	return &brewtest.MockRunner{
		ExecuteStreamFn: func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
			return streamResult(err)
		},
	}
}

// loggingKinds runs each of the three Execute kinds once through r. Each
// inner run first calls onRun, then fails with err.
var loggingKinds = []struct {
	name  string
	inner func(onRun func(), err error) Runner
	call  func(r *LoggingRunner) error
}{
	{
		name: "Execute",
		inner: func(onRun func(), err error) Runner {
			return &brewtest.MockRunner{ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
				onRun()
				return nil, err
			}}
		},
		call: func(r *LoggingRunner) error {
			_, err := r.Execute(context.Background(), "install", "wget")
			return err
		},
	},
	{
		name: "ExecuteJSON",
		inner: func(onRun func(), err error) Runner {
			return &brewtest.MockRunner{ExecuteJSONFn: func(ctx context.Context, result any, args ...string) error {
				onRun()
				return err
			}}
		},
		call: func(r *LoggingRunner) error {
			var v any
			return r.ExecuteJSON(context.Background(), &v, "info", "--json=v2", "wget")
		},
	},
	{
		name: "ExecuteStream",
		inner: func(onRun func(), err error) Runner {
			return &brewtest.MockRunner{ExecuteStreamFn: func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
				onRun()
				return streamResult(err)
			}}
		},
		call: func(r *LoggingRunner) error {
			lines, errs := r.ExecuteStream(context.Background(), "upgrade", "wget")
			for range lines {
			}
			return <-errs
		},
	},
}

// A stream error reaches both the caller and the finisher. The old code had the
// logger and the caller both read one error channel, so one of them saw nil
// and a failed install could look like a success. Repeated, so the race
// between the two readers shows up often under -race.
func TestLoggingRunnerStreamErrorReachesCallerAndFinisher(t *testing.T) {
	want := errors.New("brew install failed")
	for i := 0; i < 200; i++ {
		var finished error
		r := NewLoggingRunner(streamingRunner(want), func(args []string) func(error) {
			return func(err error) { finished = err }
		})
		_, errs := r.ExecuteStream(context.Background(), "install", "wget")
		if got := <-errs; !errors.Is(got, want) {
			t.Fatalf("iteration %d: caller got %v, want %v", i, got, want)
		}
		if !errors.Is(finished, want) {
			t.Fatalf("iteration %d: finisher got %v, want %v", i, finished, want)
		}
	}
}

// A successful stream reaches the caller as nil, and the finisher sees nil once.
func TestLoggingRunnerStreamSuccessIsNilForBoth(t *testing.T) {
	for i := 0; i < 200; i++ {
		var finished error
		finishes := 0
		r := NewLoggingRunner(streamingRunner(nil), func(args []string) func(error) {
			return func(err error) {
				finished = err
				finishes++
			}
		})
		lines, errs := r.ExecuteStream(context.Background(), "upgrade", "wget")
		for range lines {
		}
		if got := <-errs; got != nil {
			t.Fatalf("iteration %d: caller got %v, want nil", i, got)
		}
		if finished != nil || finishes != 1 {
			t.Fatalf("iteration %d: finisher got %v (%d calls), want nil once", i, finished, finishes)
		}
	}
}

// Each Execute kind starts the hook once, before the command runs, and calls
// the finisher once with the command's final error.
func TestLoggingRunnerHookStartsBeforeRunAndFinishesOnce(t *testing.T) {
	want := errors.New("brew install failed")
	for _, kind := range loggingKinds {
		t.Run(kind.name, func(t *testing.T) {
			starts := 0
			var finishes []error
			hook := func(args []string) func(error) {
				starts++
				return func(err error) { finishes = append(finishes, err) }
			}
			inner := kind.inner(func() {
				if starts != 1 {
					t.Errorf("command ran after %d starts, want 1 before it runs", starts)
				}
			}, want)
			if got := kind.call(NewLoggingRunner(inner, hook)); !errors.Is(got, want) {
				t.Fatalf("caller got %v, want %v", got, want)
			}
			if starts != 1 {
				t.Fatalf("hook started %d times, want once", starts)
			}
			if len(finishes) != 1 || !errors.Is(finishes[0], want) {
				t.Fatalf("finisher got %v, want %v once", finishes, want)
			}
		})
	}
}

// A nil hook logs nothing and leaves every Execute kind's result unchanged.
func TestLoggingRunnerNilHookChangesNothing(t *testing.T) {
	want := errors.New("brew upgrade failed")
	for _, kind := range loggingKinds {
		t.Run(kind.name, func(t *testing.T) {
			r := NewLoggingRunner(kind.inner(func() {}, want), nil)
			if got := kind.call(r); !errors.Is(got, want) {
				t.Fatalf("caller got %v, want %v", got, want)
			}
		})
	}
}

// With no hook the stream is the inner one, unchanged: the caller gets the
// inner error channel itself.
func TestLoggingRunnerStreamWithoutHookPassesInnerThrough(t *testing.T) {
	want := errors.New("brew upgrade failed")
	inner := streamingRunner(want)
	r := NewLoggingRunner(inner, nil)
	_, errs := r.ExecuteStream(context.Background(), "upgrade", "wget")
	if got := <-errs; !errors.Is(got, want) {
		t.Fatalf("caller got %v, want %v", got, want)
	}
}
