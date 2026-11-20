package gui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
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

// statusRow is the row under the operation's title, plain.
func statusRow(t *testing.T, m *Model, width int) string {
	t.Helper()
	rows := strings.Split(ansi.Strip(m.renderOperation(width, 12)), "\n")
	if len(rows) < 2 {
		t.Fatalf("the operation view has no status row:\n%s", strings.Join(rows, "\n"))
	}
	return strings.TrimRight(rows[1], " ")
}

// TestFinishedRunsEndWithTheCloseHint (v2-07 AC-02.1): the error and cancelled
// rows end with the close hint, and so does the success row during its 2 s. A
// running row keeps its cancel hint.
func TestFinishedRunsEndWithTheCloseHint(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	cases := []struct {
		name string
		op   Operation
		want string
	}{
		{"success", Operation{Title: "stop moshi-hook", Status: opSuccess, Started: start, Ended: start},
			"✓ Completed in 0s · Esc to close"},
		{"failed", Operation{Title: "stop moshi-hook", Status: opError, Err: errors.New("boom"), Started: start, Ended: start},
			"✗ Failed: boom · Esc to close"},
		{"cancelled", Operation{Title: "stop moshi-hook", Status: opCancelled, Started: start, Ended: start},
			"Cancelled · Esc to close"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := newTestModel()
			m.icons = style.UnicodeIcons
			op := c.op
			m.opState = &op
			if got := statusRow(t, m, 80); got != c.want {
				t.Fatalf("status row = %q, want %q", got, c.want)
			}
		})
	}

	m := newTestModel()
	m.opState = &Operation{Title: "stop moshi-hook", Status: opRunning, Started: time.Now()}
	if got := statusRow(t, m, 80); !strings.HasSuffix(got, "Esc to cancel") || strings.Contains(got, "close") {
		t.Fatalf("a running row should keep its cancel hint, got %q", got)
	}
}

// A long error is cut so the close hint is still on the row.
func TestLongErrorKeepsTheCloseHint(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
	m.opState = &Operation{Title: "stop moshi-hook", Status: opError,
		Err: errors.New(strings.Repeat("brew services failed ", 10)), Started: time.Now(), Ended: time.Now()}
	got := statusRow(t, m, 60)
	if !strings.HasSuffix(got, "· Esc to close") {
		t.Fatalf("the close hint is cut off: %q", got)
	}
	if w := ansi.StringWidth(got); w > 60 {
		t.Fatalf("the status row is %d cells, want at most 60: %q", w, got)
	}
}

// TestFinishedOperationFooterStartsWithEscClose (v2-07 AC-02.2): while a finished
// operation is shown, the footer's first hint is esc close; a running one keeps
// the panel's own keys first.
func TestFinishedOperationFooterStartsWithEscClose(t *testing.T) {
	m := sizedServicesModel(t)
	m.opState = &Operation{Title: "stop moshi-hook", Status: opError, Err: errors.New("boom"), Started: time.Now(), Ended: time.Now()}
	if hints := m.footerHints(); len(hints) == 0 || hints[0] != (keyHint{"esc", "close"}) {
		t.Fatalf("first footer hint = %+v, want esc close", hints)
	}
	if !strings.Contains(screenText(m), "esc close") {
		t.Fatalf("the footer should show esc close, got:\n%s", screenText(m))
	}

	m.opState = &Operation{Title: "stop moshi-hook", Status: opRunning, Started: time.Now()}
	if hints := m.footerHints(); len(hints) == 0 || hints[0] == (keyHint{"esc", "close"}) {
		t.Fatalf("a running operation should not offer esc close first, got %+v", hints)
	}
	if strings.Contains(screenText(m), "esc close") {
		t.Fatalf("a running operation shows no close hint in the footer")
	}
}
