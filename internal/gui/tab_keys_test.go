package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestHLSwitchTabs (lb19 AC-03): h and l move back and forward through the tabs,
// as the arrow keys do.
func TestHLSwitchTabs(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "l")
	if m.activeTab != 1 {
		t.Fatalf("l: activeTab = %d, want 1", m.activeTab)
	}
	m = sendKey(m, "h")
	if m.activeTab != 0 {
		t.Fatalf("h: activeTab = %d, want 0", m.activeTab)
	}
	m = sendSpecial(m, tea.KeyRight)
	if m.activeTab != 1 {
		t.Fatalf("right: activeTab = %d, want 1", m.activeTab)
	}
	m = sendSpecial(m, tea.KeyLeft)
	if m.activeTab != 0 {
		t.Fatalf("left: activeTab = %d, want 0", m.activeTab)
	}
}

// TestHLTypeIntoSearch (lb19 AC-03): on Search h and l are text. With the box
// focused they go into it; with it unfocused the first one starts the search.
func TestHLTypeIntoSearch(t *testing.T) {
	m := newTestModel()
	m = sendKey(m, "7")
	m = sendKey(m, "h")
	if m.activePanel != PanelSearch || !m.searchInput.Focused() || m.searchInput.Value() != "h" {
		t.Fatalf("h on unfocused Search: panel %v focused %v value %q, want Search focused with h",
			m.activePanel, m.searchInput.Focused(), m.searchInput.Value())
	}
	m = sendKey(m, "l")
	if got := m.searchInput.Value(); got != "hl" {
		t.Fatalf("l into focused Search: value %q, want hl", got)
	}
}
