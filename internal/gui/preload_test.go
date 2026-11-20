package gui

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
)

// preloadRunner is a mock runner that logs every brew read it answers. A read
// whose command line contains failArg fails.
func preloadRunner(calls *[]string, failArg string) *brew.MockRunner {
	read := func(args []string) ([]byte, error) {
		line := strings.Join(args, " ")
		*calls = append(*calls, line)
		if failArg != "" && strings.Contains(line, failArg) {
			return nil, errors.New("brew " + line + ": boom")
		}
		switch args[0] {
		case "tap":
			return []byte{}, nil
		case "services":
			return []byte("[]"), nil
		}
		return []byte("{}"), nil
	}
	return &brew.MockRunner{
		ExecuteFn: func(_ context.Context, args ...string) ([]byte, error) {
			return read(args)
		},
		ExecuteJSONFn: func(ctx context.Context, result any, args ...string) error {
			out, err := read(args)
			if err != nil {
				return err
			}
			return json.Unmarshal(out, result)
		},
	}
}

// preloadModel is a model with the preload on, over a runner that logs its reads.
func preloadModel(t *testing.T, cfg *config.Config, calls *[]string, failArg string) *Model {
	t.Helper()
	return New(brew.NewClient(preloadRunner(calls, failArg)), cfg)
}

// preloadConfig is the production default with the preload on: the tests here
// turn it on by hand.
func preloadConfig() *config.Config {
	cfg := config.Default()
	cfg.GUI.Preload = true
	return cfg
}

// preloadStep sends one message through Update and returns the model it leaves and
// the command it returns, whichever model type the handler returns.
func preloadStep(m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	nm, cmd := m.Update(msg)
	switch v := nm.(type) {
	case *Model:
		return v, cmd
	case Model:
		return &v, cmd
	}
	panic("step: unexpected model type")
}

// statusLoaded is the Status dashboard landing, the message that starts the
// preload.
func statusLoaded() tea.Msg {
	return DataLoadedMsg{PanelID: PanelStatus, Items: []string{}}
}

// landPreload runs the preload from the Status landing the way the program does:
// each read runs, its result goes back through Update, and the reads Update starts
// run next. It returns the panels in the order their reads landed, and fails
// if more than preloadWidth reads are ever in flight.
func landPreload(t *testing.T, m *Model, cmd tea.Cmd) (*Model, []PanelID) {
	t.Helper()
	pending := drainCmds(cmd)
	var order []PanelID
	for len(pending) > 0 {
		msg, ok := pending[0].(preloadedMsg)
		if !ok {
			t.Fatalf("pending message %T is not a preload read", pending[0])
		}
		pending = pending[1:]
		order = append(order, PanelID(msg.PanelID))
		var next tea.Cmd
		m, next = preloadStep(m, msg)
		pending = append(pending, drainCmds(next)...)
		if len(pending) > preloadWidth {
			t.Fatalf("%d preload reads in flight, want at most %d", len(pending), preloadWidth)
		}
		if m.preload.inFlight != len(pending) {
			t.Fatalf("inFlight = %d with %d reads pending", m.preload.inFlight, len(pending))
		}
	}
	return m, order
}

// TestPreloadOrderAndTwoInFlight covers AC-01.2 and AC-01.3: the panels load in
// the order Formulae, Outdated, Casks, Taps, Services, and never more than two
// reads are in flight at once.
func TestPreloadOrderAndTwoInFlight(t *testing.T) {
	var calls []string
	m := preloadModel(t, preloadConfig(), &calls, "")

	m, cmd := preloadStep(m, statusLoaded())
	if !m.preload.started {
		t.Fatal("the first Status load did not start the preload")
	}
	m, order := landPreload(t, m, cmd)

	want := []PanelID{PanelFormulae, PanelOutdated, PanelCasks, PanelTaps, PanelServices}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("preload order = %v, want %v", order, want)
	}
	for _, id := range want {
		if p := m.panels[id]; !p.loaded || p.loading {
			t.Errorf("%s: loaded=%v loading=%v after the preload, want loaded and not loading", id, p.loaded, p.loading)
		}
	}
	if m.preload.inFlight != 0 || len(m.preload.queue) != 0 {
		t.Errorf("preload left inFlight=%d queue=%v", m.preload.inFlight, m.preload.queue)
	}
}

// TestPreloadStartsWithTwoReads checks the cap by state, straight after the
// Status landing: two reads start, and the other three panels wait in order.
func TestPreloadStartsWithTwoReads(t *testing.T) {
	var calls []string
	m := preloadModel(t, preloadConfig(), &calls, "")

	m, cmd := preloadStep(m, statusLoaded())
	if got := m.preload.queue; !reflect.DeepEqual(got, []PanelID{PanelCasks, PanelTaps, PanelServices}) {
		t.Errorf("queue after the start = %v, want Casks, Taps, Services", got)
	}
	if m.preload.inFlight != 2 {
		t.Errorf("inFlight after the start = %d, want 2", m.preload.inFlight)
	}
	for _, id := range []PanelID{PanelFormulae, PanelOutdated, PanelCasks, PanelTaps, PanelServices} {
		if !m.panels[id].loading {
			t.Errorf("%s is not marked loading although its read is queued or in flight", id)
		}
	}
	if cmd == nil {
		t.Fatal("the Status landing returned no read")
	}
}

// TestPreloadSkipsLoadedPanels: a panel that has loaded before the Status
// landing is not read again.
func TestPreloadSkipsLoadedPanels(t *testing.T) {
	var calls []string
	m := preloadModel(t, preloadConfig(), &calls, "")
	m.panels[PanelCasks].loaded = true
	m.panels[PanelCasks].items = []string{}

	m, cmd := preloadStep(m, statusLoaded())
	if got := m.preload.queue; !reflect.DeepEqual(got, []PanelID{PanelTaps, PanelServices}) {
		t.Errorf("queue = %v, want Taps, Services: a loaded panel must not be queued", got)
	}
	m, order := landPreload(t, m, cmd)
	for _, id := range order {
		if id == PanelCasks {
			t.Error("Casks was read although it had loaded")
		}
	}
}

// TestPreloadDedupesWithAUserSwitch covers AC-01.4: a panel the user opens while
// it is queued is read once, and it is read next.
func TestPreloadDedupesWithAUserSwitch(t *testing.T) {
	var calls []string
	m := preloadModel(t, preloadConfig(), &calls, "")

	m, cmd := preloadStep(m, statusLoaded())
	pending := drainCmds(cmd)
	if len(pending) != 2 {
		t.Fatalf("%d reads started, want 2", len(pending))
	}

	// Taps is queued behind two reads in flight. The switch moves it to the
	// front of the queue; it starts no read of its own.
	m, switchCmd := preloadStep(m, tea.KeyPressMsg{Code: '5', Text: "5"})
	if switchCmd != nil && len(drainCmds(switchCmd)) != 0 {
		t.Error("switching to a queued panel started a read while the slots were full")
	}
	if got := m.preload.queue; len(got) == 0 || got[0] != PanelTaps {
		t.Errorf("queue after the switch = %v, want Taps first", got)
	}

	m, order := landPreloadFrom(t, m, pending)
	if n := countPanel(order, PanelTaps); n != 1 {
		t.Errorf("Taps was read %d times, want once", n)
	}
	if len(order) != 5 {
		t.Errorf("%d reads landed, want 5: %v", len(order), order)
	}
	if m.panels[PanelTaps].loading {
		t.Error("Taps is still marked loading after its read landed")
	}
}

// landPreloadFrom is landPreload for reads that have already started.
func landPreloadFrom(t *testing.T, m *Model, pending []tea.Msg) (*Model, []PanelID) {
	t.Helper()
	var order []PanelID
	for len(pending) > 0 {
		msg, ok := pending[0].(preloadedMsg)
		if !ok {
			t.Fatalf("pending message %T is not a preload read", pending[0])
		}
		pending = pending[1:]
		order = append(order, PanelID(msg.PanelID))
		var next tea.Cmd
		m, next = preloadStep(m, msg)
		pending = append(pending, drainCmds(next)...)
	}
	return m, order
}

func countPanel(order []PanelID, id PanelID) int {
	n := 0
	for _, o := range order {
		if o == id {
			n++
		}
	}
	return n
}

// TestPreloadFailedReadDoesNotStopTheOthers covers AC-02.2 at the scheduler:
// one panel's failed read leaves its error on that panel, and the reads after
// it still run.
func TestPreloadFailedReadDoesNotStopTheOthers(t *testing.T) {
	var calls []string
	m := preloadModel(t, preloadConfig(), &calls, "outdated")

	m, cmd := preloadStep(m, statusLoaded())
	m, order := landPreload(t, m, cmd)
	if len(order) != 5 {
		t.Errorf("%d reads landed, want all 5 after a failure: %v", len(order), order)
	}
	if m.panels[PanelOutdated].err == nil {
		t.Error("the failed Outdated read left no error on its panel")
	}
	for _, id := range []PanelID{PanelFormulae, PanelCasks, PanelTaps, PanelServices} {
		if p := m.panels[id]; p.err != nil || !p.loaded {
			t.Errorf("%s: err=%v loaded=%v, want loaded with no error", id, p.err, p.loaded)
		}
	}
}

// TestPreloadOffStaysLazy covers AC-01.5: with gui.preload off the Status
// landing starts no read, and a switch to a panel reads it, as it always has.
func TestPreloadOffStaysLazy(t *testing.T) {
	var calls []string
	cfg := preloadConfig()
	cfg.GUI.Preload = false
	m := preloadModel(t, cfg, &calls, "")

	m, cmd := preloadStep(m, statusLoaded())
	if cmd != nil {
		t.Errorf("the Status landing returned a command with the preload off: %v", drainCmds(cmd))
	}
	if m.preload.started || len(m.preload.queue) != 0 || m.preload.inFlight != 0 {
		t.Errorf("preload state changed with the preload off: %+v", m.preload)
	}
	for _, id := range preloadOrder {
		if m.panels[id].loading || m.panels[id].loaded {
			t.Errorf("%s was touched by a preload that is off", id)
		}
	}

	m, cmd = preloadStep(m, tea.KeyPressMsg{Code: '2', Text: "2"})
	if cmd == nil {
		t.Fatal("switching to Formulae with the preload off started no read")
	}
	if !m.panels[PanelFormulae].loading {
		t.Error("Formulae is not loading after the switch")
	}
	if n := len(drainCmds(cmd)); n != 1 {
		t.Errorf("the switch started %d reads, want 1", n)
	}
}

// TestPreloadWaitsForUpdateOnStart covers AC-01.7 for update_on_start: the
// preload starts after the brew update, once the Status load that follows it
// lands, and not before.
func TestPreloadWaitsForUpdateOnStart(t *testing.T) {
	var calls []string
	cfg := preloadConfig()
	cfg.Brew.UpdateOnStart = true
	m := preloadModel(t, cfg, &calls, "")

	m, _ = preloadStep(m, TaskCompletedMsg{ID: "brew-update"})
	if m.preload.started || m.preload.inFlight != 0 || len(m.preload.queue) != 0 {
		t.Fatalf("the preload started before Status landed: %+v", m.preload)
	}
	for _, id := range preloadOrder {
		if m.panels[id].loading {
			t.Errorf("%s is loading before Status landed", id)
		}
	}

	m, cmd := preloadStep(m, statusLoaded())
	if !m.preload.started || m.preload.inFlight != 2 {
		t.Errorf("after Status landed: started=%v inFlight=%d, want started with 2 reads",
			m.preload.started, m.preload.inFlight)
	}
	m, order := landPreload(t, m, cmd)
	if len(order) != 5 {
		t.Errorf("%d reads landed after update_on_start, want 5", len(order))
	}
}

// TestPreloadLandingDoesNotCountAsRefresh: a refresh counts the DataLoadedMsg it
// waits for. A preload read landing in the middle of a refresh must not count.
func TestPreloadLandingDoesNotCountAsRefresh(t *testing.T) {
	var calls []string
	m := preloadModel(t, preloadConfig(), &calls, "")
	m.refreshing = 1

	m, _ = preloadStep(m, preloadedMsg{PanelID: PanelFormulae, Items: []string{}})
	if m.refreshing != 1 {
		t.Errorf("refreshing = %d after a preload landing, want 1", m.refreshing)
	}
	m, _ = preloadStep(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{}})
	if m.refreshing != 0 {
		t.Errorf("refreshing = %d after the refresh's own landing, want 0", m.refreshing)
	}
}

// preloadScreen is a model at 140x40 that has its Status dashboard landed and
// the preload running, with the first two reads started and not yet landed.
func preloadScreen(t *testing.T, cfg *config.Config, calls *[]string, failArg string) (*Model, []tea.Msg) {
	t.Helper()
	m := preloadModel(t, cfg, calls, failArg)
	m, _ = preloadStep(m, tea.WindowSizeMsg{Width: 140, Height: 40})
	m, cmd := preloadStep(m, statusLoaded())
	return m, drainCmds(cmd)
}

// TestPreloadKeysWorkMidPreload covers AC-02.1: keys work while panels preload.
// Casks is queued and loading; a key moves to it at once, and the rest of the
// keys still do what they do.
func TestPreloadKeysWorkMidPreload(t *testing.T) {
	var calls []string
	m, pending := preloadScreen(t, preloadConfig(), &calls, "")

	m, _ = preloadStep(m, tea.KeyPressMsg{Code: '3', Text: "3"})
	if m.activePanel != PanelCasks {
		t.Fatalf("key 3 left the panel at %v, want Casks", m.activePanel)
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Loading...") {
		t.Errorf("Casks is loading, but its main area does not say so:\n%s", view)
	}

	m, _ = preloadStep(m, tea.KeyPressMsg{Code: '?', Text: "?"})
	if !m.showHelp {
		t.Fatal("? did not open help while the preload was running")
	}
	m, _ = preloadStep(m, tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.showHelp {
		t.Fatal("Esc did not close help while the preload was running")
	}

	m, _ = landPreloadFrom(t, m, pending)
	if view := ansi.Strip(m.View().Content); strings.Contains(view, "Loading...") {
		t.Errorf("Casks has landed but its main area still says Loading...:\n%s", view)
	}
}

// TestPreloadMouseWorksMidPreload covers AC-02.1 for the mouse: a click on the
// Services box while the preload runs switches to Services at once.
func TestPreloadMouseWorksMidPreload(t *testing.T) {
	var calls []string
	cfg := preloadConfig()
	m, pending := preloadScreen(t, cfg, &calls, "")
	m.cfg.GUI.Mouse = true

	heights := m.computeContentHeights()
	y := 1
	for i := 0; i < int(PanelServices); i++ {
		y += heights[i] + 2
	}
	m, _ = preloadStep(m, tea.MouseClickMsg{X: 2, Y: y, Button: tea.MouseLeft})
	if m.activePanel != PanelServices {
		t.Fatalf("click at row %d left the panel at %v, want Services (heights=%v)", y, m.activePanel, heights)
	}
	m, _ = landPreloadFrom(t, m, pending)
	if !m.panels[PanelServices].loaded {
		t.Error("Services did not load after the click")
	}
}

// TestPreloadFailedReadShowsItsErrorOnItsPanelOnly covers AC-02.2: a failed
// read shows its error on its own panel, and the other panels still load.
func TestPreloadFailedReadShowsItsErrorOnItsPanelOnly(t *testing.T) {
	var calls []string
	m, pending := preloadScreen(t, preloadConfig(), &calls, "outdated")
	m, _ = landPreloadFrom(t, m, pending)

	if dash := ansi.Strip(m.renderDashboard(140)); strings.Contains(dash, "boom") {
		t.Errorf("the failed Outdated read leaked onto the dashboard:\n%s", dash)
	}
	// Switching to the failed panel retries its read, as it always has; the
	// retry's failure is what the panel then shows.
	m, cmd := preloadStep(m, tea.KeyPressMsg{Code: '4', Text: "4"})
	for _, msg := range drainCmds(cmd) {
		m, _ = preloadStep(m, msg)
	}
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "boom") {
		t.Errorf("the Outdated panel does not show its error:\n%s", view)
	}
	if !m.panels[PanelTaps].loaded || m.panels[PanelTaps].err != nil {
		t.Error("Taps did not load after the Outdated read failed")
	}
}

// TestPreloadFillsTheDashboardCounts covers AC-02.3: the Outdated count is not
// "not loaded — press R" while preloaded; it reads loading, then the count.
func TestPreloadFillsTheDashboardCounts(t *testing.T) {
	var calls []string
	m, pending := preloadScreen(t, preloadConfig(), &calls, "")

	dash := ansi.Strip(m.renderDashboard(140))
	if !strings.Contains(dash, "loading…") || strings.Contains(dash, "press R") {
		t.Errorf("while Outdated is queued the dashboard should say loading…, got:\n%s", dash)
	}

	m, _ = landPreloadFrom(t, m, pending)
	dash = ansi.Strip(m.renderDashboard(140))
	if strings.Contains(dash, "not loaded") || strings.Contains(dash, "loading…") {
		t.Errorf("after the preload the dashboard should show every count, got:\n%s", dash)
	}
	if !strings.Contains(dash, "0 packages") {
		t.Errorf("the Outdated count did not fill in, got:\n%s", dash)
	}
	if !m.status.Outdated.Known {
		t.Error("the status summary's Outdated count is still unknown after its read landed")
	}
}

// TestPreloadOffDashboardSaysPressR covers AC-02.3 with the preload off: the
// Outdated count is not loaded, and R is still how to load it.
func TestPreloadOffDashboardSaysPressR(t *testing.T) {
	var calls []string
	cfg := preloadConfig()
	cfg.GUI.Preload = false
	m := preloadModel(t, cfg, &calls, "")
	m, _ = preloadStep(m, statusLoaded())
	dash := ansi.Strip(m.renderDashboard(140))
	if !strings.Contains(dash, "not loaded — press R") {
		t.Errorf("with the preload off the Outdated count should say press R, got:\n%s", dash)
	}
}

// TestProductionDefaultPreloads pins the production default: the background
// preload is on unless gui.preload or LAZYBREW_PRELOAD turns it off.
func TestProductionDefaultPreloads(t *testing.T) {
	if !config.Default().GUI.Preload {
		t.Fatal("production default gui.preload = false, want true")
	}
}
