package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

func TestMouseDisabledNoOp(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = false
	before := m.activePanel
	nm, _ := m.Update(tea.MouseClickMsg{
		X: 2, Y: 5,
		Button: tea.MouseLeft,
	})
	got := nm.(Model)
	if got.activePanel != before {
		t.Fatalf("mouse disabled should not change panel, got %v", got.activePanel)
	}
}

func TestMouseClickSidebarSwitchesPanel(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	// Status is panel 0 at top. Formulae is panel 1 below it.
	// Click well into the second sidebar box.
	heights := m.computeContentHeights()
	y := heights[0] + 2 + 1 // past first box border+content, into second box title
	nm, _ := m.Update(tea.MouseClickMsg{
		X: 2, Y: y,
		Button: tea.MouseLeft,
	})
	got := nm.(Model)
	if got.activePanel != PanelFormulae {
		t.Fatalf("expected click to activate Formulae, got %v (y=%d heights=%v)", got.activePanel, y, heights)
	}
}

func TestMouseClickSidebarSelectsRow(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	// Put items on Formulae and make it active with known geometry.
	cmd := m.switchPanel(PanelFormulae)
	_ = cmd
	p := m.panels[PanelFormulae]
	p.items = []string{"alpha  1", "beta  2", "gamma  3"}
	p.formulae = []brew.Formula{{Name: "alpha"}, {Name: "beta"}, {Name: "gamma"}}
	p.selected = 0
	p.offset = 0
	p.loading = false

	heights := m.computeContentHeights()
	// Formulae is index 1
	y0 := headerRows + heights[0] + 2
	// The title sits in the top border, so the first item is the first body
	// row: screen y = y0 + 1 (top border) + 0.
	yItem0 := y0 + 1
	yItem1 := y0 + 2

	nm, _ := m.Update(tea.MouseClickMsg{
		X: 3, Y: yItem1,
		Button: tea.MouseLeft,
	})
	got := nm.(Model)
	if got.panels[PanelFormulae].selected != 1 {
		t.Fatalf("expected select beta (1), got %d (yItem0=%d yItem1=%d)", got.panels[PanelFormulae].selected, yItem0, yItem1)
	}
}

// TestMouseWheelMovesSelection: a wheel notch over a sidebar box moves that
// box's selection one row, and the box is the one the pointer is over.
func TestMouseWheelMovesSelection(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"a", "b", "c"}
	p.selected = 1
	p.offset = 0

	// The Formulae box is the second box; its first body row is one row below
	// its top border.
	heights := m.computeContentHeights()
	x, y := 2, headerRows+heights[0]+2+1
	nm, _ := m.Update(tea.MouseWheelMsg{X: x, Y: y, Button: tea.MouseWheelDown})
	got := nm.(Model)
	if got.panels[PanelFormulae].selected != 2 {
		t.Fatalf("wheel down: selected=%d want 2", got.panels[PanelFormulae].selected)
	}
	nm, _ = got.Update(tea.MouseWheelMsg{X: x, Y: y, Button: tea.MouseWheelUp})
	got = nm.(Model)
	if got.panels[PanelFormulae].selected != 1 {
		t.Fatalf("wheel up: selected=%d want 1", got.panels[PanelFormulae].selected)
	}
}

// TestWheelOverAnotherBoxFocusesIt (AC-04.1): a wheel notch over another sidebar
// box focuses that box, as a click does, and moves its list.
func TestWheelOverAnotherBoxFocusesIt(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	heights := m.computeContentHeights()
	casksY := headerRows + (heights[0] + 2) + (heights[1] + 2) + 1
	nm, _ := m.Update(tea.MouseWheelMsg{X: 2, Y: casksY, Button: tea.MouseWheelDown})
	if got := nm.(Model).activePanel; got != PanelCasks {
		t.Fatalf("a wheel notch over the Casks box focused panel %v, want Casks", got)
	}
}

// TestWheelOverMainScrollsItsContent (AC-04.1): a wheel notch over the main
// panel scrolls its tab content three rows, and the sidebar does not move.
func TestWheelOverMainScrollsItsContent(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.cfg.GUI.Mouse = true
	m = sendKey(m, "2")
	var caveats []string
	for i := 0; i < 60; i++ {
		caveats = append(caveats, fmt.Sprintf("line-%02d", i))
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"foo  1.0"},
		Formulae: []brew.Formula{{Name: "foo", Version: "1.0", Caveats: strings.Join(caveats, "\n")}}})
	for range formulaCaveatsTab {
		m = sendKey(m, "l")
	}
	if m.activeTab != formulaCaveatsTab {
		t.Fatalf("fixture: want the Caveats tab, got %d", m.activeTab)
	}
	if !strings.Contains(sgr.ReplaceAllString(m.renderMainPanelIn(m.screenLayout()), ""), "line-00") {
		t.Fatal("fixture: the first caveat line should show before scrolling")
	}

	l := m.screenLayout()
	x, y := l.sideW+5, headerRows+1+m.tabBarRows()+1
	nm, _ := m.Update(tea.MouseWheelMsg{X: x, Y: y, Button: tea.MouseWheelDown})
	got := nm.(Model)
	main := sgr.ReplaceAllString(got.renderMainPanelIn(got.screenLayout()), "")
	if strings.Contains(main, "line-00") || !strings.Contains(main, "line-03") {
		t.Fatalf("a notch should scroll the content three rows, got:\n%s", main)
	}
	if got.panels[PanelFormulae].selected != m.panels[PanelFormulae].selected {
		t.Fatal("a wheel notch over the main panel must not move the sidebar's selection")
	}
}

func TestMouseClickTabBar(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	m.activePanel = PanelFormulae
	m.activeTab = 0
	m.tabs = panelTabs[PanelFormulae]
	m.panels[PanelFormulae].loading = false
	m.panels[PanelFormulae].items = []string{"x"}
	m.panels[PanelFormulae].formulae = []brew.Formula{{Name: "x"}}

	sw := sidebarWidth(m.cfg, m.width)
	// Main content: past sidebar, inside border, on tab bar row (y=2 ≈ breadcrumb+border)
	// tab 0 is Info, tab 1 is Deps — click near the right of first tabs
	// Use handleTabBarClick directly for stability.
	nm, cmd := m.handleTabBarClick(0) // first tab = Info (already active)
	_ = cmd
	if nm.(Model).activeTab != formulaInfoTab {
		t.Fatalf("click Info: tab=%d", nm.(Model).activeTab)
	}
	// Brute-force find the Deps tab's x on the tab bar
	found := false
	for x := 0; x < 80; x++ {
		mm := m
		mm.activeTab = formulaInfoTab
		out, _ := mm.handleTabBarClick(x)
		if out.(Model).activeTab == formulaDepsTab {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to hit Deps tab somewhere on the tab bar")
	}
	_ = sw
}

// TestMouseClickFormulaeContentSelectsNothing (replaces
// TestMouseClickFormulaeTableRow, v2-03b AC-02): the List table is gone, so a
// click on the main content selects no row. The sidebar list is where a row is
// picked.
func TestMouseClickFormulaeContentSelectsNothing(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	m = sendKey(m, "2")
	m = updateModel(m, DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"a", "b", "c"},
		Formulae: []brew.Formula{{Name: "a"}, {Name: "b"}, {Name: "c"}},
	})
	m.panels[PanelFormulae].selected = 0

	sw := sidebarWidth(m.cfg, m.width)
	// Absolute row inside the main content: header, border, tab bar, then one more.
	y := headerRows + 1 + m.tabBarRows() + 1
	nm, _ := m.handleMouseClick(sw+5, y)
	if got := nm.(Model).panels[PanelFormulae].selected; got != 0 {
		t.Fatalf("a click on the content selected row %d, want 0", got)
	}
}

func TestSelectIndexClampsAndScrolls(t *testing.T) {
	p := &panelData{
		items:       []string{"0", "1", "2", "3", "4", "5"},
		selected:    0,
		offset:      0,
		visibleRows: 3,
	}
	p.selectIndex(4)
	if p.selected != 4 {
		t.Fatalf("selected=%d", p.selected)
	}
	if p.offset != 2 { // 4 - 3 + 1
		t.Fatalf("offset=%d want 2", p.offset)
	}
	p.selectIndex(100)
	if p.selected != 5 {
		t.Fatalf("clamp high selected=%d", p.selected)
	}
	p.selectIndex(-3)
	if p.selected != 0 {
		t.Fatalf("clamp low selected=%d", p.selected)
	}
}

func TestMouseDefaultConfigEnabled(t *testing.T) {
	cfg := testConfig()
	if !cfg.GUI.Mouse {
		t.Fatal("default config should enable mouse")
	}
}
