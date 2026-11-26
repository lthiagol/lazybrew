package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// tapOutcome says what a key on the Taps list did: the confirm, menu or input it
// opened, the toast it showed, the task it queued, or nothing. It checks the
// modal and the toast before it runs the command, so a toast's timer is not run.
func tapOutcome(m *Model, toastsBefore int, cmd tea.Cmd) string {
	switch c := m.activeModal.(type) {
	case *modal.ConfirmModal:
		return "confirm " + c.Action()
	case *modal.MenuModal:
		return "menu"
	case *modal.InputModal:
		return "input"
	}
	if len(m.toasts) > toastsBefore {
		return "toast " + lastToast(m)
	}
	if cmd != nil {
		if _, ok := cmd().(TaskStartedMsg); ok {
			return "task"
		}
	}
	return "nothing"
}

// TestTapActionsByState is the state × action matrix for the Taps list (v2-10
// AC-04.2): every tap state against every key the Taps list answers to. Each
// cell says what the key does, on the mock runner.
func TestTapActionsByState(t *testing.T) {
	const official = "homebrew/core"
	const user = "user/tool"
	states := []struct {
		name string
		tap  brew.Tap
		// unread is the trust state not loaded (a name-only Taps list).
		unread bool
		// want is the outcome of each key, in the order of keys.
		want []string
	}{
		{
			name: "untrusted",
			tap:  brew.Tap{Name: user, Installed: true},
			want: []string{"task", "menu", "confirm Untap", "confirm Repair", "input"},
		},
		{
			name: "trusted",
			tap:  brew.Tap{Name: user, Installed: true, Trusted: true},
			want: []string{"confirm Untrust", "menu", "confirm Untap", "confirm Repair", "input"},
		},
		{
			name: "official",
			tap:  brew.Tap{Name: official, Installed: true, IsOfficial: true},
			want: []string{
				"toast homebrew/core is official — no trust needed",
				"toast homebrew/core is official — no trust needed",
				"toast Cannot untap official taps",
				"confirm Repair",
				"input",
			},
		},
		{
			name:   "unread",
			tap:    brew.Tap{Name: user, Installed: true},
			unread: true,
			want: []string{
				"toast trust state not loaded — press R",
				"toast trust state not loaded — press R",
				"confirm Untap",
				"confirm Repair",
				"input",
			},
		},
	}
	keys := []string{"t", "T", "x", "r", "a"}

	for _, s := range states {
		t.Run(s.name, func(t *testing.T) {
			for i, key := range keys {
				m, calls := tapsActionModel(t, []brew.Tap{s.tap})
				m.panels[PanelTaps].trustUnknown = s.unread
				toasts := len(m.toasts)
				m, cmd := pressKeyOn(m, key)
				if got := tapOutcome(m, toasts, cmd); got != s.want[i] {
					t.Errorf("%s: %q does %q, want %q", s.name, key, got, s.want[i])
				}
				// Only the task starts brew, and only it: a confirm or menu runs nothing.
				if s.want[i] == "task" {
					// The task's reader runs its brew call.
					if r := m.tasks.RunNext(); r != nil {
						r()
					}
					if !hasCall(*calls, "trust", user) {
						t.Errorf("%s: %q should run `brew trust %s`, got %v", s.name, key, user, *calls)
					}
				} else if len(*calls) != 0 {
					t.Errorf("%s: %q ran brew before any confirm: %v", s.name, key, *calls)
				}
			}
		})
	}
}
