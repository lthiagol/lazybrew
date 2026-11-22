package gui

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// bgOpen matches the opening escape of a background colour in a styled row.
var bgOpen = regexp.MustCompile("\x1b\\[48[;:0-9]*m")

// TestSidebarWidthRule (AC-01.1): the sidebar is a percentage of the terminal,
// 35% by default, with a 20-cell floor and a 60-cell cap.
func TestSidebarWidthRule(t *testing.T) {
	cfg := testConfig()
	if got := sidebarWidth(cfg, 140); got != 49 {
		t.Errorf("default sidebar at 140 columns = %d, want 49 (35%%)", got)
	}
	if got := sidebarWidth(cfg, 300); got != 60 {
		t.Errorf("sidebar at 300 columns = %d, want the 60-cell cap", got)
	}
	if got := sidebarWidth(cfg, 40); got != 20 {
		t.Errorf("sidebar at 40 columns = %d, want the 20-cell floor", got)
	}
}

// TestFormulaRowShowsVersionArrowAndPin (AC-01.2): a row is built from the
// formula: the status glyph, the name, `1.2 → 1.3` right-aligned, and the pin.
func TestFormulaRowShowsVersionArrowAndPin(t *testing.T) {
	p := &panelData{id: PanelFormulae, formulae: []brew.Formula{
		{Name: "bash", Version: "5.2", NewVersion: "5.3", Outdated: true, Pinned: true},
	}}
	p.items = []string{"bash  5.2"}
	row := plainLine(p.sidebarRow(0, 30, nil, style.UnicodeIcons, true))
	if w := lipgloss.Width(row); w != 30 {
		t.Fatalf("row is %d cells, want 30: %q", w, row)
	}
	if !strings.HasPrefix(row, "▸ "+style.UnicodeIcons.Upgradable+" bash") {
		t.Errorf("row should start with the marker, the outdated glyph and the name: %q", row)
	}
	if !strings.Contains(row, "5.2 → 5.3") {
		t.Errorf("row should show the version arrow: %q", row)
	}
	if !strings.HasSuffix(row, style.UnicodeIcons.Pinned) {
		t.Errorf("a pinned formula ends with the pin glyph: %q", row)
	}
}

// TestSidebarRowCutsNameWithEllipsis (AC-01.2): a long name is cut with an
// ellipsis, and the version stays whole.
func TestSidebarRowCutsNameWithEllipsis(t *testing.T) {
	p := &panelData{id: PanelCasks, casks: []brew.Cask{
		{Name: "an-extremely-long-cask-name", Version: "1.0"},
	}}
	p.items = []string{"an-extremely-long-cask-name"}
	row := plainLine(p.sidebarRow(0, 22, nil, style.UnicodeIcons, true))
	if w := lipgloss.Width(row); w != 22 {
		t.Fatalf("row is %d cells, want 22: %q", w, row)
	}
	if !strings.Contains(row, "…") {
		t.Errorf("the long name should be cut with an ellipsis: %q", row)
	}
	if !strings.HasSuffix(row, " 1.0") {
		t.Errorf("the version should stay whole and flush right: %q", row)
	}
}

// TestSelectedRowHighlightOnlyWhenFocused (AC-01.3): the selected row of the
// focused box takes the surface background across its width; an unfocused box
// shows the marker without the highlight.
func TestSelectedRowHighlightOnlyWhenFocused(t *testing.T) {
	p := &panelData{id: PanelFormulae, formulae: []brew.Formula{{Name: "bash", Version: "5.2"}}}
	p.items = []string{"bash  5.2"}

	focused := p.sidebarRow(0, 30, nil, style.UnicodeIcons, true)
	unfocused := p.sidebarRow(0, 30, nil, style.UnicodeIcons, false)
	if !bgOpen.MatchString(focused) {
		t.Errorf("the focused selected row should carry the highlight background: %q", focused)
	}
	if bgOpen.MatchString(unfocused) {
		t.Errorf("an unfocused box must not highlight its selected row: %q", unfocused)
	}
	if plainLine(focused) != plainLine(unfocused) {
		t.Errorf("focus changes the colour only, not the text:\n%q\n%q", plainLine(focused), plainLine(unfocused))
	}
}

// TestOutdatedRowsCountFormulaeThenCasks (AC-01.2): the Outdated list is the
// formulae then the casks, and each row shows `installed → latest`.
func TestOutdatedRowsCountFormulaeThenCasks(t *testing.T) {
	p := &panelData{id: PanelOutdated,
		formulae: []brew.Formula{{Name: "git", Version: "2.1", NewVersion: "2.2"}},
		casks:    []brew.Cask{{Name: "iterm", Version: "3.4", NewVersion: "3.5"}},
	}
	p.items = []string{"git", "iterm"}
	if got := plainLine(p.sidebarRow(0, 30, nil, style.UnicodeIcons, false)); !strings.Contains(got, "2.1 → 2.2") {
		t.Errorf("row 0 should be the formula: %q", got)
	}
	if got := plainLine(p.sidebarRow(1, 30, nil, style.UnicodeIcons, false)); !strings.Contains(got, "3.4 → 3.5") {
		t.Errorf("row 1 should be the cask: %q", got)
	}
}

// TestTrustedTapRowDrawsLock (lb25 fix r2, u5): a trusted tap's row leads with the
// set's locked mark, in both sets, never ✓. Unicode's locked mark is the ✔ brew prints.
func TestTrustedTapRowDrawsLock(t *testing.T) {
	taps := &panelData{id: PanelTaps, taps: []brew.Tap{{Name: "lt/tap", Trusted: true}}}
	taps.setTapRows()
	for name, icons := range map[string]style.Icons{"nerd": style.NerdIcons, "unicode": style.UnicodeIcons} {
		got := plainLine(taps.sidebarRow(0, 30, nil, icons, false))
		if want := icons.Locked + " lt/tap"; !strings.Contains(got, want) || strings.Contains(got, "✓") {
			t.Errorf("%s set: trusted tap row %q should draw %q and no ✓", name, got, icons.Locked)
		}
	}
}

// TestTapAndServiceRowsShowTheirState (AC-01.2): a tap shows its trust, a
// service its running state, both as the right-hand word.
func TestTapAndServiceRowsShowTheirState(t *testing.T) {
	taps := &panelData{id: PanelTaps, taps: []brew.Tap{{Name: "lt/tap"}}}
	taps.setTapRows()
	if got := plainLine(taps.sidebarRow(0, 30, nil, style.UnicodeIcons, false)); !strings.HasSuffix(got, "untrusted") {
		t.Errorf("an untrusted tap ends with untrusted: %q", got)
	}

	svc := &panelData{id: PanelServices, services: []brew.Service{{Name: "moshi", Status: brew.ServiceStarted}}}
	svc.items = []string{"moshi"}
	got := plainLine(svc.sidebarRow(0, 30, nil, style.UnicodeIcons, false))
	if !strings.Contains(got, style.UnicodeIcons.Running) || !strings.HasSuffix(got, "started") {
		t.Errorf("a started service shows the running glyph and started: %q", got)
	}
}

// TestLeavesFilterKeepsFormulaeAligned: the leaves filter narrows the typed
// formulae with the items, so the selected row is the selected package.
func TestLeavesFilterKeepsFormulaeAligned(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "2")
	m = updateModel(m, DataLoadedMsg{
		PanelID:  PanelFormulae,
		Items:    []string{"bash  5.2", "dep  1.0"},
		Formulae: []brew.Formula{{Name: "bash", Version: "5.2"}, {Name: "dep", Version: "1.0"}},
	})
	m = updateModel(m, MutationResultMsg{Name: "leaves", Leaves: []string{"bash"}})

	p := m.panels[PanelFormulae]
	if len(p.items) != 1 || len(p.formulae) != 1 {
		t.Fatalf("leaves should keep one row in items and formulae, got %d and %d", len(p.items), len(p.formulae))
	}
	if f := p.selectedFormula(); f == nil || f.Name != "bash" {
		t.Fatalf("the selected formula should be the leaf bash, got %+v", f)
	}
	m = updateModel(m, MutationResultMsg{Name: "leaves", Leaves: []string{"bash", "dep"}})
	p = m.panels[PanelFormulae]
	if len(p.items) != 2 || len(p.formulae) != 2 {
		t.Fatalf("leaves off should restore both lists, got %d and %d", len(p.items), len(p.formulae))
	}
}

// TestFormulaeSidebarKeepsSelectedInWindow (replaces
// TestFormulaeTableKeepsSelectedInWindow): with more formulae than the box
// shows, the selected one stays on screen and the first ones scroll away.
func TestFormulaeSidebarKeepsSelectedInWindow(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "2")
	fs := make([]brew.Formula, 30)
	items := make([]string, 30)
	for i := range fs {
		fs[i] = brew.Formula{Name: "pkg" + strconv.Itoa(i), Version: "1.0"}
		items[i] = fs[i].Name
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: items, Formulae: fs})
	for range 25 {
		m = sendKey(m, "j")
	}
	side := sgr.ReplaceAllString(m.renderSidebar(), "")
	selected := false
	for _, row := range strings.Split(side, "\n") {
		if strings.Contains(row, "pkg25 ") {
			selected = strings.Contains(row, "▸")
		}
	}
	if !selected {
		t.Fatalf("selected pkg25 should be in the sidebar window with its marker:\n%s", side)
	}
	if strings.Contains(side, "pkg0 ") {
		t.Fatalf("the window should have scrolled past pkg0:\n%s", side)
	}
}
