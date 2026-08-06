package gui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
)

func TestMouseDisabledNoOp(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = false
	before := m.activePanel
	nm, _ := m.Update(tea.MouseMsg{
		X: 2, Y: 5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
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
	nm, _ := m.Update(tea.MouseMsg{
		X: 2, Y: y,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
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
	y0 := heights[0] + 2
	// title at inner 0, first item at inner 1 → screen y = y0 + 1 (top border) + 1 (title) = y0+2
	yItem0 := y0 + 1 + 1
	yItem1 := y0 + 1 + 2

	nm, _ := m.Update(tea.MouseMsg{
		X: 3, Y: yItem1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	got := nm.(Model)
	if got.panels[PanelFormulae].selected != 1 {
		t.Fatalf("expected select beta (1), got %d (yItem0=%d yItem1=%d)", got.panels[PanelFormulae].selected, yItem0, yItem1)
	}
}

func TestMouseWheelMovesSelection(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	m.activePanel = PanelFormulae
	p := m.panels[PanelFormulae]
	p.items = []string{"a", "b", "c"}
	p.selected = 1
	p.offset = 0

	nm, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	got := nm.(Model)
	if got.panels[PanelFormulae].selected != 2 {
		t.Fatalf("wheel down: selected=%d want 2", got.panels[PanelFormulae].selected)
	}
	nm, _ = got.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	got = nm.(Model)
	if got.panels[PanelFormulae].selected != 1 {
		t.Fatalf("wheel up: selected=%d want 1", got.panels[PanelFormulae].selected)
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
	// tab 0 is List, tab 1 is Info — click near the right of first tabs
	// Use handleTabBarClick directly for stability.
	nm, cmd := m.handleTabBarClick(0) // first tab = List (already active)
	_ = cmd
	if nm.(Model).activeTab != 0 {
		t.Fatalf("click List: tab=%d", nm.(Model).activeTab)
	}
	// Width of " List " active tab
	m2 := nm.(Model)
	// Click far enough to hit Info (second tab)
	// Accumulate: measure via repeated clicks advancing
	x := 0
	for i, tab := range m2.tabs {
		label := " " + tab.name + " "
		// approximate using inactive width for all after first
		w := len(label) + 2 // padding fudge; handleTabBarClick uses lipgloss
		_ = w
		if i == 1 {
			nm, _ = m2.handleTabBarClick(x + 1)
			break
		}
		// use real widths through a dry run of handleTabBarClick logic
		break
	}
	// Simpler: call handleTabBarClick with x past first tab using lipgloss in test
	m3 := m
	m3.activeTab = 0
	// Brute-force find Info tab x
	found := false
	for x := 0; x < 80; x++ {
		mm := m3
		mm.activeTab = 0
		out, _ := mm.handleTabBarClick(x)
		if out.(Model).activeTab == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to hit Info tab somewhere on the tab bar")
	}
	_ = sw
}

func TestMouseClickFormulaeTableRow(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.cfg.GUI.Mouse = true
	m.activePanel = PanelFormulae
	m.activeTab = 0
	m.tabs = panelTabs[PanelFormulae]
	p := m.panels[PanelFormulae]
	p.loading = false
	p.items = []string{"a", "b", "c"}
	p.formulae = []brew.Formula{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	p.selected = 0
	p.offset = 0

	sw := sidebarWidth(m.cfg, m.width)
	// content starts roughly: border(1)+breadcrumb(1)+tabBar(~2) = 4 from top of main
	// header+sep = 2 → first body row at contentY+2
	// Use handleContentClick directly
	nm, _ := m.handleContentClick(5, 2, 50, 12) // cy=2 → first body row
	got := nm.(Model)
	if got.panels[PanelFormulae].selected != 0 {
		// first body row is index 0
		t.Fatalf("cy=2 selected=%d want 0", got.panels[PanelFormulae].selected)
	}
	nm, _ = m.handleContentClick(5, 3, 50, 12) // second body row
	got = nm.(Model)
	if got.panels[PanelFormulae].selected != 1 {
		t.Fatalf("cy=3 selected=%d want 1", got.panels[PanelFormulae].selected)
	}
	_ = sw
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
	cfg := config.Default()
	if !cfg.GUI.Mouse {
		t.Fatal("default config should enable mouse")
	}
}
