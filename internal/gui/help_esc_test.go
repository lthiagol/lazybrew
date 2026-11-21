package gui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// TestHelpEscClosesHelpBeforeCancellingATask (v2-05 fix F-r1-03): with help open
// over a running task, the first esc closes help and leaves the task running;
// the second esc cancels it.
func TestHelpEscClosesHelpBeforeCancellingATask(t *testing.T) {
	m := newTestModel()
	ctxs := make(chan context.Context, 1)
	if _, err := m.tasks.Enqueue(&task.Task{
		ID:    "upgrade:bash",
		Title: "Upgrade bash",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			ctxs <- ctx
			return make(chan string), make(chan error), nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	m.tasks.RunNext() // starts the task; the returned read command is not needed here
	ctx := <-ctxs
	m.opState = &Operation{Title: "Upgrade bash", Status: opRunning}

	m = sendKey(m, "?")
	m = sendSpecial(m, tea.KeyEsc)
	if m.showHelp {
		t.Fatal("the first esc should close help")
	}
	if ctx.Err() != nil {
		t.Fatal("the first esc must not cancel the running task")
	}
	if m.opState == nil || m.opState.Status != opRunning {
		t.Fatal("the operation should still be running after the first esc")
	}

	sendSpecial(m, tea.KeyEsc)
	if ctx.Err() == nil {
		t.Fatal("the second esc should cancel the running task")
	}
}
