package gui

// Tests for the r3 review round: each test names the finding it locks in
// a comment.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/task"
)

// --- R3-9 / R3-3: tab / shift+tab must sync visibleRows --------------------

// TestPanelCycleSyncsVisibleRows locks r3 finding R3-3: nextPanel and
// prevPanel are panel switches, so they must re-derive visibleRows like
// switchPanel does. At height 30 the sidebar layout hits the cramped
// branch (the active panel gets 6 content rows, inactive ones 1), so a
// stale visibleRows after a tab cycle is directly observable.
func TestPanelCycleSyncsVisibleRows(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 30})

	check := func(stage string) {
		t.Helper()
		heights := m.computeContentHeights()
		for i := range m.panels {
			want := max(0, heights[i]-1)
			if m.panels[i].visibleRows != want {
				t.Errorf("%s: panel %d: visibleRows = %d, want %d", stage, i, m.panels[i].visibleRows, want)
			}
		}
		if got := m.panels[m.activePanel].visibleRows; got < 5 {
			t.Fatalf("%s: expected the cramped branch to give the active panel >=5 rows, got %d", stage, got)
		}
	}

	m = sendSpecial(m, tea.KeyTab)
	check("after Tab")

	m = sendSpecial(m, tea.KeyShiftTab)
	check("after Shift+Tab")
}

// --- R3-9 / R3-1: queue full at the FIRST batch enqueue --------------------

// TestBatchUpgradeQueueFullAtFirstEnqueueSurfacesToast locks r3 finding
// R3-1: when the very first Enqueue fails with ErrQueueFull, batchUpgrade
// used to return silently (no toast, no counter, no feedback).
func TestBatchUpgradeQueueFullAtFirstEnqueueSurfacesToast(t *testing.T) {
	m := newTestModel()

	// Make the manager busy: a blocker task that never completes.
	blocker := &task.Task{
		ID: "blocker",
		Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
			return make(chan string), make(chan error), nil
		},
	}
	if _, err := m.tasks.Enqueue(blocker); err != nil {
		t.Fatal(err)
	}
	m.tasks.RunNext()

	// Fill the queue to capacity while the blocker runs (the running
	// blocker itself counts against the cap, so DefaultMaxQueue-1
	// fillers fill the queue completely).
	for i := 0; i < task.DefaultMaxQueue-1; i++ {
		i := i
		filler := &task.Task{
			ID: fmt.Sprintf("filler-%d", i),
			Run: func(ctx context.Context) (<-chan string, <-chan error, error) {
				return closedCh(), nil, nil
			},
		}
		if _, err := m.tasks.Enqueue(filler); err != nil {
			t.Fatalf("filler %d: unexpected enqueue error: %v", i, err)
		}
	}

	// One selected outdated package: the very first batch enqueue must
	// now hit the full queue.
	panel := m.panels[PanelOutdated]
	panel.items = []string{"pkg  1.0"}
	panel.formulae = append(panel.formulae, brew.Formula{Name: "pkg"})
	m.batch.selected[0] = true

	nm, cmd := m.batchUpgrade()
	rm := nm.(Model)
	if rm.toast == nil || !strings.Contains(rm.toast.Message, "Queue full") {
		t.Fatalf("expected a queue-full warning toast when no batch task could be enqueued, got %+v", rm.toast)
	}
	if rm.batchCount != 0 || len(rm.batchIDs) != 0 || rm.batchTotal != 0 {
		t.Fatalf("a fully-refused batch must leave no counter state, got count=%d ids=%v total=%d",
			rm.batchCount, rm.batchIDs, rm.batchTotal)
	}
	if cmd == nil {
		t.Fatal("expected the toast command to be returned")
	}
	for _, msg := range execCmds(cmd) {
		if _, ok := msg.(TaskStartedMsg); ok {
			t.Fatal("no batch task may start when the first enqueue was refused")
		}
	}
}

// --- R3-9 / R3-2: same-named formula and cask in one batch -----------------

// TestBatchUpgradeDuplicateNamesCountOccurrences locks r3 finding R3-2:
// a formula and a cask sharing a name produce two batch tasks with the
// same ID; the bool batchIDs map collapsed them into one entry and the
// second completion could not decrement the counter.
func TestBatchUpgradeDuplicateNamesCountOccurrences(t *testing.T) {
	m := newTestModel()
	panel := m.panels[PanelOutdated]
	panel.items = []string{"dup  1.0", "dup  2.0"}
	panel.formulae = append(panel.formulae, brew.Formula{Name: "dup"})
	panel.casks = append(panel.casks, brew.Cask{Name: "dup"})
	m.batch.selected[0] = true
	m.batch.selected[1] = true

	nm, _ := m.batchUpgrade()
	rm := nm.(Model)
	if rm.batchCount != 2 {
		t.Fatalf("expected batchCount 2 for two selected rows, got %d", rm.batchCount)
	}
	// R4-4: the ID carries the batch's sequence number.
	if rm.batchIDs["batch-upgrade:1:dup"] != 2 {
		t.Fatalf("expected the shared ID to be counted twice, got %v", rm.batchIDs)
	}
	if rm.batchTotal != 2 {
		t.Fatalf("expected batchTotal 2, got %d", rm.batchTotal)
	}

	// First duplicate completes: the ID stays with count 1 and the
	// counter still decrements.
	m = updateModel(&rm, TaskCompletedMsg{ID: "batch-upgrade:1:dup", Title: "Upgrade dup"})
	if m.batchCount != 1 {
		t.Fatalf("first duplicate completion must decrement batchCount, got %d", m.batchCount)
	}
	if m.batchIDs["batch-upgrade:1:dup"] != 1 {
		t.Fatalf("expected the shared ID to remain with count 1, got %v", m.batchIDs)
	}

	// Second duplicate completes: the counter must drain to 0.
	m = updateModel(m, TaskCompletedMsg{ID: "batch-upgrade:1:dup", Title: "Upgrade dup"})
	if m.batchCount != 0 {
		t.Fatalf("second duplicate completion must decrement batchCount to 0, got %d", m.batchCount)
	}
	if len(m.batchIDs) != 0 {
		t.Fatalf("expected batchIDs drained, got %v", m.batchIDs)
	}
	if m.batchTotal != 0 {
		t.Fatalf("expected batchTotal reset when the batch drains, got %d", m.batchTotal)
	}
}

// --- R3-5: batchCount is surfaced in the Operation overlay -----------------

// TestBatchOverlayTitleShowsProgress locks r3 finding R3-5: batch tasks
// surface the batch counter in the overlay title; non-batch tasks keep
// their plain title.
func TestBatchOverlayTitleShowsProgress(t *testing.T) {
	m := newTestModel()
	m.batchCount = 3
	m.batchTotal = 10
	m.batchIDs = map[string]int{"batch-upgrade:foo": 1, "batch-upgrade:bar": 1, "batch-upgrade:dup": 1}

	m = updateModel(m, TaskStartedMsg{ID: "batch-upgrade:foo", Title: "Upgrade foo"})
	if m.opState == nil {
		t.Fatal("expected the operation overlay to open")
	}
	want := "Upgrade foo (7/10 done)"
	if m.opState.Title != want {
		t.Fatalf("overlay title = %q, want %q", m.opState.Title, want)
	}

	m = updateModel(m, TaskStartedMsg{ID: "doctor", Title: "Doctor"})
	if m.opState.Title != "Doctor" {
		t.Fatalf("non-batch task title must stay plain, got %q", m.opState.Title)
	}
	if m.batchCount != 3 {
		t.Fatalf("task starts must not touch batchCount, got %d", m.batchCount)
	}
}

// --- R3-4: the C toggle re-clamps the main list cursor ---------------------

// TestCommandLogToggleKeepsCursorVisible locks r3 finding R3-4: expanding
// the command log shrinks the main list without a resize, which used to
// leave offset+visible excluding the selected row until the next
// normalize-triggering event. At height 24 the sidebar is in the cramped
// branch (visibleRows falls back), so the main list height is the binding
// window: 16 rows collapsed, 13 expanded.
func TestCommandLogToggleKeepsCursorVisible(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m.logCollapsed = true
	m.activePanel = PanelFormulae

	p := m.panels[PanelFormulae]
	for i := 0; i < 30; i++ {
		p.items = append(p.items, fmt.Sprintf("pkg%02d  1.0", i))
	}
	// Scrolled to the bottom under the collapsed (16-row) window.
	p.selected = 29
	p.offset = 14

	m = sendKey(m, "C") // expand the log: the main window shrinks to 13

	if m.logCollapsed {
		t.Fatal("C must expand the command log")
	}
	_, mainH := m.logLayoutHeights(m.height - 4)
	if mainH != 13 {
		t.Fatalf("expected main content height 13 with the log expanded, got %d", mainH)
	}
	vc := min(p.visibleCount(), mainH)
	if p.selected < p.offset || p.selected >= p.offset+vc {
		t.Fatalf("cursor %d must be inside [offset %d, offset+%d) after the C toggle",
			p.selected, p.offset, vc)
	}
	if p.offset != 17 {
		t.Fatalf("expected offset clamped to 17, got %d", p.offset)
	}
}

// --- R3-8: untrust tasks are titled "Untrust: <name>" ----------------------

// TestUntrustTaskTitle locks r3 finding R3-8: the untrust arm of the
// trust menu must not title its task "Trust: <name>".
func TestUntrustTaskTitle(t *testing.T) {
	m := newTestModel()
	m.panels[PanelTaps].items = []string{"homebrew/foo"}
	m.panels[PanelTaps].selected = 0

	cmd := m.executeTrustAction(1)
	ts, ok := firstTaskStarted(execCmds(cmd))
	if !ok {
		t.Fatal("expected the untrust task to start")
	}
	if ts.ID != "untrust:homebrew/foo" {
		t.Fatalf("expected task ID untrust:homebrew/foo, got %q", ts.ID)
	}
	if ts.Title != "Untrust: homebrew/foo" {
		t.Fatalf("expected title %q, got %q", "Untrust: homebrew/foo", ts.Title)
	}

	// The trust arm keeps its title.
	m = newTestModel()
	m.panels[PanelTaps].items = []string{"homebrew/foo"}
	m.panels[PanelTaps].selected = 0
	cmd = m.executeTrustAction(0)
	ts, ok = firstTaskStarted(execCmds(cmd))
	if !ok {
		t.Fatal("expected the trust task to start")
	}
	if ts.ID != "trust:homebrew/foo" || ts.Title != "Trust: homebrew/foo" {
		t.Fatalf("expected trust:homebrew/foo / \"Trust: homebrew/foo\", got %q / %q", ts.ID, ts.Title)
	}
}
