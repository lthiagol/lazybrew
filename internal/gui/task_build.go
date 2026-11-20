package gui

import (
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
