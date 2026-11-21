package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// filterModel is the model on the Formulae panel with the named formulae loaded,
// at w by h.
func filterModel(t *testing.T, w, h int, names ...string) *Model {
	t.Helper()
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
	m = sendKey(m, "2")
	msg := DataLoadedMsg{PanelID: PanelFormulae}
	for _, n := range names {
		msg.Items = append(msg.Items, n+"  1.0  bottled")
		msg.Formulae = append(msg.Formulae, brew.Formula{Name: n, Version: "1.0", Tap: "homebrew/core"})
	}
	return updateModel(m, msg)
}

// typeText sends each character of s as a key press.
func typeText(m *Model, s string) *Model {
	for _, r := range s {
		m = sendKey(m, string(r))
	}
	return m
}

// TestSlashFiltersFormulaeNotSearch: on Formulae, / opens the local filter and
// leaves the Search box alone.
func TestSlashFiltersFormulaeNotSearch(t *testing.T) {
	m := sendKey(filterModel(t, 120, 40, "ripgrep", "wget"), "/")
	if m.activePanel != PanelFormulae || !m.filterInput.Focused() {
		t.Fatalf("/ on Formulae should open the filter on Formulae, panel %v", m.activePanel)
	}
	if m.searchInput.Focused() {
		t.Fatal("/ on Formulae must not focus the Search box")
	}
}

// TestSlashStillSearchesOnOtherPanels: / on any other panel jumps to Search, and
// from Formulae 7 then / reaches it.
func TestSlashStillSearchesOnOtherPanels(t *testing.T) {
	m := sendKey(updateModel(newTestModel(), tea.WindowSizeMsg{Width: 120, Height: 40}), "/")
	if m.activePanel != PanelSearch || !m.searchInput.Focused() {
		t.Fatalf("/ on Status should open Search, panel %v", m.activePanel)
	}

	m = sendKey(filterModel(t, 120, 40, "ripgrep"), "7")
	m = sendKey(m, "/")
	if m.activePanel != PanelSearch || !m.searchInput.Focused() || m.filterInput.Focused() {
		t.Fatalf("7 then / should reach Search, panel %v", m.activePanel)
	}
}

// TestFilterNarrowsAsYouType: typing narrows the list, the cursor lands on the
// first match, and the title count reads matches/total.
func TestFilterNarrowsAsYouType(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget", "rip-tool"), "/"), "ri")
	p := m.panels[PanelFormulae]
	if p.query != "ri" || len(p.items) != 2 {
		t.Fatalf("query %q with %d items, want ri with 2", p.query, len(p.items))
	}
	if p.selectedName() != "ripgrep" {
		t.Fatalf("cursor on %q, want the first match ripgrep", p.selectedName())
	}
	view := screenText(m)
	if !strings.Contains(view, "/ ri") || !strings.Contains(view, "2/3") {
		t.Fatalf("the filter row and 2/3 should show:\n%s", view)
	}
}

// TestFilterInputTakesEveryKey: while the input has focus, letters, digits, ?
// and q are typed, and only ctrl+c quits.
func TestFilterInputTakesEveryKey(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget"), "/"), "yq1?")
	if m.panels[PanelFormulae].query != "yq1?" {
		t.Fatalf("query = %q, want the typed keys", m.panels[PanelFormulae].query)
	}
	if m.activePanel != PanelFormulae || m.showHelp {
		t.Fatal("digits and ? must not switch panels or open help")
	}

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("ctrl+c should quit while the filter has focus")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c should quit while the filter has focus")
	}
}

// TestFilterUpDownMoveTheCursorOnly: up and down move the selection and leave the
// query as it is.
func TestFilterUpDownMoveTheCursorOnly(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget", "rip-tool"), "/"), "r")
	p := m.panels[PanelFormulae]
	m = sendSpecial(m, tea.KeyDown)
	if p.selectedName() != "rip-tool" || p.query != "r" {
		t.Fatalf("down: cursor on %q, query %q; want rip-tool, r", p.selectedName(), p.query)
	}
	m = sendSpecial(m, tea.KeyUp)
	if p.selectedName() != "ripgrep" || p.query != "r" {
		t.Fatalf("up: cursor on %q, query %q; want ripgrep, r", p.selectedName(), p.query)
	}
}

// TestFilterEnterKeepsTheFilter: Enter leaves the input with the filter on, and /
// opens the input again with the query in it.
func TestFilterEnterKeepsTheFilter(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget", "rip-tool"), "/"), "wg")
	m = sendSpecial(m, tea.KeyEnter)
	if m.filterInput.Focused() {
		t.Fatal("Enter should leave the input")
	}
	if p := m.panels[PanelFormulae]; p.query != "wg" || len(p.items) != 1 {
		t.Fatalf("after Enter: query %q with %d items, want wg with 1", p.query, len(p.items))
	}
	m = sendKey(m, "/")
	if got := m.filterInput.Value(); got != "wg" {
		t.Fatalf("/ should reopen the input with wg, got %q", got)
	}
}

// TestFilterEscInInputClears: Esc in the input clears the query and the filter.
func TestFilterEscInInputClears(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget", "rip-tool"), "/"), "wg")
	m = sendSpecial(m, tea.KeyEsc)
	p := m.panels[PanelFormulae]
	if m.filterInput.Focused() || p.query != "" || len(p.items) != 3 {
		t.Fatalf("Esc: focused %v, query %q, %d items; want unfocused, none, 3", m.filterInput.Focused(), p.query, len(p.items))
	}
}

// TestFilterEscClearsActiveFilter: Esc on an active filter that has no focus
// clears it too.
func TestFilterEscClearsActiveFilter(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget", "rip-tool"), "/"), "wg")
	m = sendSpecial(m, tea.KeyEnter)
	m = sendSpecial(m, tea.KeyEsc)
	if p := m.panels[PanelFormulae]; p.query != "" || len(p.items) != 3 {
		t.Fatalf("Esc on the active filter: query %q, %d items; want none, 3", p.query, len(p.items))
	}
}

// TestFilterNoMatchShowsEmptyState: a query nothing matches says so, and Esc
// brings the list back.
func TestFilterNoMatchShowsEmptyState(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget"), "/"), "zzz")
	if view := screenText(m); !strings.Contains(view, `No matches for "zzz"`) {
		t.Fatalf("the empty state should show:\n%s", view)
	}
	m = sendSpecial(m, tea.KeyEsc)
	if view := screenText(m); strings.Contains(view, "No matches for") {
		t.Fatalf("Esc should clear the empty state:\n%s", view)
	}
}

// TestFilterFollowsItsPanel: switching panel blurs the input but keeps each
// panel's own query, so coming back shows the same narrowing.
func TestFilterFollowsItsPanel(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget"), "/"), "wg")
	m = sendSpecial(m, tea.KeyEnter)
	m = sendKey(m, "1")
	if m.filterInput.Focused() {
		t.Fatal("switching panel should blur the filter")
	}
	m = sendKey(m, "2")
	if p := m.panels[PanelFormulae]; p.query != "wg" || len(p.items) != 1 {
		t.Fatalf("back on Formulae: query %q, %d items; want wg, 1", p.query, len(p.items))
	}
}

// TestFilterFramesFitTheTerminal: the filter's states draw a frame of exactly
// the terminal's size, in both themes.
func TestFilterFramesFitTheTerminal(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for _, size := range []struct{ w, h int }{{80, 24}, {140, 40}} {
		for name, theme := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
			style.ApplyTheme(theme)
			open := sendKey(filterModel(t, size.w, size.h, "ripgrep", "rip-tool", "wget"), "/")
			withQuery := typeText(sendKey(filterModel(t, size.w, size.h, "ripgrep", "rip-tool", "wget"), "/"), "ri")
			noMatch := typeText(sendKey(filterModel(t, size.w, size.h, "ripgrep", "rip-tool", "wget"), "/"), "zzz")
			entered := sendSpecial(typeText(sendKey(filterModel(t, size.w, size.h, "ripgrep", "rip-tool", "wget"), "/"), "ri"), tea.KeyEnter)
			cases := []struct {
				state string
				m     *Model
				want  string
			}{
				{"open", open, ""},
				{"query", withQuery, "2/3"},
				{"no match", noMatch, `No matches for "zzz"`},
				{"after Enter", entered, "/ ri"},
			}
			for _, c := range cases {
				view := c.m.View().Content
				assertFrameSize(t, view, size.w, size.h)
				if c.want != "" && !strings.Contains(ansi.Strip(view), c.want) {
					t.Errorf("%s %dx%d %s: want %q in the frame:\n%s", name, size.w, size.h, c.state, c.want, ansi.Strip(view))
				}
			}
		}
	}
}

// modelOf is the Model a tea.Model returned by Update holds.
func modelOf(tm tea.Model) *Model {
	m := tm.(Model)
	return &m
}

// TestFilterClickMapsToFilteredRows: a click on the filter row selects nothing,
// and a click on a list row selects that filtered row.
func TestFilterClickMapsToFilteredRows(t *testing.T) {
	m := typeText(sendKey(filterModel(t, 120, 40, "ripgrep", "wget", "rip-tool"), "/"), "r")
	heights := m.computeContentHeights()
	top := headerRows + heights[PanelStatus] + 2 // the Status box, then the Formulae border
	nm, _ := m.handleMouseClick(2, top+1)        // the filter row
	m = modelOf(nm)
	if m.panels[PanelFormulae].selected != 0 {
		t.Fatalf("a click on the filter row moved the cursor to %d", m.panels[PanelFormulae].selected)
	}
	nm, _ = m.handleMouseClick(2, top+3) // the second filtered row
	m = modelOf(nm)
	if got := m.panels[PanelFormulae].selectedName(); got != "rip-tool" {
		t.Fatalf("a click on the second filtered row selected %q, want rip-tool", got)
	}
}
