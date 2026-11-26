package brew

// Tests for LB-48: exec's copy goroutines can outlive cmd.Wait after
// exec.ErrWaitDelay, and a brew that never started keeps its cause.

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"testing"
)

// TestLineWriterDropsWritesAfterFlush: once flushed, a late write from a copy
// still running after WaitDelay sends nothing, so it never sends on the
// stream channel ExecuteStream closes after flush.
func TestLineWriterDropsWritesAfterFlush(t *testing.T) {
	out := make(chan string, 10)
	w := &lineWriter{ctx: context.Background(), out: out, keep: &bytes.Buffer{}}
	if _, err := w.Write([]byte("one\ntwo")); err != nil {
		t.Fatal(err)
	}
	w.flush()
	close(out)
	n, err := w.Write([]byte("late\n"))
	if err != nil || n != 5 {
		t.Fatalf("late Write = (%d, %v), want (5, nil)", n, err)
	}
	var got []string
	for line := range out {
		got = append(got, line)
	}
	if len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Errorf("lines = %q, want [one two]", got)
	}
	if k := w.kept(); k != "one\ntwo" {
		t.Errorf("kept = %q, want the text written before flush", k)
	}
}

// TestClosingBufferDropsWritesAfterClose: Execute's buffers ignore a copy that
// outlives cmd.Run, so the bytes it returns never change under it.
func TestClosingBufferDropsWritesAfterClose(t *testing.T) {
	var b closingBuffer
	_, _ = b.Write([]byte("out"))
	got := b.close()
	if n, err := b.Write([]byte("late")); err != nil || n != 4 {
		t.Fatalf("late Write = (%d, %v), want (4, nil)", n, err)
	}
	if string(got) != "out" || string(b.close()) != "out" {
		t.Errorf("close = %q, want out", got)
	}
}

// TestExecuteStartFailureKeepsCause: a brew path that no longer exists
// surfaces the real cause, not "exited with code 0".
func TestExecuteStartFailureKeepsCause(t *testing.T) {
	r := &DefaultRunner{brewPath: t.TempDir() + "/missing-brew"}
	_, err := r.Execute(context.Background(), "info")
	if err == nil {
		t.Fatal("Execute succeeded with a missing binary")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want it to wrap fs.ErrNotExist", err)
	}
	var exit *BrewExitError
	if errors.As(err, &exit) {
		t.Errorf("err = %v is a BrewExitError; brew never ran", err)
	}
}

// TestExecuteCanceledBeforeStartIsContextError: a ctx done before brew starts
// is reported as that ctx error.
func TestExecuteCanceledBeforeStartIsContextError(t *testing.T) {
	r := &DefaultRunner{brewPath: "/bin/true"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := r.Execute(ctx, "info")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
