package gui

import (
	"context"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestSlowRemarkCannotRestoreAStaleMark (LB-20 b): two uninstalls finish back to
// back, so two re-marks are issued. The first read the cask list before the
// second uninstall and is delivered last. It must not restore the mark the
// second re-mark cleared.
func TestSlowRemarkCannotRestoreAStaleMark(t *testing.T) {
	r := searchRunner(nil)
	cask := true
	readJSON := r.ExecuteJSONFn
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		if !cask {
			return jsonInto(result, `{"casks": []}`)
		}
		return readJSON(ctx, result, args...)
	}
	m := newTestModel()
	m.client = brew.NewClient(r)
	m.activePanel = PanelSearch
	rows := []brew.SearchResult{{Name: "claude-code", IsCask: true, Installed: true}}
	m = updateModel(m, SearchDoneMsg{Results: searchItems(rows), Raw: rows})

	slow := m.remarkSearch()
	slowMsg := slow()
	// The second uninstall invalidates the cask list it reads, as a write does.
	cask = false
	m.client.Cache.InvalidateFor("uninstall")
	fastMsg := m.remarkSearch()()

	m = updateModel(m, fastMsg)
	m = updateModel(m, slowMsg)

	if m.searchResults[0].Installed {
		t.Error("a slow re-mark from before the second uninstall restored the installed mark")
	}
	if item := m.panels[PanelSearch].items[0]; strings.Contains(item, "installed") {
		t.Errorf("the row's item shows installed after a stale re-mark: %q", item)
	}
}
