package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

func TestSortFormulae(t *testing.T) {
	fs := []brew.Formula{
		{Name: "zebra", Version: "1"},
		{Name: "alpha", Version: "1", Pinned: true},
		{Name: "beta", Version: "1", Outdated: true, NewVersion: "2"},
		{Name: "gamma", Version: "1", Outdated: true, Pinned: true, NewVersion: "2"},
		{Name: "Alpha", Version: "1"}, // case-insensitive with zebra/alpha order
	}
	sortFormulae(fs)
	// outdated first (gamma, beta — both outdated; gamma before beta by name among outdated? gamma vs beta: beta first)
	// Within outdated: pinned first → gamma (outdated+pinned), beta (outdated)
	// Then non-outdated pinned: alpha
	// Then remaining A-Z: Alpha, zebra — "Alpha" and "alpha" same lower; stable among equals after ranks
	got := make([]string, len(fs))
	for i, f := range fs {
		got[i] = f.Name
	}
	// outdated+pinned, outdated, pinned, then alpha names, zebra
	if got[0] != "gamma" {
		t.Fatalf("first should be outdated+pinned gamma, got %v", got)
	}
	if got[1] != "beta" {
		t.Fatalf("second should be outdated beta, got %v", got)
	}
	if got[2] != "alpha" {
		t.Fatalf("third should be pinned alpha, got %v", got)
	}
	// remaining Alpha then zebra (A before z)
	rest := got[3:]
	if len(rest) != 2 || strings.ToLower(rest[0]) != "alpha" || rest[1] != "zebra" {
		t.Fatalf("remaining want Alpha,zebra got %v", rest)
	}
}

// TestFormulaeTabsOrder: the Formulae bar is Info first; the List table is gone
// (v2-03b AC-02), so the default tab is the selected formula's details.
func TestFormulaeTabsOrder(t *testing.T) {
	tabs := panelTabs[PanelFormulae]
	want := []string{"Info", "Deps", "Used By", "Caveats", "Files"}
	if len(tabs) != len(want) {
		t.Fatalf("len=%d want %d", len(tabs), len(want))
	}
	for i, n := range want {
		if tabs[i].name != n {
			t.Errorf("tab[%d]=%q want %q", i, tabs[i].name, n)
		}
	}
	if formulaInfoTab != 0 {
		t.Errorf("Info must be the first tab, got index %d", formulaInfoTab)
	}
}

// TestNeedsTabFetchShifted: the tab indices moved with the List tab's removal;
// the fetching tabs are named by their constants, not by number.
func TestNeedsTabFetchShifted(t *testing.T) {
	if needsTabFetch(PanelFormulae, formulaInfoTab) {
		t.Error("Info must not fetch")
	}
	if !needsTabFetch(PanelFormulae, formulaDepsTab) {
		t.Error("Deps must fetch")
	}
	if !needsTabFetch(PanelFormulae, formulaUsesTab) {
		t.Error("Used By must fetch")
	}
	if needsTabFetch(PanelFormulae, formulaCaveatsTab) {
		t.Error("Caveats must not fetch")
	}
	if !needsTabFetch(PanelFormulae, formulaFilesTab) {
		t.Error("Files must fetch")
	}
}

// TestFormulaeDefaultTabIsInfo (replaces TestRenderContentListTabIsTable): a
// panel opens on Info, and Info draws the selected formula, not a table.
func TestFormulaeDefaultTabIsInfo(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 120, Height: 40})
	m = sendKey(m, "2")
	if m.activeTab != formulaInfoTab {
		t.Fatalf("Formulae should open on Info, got tab %d", m.activeTab)
	}
	m = updateModel(m, DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"foo  1.0"},
		Formulae: []brew.Formula{{Name: "foo", Version: "1.0", Tap: "homebrew/core"}},
	})
	out := m.renderContent(50, 15)
	if !strings.Contains(out, "Version:") || !strings.Contains(out, "foo") {
		t.Fatalf("Info should show the formula's details, not a table:\n%s", out)
	}
}

// TestFormulaInfoShowsTap (replaces TestShortenTap): the table's Tap column
// moved to Info, in full.
func TestFormulaInfoShowsTap(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 120, Height: 40})
	m = sendKey(m, "2")
	m = updateModel(m, DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"foo  1.0"},
		Formulae: []brew.Formula{{Name: "foo", Version: "1.0", Tap: "homebrew/core"}},
	})
	out := m.renderContent(60, 15)
	if !strings.Contains(out, "homebrew/core") {
		t.Fatalf("Info should show the tap in full, got:\n%s", out)
	}
}

func TestTabActiveStyleHasBottomBorder(t *testing.T) {
	style.ApplyTheme(style.DarkTheme())
	// Active uses bottom border character ━
	s := style.TabActive.Render(" List ")
	if s == style.TabInactive.Render(" List ") {
		t.Fatal("active and inactive tab styles should differ")
	}
	// Bold + accent on active
	if !style.TabActive.GetBold() {
		t.Error("TabActive should be bold")
	}
	if style.TabActive.GetForeground() != style.AccentColor {
		t.Error("TabActive should use AccentColor")
	}
}
