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
			// The manager runs Run off Update; the goroutine keeps Run quick
			// anyway, whoever calls it.
			errCh := make(chan error, 1)
			go func() { errCh <- run(ctx) }()
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
