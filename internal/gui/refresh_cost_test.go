package gui

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// TestAutomaticRefreshKeepsDiagnosticsCache: brew doctor and brew config are
// the slowest reads. A refresh after a task or on the auto-refresh tick reuses
// them; only the R key runs them again.
func TestAutomaticRefreshKeepsDiagnosticsCache(t *testing.T) {
	var mu sync.Mutex
	calls := map[string]int{}
	runner := &brewtest.MockRunner{ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
		key := strings.Join(args, " ")
		mu.Lock()
		calls[key]++
		mu.Unlock()
		switch key {
		case "doctor":
			return []byte("Your system is ready to brew."), nil
		case "config":
			return []byte("HOMEBREW_VERSION: 4.0.0"), nil
		}
		return []byte{}, nil
	}}
	client := brew.NewClient(runner)
	client.Cache.SetWithTTL(brew.KeyDoctorResult, []brew.DoctorWarning{}, time.Minute)
	client.Cache.SetWithTTL(brew.KeyConfig, &brew.BrewConfig{}, time.Minute)

	m := New(client, testConfig())
	m.activePanel = PanelStatus

	_, cmd := m.Update(RefreshMsg{})
	drainCmds(cmd)
	mu.Lock()
	if calls["doctor"] != 0 || calls["config"] != 0 {
		t.Errorf("an automatic refresh ran doctor %d times and config %d times, want 0 and 0", calls["doctor"], calls["config"])
	}
	mu.Unlock()

	// The R key is the explicit request: it reloads diagnostics too.
	_, keyCmd := m.Update(tea.KeyPressMsg{Code: 'R', Text: "R"})
	refreshes := collectRefreshMsgs(keyCmd)
	if len(refreshes) != 1 {
		t.Fatalf("R produced %d RefreshMsg, want 1", len(refreshes))
	}
	_, cmd = m.Update(refreshes[0])
	drainCmds(cmd)
	mu.Lock()
	defer mu.Unlock()
	if calls["doctor"] != 1 || calls["config"] != 1 {
		t.Errorf("R ran doctor %d times and config %d times, want 1 and 1", calls["doctor"], calls["config"])
	}
}

// slowRunner returns canned output for each command after d, and counts the
// commands it ran.
func slowRunner(d time.Duration, outputs map[string]string) *brewtest.MockRunner {
	return &brewtest.MockRunner{ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
		time.Sleep(d)
		return []byte(outputs[strings.Join(args, " ")]), nil
	}}
}

// TestOutdatedPanelReadsConcurrently: the formula and cask outdated lists are
// read at the same time. Two sequential reads of d each take 2d; concurrent
// reads take about d.
func TestOutdatedPanelReadsConcurrently(t *testing.T) {
	const d = 300 * time.Millisecond
	runner := slowRunner(d, map[string]string{
		"outdated --json=v2 --formula": `{"formulae":[]}`,
		"outdated --json=v2 --cask":    `{"casks":[]}`,
	})
	start := time.Now()
	fetchPanelData(brew.NewClient(runner), PanelOutdated)()
	if took := time.Since(start); took >= 2*d {
		t.Errorf("the outdated panel took %v, want under %v (the two reads should overlap)", took, 2*d)
	}
}

// TestServicesPanelReadsConcurrently: the services list and the services info
// are read at the same time.
func TestServicesPanelReadsConcurrently(t *testing.T) {
	const d = 300 * time.Millisecond
	runner := slowRunner(d, map[string]string{
		"services list --json":       `[]`,
		"services info --all --json": `[]`,
	})
	start := time.Now()
	fetchPanelData(brew.NewClient(runner), PanelServices)()
	if took := time.Since(start); took >= 2*d {
		t.Errorf("the services panel took %v, want under %v (the two reads should overlap)", took, 2*d)
	}
}

// TestInvalidatePanelCacheDropsItsKeys: each panel drops the cache keys its
// brew group names, and no other key. The status dashboard keeps the outdated
// entries, and its diagnostics come from the diagnostics group on R only.
func TestInvalidatePanelCacheDropsItsKeys(t *testing.T) {
	all := []brew.CacheKey{
		brew.KeyFormulaeList, brew.KeyCasksList,
		brew.KeyOutdatedFormulae, brew.KeyOutdatedCasks,
		brew.KeyTapsList, brew.KeyServicesList, brew.KeyServicesInfo,
		brew.KeyDoctorResult, brew.KeyConfig,
	}
	cases := []struct {
		panel PanelID
		keys  []brew.CacheKey
	}{
		{PanelFormulae, []brew.CacheKey{brew.KeyFormulaeList}},
		{PanelCasks, []brew.CacheKey{brew.KeyCasksList}},
		{PanelOutdated, []brew.CacheKey{brew.KeyOutdatedFormulae, brew.KeyOutdatedCasks}},
		{PanelTaps, []brew.CacheKey{brew.KeyTapsList}},
		{PanelServices, []brew.CacheKey{brew.KeyServicesList, brew.KeyServicesInfo}},
		{PanelStatus, []brew.CacheKey{brew.KeyFormulaeList, brew.KeyCasksList, brew.KeyTapsList, brew.KeyServicesList}},
	}
	for _, tc := range cases {
		t.Run(tc.panel.String(), func(t *testing.T) {
			client := brew.NewClient(brewtest.NewMockRunner())
			for _, key := range all {
				client.Cache.SetWithTTL(key, "seeded", time.Minute)
			}
			m := New(client, testConfig())
			m.invalidatePanelCache(tc.panel)

			dropped := map[brew.CacheKey]bool{}
			for _, key := range tc.keys {
				dropped[key] = true
			}
			for _, key := range all {
				_, cached := client.Cache.Get(key)
				if dropped[key] && cached {
					t.Errorf("%s: %s still cached", tc.panel, key)
				}
				if !dropped[key] && !cached {
					t.Errorf("%s: %s was dropped but is not this panel's", tc.panel, key)
				}
			}
		})
	}
}
