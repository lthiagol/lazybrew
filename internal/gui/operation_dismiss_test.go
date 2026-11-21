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
			"✔ Completed in 0s · Esc to close · y copy"},
		{"failed", Operation{Title: "stop moshi-hook", Status: opError, Err: errors.New("boom"), Started: start, Ended: start},
			"✗ Failed: boom · Esc to close · y copy"},
		{"cancelled", Operation{Title: "stop moshi-hook", Status: opCancelled, Started: start, Ended: start},
			"Cancelled · Esc to close · y copy"},
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
	if got := statusRow(t, m, 80); !strings.HasSuffix(got, "Esc to cancel · y copy") || strings.Contains(got, "close") {
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
	if hints := m.footerHints().all(); len(hints) == 0 || hints[0] != keyClose {
		t.Fatalf("first footer hint = %+v, want esc close", hints)
	}
	if !strings.Contains(screenText(m), "esc close") {
		t.Fatalf("the footer should show esc close, got:\n%s", screenText(m))
	}

	m.opState = &Operation{Title: "stop moshi-hook", Status: opRunning, Started: time.Now()}
	if hints := m.footerHints().all(); len(hints) == 0 || hints[0] == keyClose {
		t.Fatalf("a running operation should not offer esc close first, got %+v", hints)
	}
	if strings.Contains(screenText(m), "esc close") {
		t.Fatalf("a running operation shows no close hint in the footer")
	}
}

// finishedServicesRun is the Services panel with a successful stop shown: the
// run is done and its operation is on screen.
func finishedServicesRun(t *testing.T) *Model {
	t.Helper()
	m := sizedServicesModel(t)
	m = startOp(m, "service-stop:moshi-hook", "stop moshi-hook")
	return finishOp(m, "service-stop:moshi-hook", "stop moshi-hook", nil)
}

// TestMovingAwayClosesAFinishedOperation (v2-07 AC-03.1): Tab, Shift+Tab, a
// panel number and a click on another sidebar box each clear a finished
// operation, and the panel moved to shows in the main area.
func TestMovingAwayClosesAFinishedOperation(t *testing.T) {
	moves := map[string]func(m *Model) *Model{
		"tab": func(m *Model) *Model {
			return updateModel(m, tea.KeyPressMsg{Code: tea.KeyTab})
		},
		"shift+tab": func(m *Model) *Model {
			return updateModel(m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
		},
		"1": func(m *Model) *Model { return sendKey(m, "1") },
		"click": func(m *Model) *Model {
			m.cfg.GUI.Mouse = true
			y := m.computeContentHeights()[0] + 2 + 1
			return updateModel(m, tea.MouseClickMsg{X: 2, Y: y, Button: tea.MouseLeft})
		},
	}
	for name, move := range moves {
		t.Run(name, func(t *testing.T) {
			m := finishedServicesRun(t)
			if !strings.Contains(screenText(m), "Completed in") {
				t.Fatalf("the finished run should show before the move")
			}
			before := m.activePanel
			m = move(m)
			if m.activePanel == before {
				t.Fatalf("the move did not change panel")
			}
			if m.opState != nil || strings.Contains(screenText(m), "Completed in") {
				t.Fatalf("moving away should clear the finished run, got:\n%s", screenText(m))
			}
		})
	}
}

// TestMovingAwayLeavesARunningOperation (v2-07 AC-03.1): a running operation is
// not cleared by moving; it keeps its view until it ends.
func TestMovingAwayLeavesARunningOperation(t *testing.T) {
	m := sizedServicesModel(t)
	m = startOp(m, "service-stop:moshi-hook", "stop moshi-hook")
	m = sendKey(m, "1")
	if m.opState == nil || !m.opState.Running() {
		t.Fatal("moving away cleared a running operation")
	}
	if !strings.Contains(screenText(m), "Running") {
		t.Fatalf("the running view should stay on screen, got:\n%s", screenText(m))
	}
}

// A key on the panel already shown is not a move, so the finished run stays.
func TestKeyOnTheCurrentPanelKeepsTheFinishedOperation(t *testing.T) {
	m := finishedServicesRun(t)
	m = sendKey(m, "6")
	if m.opState == nil {
		t.Fatal("pressing the current panel's number cleared the finished run")
	}
}

// infoStatusRow is the Status row of the Services Info in the main area. The
// sidebar's own Status box is above it, so the search starts at the Info's title.
func infoStatusRow(m *Model) string {
	text := screenText(m)
	i := strings.Index(text, "Services › Info")
	if i < 0 {
		return ""
	}
	for _, line := range strings.Split(text[i:], "\n") {
		if strings.Contains(line, "Status") {
			return line
		}
	}
	return ""
}

// TestServicesInfoDrawsTheReloadOnceTheRunCloses (v2-07 AC-03.2): the stop's
// completion reloads the Services panel, and when the run closes the Info shows
// the reloaded state.
func TestServicesInfoDrawsTheReloadOnceTheRunCloses(t *testing.T) {
	m := finishedServicesRun(t)
	stopped, _ := serviceStatus(brew.ServiceNone, style.UnicodeIcons)

	m = updateModel(m, RefreshMsg{})
	m = updateModel(m, dataLoaded([]brew.Service{stoppedService("moshi-hook")}))
	m = updateModel(m, opAutoCloseMsg{op: m.opState})
	if m.opState != nil {
		t.Fatal("the run should have closed")
	}
	if !strings.Contains(screenText(m), "Services › Info") {
		t.Fatalf("the Services Info should be in the main area, got:\n%s", screenText(m))
	}
	if row := infoStatusRow(m); !strings.Contains(row, stopped) {
		t.Fatalf("the Info's Status row = %q, want the reloaded %q", row, stopped)
	}
	if strings.Contains(screenText(m), "Loading...") {
		t.Fatalf("the reload has landed, the Info should not still be loading:\n%s", screenText(m))
	}
}

// TestCompletedRowMatchesTheStepTicks (lb25 fix r1): a finished run ticks its
// steps and its Completed row with the same mark, the ✔ brew prints, at the
// harness sizes.
func TestCompletedRowMatchesTheStepTicks(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	sizes := []struct {
		name string
		msg  tea.WindowSizeMsg
	}{
		{"80x24", tea.WindowSizeMsg{Width: 80, Height: 24}},
		{"140x40", tea.WindowSizeMsg{Width: 140, Height: 40}},
	}
	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			m := updateModel(newTestModel(), size.msg)
			m.opState = operationFrom(Operation{Title: "upgrade bash", Status: opSuccess, Started: start, Ended: start.Add(34 * time.Second)}, upgradeTranscript)
			out := screenText(m)
			if !strings.Contains(out, "✔ Completed in 34s") {
				t.Fatalf("the Completed row should carry ✔, got:\n%s", out)
			}
			if !strings.Contains(out, "✔ Fetching bash") {
				t.Fatalf("a finished step should carry ✔, got:\n%s", out)
			}
		})
	}
}
