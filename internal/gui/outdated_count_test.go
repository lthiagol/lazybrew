package gui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
