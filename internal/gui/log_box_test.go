package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestLogIsSiblingBoxBelowMainPanel locks AC-06.1: the main panel and the
// Log are two boxes stacked in the same column, each with its own border,
// not a Log box inside the main panel's border.
func TestLogIsSiblingBoxBelowMainPanel(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.commandLog.Append("visible-cmd")
	l := m.screenLayout()
	if l.logH == 0 {
		t.Fatal("fixture: the Log should be shown at 140x40")
	}

	rows := plainRows(m.View().Content)
	col := l.sideW
	var corners []int
	for r := range rows {
		if rows[r][col] == '╭' {
			corners = append(corners, r)
		}
	}
	if len(corners) != 2 || corners[0] != 0 || corners[1] != l.mainH {
		t.Fatalf("top-left corners of the main panel and Log are at rows %v, want [0 %d]", corners, l.mainH)
	}
	if rows[l.mainH-1][col] != '╰' {
		t.Fatalf("main panel bottom-left corner missing at row %d", l.mainH-1)
	}
	if !strings.HasPrefix(string(rows[l.mainH+1][col:]), "│Log") {
		t.Fatalf("Log title should sit inside the Log box at row %d, got %q", l.mainH+1, string(rows[l.mainH+1][col:]))
	}
}

// TestClickOnLogBoxDoesNothing: the Log is not part of the main panel, so a
// click on its rows changes neither the panel nor the list selection.
func TestClickOnLogBoxDoesNothing(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.cfg.GUI.Mouse = true
	m.activePanel = PanelFormulae
	m.tabs = panelTabs[PanelFormulae]
	p := m.panels[PanelFormulae]
	p.items = []string{"a", "b", "c"}
	p.formulae = []brew.Formula{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	p.selected = 0

	l := m.screenLayout()
	nm, _ := m.Update(tea.MouseClickMsg{X: l.sideW + 5, Y: l.mainH + 2, Button: tea.MouseLeft})
	got := nm.(Model)
	if got.panels[PanelFormulae].selected != 0 || got.activePanel != PanelFormulae {
		t.Fatalf("click in the Log box changed selection to %d on panel %v",
			got.panels[PanelFormulae].selected, got.activePanel)
	}
}

// TestMouseHitsWhatIsDrawn locks AC-06.4: a click on a Formulae row, a tab
// and a sidebar row lands on the element the frame drew at that position,
// at the current geometry.
func TestMouseHitsWhatIsDrawn(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.cfg.GUI.Mouse = true
	m.activePanel = PanelFormulae
	m.tabs = panelTabs[PanelFormulae]
	p := m.panels[PanelFormulae]
	p.items = []string{"alpha", "bravo", "charlie"}
	p.formulae = []brew.Formula{{Name: "alpha"}, {Name: "bravo"}, {Name: "charlie"}}
	p.selected = 0

	l := m.screenLayout()
	// Inside the main border: breadcrumb, tab bar, then the table header
	// and separator, so body row k is drawn at y = 1 + 1 + tabBar + 2 + k.
	bravoY := 1 + 1 + m.tabBarRows() + 2 + 1
	rows := plainRows(m.View().Content)
	if !strings.HasPrefix(strings.TrimLeft(string(rows[bravoY][l.sideW+1:]), " "), "bravo") {
		t.Fatalf("fixture: row %d should draw bravo, got %q", bravoY, string(rows[bravoY][l.sideW+1:]))
	}
	nm, _ := m.Update(tea.MouseClickMsg{X: l.sideW + 5, Y: bravoY, Button: tea.MouseLeft})
	if got := nm.(Model).panels[PanelFormulae].selected; got != 1 {
		t.Fatalf("click on the bravo row selected %d, want 1", got)
	}

	// The bottom sidebar box is drawn on the last body row.
	nm, _ = m.Update(tea.MouseClickMsg{X: 2, Y: l.bodyH - 1, Button: tea.MouseLeft})
	if got := nm.(Model).activePanel; got != PanelSearch {
		t.Fatalf("click on the bottom sidebar row activated panel %v, want Search", got)
	}
}
