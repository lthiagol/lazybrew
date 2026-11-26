package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// outdatedFormulaModel is a model on a loaded Outdated panel listing one
// formula, whose writes r records.
func outdatedFormulaModel(r *searchUninstallRunner, f brew.Formula) *Model {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 120, Height: 40})
	m.client = brew.NewClient(r.mock())
	m.activePanel = PanelOutdated
	p := m.panels[PanelOutdated]
	p.loaded = true
	p.formulae = []brew.Formula{f}
	p.items = []string{f.Name + "  1.0"}
	p.selected = 0
	return m
}

// TestOutdatedTapUpgradeWritesTheFullName: upgrade from Outdated names the tap
// formula in full, so brew does not resolve it to homebrew/core's tool.
func TestOutdatedTapUpgradeWritesTheFullName(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{}}
	m := outdatedFormulaModel(r, brew.Formula{Name: "tool", FullName: "user/tap/tool"})

	m = sendKey(m, "u")
	if m.activeModal == nil || !strings.Contains(m.View().Content, "Upgrade tool?") {
		t.Fatalf("u opened no upgrade dialog for tool:\n%s", m.View().Content)
	}
	confirmAndRun(t, m)
	wantWrites(t, r, "upgrade user/tap/tool")
}

// TestOutdatedBatchUpgradeWritesTheFullName: the batch path gives brew the same
// full name, and its task ID carries it, so a core formula of the same short
// name cannot share the ID.
func TestOutdatedBatchUpgradeWritesTheFullName(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{}}
	m := outdatedFormulaModel(r, brew.Formula{Name: "tool", FullName: "user/tap/tool"})
	m.batch.selected[0] = true

	nm, cmd := m.batchUpgrade()
	rm := nm.(Model)
	ts, ok := firstTaskStarted(execCmds(cmd))
	if !ok || ts.ID != "batch-upgrade:1:user/tap/tool" {
		t.Fatalf("batch task = %+v (started %v), want batch-upgrade:1:user/tap/tool", ts, ok)
	}
	if comp, ok := rm.tasks.RunNext()().(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("expected a clean completion, got %+v", comp)
	}
	wantWrites(t, r, "upgrade user/tap/tool")
}

// TestInstalledTapFormulaUninstallWritesTheFullName: uninstall from the
// Formulae panel names the tap formula in full; the dialog keeps the short name.
func TestInstalledTapFormulaUninstallWritesTheFullName(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{"tool": true}}
	m := formulaeModel(r, "tool")
	m.panels[PanelFormulae].formulae[0].FullName = "user/tap/tool"

	m, cmd := pressKeyOn(m, "x")
	if cmd == nil {
		t.Fatal("x started no dependency check")
	}
	m = updateModel(m, cmd())
	if m.activeModal == nil || !strings.Contains(m.View().Content, "Uninstall tool?") {
		t.Fatalf("dialog does not name tool:\n%s", m.View().Content)
	}
	confirmAndRun(t, m)
	wantWrites(t, r, "uninstall user/tap/tool")
}

// TestCoreFormulaUpgradeWritesTheShortName: a formula whose full name is its
// name (homebrew/core) is written as before.
func TestCoreFormulaUpgradeWritesTheShortName(t *testing.T) {
	r := &searchUninstallRunner{installed: map[string]bool{}}
	m := outdatedFormulaModel(r, brew.Formula{Name: "bash", FullName: "bash"})

	m = sendKey(m, "u")
	confirmAndRun(t, m)
	wantWrites(t, r, "upgrade bash")
}
