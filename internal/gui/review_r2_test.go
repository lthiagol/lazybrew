package gui

// Tests for the r2 review round: each test names the finding it locks in
// a comment.

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// startBatch enqueues the selected outdated packages and returns the
// resulting model; the caller drives the returned command(s) itself.
func seedOutdatedBatch(m *Model, n int) {
	panel := m.panels[PanelOutdated]
	for i := 0; i < n; i++ {
		name := "pkg" + string(rune('a'+i))
		panel.items = append(panel.items, name+"  1.0")
		panel.formulae = append(panel.formulae, brew.Formula{Name: name})
		m.batch.selected[i] = true
	}
}

// execCmds executes a returned command (unwrapping tea.Batch) and reports
// the first TaskStartedMsg it can produce, plus the raw messages seen.
func execCmds(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	switch b := msg.(type) {
	case tea.BatchMsg:
		var msgs []tea.Msg
		for _, c := range b {
			msgs = append(msgs, execCmds(c)...)
		}
		return msgs
	default:
		return []tea.Msg{msg}
	}
}

// step delivers msg to the model's Update loop and returns the model plus
// the message produced by the command Update handed back — mimicking the
// production Bubble Tea loop for exactly one hop.
func step(m *Model, msg tea.Msg) (*Model, tea.Msg) {
	nm, cmd := m.Update(msg)
	mm := nm.(Model)
	m = &mm
	var next tea.Msg
	if cmd != nil {
		next = cmd()
	}
	return m, next
}

func firstTaskStarted(msgs []tea.Msg) (TaskStartedMsg, bool) {
	for _, msg := range msgs {
		if ts, ok := msg.(TaskStartedMsg); ok {
			return ts, true
		}
	}
	return TaskStartedMsg{}, false
}

// --- NEW-1 / NEW-9: batchUpgrade must actually START the batch ------------

func TestBatchUpgradeSuccessPathStartsBatch(t *testing.T) {
	m := newTestModel()
	seedOutdatedBatch(m, 2)

	nm, cmd := m.batchUpgrade()
	rm := nm.(Model)
	if rm.batchCount != 2 {
		t.Fatalf("expected batchCount 2, got %d", rm.batchCount)
	}

	// NEW-9: the success path must start the batch, not just enqueue it —
	// the first task must begin immediately.
	ts, ok := firstTaskStarted(execCmds(cmd))
	if !ok {
		t.Fatal("expected the batch's first task to start (TaskStartedMsg) from the success path")
	}
	// R4-4: batch IDs carry a per-batch sequence number
	// ("batch-upgrade:<seq>:<name>").
	if !strings.HasPrefix(ts.ID, "batch-upgrade:1:") || !strings.Contains(ts.ID, "pkg") {
		t.Fatalf("expected a batch task to start, got ID %q", ts.ID)
	}
}

func TestBatchUpgradeQueueFullStillStartsBatch(t *testing.T) {
	m := newTestModel()
	seedOutdatedBatch(m, task.DefaultMaxQueue+2)

	nm, cmd := m.batchUpgrade()
	rm := nm.(Model)
	if rm.batchCount != task.DefaultMaxQueue {
		t.Fatalf("expected batchCount capped at %d, got %d", task.DefaultMaxQueue, rm.batchCount)
	}
	if rm.toast == nil || !strings.Contains(rm.toast.Message, "Queue full") {
		t.Fatalf("expected a queue-full warning toast, got %+v", rm.toast)
	}

	// NEW-1 (HIGH): the queue-full branch used to return only the toast,
	// parking every enqueued task with no RunNext — the whole batch
	// silently stalled. The toast and the batch start must BOTH run.
	msgs := execCmds(cmd)
	ts, ok := firstTaskStarted(msgs)
	if !ok {
		t.Fatalf("queue-full path must still start the batch (RunNext), got messages %T", msgs)
	}
	// R4-4: batch IDs carry a per-batch sequence number.
	if !strings.HasPrefix(ts.ID, "batch-upgrade:1:") || !strings.Contains(ts.ID, "pkg") {
		t.Fatalf("expected a batch task to start, got ID %q", ts.ID)
	}
}

// --- NEW-2 / NEW-9: failed batch tasks decrement the counter --------------

func TestFailedBatchTaskDecrementsCounter(t *testing.T) {
	// Fail only the FIRST upgrade call: map iteration order makes the
	// batch's start order nondeterministic, so key the failure on the
	// call sequence, not the package name.
	var calls atomic.Int32
	r := brew.NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		n := calls.Add(1)
		ch := make(chan string, 1)
		close(ch)
		errCh := make(chan error, 1)
		if n == 1 {
			errCh <- errors.New("boom")
		} else {
			close(errCh)
		}
		return ch, errCh
	}
	m := New(brew.NewClient(r), config.Default())
	seedOutdatedBatch(m, 2)

	nm, cmd := m.batchUpgrade()
	mm := nm.(Model)
	m = &mm
	ts, ok := firstTaskStarted(execCmds(cmd))
	if !ok {
		t.Fatal("expected the first batch task to start")
	}
	firstID := ts.ID
	if m.batchCount != 2 {
		t.Fatalf("expected batchCount 2, got %d", m.batchCount)
	}

	// Pump the running task through the real manager so it FAILS (the
	// mock runner's upgrade for this package returns an error).
	m, msg1 := step(m, ts)
	comp1, ok := msg1.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg from the failing task, got %T", msg1)
	}
	if comp1.Err == nil {
		t.Fatal("expected the first batch task to fail")
	}
	if comp1.ID != firstID {
		t.Fatalf("expected the failure on %q, got %q", firstID, comp1.ID)
	}

	// The error path used to return before the batchIDs decrement,
	// leaving batchCount stuck and the ID orphaned (r2 finding NEW-2).
	nm2, cmd2 := m.Update(comp1)
	mm2 := nm2.(Model)
	m = &mm2
	if m.batchCount != 1 {
		t.Fatalf("failed batch task must decrement batchCount, got %d", m.batchCount)
	}
	if m.batchIDs[firstID] != 0 {
		t.Fatal("failed batch task ID must be removed from batchIDs")
	}
	if len(m.batchIDs) != 1 {
		t.Fatalf("remaining batch task must still be tracked, got %v", m.batchIDs)
	}
	if m.toast == nil || !strings.Contains(m.toast.Message, "boom") {
		t.Fatalf("expected an error toast for the failed task, got %+v", m.toast)
	}

	// Mid-batch failure counter behavior: an unrelated completion must
	// not skew the still-running batch's counter.
	m = updateModel(m, TaskCompletedMsg{ID: "doctor", Title: "Doctor"})
	if m.batchCount != 1 {
		t.Fatalf("unrelated completion must not skew batchCount after a failure, got %d", m.batchCount)
	}

	// Drive the queue forward (the error path returns RunNext) and finish
	// the second task successfully.
	ts2, ok := firstTaskStarted(execCmds(cmd2))
	if !ok {
		t.Fatal("expected the second batch task to start after the failure")
	}
	m, msg2 := step(m, ts2)
	comp2, ok := msg2.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg for the second task, got %T", msg2)
	}
	if comp2.Err != nil {
		t.Fatalf("unexpected error on the second task: %v", comp2.Err)
	}
	m = updateModel(m, comp2)
	if m.batchCount != 0 {
		t.Fatalf("expected batchCount 0 after the batch finished, got %d", m.batchCount)
	}
	if len(m.batchIDs) != 0 {
		t.Fatalf("expected batchIDs drained, got %v", m.batchIDs)
	}
}

// --- NEW-3: task IDs are unique per operation ------------------------------

func TestBatchUpgradeIDsArePrefixed(t *testing.T) {
	m := newTestModel()
	seedOutdatedBatch(m, 1)

	nm, _ := m.batchUpgrade()
	rm := nm.(Model)
	// R4-4: the ID carries the batch's sequence number.
	if rm.batchIDs["batch-upgrade:1:pkga"] != 1 {
		t.Fatalf("expected batchIDs keyed by the prefixed task ID, got %v", rm.batchIDs)
	}

	// A completion carrying the bare package name (a manual task's old
	// ID scheme, or any other unrelated task) must NOT decrement the
	// batch counter (r2 finding NEW-3).
	m = updateModel(&rm, TaskCompletedMsg{ID: "pkga", Title: "Something pkga"})
	if m.batchCount != 1 {
		t.Fatalf("bare-name completion must not decrement batchCount, got %d", m.batchCount)
	}
	if m.batchIDs["batch-upgrade:1:pkga"] != 1 {
		t.Fatal("bare-name completion must not remove the batch ID")
	}
}

func TestMutationTaskIDsAreOpPrefixed(t *testing.T) {
	for _, tc := range []struct {
		mut   mutationType
		want  string
		label string
	}{
		{mutInstall, "install:", "install"},
		{mutUninstall, "uninstall:", "uninstall"},
		{mutReinstall, "reinstall:", "reinstall"},
		{mutUpgrade, "upgrade:", "upgrade"},
		{mutZap, "zap:", "zap"},
		{mutFetch, "fetch:", "fetch"},
	} {
		m := newTestModel()
		m.activePanel = PanelFormulae
		p := m.panels[PanelFormulae]
		p.items = []string{"foo  1.0"}
		p.selected = 0

		_, cmd := m.doMutation(tc.mut, tc.label)
		ts, ok := firstTaskStarted(execCmds(cmd))
		if !ok {
			t.Fatalf("%s: expected TaskStartedMsg", tc.label)
		}
		if !strings.HasPrefix(ts.ID, tc.want) {
			t.Fatalf("%s: task ID %q must be prefixed %q", tc.label, ts.ID, tc.want)
		}
	}

	// Non-doMutation write ops get the same treatment.
	check := func(name string, cmd tea.Cmd, prefix string) {
		t.Helper()
		ts, ok := firstTaskStarted(execCmds(cmd))
		if !ok {
			t.Fatalf("%s: expected TaskStartedMsg", name)
		}
		if !strings.HasPrefix(ts.ID, prefix) {
			t.Fatalf("%s: task ID %q must be prefixed %q", name, ts.ID, prefix)
		}
	}

	m := newTestModel()
	m.activePanel = PanelFormulae
	m.panels[PanelFormulae].items = []string{"foo  1.0"}
	m.panels[PanelFormulae].selected = 0
	_, cmd := m.togglePin()
	check("togglePin", cmd, "pin:")

	m = newTestModel()
	m.activePanel = PanelServices
	m.panels[PanelServices].items = []string{"redis  started"}
	m.panels[PanelServices].selected = 0
	_, cmd = m.serviceAction("restart")
	check("serviceAction", cmd, "service-restart:")

	m = newTestModel()
	m.activePanel = PanelTaps
	m.panels[PanelTaps].items = []string{"homebrew/foo"}
	m.panels[PanelTaps].selected = 0
	_, cmd = m.executeUntap()
	check("executeUntap", cmd, "untap:")

	m = newTestModel()
	m.panels[PanelTaps].items = []string{"homebrew/foo"}
	m.panels[PanelTaps].selected = 0
	_, cmd = m.executeRepair()
	check("executeRepair", cmd, "repair:")

	m = newTestModel()
	m.panels[PanelTaps].items = []string{"homebrew/foo"}
	m.panels[PanelTaps].selected = 0
	cmd = m.executeTrustAction(0)
	check("executeTrustAction", cmd, "trust:")
}

// --- NEW-4: "i install" hint on Formulae/Casks has a handler --------------

func TestInstallKeyIOnFormulaeAndCasks(t *testing.T) {
	for _, panel := range []PanelID{PanelFormulae, PanelCasks} {
		m := newTestModel()
		m.activePanel = panel
		m.panels[panel].items = []string{"foo  1.0"}
		m.panels[panel].selected = 0

		m = sendKey(m, "i")
		if m.activeModal == nil {
			t.Errorf("panel %v: i must open the install confirm (r2 finding NEW-4)", panel)
		}
	}
}

func TestCaskInstallFromCasksPanelUsesCaskWriter(t *testing.T) {
	var gotArgs [][]string
	r := brew.NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		gotArgs = append(gotArgs, args)
		ch := make(chan string)
		close(ch)
		errCh := make(chan error, 1)
		close(errCh)
		return ch, errCh
	}
	m := New(brew.NewClient(r), config.Default())
	m.activePanel = PanelCasks
	m.panels[PanelCasks].items = []string{"firefox  1.0"}
	m.panels[PanelCasks].selected = 0

	nm, _ := m.confirmMutation(mutInstall, "Install")
	mm := nm.(Model)
	mm.activeModal = nil
	res, cmd := mm.handleModalResult(&modal.ConfirmResult{Confirmed: true}, nil)
	nm2 := res.(Model)
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected the install to be enqueued")
	}
	if msg := nm2.tasks.RunNext()(); msg.(TaskCompletedMsg).Err != nil {
		t.Fatal("unexpected task error")
	}
	found := false
	for _, args := range gotArgs {
		if len(args) == 3 && args[0] == "install" && args[1] == "--cask" && args[2] == "firefox" {
			found = true
		}
	}
	if !found {
		// NEW-4: installing from the Casks panel must not fall back to
		// the formula writer.
		t.Fatalf("expected `brew install --cask firefox`, got %v", gotArgs)
	}
}
