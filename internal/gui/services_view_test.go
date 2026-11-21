package gui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// servicesList and servicesInfo are the Services panel's two reads for three
// services: dbus (stopped, nothing registered), moshi-hook (running) and backup
// (scheduled, with an interval and a cron schedule).
func servicesList() []brew.Service {
	return []brew.Service{
		{Name: "dbus", Status: brew.ServiceNone},
		{Name: "moshi-hook", Status: brew.ServiceStarted, User: "agent", File: "/x/moshi-hook.service"},
		{Name: "backup", Status: brew.ServiceScheduled, User: "agent", File: "/x/backup.service"},
	}
}

func servicesInfo() []brew.ServiceInfo {
	pid := 800
	zero := 0
	interval := 3600
	return []brew.ServiceInfo{
		{Name: "dbus", ServiceName: "sh.brew.dbus", Status: brew.ServiceNone, File: "/x/dbus.service"},
		{
			Name: "moshi-hook", ServiceName: "sh.brew.moshi-hook", Running: true, Loaded: true, Registered: true,
			PID: &pid, User: "agent", Status: brew.ServiceStarted, File: "/x/moshi-hook.service",
			Command: "moshi-hook serve", WorkingDir: "/home/linuxbrew/.linuxbrew/var",
			LogPath:      "/home/linuxbrew/.linuxbrew/var/log/moshi-hook.log",
			ErrorLogPath: "/home/linuxbrew/.linuxbrew/var/log/moshi-hook.log",
		},
		{
			Name: "backup", ServiceName: "sh.brew.backup", Loaded: true, Registered: true, Schedulable: true,
			ExitCode: &zero, User: "agent", Status: brew.ServiceScheduled, File: "/x/backup.service",
			Command: "/bin/backup", Interval: &interval, Cron: map[string]any{"hour": 3.0, "minute": 0.0},
		},
	}
}

// loadServices gives a sized model a loaded Services panel and opens it.
func loadServices(m *Model, services []brew.Service, info []brew.ServiceInfo) *Model {
	p := m.panels[PanelServices]
	p.services = services
	p.serviceInfo = info
	p.items = make([]string, len(services))
	for i, s := range services {
		p.items[i] = s.Name
	}
	p.loaded = true
	return sendKey(m, "6")
}

// selectService moves the Services cursor to the named service.
func selectService(m *Model, name string) *Model {
	for i, s := range m.panels[PanelServices].services {
		if s.Name == name {
			m.panels[PanelServices].selected = i
		}
	}
	return m
}

// infoText is the Info tab for the selected service, as plain text.
func infoText(m *Model, width int) string {
	return ansi.Strip(renderServiceInfo(m.panels[PanelServices], width, style.UnicodeIcons))
}

func TestServicesTabIsInfo(t *testing.T) {
	if got := panelTabs[PanelServices][0].name; got != "Info" {
		t.Errorf("Services tab = %q, want Info", got)
	}
}

func TestServiceInfoShowsDetailAndRunningActions(t *testing.T) {
	m := selectService(loadServices(newTestModel(), servicesList(), servicesInfo()), "moshi-hook")
	text := infoText(m, 100)

	for _, want := range []string{"running", "PID", "800", "User", "agent", "Loaded", "yes",
		"Unit file", "Command", "moshi-hook serve", "Working dir", "Log", "/home/linuxbrew/.linuxbrew/var/log/moshi-"} {
		if !strings.Contains(text, want) {
			t.Errorf("Info of a running service lacks %q:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "S stop") || !strings.Contains(text, "r restart") {
		t.Errorf("a running service's key line lacks S stop and r restart:\n%s", text)
	}
	if strings.Contains(text, "s start") {
		t.Errorf("a running service offers start:\n%s", text)
	}
	// At 100 cells the log path is longer than the right column, so the view
	// keeps its single column and the path is whole on its row (F-r4-01).
	if !strings.Contains(text, "/home/linuxbrew/.linuxbrew/var/log/moshi-hook.log") {
		t.Errorf("the log path should be whole on one row:\n%s", text)
	}
	if strings.Count(text, "Log") != 1 {
		t.Errorf("identical log and error log should show one Log row:\n%s", text)
	}
}

func TestServiceInfoStoppedOffersStartAndRun(t *testing.T) {
	m := selectService(loadServices(newTestModel(), servicesList(), servicesInfo()), "dbus")
	text := infoText(m, 100)

	if !strings.Contains(text, "s start") || !strings.Contains(text, "f run") {
		t.Errorf("a stopped service's key line lacks s start and f run:\n%s", text)
	}
	if strings.Contains(text, "S stop") || strings.Contains(text, "r restart") {
		t.Errorf("a stopped service offers stop or restart:\n%s", text)
	}
}

// Null fields in brew's info are omitted, not drawn as empty rows.
func TestServiceInfoOmitsNullFields(t *testing.T) {
	m := selectService(loadServices(newTestModel(), servicesList(), servicesInfo()), "dbus")
	text := infoText(m, 100)

	for _, gone := range []string{"PID", "Exit code", "Command", "Working dir", "Log", "Interval", "Cron"} {
		if strings.Contains(text, gone) {
			t.Errorf("a null %s was drawn:\n%s", gone, text)
		}
	}
}

func TestServiceInfoShowsSchedule(t *testing.T) {
	m := selectService(loadServices(newTestModel(), servicesList(), servicesInfo()), "backup")
	text := infoText(m, 100)

	for _, want := range []string{"scheduled", "every 3600s", "hour 3, minute 0", "Exit code", "0"} {
		if !strings.Contains(text, want) {
			t.Errorf("Info of a scheduled service lacks %q:\n%s", want, text)
		}
	}
}

// Before `brew services info` has loaded, the Info tab shows the list's fields,
// never a blank pane.
func TestServiceInfoFallsBackToListBeforeInfoLoads(t *testing.T) {
	m := selectService(loadServices(newTestModel(), servicesList(), nil), "moshi-hook")
	text := infoText(m, 100)

	for _, want := range []string{"running", "User", "agent", "File", "Exit code", "S stop", "r restart"} {
		if !strings.Contains(text, want) {
			t.Errorf("the list fallback lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "PID") {
		t.Errorf("the list fallback drew info-only fields:\n%s", text)
	}
}

func TestServicesFooterLeadsWithStateActions(t *testing.T) {
	m := loadServices(newTestModel(), servicesList(), servicesInfo())

	selectService(m, "moshi-hook")
	running := m.footerHints()
	if running[0].key != "S" || running[1].key != "r" {
		t.Errorf("running footer starts %q %q, want S and r", running[0].key, running[1].key)
	}

	selectService(m, "dbus")
	stopped := m.footerHints()
	if stopped[0].key != "s" || stopped[1].key != "f" {
		t.Errorf("stopped footer starts %q %q, want s and f", stopped[0].key, stopped[1].key)
	}
	for _, h := range stopped {
		if h.key == "S" || h.key == "r" {
			t.Errorf("stopped footer offers %q", h.key)
		}
	}
}

// The footer and the Info key line are both made from the same state, so a
// service's state cannot show different actions in the two places.
func TestServicesFooterAndInfoAgreeOnActions(t *testing.T) {
	m := loadServices(newTestModel(), servicesList(), servicesInfo())
	for _, name := range []string{"dbus", "moshi-hook", "backup"} {
		selectService(m, name)
		text := infoText(m, 100)
		for _, h := range servicesHints(m.panels[PanelServices]) {
			// j/k, h/l and c are keys of the panel, not of the service's state.
			if h.key == "j/k" || h.key == "h/l" || h.key == "c" {
				continue
			}
			if !strings.Contains(text, h.key+" "+h.desc) {
				t.Errorf("%s: footer offers %q %q, the Info key line lacks it:\n%s", name, h.key, h.desc, text)
			}
		}
	}
}
