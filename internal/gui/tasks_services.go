package gui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

func (m Model) serviceAction(action string) (tea.Model, tea.Cmd) {
	s := m.panels[PanelServices].selectedService()
	if s == nil {
		return m, nil
	}
	// The state decides: starting a running service, or stopping or restarting
	// one that is not loaded, would change nothing, so it says so instead.
	switch action {
	case "start":
		if serviceRunning(*s) {
			toast := m.setToast(s.Name+" is already running", modal.ToastInfo)
			return m, toast
		}
	case "stop", "restart":
		if !serviceStoppable(*s) {
			toast := m.setToast(s.Name+" is not running", modal.ToastInfo)
			return m, toast
		}
		return m.confirmServiceAction(action, s.Name)
	}
	return m.runServiceTask(action, s.Name)
}

// confirmServiceAction opens the confirm for stop or restart. The service is
// kept with the pending action, so the confirm acts on the one it was opened
// for.
func (m Model) confirmServiceAction(action, name string) (tea.Model, tea.Cmd) {
	m.pendingAction = "service-" + action
	m.pendingService = name
	if action == "stop" {
		m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Stop", "Stop "+name+"?").
			WithDetail("The service stops until you start it again.")
	} else {
		m.activeModal = modal.NewConfirmModal(modal.ActionUpgrade, "Restart", "Restart "+name+"?").
			WithDetail("The service stops and starts again.")
	}
	return m, m.activeModal.Init()
}

// isServiceTask is whether a task is a services write, which reloads the
// Services panel when it ends.
func isServiceTask(id string) bool {
	return strings.HasPrefix(id, "service-") || id == "services-cleanup"
}

// serviceReload asks for the Services list and its info again when a services
// task fails or is cancelled (AC-03.3). It is nil for every other task. A
// successful task reloads through RefreshMsg, as every task does.
func serviceReload(id string) tea.Cmd {
	if !isServiceTask(id) {
		return nil
	}
	return func() tea.Msg { return servicesReloadMsg{} }
}

// runServiceTask enqueues a services write for name through the task manager.
func (m Model) runServiceTask(action, name string) (tea.Model, tea.Cmd) {
	client := m.client
	t := errTask("service-"+action+":"+name, action+" "+name, func(ctx context.Context) error {
		switch action {
		case "start":
			return client.ServicesWrite.Start(ctx, name)
		case "stop":
			return client.ServicesWrite.Stop(ctx, name)
		case "restart":
			return client.ServicesWrite.Restart(ctx, name)
		case "run":
			return client.ServicesWrite.Run(ctx, name)
		default:
			// An unrecognized action must still resolve the task: an
			// empty errCh left the task "running" forever (r1 finding 16).
			return fmt.Errorf("unknown service action: %s", action)
		}
	})

	c := m.enqueue(t)
	return m, c
}

// serviceCleanup opens the confirm dialog; the actual cleanup runs through
// the task manager (executeServicesCleanup) so it is serialized with every
// other write op — running it directly inside the confirm callback used to
// bypass task.Manager (r1 finding 5).
func (m Model) serviceCleanup() (tea.Model, tea.Cmd) {
	m.pendingAction = "services-cleanup"
	m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Cleanup Services", "Remove stale service files?").
		WithDetail("Removes the service files of formulae that are no longer installed.")
	return m, m.activeModal.Init()
}

func (m Model) executeServicesCleanup() (tea.Model, tea.Cmd) {
	client := m.client
	t := errTask("services-cleanup", "Services cleanup", func(ctx context.Context) error {
		return client.ServicesWrite.Cleanup(ctx)
	})

	c := m.enqueue(t)
	return m, c
}
