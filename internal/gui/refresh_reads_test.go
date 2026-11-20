package gui

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestRefreshReadsEachListOnce: one RefreshMsg, with the formulae, casks, taps
// and services panels loaded, runs each list command once in total. The status
// dashboard and the panels read the same lists at the same time; the reads
// must share one brew call per list. The list commands are held until every
// read is in flight, so the overlap is certain and not left to the scheduler.
func TestRefreshReadsEachListOnce(t *testing.T) {
	var mu sync.Mutex
	calls := map[string]int{}
	release := make(chan struct{})
	outputs := map[string]string{
		"info --json=v2 --installed":        `{"formulae":[]}`,
		"info --json=v2 --installed --cask": `{"casks":[]}`,
		"tap":                               "",
		"services list --json":              `[]`,
		"services info --all --json":        `[]`,
		"doctor":                            "Your system is ready to brew.",
		"config":                            "HOMEBREW_VERSION: 4.0.0",
	}
	holds := map[string]bool{
		"info --json=v2 --installed":        true,
		"info --json=v2 --installed --cask": true,
		"tap":                               true,
		"services list --json":              true,
	}
	runner := &brew.MockRunner{ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
		key := strings.Join(args, " ")
		mu.Lock()
		calls[key]++
		mu.Unlock()
		if holds[key] {
			<-release
		}
		return []byte(outputs[key]), nil
	}}

	m := New(brew.NewClient(runner), testConfig())
	m.activePanel = PanelFormulae
	for _, id := range []PanelID{PanelFormulae, PanelCasks, PanelTaps, PanelServices} {
		m.panels[id].items = []string{}
	}

	_, cmd := m.Update(RefreshMsg{})
	if cmd == nil {
		t.Fatal("RefreshMsg returned nil cmd")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("RefreshMsg cmd is not a batch")
	}

	var wg sync.WaitGroup
	for _, c := range batch {
		if c == nil {
			continue
		}
		wg.Add(1)
		go func(c tea.Cmd) {
			defer wg.Done()
			c()
		}(c)
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	for key := range holds {
		if calls[key] != 1 {
			t.Errorf("`brew %s` ran %d times for one refresh, want 1", key, calls[key])
		}
	}
}
