package gui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// enqueue hands t to the task manager. It returns the command that starts the
// task, or the toast that says why it did not start: the queue is full, or
// another task is running and t waits behind it.
func (m *Model) enqueue(t *task.Task) tea.Cmd {
	c, _ := m.tryEnqueue(t)
	return c
}

// tryEnqueue is enqueue for a caller that must react to a refused task. The
// error is the queue-full error; the command is the warning toast for it.
func (m *Model) tryEnqueue(t *task.Task) (tea.Cmd, error) {
	res, err := m.tasks.Enqueue(t)
	if err != nil {
		return m.setToast("Queue full: "+err.Error(), modal.ToastWarning), err
	}
	if res == task.EnqueueQueued {
		return m.setToast("Queued behind running task", modal.ToastInfo), nil
	}
	return m.tasks.RunNext(), nil
}

// errTask is a task whose work is one call that returns its outcome. It has no
// output lines: the error, nil or not, is the task's result.
func errTask(id, title string, run func(ctx context.Context) error) *task.Task {
	return &task.Task{
		ID:    id,
		Title: title,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			errCh := make(chan error, 1)
			errCh <- run(ctx)
			return closedCh(), errCh, nil
		},
	}
}

// streamTask is a task whose work streams its output from a brew call. A nil
// line channel, which brew returns when there is nothing to show, is an empty
// stream.
func streamTask(id, title string, run func(ctx context.Context) (<-chan string, <-chan error)) *task.Task {
	return &task.Task{
		ID:    id,
		Title: title,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ch, errCh := run(ctx)
			if ch == nil {
				ch = closedCh()
			}
			return ch, errCh, nil
		},
	}
}

// linesTask is a task whose output is computed in full, then sent one line at
// a time. The send is buffered by one and guarded by ctx: a send that finds the
// buffer full waits for the manager to read, or for the task to be cancelled,
// so a manager that stops reading cannot leak the goroutine. The error reaches
// errCh once, after the last line is sent, and the manager reads lines to the
// close of out even after the error is in.
func linesTask(id, title string, fetch func(ctx context.Context) ([]string, error)) *task.Task {
	return &task.Task{
		ID:    id,
		Title: title,
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			out := make(chan string, 1)
			errCh := make(chan error, 1)
			go func() {
				defer close(out)
				lines, err := fetch(ctx)
				if err == nil {
					err = sendLines(ctx, out, lines)
				}
				errCh <- err
			}()
			return out, errCh, nil
		},
	}
}

// sendLines sends lines to out in order, and stops early when ctx is cancelled.
// It returns ctx.Err(), so a cancelled task always ends with an error, and nil
// only when every line went out and the task was not cancelled.
func sendLines(ctx context.Context, out chan<- string, lines []string) error {
	for _, line := range lines {
		select {
		case out <- line:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return ctx.Err()
}
