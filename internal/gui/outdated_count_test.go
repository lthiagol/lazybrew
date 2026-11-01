package gui

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
)

// collectRefreshMsgs drains a command (and every command in a batch it
// produces) and returns the RefreshMsg values it emits.
func collectRefreshMsgs(cmd tea.Cmd) []RefreshMsg {
	var out []RefreshMsg
	for _, msg := range drainCmds(cmd) {
		if refresh, ok := msg.(RefreshMsg); ok {
			out = append(out, refresh)
		}
	}
	return out
}

// TestNonRRefreshEmittersStayCacheOnly pins M9's invariant: only an explicit
// R may load the outdated count. The install, uninstall and task-completion
// emitters must emit a *plain* refresh.
//
// The assertion is on the emitted message (LoadOutdated == false), not on the
// absence of a brew call, so the test bites exactly when an emitter is
// changed to RefreshMsg{LoadOutdated: true}.
func TestNonRRefreshEmittersStayCacheOnly(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.Msg
	}{
		{"install completion", TaskCompletedMsg{ID: "ripgrep", Title: "Install ripgrep"}},
		{"uninstall completion", TaskCompletedMsg{ID: "ripgrep", Title: "Uninstall ripgrep"}},
		{"install mutation result", MutationResultMsg{Name: "ripgrep", Type: mutInstall}},
		{"uninstall mutation result", MutationResultMsg{Name: "ripgrep", Type: mutUninstall}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel()
			_, cmd := m.Update(tc.msg)
			if cmd == nil {
				t.Fatalf("%s: expected a command, got nil", tc.name)
			}
			refreshes := collectRefreshMsgs(cmd)
			if len(refreshes) == 0 {
				t.Fatalf("%s: expected a RefreshMsg, got none", tc.name)
			}
			for _, refresh := range refreshes {
				if refresh.LoadOutdated {
					t.Errorf("%s: the emitted refresh must stay plain, got RefreshMsg{LoadOutdated: true}", tc.name)
				}
			}
		})
	}
}

// TestStatusRefreshKeepsOutdatedCache covers the cache-survival behaviour:
// a status refresh (the path used after an install and on the auto-refresh
// tick) must not throw away the outdated cache entries, so the dashboard
// keeps rendering the real count instead of "not loaded".
//
// The test fails if the two outdated cache keys are put back into
// invalidatePanelCache's PanelStatus case.
func TestStatusRefreshKeepsOutdatedCache(t *testing.T) {
	var outdatedCalls int32
	client := brew.NewClient(outdatedCountingRunner(&outdatedCalls))
	client.Cache.SetWithTTL(brew.KeyOutdatedFormulae, []brew.Formula{{Name: "ripgrep"}}, time.Minute)
	client.Cache.SetWithTTL(brew.KeyOutdatedCasks, []brew.Cask{}, time.Minute)

	m := New(client, config.Default())
	m.activePanel = PanelStatus
	m.cfg.GUI.AutoRefreshSeconds = 0

	_, cmd := m.Update(RefreshMsg{})
	if cmd == nil {
		t.Fatal("RefreshMsg returned nil cmd")
	}
	msgs := drainCmds(cmd)

	if _, ok := client.Cache.Get(brew.KeyOutdatedFormulae); !ok {
		t.Error("a status refresh must not invalidate the outdated formulae cache entry")
	}
	if _, ok := client.Cache.Get(brew.KeyOutdatedCasks); !ok {
		t.Error("a status refresh must not invalidate the outdated casks cache entry")
	}
	if got := atomic.LoadInt32(&outdatedCalls); got != 0 {
		t.Errorf("a plain RefreshMsg must not shell `brew outdated`, got %d calls", got)
	}

	var found bool
	for _, msg := range msgs {
		dMsg, ok := msg.(DataLoadedMsg)
		if !ok || dMsg.PanelID != PanelStatus {
			continue
		}
		found = true
		joined := strings.Join(dMsg.Items, "\n")
		if !strings.Contains(joined, "Outdated:  1 packages") {
			t.Errorf("the dashboard must still render the cached count, got:\n%s", joined)
		}
	}
	if !found {
		t.Error("expected a DataLoadedMsg for the status panel")
	}
}

// TestRKeyLoadsOutdatedWhenUpdateOnStart covers the edge case where
// brew.update_on_start is true: the first R both runs `brew update` and
// must still load the outdated count, so the dashboard is not stuck at
// "not loaded" until R is pressed again.
//
// The load flag belongs to the R path only: the brew-update completion
// handler also serves the Init path and must stay plain.
func TestRKeyLoadsOutdatedWhenUpdateOnStart(t *testing.T) {
	m := newTestModel()
	m.cfg.Brew.UpdateOnStart = true
	m.cfg.GUI.AutoRefreshSeconds = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	if cmd == nil {
		t.Fatal("R returned nil cmd")
	}
	msgs := drainCmds(cmd)

	var foundRefresh, foundTask bool
	for _, msg := range msgs {
		switch v := msg.(type) {
		case RefreshMsg:
			foundRefresh = true
			if !v.LoadOutdated {
				t.Error("R must request the outdated load even when it also runs brew update")
			}
		case TaskStartedMsg:
			if v.ID == "brew-update" {
				foundTask = true
			}
		}
	}
	if !foundRefresh {
		t.Error("expected a RefreshMsg from the R key")
	}
	if !foundTask {
		t.Error("expected the brew-update task to start")
	}
}

// TestBrewUpdateCompletionStaysPlain guards the constraint that keeps the
// load flag on the R path: the brew-update completion handler serves both
// the R key and the Init path, so it must emit a plain refresh — otherwise
// the count would be loaded at startup, which M9 forbids.
func TestBrewUpdateCompletionStaysPlain(t *testing.T) {
	m := newTestModel()
	m.cfg.Brew.UpdateOnStart = true

	_, cmd := m.Update(TaskCompletedMsg{ID: "brew-update", Title: "brew update"})
	if cmd == nil {
		t.Fatal("brew-update completion returned nil cmd")
	}
	refreshes := collectRefreshMsgs(cmd)
	if len(refreshes) == 0 {
		t.Fatal("expected a RefreshMsg, got none")
	}
	for _, refresh := range refreshes {
		if refresh.LoadOutdated {
			t.Error("the brew-update completion must stay plain, got RefreshMsg{LoadOutdated: true}")
		}
	}
}
