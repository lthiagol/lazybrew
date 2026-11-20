package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestNeverLoadedPanelSaysNotLoaded locks AC-05: a panel that has never been
// fetched says "not loaded" in the sidebar, while a fetched panel with no
// items shows 0 and a fetched panel with items lists them.
func TestNeverLoadedPanelSaysNotLoaded(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	casks := m.panels[PanelCasks]

	// The state is the box's border label, so a collapsed box still shows it.
	if got := m.sidebarLabel(casks).full; !strings.Contains(got, "not loaded") {
		t.Fatalf("fresh Casks panel should say not loaded, got %q", got)
	}
	if side := m.renderSidebar(); !strings.Contains(side, "not loaded") {
		t.Fatal("fresh sidebar should show not loaded for the panels that were never fetched")
	}

	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{}})
	got := m.sidebarLabel(m.panels[PanelCasks]).full
	if plainLine(got) != "0" {
		t.Fatalf("fetched Casks panel with no items should show 0, got %q", got)
	}
	if strings.Contains(got, "not loaded") {
		t.Fatalf("fetched Casks panel must not say not loaded, got %q", got)
	}

	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Items: []string{"alacritty  0.14"}})
	if got := m.panels[PanelCasks].renderSidebarContent(30, 5, nil); !strings.Contains(got, "alacritty") {
		t.Fatalf("fetched Casks panel should list its items, got %q", got)
	}
}

// TestFailedFirstFetchIsLoadedNotNotLoaded: a fetch that failed is still a
// fetch, so the panel shows its error rather than claiming it was never
// loaded.
func TestFailedFirstFetchIsLoadedNotNotLoaded(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelCasks, Err: assertAnError})

	got := m.panels[PanelCasks].renderSidebarContent(30, 5, nil)
	if strings.Contains(got, "not loaded") {
		t.Fatalf("a failed fetch must not say not loaded, got %q", got)
	}
	if !strings.Contains(got, "test error") {
		t.Fatalf("a failed fetch should show its error, got %q", got)
	}
}
