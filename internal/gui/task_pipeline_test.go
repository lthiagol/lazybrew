package gui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// FC1: an errTask whose brew call blocks never holds Update. Enqueue returns
// at once, the start returns the reader, and Esc cancels the call while it
// still blocks.
func TestBlockingErrTaskDoesNotBlockUpdate(t *testing.T) {
	m := newTestModel()
	release := make(chan struct{})
	defer close(release)
	tk := errTask("tap-slow", "Tap slow", func(ctx context.Context) error {
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	got := make(chan tea.Cmd, 1)
	go func() { got <- m.enqueue(tk) }()
	var start tea.Cmd
	select {
	case start = <-got:
	case <-time.After(5 * time.Second):
		t.Fatal("enqueue blocked on the task's brew call")
	}
	started, ok := start().(TaskStartedMsg)
	if !ok {
		t.Fatalf("want TaskStartedMsg, got %#v", started)
	}
	nm, reader := m.Update(started)
	m2 := nm.(Model)
	if m2.opState == nil || !m2.opState.Running() {
		t.Fatal("the Operation shows running while the call blocks")
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- reader() }()
	updateModel(&m2, tea.KeyPressMsg{Code: tea.KeyEscape})
	select {
	case msg := <-done:
		c, ok := msg.(TaskCompletedMsg)
		if !ok || !errors.Is(c.Err, context.Canceled) {
			t.Fatalf("want a cancelled completion, got %#v", msg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Esc did not cancel the blocking call")
	}
}

// FC2: a task's completion still on its way to Update keeps the manager busy:
// a write confirmed in that window is queued, not started beside it.
func TestWriteConfirmedBeforeCompletionIsHandledIsQueued(t *testing.T) {
	m := newTestModel()
	a := errTask("pin:a", "Toggle pin a", func(context.Context) error { return nil })
	started := m.enqueue(a)()
	nm, reader := m.Update(started)
	m2 := nm.(Model)
	if _, ok := reader().(TaskCompletedMsg); !ok {
		t.Fatal("want A's completion")
	}
	b := errTask("pin:b", "Toggle pin b", func(context.Context) error { return nil })
	m2.enqueue(b)
	if got := lastToastText(&m2); got != "Queued behind running task" {
		t.Fatalf("toast = %q, want the queued toast", got)
	}
	if m2.tasks.Idle() {
		t.Fatal("the manager is idle before A's completion is handled")
	}
}

// FB4: a failed or cancelled package write refreshes the lists quietly and
// re-marks Search; a write that changes no package does not.
func TestFailedPackageWriteRefreshesQuietly(t *testing.T) {
	cases := []struct {
		id   string
		err  error
		want bool
	}{
		{"upgrade-all", errors.New("exit 1"), true},
		{"install:foo", context.Canceled, true},
		{"batch-upgrade:1:foo", errors.New("exit 1"), true},
		{"autoremove", context.Canceled, true},
		{"fetch:foo", errors.New("exit 1"), false},
		{"pin:foo", errors.New("exit 1"), false},
	}
	for _, tc := range cases {
		m := newTestModel()
		_, cmd := m.Update(TaskCompletedMsg{ID: tc.id, Title: "T", Err: tc.err})
		var quiet bool
		for _, r := range collectRefreshMsgs(cmd) {
			quiet = quiet || r.Quiet
		}
		if quiet != tc.want {
			t.Errorf("%s: quiet refresh = %v, want %v", tc.id, quiet, tc.want)
		}
	}
}

// FB4: the quiet refresh reloads a loaded Outdated panel and ends with no
// "Data refreshed" toast over the error.
func TestQuietRefreshReloadsWithoutToast(t *testing.T) {
	m := newTestModel()
	m.panels[PanelOutdated].items = []string{"foo  1.0"}
	m = updateModel(m, TaskCompletedMsg{ID: "upgrade-all", Title: "Upgrading all packages", Err: errors.New("exit 1")})
	errToast := lastToastText(m)
	m = updateModel(m, RefreshMsg{Quiet: true})
	if !m.panels[PanelOutdated].loading {
		t.Fatal("the loaded Outdated panel is not reloaded")
	}
	for m.refreshing > 0 {
		m.countRefreshLanding(true)
	}
	if got := lastToastText(m); got != errToast {
		t.Fatalf("toast = %q, want the error toast %q", got, errToast)
	}

	// A plain refresh still says it is done.
	m = updateModel(m, RefreshMsg{})
	for m.refreshing > 0 {
		m.countRefreshLanding(true)
	}
	if got := lastToastText(m); got != "Data refreshed" {
		t.Fatalf("toast = %q, want Data refreshed", got)
	}
}

// FB6: a failed or cancelled update is not fresh; a cancel is not an error.
func TestUnfinishedUpdateIsNotFresh(t *testing.T) {
	for _, err := range []error{errors.New("exit 1"), context.Canceled} {
		m := newTestModel()
		m.isUpdating = true
		m = updateModel(m, TaskCompletedMsg{ID: "brew-update", Title: "brew update", Err: err})
		if !m.lastUpdate.IsZero() {
			t.Errorf("%v: lastUpdate set by an unfinished update", err)
		}
		if !m.updateSettled || !m.updateRefreshed {
			t.Errorf("%v: the update did not settle", err)
		}
		if errors.Is(err, context.Canceled) && lastToastText(m) != "Update cancelled" {
			t.Errorf("cancel toast = %q, want Update cancelled", lastToastText(m))
		}
	}
	m := newTestModel()
	m = updateModel(m, TaskCompletedMsg{ID: "brew-update", Title: "brew update"})
	if m.lastUpdate.IsZero() {
		t.Error("a finished update is fresh")
	}
}

// FB6: a queued update does not collect the running task's output.
func TestQueuedUpdateIgnoresOtherOutput(t *testing.T) {
	m := newTestModel()
	m.isUpdating = true
	m = updateModel(m, TaskOutputMsg{ID: "upgrade-all", Lines: []string{"==> Upgrading foo"}})
	m = updateModel(m, TaskOutputMsg{ID: "brew-update", Lines: []string{"Updated 1 tap"}})
	if len(m.updateOutput) != 1 || m.updateOutput[0] != "Updated 1 tap" {
		t.Fatalf("updateOutput = %v, want only the update's line", m.updateOutput)
	}
}
