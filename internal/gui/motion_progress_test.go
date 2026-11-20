package gui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/config"
)

// batchModel is a model with a running batch of ten packages, none done yet,
// with motion on or off.
func batchModel(t *testing.T, motion string) (*Model, *Operation) {
	t.Helper()
	t.Setenv(config.AnimationsEnv, motion)
	m := newTestModel()
	op := &Operation{Title: "Upgrade", Status: opRunning, Started: time.Now(), BatchTotal: 10}
	m.opState = op
	return m, op
}

// filled counts the filled cells of the progress bar in a rendered row.
func filled(view string) int {
	return strings.Count(ansi.Strip(view), "━")
}

// pumpBar delivers the bar's frames, as the program would: each frame goes
// through Update, and the command it returns is the next frame. It runs at
// most n frames and returns the model and the command still pending.
func pumpBar(m *Model, cmd tea.Cmd, n int) (*Model, tea.Cmd) {
	for i := 0; i < n && cmd != nil; i++ {
		nm, next := m.Update(cmd())
		switch v := nm.(type) {
		case Model:
			m = &v
		case *Model:
			m = v
		}
		cmd = next
	}
	return m, cmd
}

func TestProgressBarSpringsToTheBatchShare(t *testing.T) {
	m, op := batchModel(t, "1")
	m = updateModel(m, probeMsg{})
	op.BatchDone = 5
	cmd := m.motionTargets()
	if cmd == nil {
		t.Fatal("the bar did not start moving when the count changed")
	}

	first := m.operationBar(op, 100)
	if filled(first) != 0 {
		t.Fatalf("first frame bar has %d filled cells; want the fill it had, 0", filled(first))
	}
	if !strings.HasSuffix(ansi.Strip(first), "5/10") {
		t.Fatalf("first frame count is %q; want 5/10", ansi.Strip(first))
	}

	m, cmd = pumpBar(m, cmd, 3)
	mid := m.operationBar(op, 100)
	off, _ := batchModel(t, "0")
	off.opState.BatchDone = 5
	full := filled(off.operationBar(off.opState, 100))
	if n := filled(mid); n <= 0 || n >= full {
		t.Fatalf("mid-flight bar has %d filled cells; want between 0 and the batch's %d", n, full)
	}
	if !strings.HasSuffix(ansi.Strip(mid), "5/10") {
		t.Fatalf("mid-flight count is %q; want 5/10", ansi.Strip(mid))
	}

	m, cmd = pumpBar(m, cmd, 120)
	if cmd != nil {
		t.Fatal("the bar kept animating after it reached its share")
	}
	if got, want := m.operationBar(op, 100), off.operationBar(off.opState, 100); got != want {
		t.Fatalf("settled bar differs from the bar with animations off:\n%q\n%q", got, want)
	}
}

func TestProgressBarIsStaticWithAnimationsOff(t *testing.T) {
	m, op := batchModel(t, "0")
	op.BatchDone = 5
	if cmd := m.motionTargets(); cmd != nil {
		t.Fatal("the bar scheduled frames with animations off")
	}
	if got := filled(m.operationBar(op, 100)); got == 0 {
		t.Fatal("the bar with animations off does not show the batch's share at once")
	}
}

func TestNewBatchStartsItsBarEmpty(t *testing.T) {
	m, op := batchModel(t, "1")
	op.BatchDone = 10
	m, _ = pumpBar(m, m.motionTargets(), 120)
	if got := filled(m.operationBar(op, 100)); got != operationBarWidth {
		t.Fatalf("a finished batch's bar has %d filled cells; want all %d", got, operationBarWidth)
	}

	next := &Operation{Title: "Upgrade again", Status: opRunning, Started: time.Now(), BatchTotal: 4}
	m.opState = next
	if cmd := m.motionTargets(); cmd != nil {
		t.Fatal("a new batch at 0 of 4 moved its bar")
	}
	if got := filled(m.operationBar(next, 100)); got != 0 {
		t.Fatalf("a new batch's bar has %d filled cells; want 0", got)
	}
}

func TestProgressBarCountIsExactEveryFrame(t *testing.T) {
	m, op := batchModel(t, "1")
	m = updateModel(m, probeMsg{})
	op.BatchDone = 7
	cmd := m.motionTargets()
	for i := 0; i < 30 && cmd != nil; i++ {
		m, cmd = pumpBar(m, cmd, 1)
		if !strings.HasSuffix(ansi.Strip(m.operationBar(op, 100)), "7/10") {
			t.Fatalf("frame %d count is not 7/10: %q", i, ansi.Strip(m.operationBar(op, 100)))
		}
	}
}
