package gui

import (
	"context"
	"errors"
	"testing"
	"time"

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
