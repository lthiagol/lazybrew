package gui

// Tests for the r4 review round (final): each test names the finding it
// locks in a comment.

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// asModel normalizes the tea.Model values returned by the various
// (value- and pointer-receiver) helpers back into a *Model.
func asModel(nm tea.Model) *Model {
	switch v := nm.(type) {
	case Model:
		return &v
	case *Model:
		return v
	}
	return nil
}

// --- R4-1: a cancelled menu must not leave pendingAction set -------------

// TestCancelledMenuClearsPendingAction locks r4 finding R4-1: a
// CANCELLED MenuResult must clear pendingAction, exactly like the
// ConfirmResult and InputResult cancel paths already do.
func TestCancelledMenuClearsPendingAction(t *testing.T) {
	for _, action := range []string{"trust-menu", "trust-formula", "trust-cask"} {
		m := newTestModel()
		m.pendingAction = action
		nm, _ := m.handleModalResult(&modal.MenuResult{SelectedIndex: 0, Cancelled: true}, nil)
		rm := asModel(nm)
		if rm.pendingAction != "" {
			t.Errorf("pendingAction %q must be cleared after a cancelled menu, got %q", action, rm.pendingAction)
		}
	}
}

// TestCancelledTrustMenuDoesNotPoisonBrewfileMenu drives the exact R4-1
// scenario through Update: cancel the trust menu with Esc, then open the
// Brewfile menu (which sets no pendingAction) and pick "Export to
// Brewfile". The stale pendingAction used to route that pick into
// executeTrustMenuAction(0) — trusting whatever tap was selected.
func TestCancelledTrustMenuDoesNotPoisonBrewfileMenu(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelTaps]
	p.items = []string{"user/tap"}
	p.selected = 0
	p.taps = []brew.Tap{{Name: "user/tap", FormulaNames: []string{"f1"}}}

	// Open the trust menu and cancel it.
	m = sendKey(m, "5")
	m = sendKey(m, "t")
	if m.activeModal == nil {
		t.Fatal("expected the trust menu to open")
	}
	m = sendSpecial(m, tea.KeyEsc)
	if m.pendingAction != "" {
		t.Fatalf("cancelled trust menu must clear pendingAction, got %q", m.pendingAction)
	}

	// Switch to Status and open the Brewfile menu, then pick entry 0.
	m = sendKey(m, "1")
	m = sendKey(m, "B")
	if m.activeModal == nil {
		t.Fatal("expected the Brewfile menu to open")
	}
	if m.pendingAction != "" {
		t.Fatalf("the Brewfile menu must not set pendingAction, got %q", m.pendingAction)
	}
	nm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = asModel(nm)

	ts, ok := firstTaskStarted(execCmds(cmd))
	if !ok {
		t.Fatal("expected the brewfile-dump task to start")
	}
	if ts.ID != "brewfile-dump" {
		t.Fatalf("the Brewfile pick must run brewfile-dump, got task %q (a trust action here means pendingAction leaked)", ts.ID)
	}
}

// --- R4-2: stale batch selection must not survive an Outdated refresh ----

// TestBatchSelectionClearedOnOutdatedRefresh locks r4 finding R4-2: the
// batch selection is index-based, so a refresh that rewrites the
// Outdated panel must drop it — stale indexes silently remapped the ●
// markers onto different rows and made a later batch upgrade target the
// wrong packages.
func TestBatchSelectionClearedOnOutdatedRefresh(t *testing.T) {
	m := newTestModel()
	seedOutdatedBatch(m, 2)
	if len(m.batch.selected) != 2 {
		t.Fatalf("setup: expected 2 selected rows, got %d", len(m.batch.selected))
	}

	// The refresh rewrites the list with different packages.
	m = updateModel(m, DataLoadedMsg{
		PanelID:  PanelOutdated,
		Items:    []string{"other  2.0"},
		Formulae: []brew.Formula{{Name: "other", Version: "1.0", NewVersion: "2.0"}},
	})
	if len(m.batch.selected) != 0 {
		t.Fatalf("batch selection must be cleared when the Outdated list is rewritten, got %d entries", len(m.batch.selected))
	}

	// Pressing u now falls back to the single-package upgrade confirm,
	// never the batch confirm (which would have promised N upgrades for
	// whatever now occupies the stale indexes).
	m = sendKey(m, "4")
	m.panels[PanelOutdated].selected = 0
	m = sendKey(m, "u")
	if m.pendingAction == "batch-upgrade" {
		t.Fatal("a cleared selection must not open the batch upgrade confirm")
	}
	if m.pendingAction != "mutation" {
		t.Fatalf("expected the single-package upgrade confirm, got pendingAction %q", m.pendingAction)
	}
}

// TestBatchSelectionKeptOnFailedOutdatedRefresh locks the complement of
// R4-2: a FAILED refresh keeps the old items, so the still-accurate
// selection is not dropped.
func TestBatchSelectionKeptOnFailedOutdatedRefresh(t *testing.T) {
	m := newTestModel()
	seedOutdatedBatch(m, 2)

	m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Err: assertAnError})
	if len(m.batch.selected) != 2 {
		t.Fatalf("a failed refresh keeps the list, so the selection must survive, got %d entries", len(m.batch.selected))
	}
}

// TestBatchUpgradeZeroValidSelectionSurfacesToast locks the R4-2
// dead-end: when every selected index has fallen out of range, the
// confirmed batch enqueues nothing — after the modal promised N
// upgrades, silence is not acceptable.
func TestBatchUpgradeZeroValidSelectionSurfacesToast(t *testing.T) {
	m := newTestModel()
	panel := m.panels[PanelOutdated]
	panel.items = []string{"foo  1.0"}
	panel.formulae = []brew.Formula{{Name: "foo"}}
	m.batch.selected = map[int]bool{7: true, 8: true}

	nm, cmd := m.batchUpgrade()
	rm := nm.(Model)
	if rm.toast == nil || !strings.Contains(rm.toast.Message, "Outdated list") {
		t.Fatalf("expected a warning toast when no selected package is still in the list, got %+v", rm.toast)
	}
	if rm.batchCount != 0 || rm.batchTotal != 0 || len(rm.batchIDs) != 0 {
		t.Fatalf("an empty batch must leave no counter state, got count=%d total=%d ids=%v",
			rm.batchCount, rm.batchTotal, rm.batchIDs)
	}
	for _, msg := range execCmds(cmd) {
		if _, ok := msg.(TaskStartedMsg); ok {
			t.Fatal("no batch task may start when no selected package is valid")
		}
	}
}

// TestRefreshDuringBatchConfirmStillSurfacesToast locks r5 finding R5-1:
// a refresh that lands while the batch confirm modal is open clears the
// selection (the R4-2 fix); confirming the modal that promised N
// upgrades must still surface the warning toast instead of ending in a
// silent dead-end.
func TestRefreshDuringBatchConfirmStillSurfacesToast(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "4")
	p := m.panels[PanelOutdated]
	p.items = []string{"foo  1.0"}
	p.formulae = []brew.Formula{{Name: "foo"}}
	p.selected = 0
	m.batch.selected[0] = true

	m = sendKey(m, "u")
	if m.pendingAction != "batch-upgrade" || m.activeModal == nil {
		t.Fatalf("expected the batch confirm modal to open, got pendingAction=%q modal=%v", m.pendingAction, m.activeModal)
	}

	// A refresh lands while the modal is up: the selection is cleared.
	m = updateModel(m, DataLoadedMsg{
		PanelID:  PanelOutdated,
		Items:    []string{"foo  1.0"},
		Formulae: []brew.Formula{{Name: "foo"}},
	})
	if len(m.batch.selected) != 0 {
		t.Fatal("setup: expected the refresh to clear the selection")
	}

	// Confirming must not be a silent dead-end. (The confirm modal
	// defaults to "No"; "y" is the explicit confirm.)
	nm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	rm := asModel(nm)
	if rm.toast == nil || !strings.Contains(rm.toast.Message, "Outdated list") {
		t.Fatalf("expected the empty-selection warning toast after confirming, got %+v", rm.toast)
	}
	for _, msg := range execCmds(cmd) {
		if _, ok := msg.(TaskStartedMsg); ok {
			t.Fatal("no batch task may start when the selection was drained")
		}
	}
}

// --- R4-3: the trust menu's submenu entries are position-independent -----

// TestTrustMenuCaskOnlyTapOpensCaskSubmenu locks r4 finding R4-3: on a
// tap with casks but NO formulae, "Trust specific cask..." lands at menu
// index 2, which the hardcoded dispatch mapped to the (unreachable)
// formula submenu — the entry was dead.
func TestTrustMenuCaskOnlyTapOpensCaskSubmenu(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelTaps]
	p.items = []string{"user/casktap"}
	p.selected = 0
	p.taps = []brew.Tap{{Name: "user/casktap", CaskNames: []string{"firefox", "chrome"}}}

	nm, _ := m.startTrustMenu()
	mm := asModel(nm)
	if mm.activeModal == nil {
		t.Fatal("expected the trust menu to open")
	}

	nm2, _ := mm.handleModalResult(&modal.MenuResult{SelectedIndex: 2}, nil)
	mm2 := asModel(nm2)
	if mm2.activeModal == nil {
		t.Fatal("selecting 'Trust specific cask...' on a cask-only tap must open the cask submenu")
	}
	if mm2.pendingAction != "trust-cask" {
		t.Fatalf("expected pendingAction 'trust-cask', got %q", mm2.pendingAction)
	}

	// The cask submenu dispatches by cask name.
	nm3, cmd := mm2.handleModalResult(&modal.MenuResult{SelectedIndex: 0}, nil)
	mm3 := asModel(nm3)
	if mm3.pendingAction != "" {
		t.Fatalf("the submenu result must clear pendingAction, got %q", mm3.pendingAction)
	}
	ts, ok := firstTaskStarted(execCmds(cmd))
	if !ok {
		t.Fatal("expected the cask trust task to start")
	}
	if ts.ID != "trust:firefox" {
		t.Fatalf("expected trust:firefox, got %q", ts.ID)
	}
}

// TestTrustMenuFormulaOnlyTapOpensFormulaSubmenu locks the mirror case:
// on a formula-only tap the formula entry is at index 2 and must open
// the formula submenu.
func TestTrustMenuFormulaOnlyTapOpensFormulaSubmenu(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelTaps]
	p.items = []string{"user/ftap"}
	p.selected = 0
	p.taps = []brew.Tap{{Name: "user/ftap", FormulaNames: []string{"ripgrep"}}}

	nm, _ := m.startTrustMenu()
	mm := asModel(nm)

	nm2, _ := mm.handleModalResult(&modal.MenuResult{SelectedIndex: 2}, nil)
	mm2 := asModel(nm2)
	if mm2.activeModal == nil {
		t.Fatal("selecting 'Trust specific formula...' on a formula-only tap must open the formula submenu")
	}
	if mm2.pendingAction != "trust-formula" {
		t.Fatalf("expected pendingAction 'trust-formula', got %q", mm2.pendingAction)
	}
}

// TestTrustMenuBothKindsKeepsSubmenuPositions locks the unchanged
// behaviour on a tap with both formulae and casks: index 2 is the
// formula submenu, index 3 the cask submenu.
func TestTrustMenuBothKindsKeepsSubmenuPositions(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelTaps]
	p.items = []string{"user/tap"}
	p.selected = 0
	p.taps = []brew.Tap{{
		Name:         "user/tap",
		FormulaNames: []string{"ripgrep"},
		CaskNames:    []string{"firefox"},
	}}

	nm, _ := m.startTrustMenu()
	mm := asModel(nm)

	nm2, _ := mm.handleModalResult(&modal.MenuResult{SelectedIndex: 3}, nil)
	mm2 := asModel(nm2)
	if mm2.pendingAction != "trust-cask" {
		t.Fatalf("index 3 on a both-kinds tap must open the cask submenu, got pendingAction %q", mm2.pendingAction)
	}

	m2 := newTestModel()
	p2 := m2.panels[PanelTaps]
	p2.items = []string{"user/tap"}
	p2.selected = 0
	p2.taps = []brew.Tap{{
		Name:         "user/tap",
		FormulaNames: []string{"ripgrep"},
		CaskNames:    []string{"firefox"},
	}}
	nm3, _ := m2.startTrustMenu()
	mm3 := asModel(nm3)
	nm4, _ := mm3.handleModalResult(&modal.MenuResult{SelectedIndex: 2}, nil)
	mm4 := asModel(nm4)
	if mm4.pendingAction != "trust-formula" {
		t.Fatalf("index 2 on a both-kinds tap must open the formula submenu, got pendingAction %q", mm4.pendingAction)
	}
}

// --- R4-4: batch task IDs are unique per batch ---------------------------

// TestBatchUpgradeIDsUniquePerBatch locks r4 finding R4-4: a task still
// queued from a superseded batch shares its package name with the next
// batch, so the bare "batch-upgrade:<name>" ID let the stale completion
// decrement the NEW batch's counter early. IDs now carry a per-batch
// sequence number.
func TestBatchUpgradeIDsUniquePerBatch(t *testing.T) {
	m := newTestModel()
	seedOutdatedBatch(m, 1)

	nm, cmd := m.batchUpgrade()
	rm := nm.(Model)
	ts, ok := firstTaskStarted(execCmds(cmd))
	if !ok {
		t.Fatal("expected the first batch's task to start")
	}
	firstID := ts.ID
	if !strings.HasPrefix(firstID, "batch-upgrade:1:") {
		t.Fatalf("expected the first batch's ID to carry sequence 1, got %q", firstID)
	}

	// The first batch's task never completes before the next batch is
	// confirmed (superseded in the queue). Re-select the same package.
	m = &rm
	m.batch.selected[0] = true
	nm2, _ := m.batchUpgrade()
	rm2 := nm2.(Model)
	if rm2.batchCount != 1 || rm2.batchTotal != 1 {
		t.Fatalf("expected the second batch to count 1/1, got count=%d total=%d", rm2.batchCount, rm2.batchTotal)
	}
	var secondID string
	for id := range rm2.batchIDs {
		secondID = id
	}
	if secondID == firstID {
		t.Fatalf("batch IDs must be unique per batch, both were %q", firstID)
	}
	if !strings.HasPrefix(secondID, "batch-upgrade:2:") {
		t.Fatalf("expected the second batch's ID to carry sequence 2, got %q", secondID)
	}

	// The superseded task completes: it must NOT consume the new batch's
	// count occurrence.
	m = updateModel(&rm2, TaskCompletedMsg{ID: firstID, Title: "Upgrade pkga"})
	if m.batchCount != 1 {
		t.Fatalf("a superseded batch's completion must not decrement the new batch's counter, got %d", m.batchCount)
	}
	if len(m.batchIDs) != 1 || m.batchTotal != 1 {
		t.Fatalf("the new batch's tracking must be untouched, got ids=%v total=%d", m.batchIDs, m.batchTotal)
	}

	// The new batch's own completion drains the counter and resets the
	// overlay total.
	m = updateModel(m, TaskCompletedMsg{ID: secondID, Title: "Upgrade pkga"})
	if m.batchCount != 0 || m.batchTotal != 0 || len(m.batchIDs) != 0 {
		t.Fatalf("the new batch's own completion must drain the batch, got count=%d total=%d ids=%v",
			m.batchCount, m.batchTotal, m.batchIDs)
	}
}
