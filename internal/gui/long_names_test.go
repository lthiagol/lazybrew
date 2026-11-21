package gui

import (
	"context"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// Names longer than the 22-rune column the rows are padded to. The display
// string cuts them, so an action that read its name back from the row would
// send the cut name to brew.
const (
	longCaskName    = "font-jetbrains-mono-nerd-font"
	longFormulaName = "some-very-long-formula-name-here-ok"
	longTapName     = "someuser/homebrew-a-very-long-tap"
	longServiceName = "homebrew.mxcl.some-very-long-service-name"
)

// recordedModel is a model over a mock runner that records every brew call,
// on panel id with loading cleared so keys reach the panel.
func recordedModel(t *testing.T, id PanelID) (*Model, *[][]string) {
	t.Helper()
	calls := &[][]string{}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		*calls = append(*calls, args)
		return []byte("[]"), nil
	}
	// Installs, upgrades and uninstalls stream their output, so the streams
	// are recorded too. The mock's own closed streams stand in for the output.
	stream := brewtest.NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		*calls = append(*calls, args)
		return stream.ExecuteStream(ctx, args...)
	}
	m := New(brew.NewClient(r), testConfig())
	m.switchPanel(id)
	p := m.panels[id]
	p.loading = false
	p.loaded = true
	return m, calls
}

// anyArgIs reports whether some recorded brew call has an argument equal to
// name, so a cut name (a prefix of it) does not match.
func anyArgIs(calls [][]string, name string) bool {
	for _, c := range calls {
		if slices.Contains(c, name) {
			return true
		}
	}
	return false
}

// pressAndRun sends key, answers the confirm it leads to with yes, and runs
// the task through the task manager. x first reads what depends on the
// package and opens its confirm on the DepCheckMsg that comes back.
func pressAndRun(t *testing.T, m *Model, key string) {
	t.Helper()
	nm, cmd := pressKeyOn(m, key)
	var msg tea.Msg
	if cmd != nil {
		msg = cmd()
	}
	if dep, ok := msg.(DepCheckMsg); ok {
		nm = updateModel(nm, dep)
	}
	if nm.activeModal != nil {
		confirmAndRun(t, nm)
		return
	}
	if _, ok := msg.(TaskStartedMsg); !ok {
		t.Fatalf("key %s queued no task, got %T", key, msg)
	}
	if comp, ok := nm.tasks.RunNext()().(TaskCompletedMsg); !ok || comp.Err != nil {
		t.Fatalf("key %s: expected a clean completion, got %+v", key, comp)
	}
}

func TestLongFormulaNameReachesPinAndUninstall(t *testing.T) {
	for _, key := range []string{"p", "x"} {
		t.Run(key, func(t *testing.T) {
			m, calls := recordedModel(t, PanelFormulae)
			p := m.panels[PanelFormulae]
			p.formulae = []brew.Formula{{Name: longFormulaName}}
			p.items = []string{p.formulae[0].Name}
			p.selected = 0

			pressAndRun(t, m, key)
			if !anyArgIs(*calls, longFormulaName) {
				t.Fatalf("key %s: no brew call names %q, got %v", key, longFormulaName, *calls)
			}
		})
	}
}

func TestLongCaskNameReachesUninstallAndUpgrade(t *testing.T) {
	for _, key := range []string{"x", "u"} {
		t.Run(key, func(t *testing.T) {
			m, calls := recordedModel(t, PanelCasks)
			p := m.panels[PanelCasks]
			p.casks = []brew.Cask{{Name: longCaskName}}
			p.items = []string{p.casks[0].Name}
			p.selected = 0

			pressAndRun(t, m, key)
			if !anyArgIs(*calls, longCaskName) {
				t.Fatalf("key %s: no brew call names %q, got %v", key, longCaskName, *calls)
			}
		})
	}
}

func TestLongNamesReachUpgradeOnOutdated(t *testing.T) {
	m, calls := recordedModel(t, PanelOutdated)
	p := m.panels[PanelOutdated]
	p.casks = []brew.Cask{{Name: longCaskName}}
	p.items = []string{p.casks[0].Name}
	p.selected = 0

	pressAndRun(t, m, "u")
	if !anyArgIs(*calls, longCaskName) {
		t.Fatalf("upgrade on Outdated: no brew call names %q, got %v", longCaskName, *calls)
	}
}

func TestLongNameReachesOutdatedFormulaRow(t *testing.T) {
	m, calls := recordedModel(t, PanelOutdated)
	p := m.panels[PanelOutdated]
	p.formulae = []brew.Formula{{Name: longFormulaName}}
	p.items = []string{p.formulae[0].Name}
	p.selected = 0

	pressAndRun(t, m, "u")
	if !anyArgIs(*calls, longFormulaName) {
		t.Fatalf("upgrade on an Outdated formula: no brew call names %q, got %v", longFormulaName, *calls)
	}
}

func TestLongTapNameReachesUntapAndTrust(t *testing.T) {
	for _, tc := range []struct {
		key string
		tap brew.Tap
	}{
		{"x", brew.Tap{Name: longTapName, Installed: true, Trusted: true}},
		{"t", brew.Tap{Name: longTapName, Installed: true}},
	} {
		t.Run(tc.key, func(t *testing.T) {
			m, calls := tapsActionModel(t, []brew.Tap{tc.tap})
			p := m.panels[PanelTaps]
			p.items[0] = tc.tap.Name

			pressAndRun(t, m, tc.key)
			if !anyArgIs(*calls, longTapName) {
				t.Fatalf("key %s: no brew call names %q, got %v", tc.key, longTapName, *calls)
			}
		})
	}
}

func TestLongServiceNameReachesStop(t *testing.T) {
	m, calls := servicesActionModel(t, []brew.Service{runningService(longServiceName)})

	pressAndRun(t, m, "S")
	if !anyArgIs(*calls, longServiceName) {
		t.Fatalf("stop: no brew call names %q, got %v", longServiceName, *calls)
	}
}

func TestLongFormulaNameReachesTabFetch(t *testing.T) {
	m, calls := recordedModel(t, PanelFormulae)
	p := m.panels[PanelFormulae]
	p.formulae = []brew.Formula{{Name: longFormulaName}}
	p.items = []string{p.formulae[0].Name}
	p.selected = 0
	m.activeTab = formulaUsesTab

	drainCmds(m.loadTabContent())
	if !anyArgIs(*calls, longFormulaName) {
		t.Fatalf("Used By tab fetch: no brew call names %q, got %v", longFormulaName, *calls)
	}
}

// The batch selection names its rows by index; the names it sends must be the
// full ones too.
func TestLongCaskNameReachesBatchUpgrade(t *testing.T) {
	m, calls := recordedModel(t, PanelOutdated)
	p := m.panels[PanelOutdated]
	p.casks = []brew.Cask{{Name: longCaskName}}
	p.items = []string{p.casks[0].Name}
	m.batch.selected[0] = true

	pressAndRun(t, m, "u")
	if !anyArgIs(*calls, longCaskName) {
		t.Fatalf("batch upgrade: no brew call names %q, got %v", longCaskName, *calls)
	}
}

func TestLongCaskNameReachesSearchInstall(t *testing.T) {
	m, calls := recordedModel(t, PanelStatus)
	m.activePanel = PanelSearch
	p := m.panels[PanelSearch]
	m.searchResults = []brew.SearchResult{{Name: longCaskName, IsCask: true}}
	p.results = m.searchResults
	p.items = []string{longCaskName + "  cask"}

	// I installs from the result list; i would type into the search box.
	m = sendKey(m, "I")
	if m.activeModal == nil {
		t.Fatal("I on Search must open the install confirm")
	}
	confirmAndRun(t, m)
	if !anyArgIs(*calls, longCaskName) {
		t.Fatalf("install from Search: no brew call names %q, got %v", longCaskName, *calls)
	}
}
