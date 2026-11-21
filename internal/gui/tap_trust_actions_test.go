package gui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// tapsActionModel is the Taps panel over a mock runner that records every brew
// call. The trust writes these tests make only ever reach the mock.
func tapsActionModel(t *testing.T, taps []brew.Tap) (*Model, *[][]string) {
	t.Helper()
	calls := &[][]string{}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		*calls = append(*calls, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), testConfig())
	m.switchPanel(PanelTaps)
	p := m.panels[PanelTaps]
	p.taps = taps
	p.items = make([]string, len(taps))
	for i, tap := range taps {
		p.items[i] = tap.Name
	}
	p.loaded = true
	p.loading = false
	p.selected = 0
	return m, calls
}

// pressKeyOn sends one key through Update and returns the model and the command
// it queued, so a test can run what the key started.
func pressKeyOn(m *Model, key string) (*Model, tea.Cmd) {
	nm, cmd := m.Update(tea.KeyPressMsg{Code: []rune(key)[0], Text: key})
	return asModel(nm), cmd
}

// runQueuedTapTask runs the task a key or a confirm queued, through the task
// manager, and fails unless it started and completed cleanly.
func runQueuedTapTask(t *testing.T, m *Model, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("no task was queued")
	}
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected TaskStartedMsg from the queued tap task")
	}
	if comp, ok := m.tasks.RunNext()().(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("expected a clean completion, got %+v", comp)
	}
}

func untrustedTap(name string) brew.Tap { return brew.Tap{Name: name, Installed: true} }

func trustedTap(name string) brew.Tap { return brew.Tap{Name: name, Installed: true, Trusted: true} }

// t on an untrusted tap trusts it at once: one trust write, and no confirm.
func TestTapTrustKeyTrustsAnUntrustedTapAtOnce(t *testing.T) {
	m, calls := tapsActionModel(t, []brew.Tap{untrustedTap("user/tool")})

	m, cmd := pressKeyOn(m, "t")
	if m.activeModal != nil {
		t.Fatalf("t on an untrusted tap opened %T; it should trust at once", m.activeModal)
	}
	runQueuedTapTask(t, m, cmd)
	if !hasCall(*calls, "trust", "user/tool") {
		t.Fatalf("expected `brew trust user/tool`, got %v", *calls)
	}
}

// t on a trusted tap asks first, with the upgrade kind, the UNTRUST chip and the
// consequence; nothing runs until the confirm is answered yes.
func TestTapTrustKeyOnTrustedTapAsksFirst(t *testing.T) {
	m, calls := tapsActionModel(t, []brew.Tap{trustedTap("user/tool")})

	m, cmd := pressKeyOn(m, "t")
	if cmd != nil {
		if _, ok := cmd().(TaskStartedMsg); ok {
			t.Fatal("t on a trusted tap started a task before the confirm")
		}
	}
	c, ok := m.activeModal.(*modal.ConfirmModal)
	if !ok {
		t.Fatalf("t on a trusted tap should open the untrust confirm, got %T", m.activeModal)
	}
	if c.Action() != "Untrust" || c.Kind() != modal.ActionUpgrade {
		t.Errorf("confirm = %q kind %v, want the Untrust chip in the upgrade kind", c.Action(), c.Kind())
	}
	if want := "Homebrew will refuse to load its formulae and casks until you trust it again"; !strings.Contains(c.Detail(), want) {
		t.Errorf("confirm detail = %q, want it to say %q", c.Detail(), want)
	}
	if len(*calls) != 0 {
		t.Fatalf("the confirm ran brew before it was answered: %v", *calls)
	}

	confirmAndRun(t, m)
	if !hasCall(*calls, "untrust", "user/tool") {
		t.Fatalf("a confirmed untrust should run `brew untrust user/tool`, got %v", *calls)
	}
}

// A cancelled untrust confirm runs nothing.
func TestTapUntrustCancelledRunsNothing(t *testing.T) {
	m, calls := tapsActionModel(t, []brew.Tap{trustedTap("user/tool")})
	m, _ = pressKeyOn(m, "t")

	m.activeModal = nil
	m.handleModalResult(&modal.ConfirmResult{Confirmed: false}, nil)
	if !noTaskQueued(m) {
		t.Fatal("a cancelled untrust confirm queued a task")
	}
	if len(*calls) != 0 {
		t.Fatalf("a cancelled untrust ran brew: %v", *calls)
	}
}

// t on the official tap says it needs no trust and enqueues nothing; T is
// refused the same way and opens no menu.
func TestTapTrustRefusedOnOfficialTap(t *testing.T) {
	official := brew.Tap{Name: "homebrew/core", IsOfficial: true, Installed: true}
	for _, key := range []string{"t", "T"} {
		m, calls := tapsActionModel(t, []brew.Tap{official})
		m, cmd := pressKeyOn(m, key)
		if got := lastToast(m); got != "homebrew/core is official — no trust needed" {
			t.Errorf("%s on an official tap: toast = %q", key, got)
		}
		if m.activeModal != nil {
			t.Errorf("%s on an official tap opened %T", key, m.activeModal)
		}
		if cmd != nil {
			if _, ok := cmd().(TaskStartedMsg); ok {
				t.Errorf("%s on an official tap started a task", key)
			}
		}
		if len(*calls) != 0 {
			t.Errorf("%s on an official tap ran brew: %v", key, *calls)
		}
	}
}

// T opens the item menu for an untrusted tap, which holds the tap-level trust
// entry and the untrust entry.
func TestTapTrustMenuKeyOpensTheMenu(t *testing.T) {
	m, _ := tapsActionModel(t, []brew.Tap{untrustedTap("user/tool")})
	m, _ = pressKeyOn(m, "T")
	if _, ok := m.activeModal.(*modal.MenuModal); !ok {
		t.Fatalf("T should open the trust menu, got %T", m.activeModal)
	}
	if view := m.activeModal.View().Content; !strings.Contains(view, "Trust entire tap: user/tool") {
		t.Fatalf("trust menu lacks the tap entry:\n%s", view)
	}
}

// The menu's untrust entry is confirmed too, so an untrust never skips the
// confirm whichever way it is asked for.
func TestTapUntrustFromTheMenuIsConfirmed(t *testing.T) {
	m, calls := tapsActionModel(t, []brew.Tap{trustedTap("user/tool")})
	m.pendingAction = "trust-menu"
	m.handleModalResult(&modal.MenuResult{SelectedIndex: 1}, nil)

	c, ok := m.activeModal.(*modal.ConfirmModal)
	if !ok {
		t.Fatalf("the menu's untrust entry should open the confirm, got %T", m.activeModal)
	}
	if c.Action() != "Untrust" {
		t.Fatalf("confirm chip = %q, want Untrust", c.Action())
	}
	if !noTaskQueued(m) || len(*calls) != 0 {
		t.Fatalf("the menu's untrust ran before the confirm: calls %v", *calls)
	}
}
