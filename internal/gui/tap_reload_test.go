package gui

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// tapsPanelWith is a Taps panel holding the named taps, the cursor on index sel.
func tapsPanelWith(names []string, sel int) *Model {
	m := newTestModel()
	p := m.panels[PanelTaps]
	p.taps = make([]brew.Tap, len(names))
	p.items = make([]string, len(names))
	for i, n := range names {
		p.taps[i] = brew.Tap{Name: n, Installed: true}
		p.items[i] = n
	}
	p.setTapRows()
	p.loaded = true
	p.selected = sel
	return m
}

// landedTaps is the DataLoadedMsg of a Taps read that returned names, in order.
func landedTaps(names ...string) DataLoadedMsg {
	msg := DataLoadedMsg{PanelID: PanelTaps}
	for _, n := range names {
		msg.Taps = append(msg.Taps, brew.Tap{Name: n, Installed: true})
		msg.Items = append(msg.Items, n)
	}
	return msg
}

func isTapsReload(msg tea.Msg) bool {
	_, ok := msg.(tapsReloadMsg)
	return ok
}

func isRefresh(msg tea.Msg) bool {
	_, ok := msg.(RefreshMsg)
	return ok
}

// A taps write that fails or is cancelled reloads the Taps list, whichever the
// write was: tap, untap, repair, trust and untrust.
func TestFailedTapWriteReloadsTheList(t *testing.T) {
	ids := []string{"tap-user/tap", "untap:user/tap", "repair:user/tap", "trust:user/tool", "untrust:user/tool"}
	for _, id := range ids {
		m := newTestModel()
		_, cmd := m.Update(TaskCompletedMsg{ID: id, Title: "Write", Err: errors.New("boom")})
		if !batchHas(cmd, isTapsReload) {
			t.Errorf("a failed %s should reload the Taps list", id)
		}
		_, cmd = m.Update(TaskCompletedMsg{ID: id, Title: "Write", Err: context.Canceled})
		if !batchHas(cmd, isTapsReload) {
			t.Errorf("a cancelled %s should reload the Taps list", id)
		}
	}
}

// A successful taps write reloads through RefreshMsg, and not through the
// failure reload, so the success toast is the last word.
func TestSuccessfulTapWriteRefreshesThroughRefresh(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(TaskCompletedMsg{ID: "trust:user/tool", Title: "Trust: user/tool"})
	if !batchHas(cmd, isRefresh) {
		t.Error("a successful trust should refresh, reloading the Taps list")
	}
	if batchHas(cmd, isTapsReload) {
		t.Error("a successful trust should not also run the failure reload")
	}
}

// A non-tap failure does not reload the Taps list.
func TestNonTapFailureDoesNotReloadTaps(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(TaskCompletedMsg{ID: "upgrade:bash", Title: "Upgrade bash", Err: errors.New("boom")})
	if batchHas(cmd, isTapsReload) {
		t.Error("a failed upgrade reloaded the Taps list")
	}
}

// The reload shows the loading state and fetches the list again.
func TestTapsReloadMsgFetchesTheList(t *testing.T) {
	m := tapsPanelWith([]string{"a/one"}, 0)
	nm, cmd := m.Update(tapsReloadMsg{})
	if cmd == nil {
		t.Fatal("the Taps reload queued no read")
	}
	if !asModel(nm).panels[PanelTaps].loading {
		t.Error("the Taps reload should show the loading state")
	}
}

// After a reload the cursor is on the same tap, even when the list moved under it.
func TestTapsReloadKeepsTheCursorOnTheSameTap(t *testing.T) {
	m := tapsPanelWith([]string{"a/one", "b/two", "c/three"}, 1)
	m = updateModel(m, landedTaps("c/three", "a/one", "b/two"))
	if got := m.panels[PanelTaps].selectedTap(); got == nil || got.Name != "b/two" {
		t.Fatalf("the cursor should stay on b/two, got %+v", got)
	}
}

// After an untap the cursor is on the tap that followed the removed one.
func TestUntapLeavesTheCursorOnTheNextTap(t *testing.T) {
	m := tapsPanelWith([]string{"a/one", "b/two", "c/three"}, 1)
	m = updateModel(m, landedTaps("a/one", "c/three"))
	if got := m.panels[PanelTaps].selectedTap(); got == nil || got.Name != "c/three" {
		t.Fatalf("after untapping b/two the cursor should be on c/three, got %+v", got)
	}
}

// Untapping the last tap leaves the cursor on the tap before it, not past the end.
func TestUntapOfTheLastTapLeavesTheCursorInRange(t *testing.T) {
	m := tapsPanelWith([]string{"a/one", "b/two"}, 1)
	m = updateModel(m, landedTaps("a/one"))
	if got := m.panels[PanelTaps].selectedTap(); got == nil || got.Name != "a/one" {
		t.Fatalf("after untapping the last tap the cursor should be on a/one, got %+v", got)
	}
}
