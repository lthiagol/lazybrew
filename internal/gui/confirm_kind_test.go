package gui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

var sgrFrame = regexp.MustCompile("\x1b\\[[0-9;:]*m")

// confirmOf returns the confirm dialog m has open.
func confirmOf(t *testing.T, m *Model) *modal.ConfirmModal {
	t.Helper()
	cm, ok := m.activeModal.(*modal.ConfirmModal)
	if !ok {
		t.Fatalf("expected a confirm modal, got %T", m.activeModal)
	}
	return cm
}

// TestConfirmKindsByCallSite locks AC-01 and F-r1-02/F-r1-03: every confirm
// dialog is opened with the kind its call site is listed under, a chip that
// names the action, a consequence line, and a question that names the package
// or tap it acts on.
func TestConfirmKindsByCallSite(t *testing.T) {
	formulae := func(m *Model) {
		m.activePanel = PanelFormulae
		m.panels[PanelFormulae].items = []string{"foo  1.0"}
		m.panels[PanelFormulae].selected = 0
	}
	taps := func(m *Model) {
		m.activePanel = PanelTaps
		m.panels[PanelTaps].items = []string{"user/tap  3 formulae"}
		m.panels[PanelTaps].selected = 0
	}
	cases := []struct {
		name   string
		open   func(m *Model) tea.Model
		kind   modal.ActionKind
		action string
		object string
	}{
		{"install", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutInstall, "Install")
			return nm
		}, modal.ActionInstall, "Install", "foo"},
		{"upgrade", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutUpgrade, "Upgrade")
			return nm
		}, modal.ActionUpgrade, "Upgrade", "foo"},
		{"upgrade all", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutUpgradeAll, "Upgrade All")
			return nm
		}, modal.ActionUpgrade, "Upgrade All", ""},
		{"reinstall", func(m *Model) tea.Model {
			formulae(m)
			nm, _ := m.confirmMutation(mutReinstall, "Reinstall")
			return nm
		}, modal.ActionUpgrade, "Reinstall", "foo"},
		{"batch upgrade", func(m *Model) tea.Model {
			m.panels[PanelOutdated].items = []string{"foo  1.0"}
			m.batch.selected = map[int]bool{0: true}
			nm, _ := m.confirmBatchUpgrade()
			return nm
		}, modal.ActionUpgrade, "Batch Upgrade", ""},
		{"untap", func(m *Model) tea.Model {
			taps(m)
			nm, _ := m.confirmUntap()
			return nm
		}, modal.ActionDanger, "Untap", "user/tap"},
		{"repair", func(m *Model) tea.Model {
			taps(m)
			nm, _ := m.confirmRepair()
			return nm
		}, modal.ActionInstall, "Repair", "user/tap"},
		{"services cleanup", func(m *Model) tea.Model {
			nm, _ := m.serviceCleanup()
			return nm
		}, modal.ActionDanger, "Cleanup Services", ""},
		{"cleanup", func(m *Model) tea.Model {
			nm, _ := m.Update(TaskCompletedMsg{ID: "cleanup-preview", Title: "Cleanup"})
			return nm
		}, modal.ActionDanger, "Cleanup", ""},
		{"autoremove", func(m *Model) tea.Model {
			nm, _ := m.Update(TaskCompletedMsg{ID: "autoremove-preview", Title: "Autoremove"})
			return nm
		}, modal.ActionDanger, "Autoremove", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cm := confirmOf(t, asModel(tc.open(newTestModel())))
			if cm.Kind() != tc.kind {
				t.Fatalf("kind = %v, want %v", cm.Kind(), tc.kind)
			}
			if cm.Action() != tc.action {
				t.Fatalf("chip names %q, want %q", cm.Action(), tc.action)
			}
			// F-r1-02: every confirm says what will happen, not only the question.
			if cm.Detail() == "" {
				t.Fatal("the dialog has no consequence line under its question")
			}
			if view := sgrFrame.ReplaceAllString(cm.View().Content, ""); tc.object != "" && !strings.Contains(view, tc.object) {
				t.Fatalf("the question does not name %q:\n%s", tc.object, view)
			}
		})
	}
}

// TestUninstallConfirmIsDanger covers the uninstall and zap dialogs, which open
// only after the dependency check's DepCheckMsg comes back.
func TestUninstallConfirmIsDanger(t *testing.T) {
	cases := []struct {
		mut    mutationType
		action string
	}{
		{mutUninstall, "Uninstall"},
		{mutZap, "Zap"},
	}
	for _, tc := range cases {
		m := newTestModel()
		m.activePanel = PanelFormulae
		m.panels[PanelFormulae].items = []string{"ripgrep  14.1.1  bottled"}
		m.panels[PanelFormulae].selected = 0

		_, cmd := m.confirmUninstall(tc.mut)
		dMsg, ok := cmd().(DepCheckMsg)
		if !ok {
			t.Fatalf("mutation %v: expected DepCheckMsg", tc.mut)
		}
		if !strings.HasPrefix(dMsg.Detail, "removes ripgrep") {
			t.Fatalf("mutation %v: detail = %q, want it to name what goes", tc.mut, dMsg.Detail)
		}
		cm := confirmOf(t, updateModel(newTestModel(), dMsg))
		if cm.Kind() != modal.ActionDanger {
			t.Fatalf("mutation %v: kind = %v, want ActionDanger", tc.mut, cm.Kind())
		}
		if cm.Action() != tc.action {
			t.Fatalf("mutation %v: chip names %q, want %q", tc.mut, cm.Action(), tc.action)
		}
	}
}
