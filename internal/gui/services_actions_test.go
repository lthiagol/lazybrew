package gui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
)

// servicesActionModel is a Services panel over a mock runner that records every
// brew call. The services writes these tests make only ever reach the mock.
func servicesActionModel(t *testing.T, services []brew.Service) (*Model, *[][]string) {
	t.Helper()
	calls := &[][]string{}
	r := brew.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		*calls = append(*calls, args)
		return []byte(""), nil
	}
	m := New(brew.NewClient(r), config.Default())
	m.switchPanel(PanelServices)
	p := m.panels[PanelServices]
	p.services = services
	p.items = make([]string, len(services))
	for i, s := range services {
		p.items[i] = presentation.FormatService(s, 0)
	}
	p.selected = 0
	return m, calls
}

// stopped and running are the two states the guards tell apart.
func runningService(name string) brew.Service {
	return brew.Service{Name: name, Status: brew.ServiceStarted, User: "agent"}
}

func stoppedService(name string) brew.Service {
	return brew.Service{Name: name, Status: brew.ServiceNone}
}

// lastToast is the message of the newest toast, or "" when there is none.
func lastToast(m *Model) string {
	if len(m.toasts) == 0 {
		return ""
	}
	return m.toasts[len(m.toasts)-1].Message
}

// noTaskQueued reports whether the task manager has nothing to start. RunNext
// returns nil on an empty queue.
func noTaskQueued(m *Model) bool {
	return m.tasks.RunNext() == nil
}

// confirmAndRun answers the open confirm with yes, as Update does (it clears the
// modal first), then runs the task the confirm queued through the task manager.
func confirmAndRun(t *testing.T, m *Model) Model {
	t.Helper()
	m.activeModal = nil
	res, cmd := m.handleModalResult(&modal.ConfirmResult{Confirmed: true}, nil)
	nm := res.(Model)
	if cmd == nil {
		t.Fatal("confirm queued no task")
	}
	if _, ok := cmd().(TaskStartedMsg); !ok {
		t.Fatal("expected TaskStartedMsg from the confirmed service task")
	}
	if comp, ok := nm.tasks.RunNext()().(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("expected a clean completion, got %+v", comp)
	}
	return nm
}

func TestServiceStartOnRunningSaysSoAndRunsNothing(t *testing.T) {
	m, _ := servicesActionModel(t, []brew.Service{runningService("moshi-hook")})

	m = sendKey(m, "s")
	if got := lastToast(m); got != "moshi-hook is already running" {
		t.Errorf("toast = %q, want %q", got, "moshi-hook is already running")
	}
	if m.activeModal != nil {
		t.Error("start on a running service opened a confirm")
	}
	if !noTaskQueued(m) {
		t.Error("start on a running service queued a task")
	}
}

func TestServiceStopOnStoppedSaysSoAndRunsNothing(t *testing.T) {
	for _, key := range []string{"S", "r"} {
		m, _ := servicesActionModel(t, []brew.Service{stoppedService("dbus")})

		m = sendKey(m, key)
		if got := lastToast(m); got != "dbus is not running" {
			t.Errorf("%s: toast = %q, want %q", key, got, "dbus is not running")
		}
		if m.activeModal != nil {
			t.Errorf("%s on a stopped service opened a confirm", key)
		}
		if !noTaskQueued(m) {
			t.Errorf("%s on a stopped service queued a task", key)
		}
	}
}

func TestServiceStopOpensDangerConfirm(t *testing.T) {
	m, _ := servicesActionModel(t, []brew.Service{runningService("moshi-hook")})

	m = sendKey(m, "S")
	c, ok := m.activeModal.(*modal.ConfirmModal)
	if !ok {
		t.Fatalf("S on a running service must open a confirm, got %T", m.activeModal)
	}
	if c.Kind() != modal.ActionDanger || c.Action() != "Stop" {
		t.Errorf("stop confirm kind %v action %q, want danger and Stop", c.Kind(), c.Action())
	}
	if c.Detail() != "The service stops until you start it again." {
		t.Errorf("stop consequence = %q", c.Detail())
	}
	if !noTaskQueued(m) {
		t.Error("the stop confirm queued a task before it was answered")
	}
}

func TestServiceRestartOpensUpgradeConfirm(t *testing.T) {
	m, _ := servicesActionModel(t, []brew.Service{runningService("moshi-hook")})

	m = sendKey(m, "r")
	c, ok := m.activeModal.(*modal.ConfirmModal)
	if !ok {
		t.Fatalf("r on a running service must open a confirm, got %T", m.activeModal)
	}
	if c.Kind() != modal.ActionUpgrade || c.Action() != "Restart" {
		t.Errorf("restart confirm kind %v action %q, want upgrade and Restart", c.Kind(), c.Action())
	}
	if c.Detail() != "The service stops and starts again." {
		t.Errorf("restart consequence = %q", c.Detail())
	}
}

// Confirming a stop runs `brew services stop` for the service the confirm was
// opened for, through the task manager.
func TestServiceStopConfirmedRunsStop(t *testing.T) {
	m, calls := servicesActionModel(t, []brew.Service{runningService("moshi-hook")})
	m = sendKey(m, "S")

	confirmAndRun(t, m)
	if !hasCall(*calls, "services", "stop", "moshi-hook") {
		t.Fatalf("expected `brew services stop moshi-hook`, got %v", *calls)
	}
}

func TestServiceRestartConfirmedRunsRestart(t *testing.T) {
	m, calls := servicesActionModel(t, []brew.Service{runningService("moshi-hook")})
	m = sendKey(m, "r")

	confirmAndRun(t, m)
	if !hasCall(*calls, "services", "restart", "moshi-hook") {
		t.Fatalf("expected `brew services restart moshi-hook`, got %v", *calls)
	}
}

// Start and run have no confirm: the key starts the task at once.
func TestServiceStartAndRunDoNotConfirm(t *testing.T) {
	for _, key := range []string{"s", "f"} {
		m, _ := servicesActionModel(t, []brew.Service{stoppedService("dbus")})

		nm, cmd := m.Update(tea.KeyPressMsg{Code: []rune(key)[0], Text: key})
		if cmd == nil {
			t.Fatalf("%s on a stopped service queued no task", key)
		}
		if _, ok := cmd().(TaskStartedMsg); !ok {
			t.Errorf("%s: expected TaskStartedMsg", key)
		}
		if nm.(Model).activeModal != nil {
			t.Errorf("%s opened a confirm", key)
		}
	}
}

// The reload after a services task ends is asked for by the task's ID, whatever
// its outcome; other tasks do not reload the Services panel.
func TestServiceTasksAskForReload(t *testing.T) {
	for _, id := range []string{"service-stop:moshi-hook", "service-restart:moshi-hook", "services-cleanup"} {
		if !isServiceTask(id) {
			t.Errorf("%q is a services task", id)
		}
		if msg, ok := serviceReload(id)().(RefreshMsg); !ok {
			t.Errorf("serviceReload(%q) = %T, want RefreshMsg", id, msg)
		}
	}
	if isServiceTask("install:bash") || serviceReload("install:bash") != nil {
		t.Error("a package task reloads the Services panel")
	}
}

// A failed services task still reloads the panel, so the list shows what is
// running now.
func TestServiceFailureStillReloads(t *testing.T) {
	m, _ := servicesActionModel(t, []brew.Service{runningService("moshi-hook")})

	_, cmd := m.Update(TaskCompletedMsg{ID: "service-stop:moshi-hook", Title: "stop moshi-hook", Err: errors.New("boom")})
	if !batchHasRefresh(cmd) {
		t.Error("a failed services task did not ask for a reload")
	}
}

// The cursor follows the service by name when the list is reloaded in another
// order, so a reorder never moves the selection to a different service.
func TestServicesSelectionKeepsNameAcrossReload(t *testing.T) {
	m, _ := servicesActionModel(t, []brew.Service{
		stoppedService("dbus"), runningService("moshi-hook"), stoppedService("pulseaudio"),
	})
	m.panels[PanelServices].selected = 1

	reordered := []brew.Service{stoppedService("pulseaudio"), runningService("moshi-hook"), stoppedService("dbus")}
	m = updateModel(m, dataLoaded(reordered))
	if got := m.panels[PanelServices].selectedService().Name; got != "moshi-hook" {
		t.Errorf("after reorder the cursor is on %q, want moshi-hook", got)
	}

	moved := []brew.Service{runningService("moshi-hook"), stoppedService("dbus"), stoppedService("pulseaudio")}
	m = updateModel(m, dataLoaded(moved))
	if got := m.panels[PanelServices].selectedService().Name; got != "moshi-hook" {
		t.Errorf("after second reorder the cursor is on %q, want moshi-hook", got)
	}
}

func dataLoaded(services []brew.Service) DataLoadedMsg {
	items := make([]string, len(services))
	for i, s := range services {
		items[i] = presentation.FormatService(s, 0)
	}
	return DataLoadedMsg{PanelID: PanelServices, Items: items, Services: services}
}

func hasCall(calls [][]string, args ...string) bool {
	for _, c := range calls {
		if strings.Join(c, " ") == strings.Join(args, " ") {
			return true
		}
	}
	return false
}

// batchHasRefresh runs the commands of a batch that finish at once and reports
// whether one of them is a RefreshMsg. A toast's timer is not waited for.
func batchHasRefresh(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		return false
	}
	for _, c := range batch {
		if c == nil {
			continue
		}
		done := make(chan tea.Msg, 1)
		go func() { done <- c() }()
		select {
		case msg := <-done:
			if _, ok := msg.(RefreshMsg); ok {
				return true
			}
		case <-time.After(200 * time.Millisecond):
		}
	}
	return false
}
