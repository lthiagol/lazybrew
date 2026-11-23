package gui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// beginUpdate marks the update running and queues it. The flag is set before
// the task is queued, so nothing that reads it sees a gap.
func (m *Model) beginUpdate() tea.Cmd {
	m.isUpdating = true
	m.updateOutput = nil
	return m.startBrewUpdate()
}

// startBrewUpdate enqueues `brew update` and starts it. It runs through
// DiagnosticsWrite.Update, so the update's cache group is dropped when it
// finishes. A dropped enqueue resets isUpdating and returns a warning toast.
func (m *Model) startBrewUpdate() tea.Cmd {
	client := m.client
	t := streamTask("brew-update", "brew update", func(ctx context.Context) (<-chan string, <-chan error) {
		return client.DiagnosticsWrite.Update(ctx)
	})
	// A dropped enqueue would leave isUpdating stuck true and block every
	// later R update, so surface the failure and reset the flag (r1 finding 7).
	c, err := m.tryEnqueue(t)
	if err != nil {
		m.isUpdating = false
	}
	return c
}

func (m *Model) executeBrewfileAction(index int) tea.Cmd {
	actions := []string{"dump", "install", "cleanup", "check", "list"}
	if index < 0 || index >= len(actions) {
		return nil
	}
	action := actions[index]
	title := "Brewfile " + action

	client := m.client
	t := streamTask("brewfile-"+action, title, func(ctx context.Context) (<-chan string, <-chan error) {
		return client.Runner.ExecuteStream(ctx, "bundle", action)
	})

	return m.enqueue(t)
}

func (m Model) brewfileMenu() (tea.Model, tea.Cmd) {
	menuItems := []string{
		"Export to Brewfile (brew bundle dump)",
		"Install from Brewfile (brew bundle install)",
		"Cleanup (brew bundle cleanup)",
		"Check (brew bundle check)",
		"List entries (brew bundle list)",
	}
	menuModal := modal.NewMenuModal("Brewfile", menuItems)
	m.pendingAction = "brewfile"
	m.activeModal = menuModal
	return m, menuModal.Init()
}

// openDiagnosticsTab switches the Status panel to the Doctor or Missing tab and
// loads it afresh. The diagnostics cache is dropped first, as R drops it, so the
// tab shows what brew says now.
func (m Model) openDiagnosticsTab(tab int) (tea.Model, tea.Cmd) {
	m.activeTab = tab
	m.invalidateCacheGroup("diagnostics")
	delete(m.tabContent, tabKey(PanelStatus, tab, statusItemName))
	cmd := m.loadTabContent()
	return m, cmd
}

// executeCleanup runs brew cleanup through the task manager, after its confirm.
func (m Model) executeCleanup() (tea.Model, tea.Cmd) {
	client := m.client
	t := streamTask("cleanup", "Cleanup", func(ctx context.Context) (<-chan string, <-chan error) {
		return client.DiagnosticsWrite.Cleanup(ctx)
	})

	c := m.enqueue(t)
	return m, c
}

// executeAutoremove runs brew autoremove through the task manager, after its confirm.
func (m Model) executeAutoremove() (tea.Model, tea.Cmd) {
	client := m.client
	t := streamTask("autoremove", "Autoremove", func(ctx context.Context) (<-chan string, <-chan error) {
		return client.DiagnosticsWrite.Autoremove(ctx)
	})

	c := m.enqueue(t)
	return m, c
}
