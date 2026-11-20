package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// tapTrustUnknownModel is the Taps panel at 140x40 with two taps read without
// their trust state (v2-10 AC-00): a name-only list, as tap-info left it.
func tapTrustUnknownModel(t *testing.T) *Model {
	t.Helper()
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{
		PanelID:      PanelTaps,
		Items:        []string{"homebrew/core", "user/tool"},
		Taps:         []brew.Tap{{Name: "homebrew/core", IsOfficial: true, Installed: true}, {Name: "user/tool", Installed: true}},
		TrustUnknown: true,
	})
	m = sendKey(m, "5")
	m.panels[PanelTaps].loading = false
	return m
}

func TestTapsViewSaysTrustUnknownOnBorderAndTapInfo(t *testing.T) {
	m := tapTrustUnknownModel(t)
	out := ansi.Strip(m.View().Content)
	if n := strings.Count(out, "trust unknown — press R"); n < 2 {
		t.Fatalf("the Taps border label and Tap Info should both say %q, found it %d times:\n%s", "trust unknown — press R", n, out)
	}
	if !strings.Contains(out, "Trusted") {
		t.Fatalf("Tap Info should have a Trusted row:\n%s", out)
	}
	if strings.Contains(out, "untrusted") {
		t.Fatalf("a tap whose trust was not read must not read as untrusted:\n%s", out)
	}
}

func TestTapRowsCarryNoTrustWordWhenTrustUnknown(t *testing.T) {
	p := &panelData{id: PanelTaps, taps: []brew.Tap{{Name: "user/tool"}}, trustUnknown: true}
	p.items = []string{"user/tool"}
	got := plainLine(p.sidebarRow(0, 30, nil, style.UnicodeIcons, false))
	if strings.Contains(got, "untrusted") || strings.Contains(got, "trusted") {
		t.Fatalf("an unread tap row carries no trust word: %q", got)
	}
	if !strings.Contains(got, "user/tool") {
		t.Fatalf("the row keeps the tap name: %q", got)
	}
}

func TestTrustKeysRefusedWhenTrustUnknown(t *testing.T) {
	m := tapTrustUnknownModel(t)
	m.panels[PanelTaps].selected = 1

	m = sendKey(m, "t")
	if got := lastToast(m); got != "trust state not loaded — press R" {
		t.Fatalf("t with the trust state unread: toast = %q", got)
	}
	if m.activeModal != nil {
		t.Fatalf("t with the trust state unread opened %T", m.activeModal)
	}
	if !noTaskQueued(m) {
		t.Fatal("t with the trust state unread enqueued a task")
	}
}

func TestUntapRepairAddStillWorkWhenTrustUnknown(t *testing.T) {
	m := tapTrustUnknownModel(t)
	m.panels[PanelTaps].selected = 1

	m = sendKey(m, "x")
	if _, ok := m.activeModal.(*modal.ConfirmModal); !ok {
		t.Fatalf("x should still open the untap confirm, got %T", m.activeModal)
	}
	m.activeModal = nil

	m = sendKey(m, "r")
	if _, ok := m.activeModal.(*modal.ConfirmModal); !ok {
		t.Fatalf("r should still open the repair confirm, got %T", m.activeModal)
	}
	m.activeModal = nil

	m = sendKey(m, "a")
	if _, ok := m.activeModal.(*modal.InputModal); !ok {
		t.Fatalf("a should still open the add-tap input, got %T", m.activeModal)
	}
}

func TestDataLoadedClearsTrustUnknownOnFullRead(t *testing.T) {
	m := tapTrustUnknownModel(t)
	m = updateModel(m, DataLoadedMsg{
		PanelID: PanelTaps,
		Items:   []string{"user/tool"},
		Taps:    []brew.Tap{{Name: "user/tool", Installed: true, Trusted: true}},
	})
	if m.panels[PanelTaps].trustUnknown {
		t.Fatal("a full tap-info read must clear the trust-unknown state")
	}
	if out := ansi.Strip(m.View().Content); strings.Contains(out, "trust unknown") {
		t.Fatalf("a full read must not say trust unknown:\n%s", out)
	}
}
