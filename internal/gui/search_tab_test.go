package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestTabCyclesPastFocusedSearchInput (lb19 AC-01): with the search input
// focused, Tab and Shift+Tab still move between panels, and the query gets no
// tab character.
func TestTabCyclesPastFocusedSearchInput(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelSearch
	m.searchInput.Focus()

	m = updateModel(m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.activePanel != PanelStatus {
		t.Fatalf("Tab from focused Search: activePanel = %v, want PanelStatus", m.activePanel)
	}
	if m.searchInput.Focused() {
		t.Error("Tab left the search input focused on another panel")
	}

	m = updateModel(m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.activePanel != PanelSearch || !m.searchInput.Focused() {
		t.Fatalf("Shift+Tab from Status: activePanel = %v focused = %v, want Search focused",
			m.activePanel, m.searchInput.Focused())
	}

	m = updateModel(m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.activePanel != PanelServices {
		t.Fatalf("Shift+Tab from focused Search: activePanel = %v, want PanelServices", m.activePanel)
	}
	if got := m.searchInput.Value(); got != "" {
		t.Errorf("query = %q, want empty: tab must not be typed into it", got)
	}
}

// TestTabCycleReturnsToStartThroughSearch (lb19 AC-01): seven Tabs from Status
// pass through Search and come back to Status, and seven Shift+Tabs do the same
// the other way.
func TestTabCycleReturnsToStartThroughSearch(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyTab}, {Code: tea.KeyTab, Mod: tea.ModShift}} {
		m := newTestModel()
		start := m.activePanel
		sawSearch := false
		for range 7 {
			m = updateModel(m, key)
			sawSearch = sawSearch || m.activePanel == PanelSearch
		}
		if !sawSearch {
			t.Errorf("%v x7 from %v never visited PanelSearch", key, start)
		}
		if m.activePanel != start {
			t.Errorf("%v x7: activePanel = %v, want %v", key, m.activePanel, start)
		}
	}
}
