package brew

// Tests for the r2 review round on DefaultRunner.ExecuteStream.

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestExecuteStreamCancelWinsOverReaderError locks r2 finding NEW-7: when
// a cancel kills the process, the reported error must be
// context.Canceled even if the stdout reader hits a read error first and
// consumes the single-error slot. A chatty script maximizes the chance
// that the reader is mid-Read when the kill lands, so both orderings
// (EOF first, read error first) are exercised across iterations.
func TestExecuteStreamCancelWinsOverReaderError(t *testing.T) {
	r := writeFakeBrew(t, "while true; do echo tick; sleep 0.02; done\n")

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		out, errCh := r.ExecuteStream(ctx, "whatever")

		time.Sleep(80 * time.Millisecond)
		cancel()

		// Drain stdout until close so the channels resolve.
		go func() {
			for range out {
			}
		}()

		select {
		case err := <-errCh:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("iteration %d: expected context.Canceled after cancel, got %v", i, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("iteration %d: ExecuteStream did not report the cancel in time", i)
		}
	}
}

// TestExecuteStreamCanceledContextBeforeStart locks the other NEW-7 arm:
// an error produced while the context is already cancelled (here the
// start failure) must still surface as context.Canceled, not the local
// error.
func TestExecuteStreamCanceledContextBeforeStart(t *testing.T) {
	r := writeFakeBrew(t, "echo hi\n")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out, errCh := r.ExecuteStream(ctx, "whatever")
	for range out {
	}

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled for an already-cancelled context, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteStream did not report the cancel in time")
	}
}
