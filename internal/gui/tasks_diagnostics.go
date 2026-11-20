package gui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

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
	m.activeModal = menuModal
	return m, menuModal.Init()
}

func (m Model) runVulns() (tea.Model, tea.Cmd) {
	client := m.client
	t := linesTask("vulns", "Vulnerability Check", func(ctx context.Context) ([]string, error) {
		output, err := client.Diagnostics.Vulns(ctx)
		if err != nil {
			return nil, err
		}
		if output == "" {
			output = "No vulnerabilities found"
		}
		// The whole report is one line, as brew wrote it.
		return []string{output}, nil
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) runMissing() (tea.Model, tea.Cmd) {
	client := m.client
	t := linesTask("missing", "Missing Dependencies", func(ctx context.Context) ([]string, error) {
		missing, err := client.Diagnostics.Missing(ctx)
		if err != nil {
			return nil, err
		}
		if len(missing) == 0 {
			return []string{"All dependencies satisfied"}, nil
		}
		lines := make([]string, 0, len(missing))
		for _, d := range missing {
			lines = append(lines, d.Formula+": missing "+d.Missing)
		}
		return lines, nil
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) runDoctor() (tea.Model, tea.Cmd) {
	client := m.client
	t := linesTask("doctor", "Doctor", func(ctx context.Context) ([]string, error) {
		warnings, err := client.Diagnostics.Doctor(ctx)
		if err != nil {
			return nil, err
		}
		if len(warnings) == 0 {
			return []string{"Your system is ready to brew."}, nil
		}
		var lines []string
		for _, w := range warnings {
			lines = append(lines, w.Title)
			if w.Details != "" {
				lines = append(lines, "  "+w.Details)
			}
		}
		return lines, nil
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) brewCleanup() (tea.Model, tea.Cmd) {
	client := m.client
	t := streamTask("cleanup-preview", "Cleanup Preview", func(ctx context.Context) (<-chan string, <-chan error) {
		return client.DiagnosticsWrite.Cleanup(ctx, true)
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) runAutoremove() (tea.Model, tea.Cmd) {
	client := m.client
	t := streamTask("autoremove-preview", "Autoremove Preview", func(ctx context.Context) (<-chan string, <-chan error) {
		return client.DiagnosticsWrite.Autoremove(ctx, true)
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) executeCleanup() (tea.Model, tea.Cmd) {
	client := m.client
	t := streamTask("cleanup", "Cleanup", func(ctx context.Context) (<-chan string, <-chan error) {
		return client.DiagnosticsWrite.Cleanup(ctx, false)
	})

	c := m.enqueue(t)
	return m, c
}

func (m Model) executeAutoremove() (tea.Model, tea.Cmd) {
	client := m.client
	t := streamTask("autoremove", "Autoremove", func(ctx context.Context) (<-chan string, <-chan error) {
		return client.DiagnosticsWrite.Autoremove(ctx, false)
	})

	c := m.enqueue(t)
	return m, c
}
