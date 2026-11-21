package gui

import (
	"context"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// uninstallConfirm opens the uninstall confirm for the selected row of panel
// and returns the dependency check it queues, with the brew calls it made.
func uninstallConfirm(t *testing.T, panel PanelID, mutType mutationType) (DepCheckMsg, [][]string) {
	t.Helper()
	calls := &[][]string{}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		*calls = append(*calls, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), testConfig())
	m.switchPanel(panel)
	p := m.panels[panel]
	p.loaded = true
	p.selected = 0
	switch panel {
	case PanelCasks:
		p.casks = []brew.Cask{{Name: "firefox"}}
		p.items = []string{"firefox"}
	case PanelFormulae:
		p.formulae = []brew.Formula{{Name: "wget"}}
		p.items = []string{"wget"}
	}

	_, cmd := m.confirmUninstall(mutType)
	msg, ok := cmd().(DepCheckMsg)
	if !ok {
		t.Fatalf("the confirm queued %T, want a DepCheckMsg", cmd())
	}
	return msg, *calls
}

// A cask has no dependents for brew uses to report, so its uninstall confirm
// does not ask brew (Q4). A formula's still does.
func TestCaskUninstallConfirmRecordsNoUsesCall(t *testing.T) {
	msg, calls := uninstallConfirm(t, PanelCasks, mutUninstall)
	for _, c := range calls {
		if len(c) > 0 && c[0] == "uses" {
			t.Errorf("cask uninstall confirm asked brew for dependents: %v", c)
		}
	}
	if msg.Detail != "removes firefox" {
		t.Errorf("cask detail = %q, want %q", msg.Detail, "removes firefox")
	}
}

func TestFormulaUninstallConfirmStillAsksForDependents(t *testing.T) {
	_, calls := uninstallConfirm(t, PanelFormulae, mutUninstall)
	for _, c := range calls {
		if len(c) > 0 && c[0] == "uses" {
			return
		}
	}
	t.Errorf("formula uninstall confirm made no uses call: %v", calls)
}
