package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// confirmOf returns the confirm dialog m has open.
func confirmOf(t *testing.T, m *Model) *modal.ConfirmModal {
	t.Helper()
	cm, ok := m.activeModal.(*modal.ConfirmModal)
	if !ok {
		t.Fatalf("expected a confirm modal, got %T", m.activeModal)
	}
	return cm
}

// TestConfirmKindsByCallSite locks AC-01: every confirm dialog is opened with
// the action kind its call site is listed under in the v2-04 receipt.
func TestConfirmKindsByCallSite(t *testing.T) {
	formulae := func(m *Model) {
		m.activePanel = PanelFormulae
		m.panels[PanelFormulae].items = []string{"foo  1.0"}
		m.panels[PanelFormulae].selected = 0
	}
	cases := []struct {
		name string
		open func(m *Model) tea.Model
		want modal.ActionKind
	}{
		{"install", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutInstall, "Install")
			return nm
		}, modal.ActionInstall},
		{"upgrade", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutUpgrade, "Upgrade")
			return nm
		}, modal.ActionUpgrade},
		{"upgrade all", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutUpgradeAll, "Upgrade All")
			return nm
		}, modal.ActionUpgrade},
		{"reinstall", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutReinstall, "Reinstall")
			return nm
		}, modal.ActionUpgrade},
		{"batch upgrade", func(m *Model) tea.Model {
			m.panels[PanelOutdated].items = []string{"foo  1.0"}
			m.batch.selected = map[int]bool{0: true}
			nm, _ := m.confirmBatchUpgrade()
			return nm
		}, modal.ActionUpgrade},
		{"untap", func(m *Model) tea.Model {
			m.activePanel = PanelTaps
			m.panels[PanelTaps].items = []string{"user/tap  3 formulae"}
			m.panels[PanelTaps].selected = 0
			nm, _ := m.confirmUntap()
			return nm
		}, modal.ActionDanger},
		{"repair", func(m *Model) tea.Model {
			m.activePanel = PanelTaps
			m.panels[PanelTaps].items = []string{"user/tap  3 formulae"}
			m.panels[PanelTaps].selected = 0
			nm, _ := m.confirmRepair()
			return nm
		}, modal.ActionInstall},
		{"services cleanup", func(m *Model) tea.Model {
			nm, _ := m.serviceCleanup()
			return nm
		}, modal.ActionDanger},
		{"cleanup", func(m *Model) tea.Model {
			nm, _ := m.Update(TaskCompletedMsg{ID: "cleanup-preview", Title: "Cleanup"})
			return nm
		}, modal.ActionDanger},
		{"autoremove", func(m *Model) tea.Model {
			nm, _ := m.Update(TaskCompletedMsg{ID: "autoremove-preview", Title: "Autoremove"})
			return nm
		}, modal.ActionDanger},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cm := confirmOf(t, asModel(tc.open(newTestModel())))
			if cm.Kind() != tc.want {
				t.Fatalf("kind = %v, want %v", cm.Kind(), tc.want)
			}
			// F-r1-02: every confirm says what will happen, not only the question.
			if cm.Detail() == "" {
				t.Fatal("the dialog has no consequence line under its question")
			}
		})
	}
}

// TestUninstallConfirmIsDanger covers the uninstall and zap dialogs, which open
// only after the dependency check's DepCheckMsg comes back.
func TestUninstallConfirmIsDanger(t *testing.T) {
	for _, mut := range []mutationType{mutUninstall, mutZap} {
		m := newTestModel()
		m.activePanel = PanelFormulae
		m.panels[PanelFormulae].items = []string{"ripgrep  14.1.1  bottled"}
		m.panels[PanelFormulae].selected = 0

		_, cmd := m.confirmUninstall(mut)
		dMsg, ok := cmd().(DepCheckMsg)
		if !ok {
			t.Fatalf("mutation %v: expected DepCheckMsg", mut)
		}
		if !strings.HasPrefix(dMsg.Detail, "removes ripgrep") {
			t.Fatalf("mutation %v: detail = %q, want it to name what goes", mut, dMsg.Detail)
		}
		nm := updateModel(newTestModel(), dMsg)
		if got := confirmOf(t, nm).Kind(); got != modal.ActionDanger {
			t.Fatalf("mutation %v: kind = %v, want ActionDanger", mut, got)
		}
	}
}
