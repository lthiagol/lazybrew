package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// errorService and scheduledService are loaded units that are not running:
// brew reports them as error and scheduled (formula_wrapper.rb, to_hash).
func errorService(name string) brew.Service {
	return brew.Service{Name: name, Status: brew.ServiceError, User: "agent"}
}

func scheduledService(name string) brew.Service {
	return brew.Service{Name: name, Status: brew.ServiceScheduled, User: "agent"}
}

// Stop and restart apply to a running service and to a loaded one that is not
// running; a stopped or unloaded one is refused (AC-03.1).
func TestServiceStoppableFollowsLoadedState(t *testing.T) {
	cases := map[brew.ServiceStatus]bool{
		brew.ServiceStarted:   true,
		brew.ServiceError:     true,
		brew.ServiceScheduled: true,
		brew.ServiceStopped:   false,
		brew.ServiceNone:      false,
	}
	for status, want := range cases {
		if got := serviceStoppable(brew.Service{Status: status}); got != want {
			t.Errorf("serviceStoppable(%q) = %v, want %v", status, got, want)
		}
	}
}

// S and r on a failed or scheduled service open their confirms, and confirming
// runs brew's stop or restart. Neither says "not running".
func TestStopAndRestartOnErrorAndScheduledConfirm(t *testing.T) {
	cases := []struct {
		name    string
		service brew.Service
	}{
		{"error", errorService("redis")},
		{"scheduled", scheduledService("backup")},
	}
	for _, c := range cases {
		for _, key := range []string{"S", "r"} {
			t.Run(c.name+"/"+key, func(t *testing.T) {
				m, calls := servicesActionModel(t, []brew.Service{c.service})

				m = sendKey(m, key)
				confirm, ok := m.activeModal.(*modal.ConfirmModal)
				if !ok {
					t.Fatalf("%s on a %s service must open a confirm, got %T", key, c.name, m.activeModal)
				}
				if key == "S" && (confirm.Action() != "Stop" || confirm.Kind() != modal.ActionDanger) {
					t.Errorf("stop confirm %q kind %v", confirm.Action(), confirm.Kind())
				}
				if key == "r" && (confirm.Action() != "Restart" || confirm.Kind() != modal.ActionUpgrade) {
					t.Errorf("restart confirm %q kind %v", confirm.Action(), confirm.Kind())
				}
				if strings.Contains(lastToast(m), "not running") {
					t.Errorf("%s on a %s service toasted %q", key, c.name, lastToast(m))
				}

				verb := map[string]string{"S": "stop", "r": "restart"}[key]
				confirmAndRun(t, m)
				if !hasCall(*calls, "services", verb, c.service.Name) {
					t.Errorf("expected `brew services %s %s`, got %v", verb, c.service.Name, *calls)
				}
			})
		}
	}
}

// s on a failed or scheduled service starts it at once: brew re-loads a loaded
// unit that is not running, so start is not refused there.
func TestStartOnErrorAndScheduledRunsWithoutConfirm(t *testing.T) {
	for _, service := range []brew.Service{errorService("redis"), scheduledService("backup")} {
		m, calls := servicesActionModel(t, []brew.Service{service})

		nm, cmd := m.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
		if cmd == nil {
			t.Fatalf("s on %s queued no task", service.Status)
		}
		if _, ok := cmd().(TaskStartedMsg); !ok {
			t.Fatalf("s on %s: expected TaskStartedMsg", service.Status)
		}
		if nm.(Model).activeModal != nil {
			t.Errorf("s on %s opened a confirm", service.Status)
		}
		nm.(Model).tasks.RunNext()()
		if !hasCall(*calls, "services", "start", service.Name) {
			t.Errorf("s on %s: expected `brew services start %s`, got %v", service.Status, service.Name, *calls)
		}
	}
}

// The hints follow the same state as the guards: a running service stops and
// restarts; a not-running one starts and runs; a loaded one that is not running
// also stops and restarts.
func TestServiceHintsFollowLoadedState(t *testing.T) {
	keys := func(s brew.Service) string {
		var out []string
		for _, h := range serviceActionHints(s) {
			out = append(out, h.key)
		}
		return strings.Join(out, " ")
	}
	cases := []struct {
		service brew.Service
		want    string
	}{
		{brew.Service{Status: brew.ServiceStarted}, "S r"},
		{brew.Service{Status: brew.ServiceNone}, "s f"},
		{brew.Service{Status: brew.ServiceStopped}, "s f"},
		{errorService("redis"), "s S r f"},
		{scheduledService("backup"), "s S r f"},
	}
	for _, c := range cases {
		if got := keys(c.service); got != c.want {
			t.Errorf("%s hints = %q, want %q", c.service.Status, got, c.want)
		}
	}
}

// The Info key line for a failed service offers start, stop and restart, the
// same set the footer leads with.
func TestServiceInfoKeyLineOffersStartStopRestartForError(t *testing.T) {
	services := []brew.Service{errorService("redis")}
	m := selectService(loadServices(newTestModel(), services, nil), "redis")
	text := infoText(m, 100)

	for _, want := range []string{"s start", "S stop", "r restart"} {
		if !strings.Contains(text, want) {
			t.Errorf("Info of a failed service lacks %q:\n%s", want, text)
		}
	}
	footer := m.footerHints().all()
	if footer[0].key != "s" || footer[1].key != "S" || footer[2].key != "r" {
		t.Errorf("footer for a failed service starts %q %q %q, want s S r",
			footer[0].key, footer[1].key, footer[2].key)
	}
}
