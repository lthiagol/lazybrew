package brew

// Tests for the r1 review round on DefaultRunner.ExecuteStream:
// single-error send (no double-send deadlock) and cancel mapping to
// context.Canceled.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFakeBrew(t *testing.T, body string) *DefaultRunner {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "brew")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := NewDefaultRunnerWithPath(script)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestExecuteStreamSingleErrorThenClose locks r1 finding 8: the stream
// must deliver exactly one error and close both channels even when the
// command both writes output and fails — a second errChan send used to
// deadlock the goroutine.
func TestExecuteStreamSingleErrorThenClose(t *testing.T) {
	r := writeFakeBrew(t, "echo line1\necho line2\necho line3\nexit 3\n")

	done := make(chan struct{})
	var lines []string
	var gotErr error

	out, errCh := r.ExecuteStream(context.Background(), "whatever")
	go func() {
		defer close(done)
		for line := range out {
			lines = append(lines, line)
		}
		gotErr = <-errCh
		// A second receive must yield the zero value immediately: the
		// channel is closed after exactly one error.
		select {
		case err, ok := <-errCh:
			if ok {
				t.Errorf("expected closed errChan after one error, got %v", err)
			}
		default:
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteStream deadlocked after the error (double send on errChan)")
	}

	if len(lines) != 3 {
		t.Fatalf("expected 3 output lines, got %v", lines)
	}
	if gotErr == nil {
		t.Fatal("expected an error from a failing command")
	}
}

// TestExecuteStreamWaitAfterReader locks the reader/Wait synchronization
// part of r1 finding 8: with StdoutPipe, cmd.Wait closes the pipe, so a
// reader racing Wait used to produce a spurious "stdout read: file already
// closed" error. A successful run must yield NO error at all.
func TestExecuteStreamNoSpuriousReadError(t *testing.T) {
	r := writeFakeBrew(t, "for i in 1 2 3 4 5; do echo line-$i; done\nexit 0\n")

	out, errCh := r.ExecuteStream(context.Background(), "whatever")
	var lines []string
	for line := range out {
		lines = append(lines, line)
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected no error on success, got %v (lines: %v)", err, lines)
		}
	default:
	}
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %v", lines)
	}
}

// TestExecuteStreamCancelReportsCanceled locks r1 finding 10: killing the
// process on cancel must report context.Canceled (which the GUI maps to
// the "Cancelled" operation state), not "signal: killed".
func TestExecuteStreamCancelReportsCanceled(t *testing.T) {
	r := writeFakeBrew(t, "sleep 30\n")

	ctx, cancel := context.WithCancel(context.Background())
	out, errCh := r.ExecuteStream(ctx, "whatever")

	// Let the process start, then cancel.
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Drain stdout until close so the channels resolve.
	for range out {
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled after cancel, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteStream did not report the cancel in time")
	}
}
