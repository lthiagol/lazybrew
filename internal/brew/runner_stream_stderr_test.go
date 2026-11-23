package brew

// Tests for brew's stderr lines on DefaultRunner.ExecuteStream (LB-43).

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

// collectStream drains a stream until its lines channel closes, then returns
// the lines and the stream's error. It fails the test if the stream hangs.
func collectStream(t *testing.T, out <-chan string, errCh <-chan error) ([]string, error) {
	t.Helper()
	var lines []string
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range out {
			lines = append(lines, line)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteStream did not close its lines channel")
	}
	return lines, <-errCh
}

// TestExecuteStreamSendsStderrLines: brew's stderr lines reach the stream, split
// like stdout. A \r ends a line too, as brew's progress redraws do.
func TestExecuteStreamSendsStderrLines(t *testing.T) {
	r := writeFakeBrew(t, "echo first >&2\nprintf 'half\\r' >&2\necho last >&2\n")

	out, errCh := r.ExecuteStream(context.Background(), "upgrade")
	lines, err := collectStream(t, out, errCh)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"first", "half", "last"}
	if !slices.Equal(lines, want) {
		t.Errorf("lines = %q, want %q", lines, want)
	}
}

// TestExecuteStreamKeepsOrderOfBothStreams: stdout and stderr share one
// channel. Their lines may interleave, but each stream keeps its own order and
// no line mixes bytes from both.
func TestExecuteStreamKeepsOrderOfBothStreams(t *testing.T) {
	const n = 8
	r := writeFakeBrew(t, fmt.Sprintf(`for i in $(seq 1 %d); do
echo out-$i
echo err-$i >&2
done
`, n))

	out, errCh := r.ExecuteStream(context.Background(), "upgrade")
	lines, err := collectStream(t, out, errCh)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr []string
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "out-"):
			stdout = append(stdout, line)
		case strings.HasPrefix(line, "err-"):
			stderr = append(stderr, line)
		default:
			t.Fatalf("line %q is neither stdout nor stderr", line)
		}
	}
	var wantOut, wantErr []string
	for i := 1; i <= n; i++ {
		wantOut = append(wantOut, fmt.Sprintf("out-%d", i))
		wantErr = append(wantErr, fmt.Sprintf("err-%d", i))
	}
	if !slices.Equal(stdout, wantOut) {
		t.Errorf("stdout lines = %q, want %q", stdout, wantOut)
	}
	if !slices.Equal(stderr, wantErr) {
		t.Errorf("stderr lines = %q, want %q", stderr, wantErr)
	}
}

// TestExecuteStreamErrorKeepsWholeStderr: a failed run still streams the lines
// of both streams, and its error is built from the whole stderr text.
func TestExecuteStreamErrorKeepsWholeStderr(t *testing.T) {
	r := writeFakeBrew(t, "echo partial-stdout\necho problem one >&2\necho problem two >&2\nexit 2\n")

	out, errCh := r.ExecuteStream(context.Background(), "upgrade")
	lines, err := collectStream(t, out, errCh)
	if err == nil {
		t.Fatal("expected an error from a failing command")
	}
	if want := "brew upgrade: problem one\nproblem two"; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}

	// The stdout and stderr lines may come in either order.
	got := slices.Clone(lines)
	slices.Sort(got)
	want := []string{"partial-stdout", "problem one", "problem two"}
	if !slices.Equal(got, want) {
		t.Errorf("lines = %q, want %q", lines, want)
	}
}

// TestExecuteStreamCancelWithStderrOutput: a cancel while brew is writing to
// stderr still closes the stream and reports context.Canceled.
func TestExecuteStreamCancelWithStderrOutput(t *testing.T) {
	r := writeFakeBrew(t, "while true; do echo tick >&2; sleep 0.005; done\n")

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		out, errCh := r.ExecuteStream(ctx, "whatever")

		// Cancel once brew is running and writing to stderr. The wait is
		// bounded, so a stream that never sends stderr lines fails here.
		for range 3 {
			select {
			case <-out:
			case <-time.After(5 * time.Second):
				t.Fatal("no stderr line arrived before the cancel")
			}
		}
		cancel()

		_, err := collectStream(t, out, errCh)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("iteration %d: expected context.Canceled after cancel, got %v", i, err)
		}
	}
}
