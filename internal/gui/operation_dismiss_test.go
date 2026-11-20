package gui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// screenText is the plain text of the frame the model draws now.
func screenText(m *Model) string {
	return ansi.Strip(m.View().Content)
}

// sizedServicesModel is the Services action model at the harness size, so View
// draws a frame.
func sizedServicesModel(t *testing.T) *Model {
	t.Helper()
	m, _ := servicesActionModel(t, []brew.Service{runningService("moshi-hook")})
	return updateModel(m, tea.WindowSizeMsg{Width: 140, Height: 40})
}

// startOp and finishOp are the two messages a task sends to the Model: the run
// begins, then it ends with err.
func startOp(m *Model, id, title string) *Model {
	return updateModel(m, TaskStartedMsg{ID: id, Title: title})
}

func finishOp(m *Model, id, title string, err error) *Model {
	return updateModel(m, TaskCompletedMsg{ID: id, Title: title, Err: err})
}

// TestSuccessfulOperationClosesOnItsTimer (v2-07 AC-01.1): a successful run shows
// its final state, and the timer's message, for that run, puts the panel back.
func TestSuccessfulOperationClosesOnItsTimer(t *testing.T) {
	m := sizedServicesModel(t)
	m = startOp(m, "service-stop:moshi-hook", "stop moshi-hook")
	m = finishOp(m, "service-stop:moshi-hook", "stop moshi-hook", nil)
	if !strings.Contains(screenText(m), "Completed in") {
		t.Fatalf("the finished run should show, got:\n%s", screenText(m))
	}
	if m.autoCloseCmd() == nil {
		t.Fatal("a successful run starts its timer")
	}

	m = updateModel(m, opAutoCloseMsg{op: m.opState})
	if m.opState != nil {
		t.Fatal("the timer should clear the finished operation")
	}
	if strings.Contains(screenText(m), "Completed in") {
		t.Fatalf("the operation should be gone, got:\n%s", screenText(m))
	}
	if !strings.Contains(screenText(m), "Services › Info") {
		t.Fatalf("the Services panel should be back in the main area, got:\n%s", screenText(m))
	}
}

// TestOperationTimerIgnoresANewerOperation (v2-07 AC-01.1): a timer names its
// operation. After a second one started, the first one's timer closes nothing,
// and neither does a timer for a running operation.
func TestOperationTimerIgnoresANewerOperation(t *testing.T) {
	m := sizedServicesModel(t)
	m = startOp(m, "service-stop:moshi-hook", "stop moshi-hook")
	m = finishOp(m, "service-stop:moshi-hook", "stop moshi-hook", nil)
	first := m.opState

	m = startOp(m, "service-start:moshi-hook", "start moshi-hook")
	second := m.opState
	if second == first || !second.Running() {
		t.Fatal("the second run should be a new, running operation")
	}
	m = updateModel(m, opAutoCloseMsg{op: first})
	if m.opState != second || !m.opState.Running() {
		t.Fatalf("a stale timer closed the running operation")
	}

	m = finishOp(m, "service-start:moshi-hook", "start moshi-hook", nil)
	m = updateModel(m, opAutoCloseMsg{op: first})
	if m.opState != second {
		t.Fatal("a stale timer closed a newer finished operation")
	}
	m = updateModel(m, opAutoCloseMsg{op: second})
	if m.opState != nil {
		t.Fatal("the newer operation's own timer should close it")
	}
}

// TestBatchDoesNotCloseMidBatch (v2-07 AC-01.2): while a batch still has tasks
// queued or running, its finished task starts no timer and a timer for it closes
// nothing. The last task's success starts the timer.
func TestBatchDoesNotCloseMidBatch(t *testing.T) {
	m := sizedServicesModel(t)
	m.batchIDs = map[string]int{"service-stop:moshi-hook": 1, "service-stop:dbus": 1}
	m.batchCount = 2
	m.batchTotal = 2

	m = startOp(m, "service-stop:moshi-hook", "stop moshi-hook")
	m = finishOp(m, "service-stop:moshi-hook", "stop moshi-hook", nil)
	if m.batchCount != 1 {
		t.Fatalf("batchCount = %d after one of two tasks, want 1", m.batchCount)
	}
	if m.autoCloseCmd() != nil {
		t.Fatal("a batch with a task still to run starts no timer")
	}
	mid := m.opState
	m = updateModel(m, opAutoCloseMsg{op: mid})
	if m.opState != mid || !strings.Contains(screenText(m), "Completed in") {
		t.Fatalf("the batch closed mid-batch, got:\n%s", screenText(m))
	}

	m = startOp(m, "service-stop:dbus", "stop dbus")
	m = finishOp(m, "service-stop:dbus", "stop dbus", nil)
	if m.batchCount != 0 {
		t.Fatalf("batchCount = %d after the batch, want 0", m.batchCount)
	}
	if m.autoCloseCmd() == nil {
		t.Fatal("the last task's success starts the timer")
	}
	m = updateModel(m, opAutoCloseMsg{op: m.opState})
	if m.opState != nil {
		t.Fatal("the batch's last operation should close after its timer")
	}
}

// A failed or cancelled run starts no timer: it stays until Esc.
func TestFailedAndCancelledOperationsStartNoTimer(t *testing.T) {
	m := sizedServicesModel(t)
	m = startOp(m, "service-stop:moshi-hook", "stop moshi-hook")
	m = finishOp(m, "service-stop:moshi-hook", "stop moshi-hook", errors.New("boom"))
	if m.autoCloseCmd() != nil {
		t.Fatal("a failed run starts a timer")
	}
	m = startOp(m, "service-stop:moshi-hook", "stop moshi-hook")
	m = finishOp(m, "service-stop:moshi-hook", "stop moshi-hook", context.Canceled)
	if m.autoCloseCmd() != nil {
		t.Fatal("a cancelled run starts a timer")
	}
}
