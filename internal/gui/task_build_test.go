package gui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// runBuilt runs a built task and returns its output and error channels.
func runBuilt(t *testing.T, tk *task.Task, ctx context.Context) (<-chan string, <-chan error) {
	t.Helper()
	out, errCh, err := tk.Run(ctx)
	if err != nil {
		t.Fatalf("Run returned %v, want nil", err)
	}
	return out, errCh
}

// drain reads out until it is closed, failing if that takes too long.
func drain(t *testing.T, out <-chan string) []string {
	t.Helper()
	var lines []string
	timeout := time.After(2 * time.Second)
	for {
		select {
		case line, ok := <-out:
			if !ok {
				return lines
			}
			lines = append(lines, line)
		case <-timeout:
			t.Fatal("output channel was not closed")
		}
	}
}

func TestErrTaskSuccess(t *testing.T) {
	tk := errTask("id", "Title", func(ctx context.Context) error { return nil })
	if tk.ID != "id" || tk.Title != "Title" {
		t.Fatalf("task = %q %q, want id Title", tk.ID, tk.Title)
	}
	out, errCh := runBuilt(t, tk, context.Background())
	if lines := drain(t, out); len(lines) != 0 {
		t.Fatalf("lines = %v, want none", lines)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestErrTaskError(t *testing.T) {
	want := errors.New("brew failed")
	out, errCh := runBuilt(t, errTask("id", "Title", func(ctx context.Context) error {
		return want
	}), context.Background())
	drain(t, out)
	if err := <-errCh; !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

func TestErrTaskPassesContext(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "marker")
	var got any
	out, errCh := runBuilt(t, errTask("id", "Title", func(c context.Context) error {
		got = c.Value(key{})
		return nil
	}), ctx)
	drain(t, out)
	<-errCh
	if got != "marker" {
		t.Fatalf("run got ctx value %v, want marker", got)
	}
}

func TestStreamTaskPassesLines(t *testing.T) {
	lines := make(chan string, 2)
	lines <- "one"
	lines <- "two"
	close(lines)
	errCh := make(chan error, 1)
	errCh <- nil

	out, gotErr := runBuilt(t, streamTask("id", "Title", func(ctx context.Context) (<-chan string, <-chan error) {
		return lines, errCh
	}), context.Background())
	if got := drain(t, out); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("lines = %v, want [one two]", got)
	}
	if err := <-gotErr; err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestStreamTaskNilLineChannel(t *testing.T) {
	errCh := make(chan error, 1)
	errCh <- errors.New("no output")

	out, gotErr := runBuilt(t, streamTask("id", "Title", func(ctx context.Context) (<-chan string, <-chan error) {
		return nil, errCh
	}), context.Background())
	if out == nil {
		t.Fatal("out is nil, want a closed channel")
	}
	if lines := drain(t, out); len(lines) != 0 {
		t.Fatalf("lines = %v, want none", lines)
	}
	if err := <-gotErr; err == nil || err.Error() != "no output" {
		t.Fatalf("err = %v, want the run's error", err)
	}
}

func TestStreamTaskNilErrorChannel(t *testing.T) {
	lines := make(chan string)
	close(lines)
	out, gotErr := runBuilt(t, streamTask("id", "Title", func(ctx context.Context) (<-chan string, <-chan error) {
		return lines, nil
	}), context.Background())
	drain(t, out)
	if gotErr != nil {
		t.Fatalf("errCh = %v, want nil passed through", gotErr)
	}
}

func TestLinesTaskSendsLinesInOrder(t *testing.T) {
	out, errCh := runBuilt(t, linesTask("id", "Title", func(ctx context.Context) ([]string, error) {
		return []string{"a", "b", "c"}, nil
	}), context.Background())
	if got := drain(t, out); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("lines = %v, want [a b c]", got)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestLinesTaskEmptyResult(t *testing.T) {
	out, errCh := runBuilt(t, linesTask("id", "Title", func(ctx context.Context) ([]string, error) {
		return nil, nil
	}), context.Background())
	if got := drain(t, out); len(got) != 0 {
		t.Fatalf("lines = %v, want none", got)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestLinesTaskError(t *testing.T) {
	want := errors.New("fetch failed")
	out, errCh := runBuilt(t, linesTask("id", "Title", func(ctx context.Context) ([]string, error) {
		return []string{"never sent"}, want
	}), context.Background())
	if got := drain(t, out); len(got) != 0 {
		t.Fatalf("lines = %v, want none after a fetch error", got)
	}
	if err := <-errCh; !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

// errWithin reads the task's error, failing if none arrives within 2 s. A task
// whose error never comes would otherwise block the test until the run times out.
func errWithin(t *testing.T, errCh <-chan error) error {
	t.Helper()
	select {
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("errCh delivered nothing within 2s")
		return nil
	}
}

// A manager that stops reading and cancels the task must not leave the
// goroutine blocked on a send: out closes, and the cancelled task ends with
// context.Canceled on errCh, so the manager's read of errCh returns.
func TestLinesTaskCancelMidStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}
	out, errCh := runBuilt(t, linesTask("id", "Title", func(context.Context) ([]string, error) {
		return lines, nil
	}), ctx)

	<-out
	cancel()
	drain(t, out)
	if err := errWithin(t, errCh); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v after cancel, want context.Canceled", err)
	}
}

// A task cancelled before it starts sending ends with context.Canceled, even
// when fetch returns lines: the lines are not all sent, or are sent by chance.
func TestLinesTaskCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out, errCh := runBuilt(t, linesTask("id", "Title", func(context.Context) ([]string, error) {
		return []string{"a", "b"}, nil
	}), ctx)
	drain(t, out)
	if err := errWithin(t, errCh); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

// hundredLines is the output of a diagnostics check large enough that a
// manager reads it a chunk at a time.
func hundredLines() []string {
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}
	return lines
}

// completionWithin reads the manager's messages from cmd, and the messages of
// each later RunNext, until a TaskCompletedMsg arrives. It fails the test if
// none arrives within d.
func completionWithin(t *testing.T, mgr *task.Manager, cmd tea.Cmd, d time.Duration) task.TaskCompletedMsg {
	t.Helper()
	done := make(chan task.TaskCompletedMsg, 1)
	go func() {
		for cmd != nil {
			if c, ok := cmd().(task.TaskCompletedMsg); ok {
				done <- c
				return
			}
			cmd = mgr.RunNext()
		}
	}()
	select {
	case c := <-done:
		return c
	case <-time.After(d):
		t.Fatalf("no TaskCompletedMsg within %v", d)
		return task.TaskCompletedMsg{}
	}
}

// The three diagnostics checks are linesTasks, and a cancel while the manager
// reads one must complete it: the Operation overlay leaves "running" and the
// queue moves on. Each shape runs a few hundred times to catch the select race
// between the cancelled context and the buffered line.
func TestLinesTaskCancelThroughManager(t *testing.T) {
	shapes := []struct{ id, title string }{
		{"doctor", "Doctor"},
		{"missing", "Missing Dependencies"},
		{"vulns", "Vulnerability Check"},
	}
	for _, s := range shapes {
		t.Run(s.id, func(t *testing.T) {
			for i := 0; i < 300; i++ {
				mgr := task.NewManager(10)
				tk := linesTask(s.id, s.title, func(context.Context) ([]string, error) {
					return hundredLines(), nil
				})
				if res, err := mgr.Enqueue(tk); err != nil || res != task.EnqueueStarted {
					t.Fatalf("Enqueue = %v, %v; want EnqueueStarted, nil", res, err)
				}
				mgr.RunNext()()
				if _, ok := mgr.RunNext()().(task.TaskOutputMsg); !ok {
					t.Fatal("first read is not a TaskOutputMsg")
				}
				mgr.CancelCurrent()
				done := completionWithin(t, mgr, mgr.RunNext(), 2*time.Second)
				if !errors.Is(done.Err, context.Canceled) {
					t.Fatalf("completed with %v, want context.Canceled", done.Err)
				}
			}
		})
	}
}
