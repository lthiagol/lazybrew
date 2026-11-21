package gui

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// all is the row's keys in the order the footer lists them. The tests read the
// row flat; the footer itself keeps the two groups apart.
func (r footerRow) all() []keyHint {
	return append(append([]keyHint{}, r.panel...), r.global...)
}

// footerPanel and footerGlobal are a representative row: Formulae's own actions,
// then the global keys with the arrow hint in front (lb29 AC-02 makes that hint
// real; the renderer is tested on the shape it will have).
var footerPanel = []keyHint{fInstall, fRemove, fReinstall, keyUpgradeSelected, keyUpgradeAll, fPin, fLeaves, fFetch, keyCopy}

var footerGlobal = []keyHint{
	{key: "←↑↓→", keys: "up down left right", desc: "navigate"},
	keyNext, keyJump, keySearch, keyRefresh, keyLog, keyHelp, keyQuit,
}

// footerWidths is every width the footer is checked at: the narrowest a
// terminal the TUI supports draws, to a width that holds the whole row.
func footerWidths() []int {
	var ws []int
	for w := 60; w <= 200; w++ {
		ws = append(ws, w)
	}
	return ws
}

// TestFooterFitsAtEveryWidth: at each width the row is no wider than the width,
// it never ends with a separator, and it says "? more" exactly when some key was
// dropped, in place of "? help".
func TestFooterFitsAtEveryWidth(t *testing.T) {
	for _, w := range footerWidths() {
		row, more := fitFooter(footerRow{panel: footerPanel, global: footerGlobal}, w)
		line := renderFooter(row, more)
		if got := lipgloss.Width(line); got > w {
			t.Errorf("width %d: the row is %d cells wide", w, got)
		}
		plain := strings.TrimRight(plainLine(line), " ")
		if strings.HasSuffix(plain, "·") || strings.HasSuffix(plain, "│") {
			t.Errorf("width %d: the row ends with a separator: %q", w, plain)
		}
		if more && !strings.HasSuffix(plain, "? more") {
			t.Errorf("width %d: keys were dropped but the row does not end with ? more: %q", w, plain)
		}
		if more && strings.Contains(plain, "? help") {
			t.Errorf("width %d: keys were dropped but ? help is still shown: %q", w, plain)
		}
		if !more && !strings.Contains(plain, "? help") {
			t.Errorf("width %d: nothing was dropped but ? help is not shown: %q", w, plain)
		}
		if !more && len(row.panel) != len(footerPanel) {
			t.Errorf("width %d: panel actions dropped without a drop", w)
		}
	}
}

// TestFooterDropsInTheStatedOrder holds the drops to the lb29 rules, narrowing
// from 200 to 60 cells: panel actions go from the right end, and at least the
// first three stay while a droppable global key remains; the droppable global
// keys go in footerDropOrder; the arrow hint and ? help never go.
func TestFooterDropsInTheStatedOrder(t *testing.T) {
	droppable := 0
	for _, key := range footerDropOrder {
		if slices.ContainsFunc(footerGlobal, func(h keyHint) bool { return h.key == key }) {
			droppable++
		}
	}
	if droppable != len(footerDropOrder) {
		t.Fatalf("footerDropOrder names %d keys, the row has %d of them", len(footerDropOrder), droppable)
	}

	for w := 200; w >= 60; w-- {
		row, _ := fitFooter(footerRow{panel: footerPanel, global: footerGlobal}, w)

		if !slices.Equal(row.panel, footerPanel[:len(row.panel)]) {
			t.Errorf("width %d: the panel actions shown are not a prefix of the list: %v", w, row.panel)
		}
		if !slices.Contains(row.global, footerGlobal[0]) {
			t.Errorf("width %d: the arrow hint was dropped", w)
		}

		// Removed global keys are a prefix of footerDropOrder: once one is
		// shown, every later one is shown too.
		shown := false
		for _, key := range footerDropOrder {
			present := slices.ContainsFunc(row.global, func(h keyHint) bool { return h.key == key })
			if shown && !present {
				t.Errorf("width %d: %q was dropped before a key after it in footerDropOrder", w, key)
			}
			shown = shown || present
		}

		anyDroppable := shown
		if anyDroppable && len(row.panel) < min(footerMinPanel, len(footerPanel)) {
			t.Errorf("width %d: %d panel actions shown while a global key remains, want at least %d",
				w, len(row.panel), footerMinPanel)
		}
	}
}

// TestFooterSeparatorsAreMuted: the group rule and the dot between hints are
// drawn in the muted role on every theme; keys in the accent.
func TestFooterSeparatorsAreMuted(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for _, th := range []*style.Theme{style.DarkTheme(), style.LightTheme()} {
		style.ApplyTheme(th)
		line := renderFooter(footerRow{panel: footerPanel, global: footerGlobal}, false)
		if !strings.Contains(line, sgrColor("38", th.Muted)+"m·") {
			t.Errorf("dot separator is not in the muted role")
		}
		if !strings.Contains(line, sgrColor("38", th.Muted)+"m│") {
			t.Errorf("group rule is not in the muted role")
		}
		if !strings.Contains(line, sgrColor("38", th.Accent)+"m") {
			t.Errorf("keys are not in the accent role")
		}
	}
}

// TestFooterNeverClipsAPanel (lb29 AC-03): on each panel at 80, 100 and 140 columns
// the row the renderer draws fits the width, so the frame cuts no hint, and the row
// ends with a whole hint: "? more", or the last key the row kept.
func TestFooterNeverClipsAPanel(t *testing.T) {
	digits := map[PanelID]string{
		PanelStatus: "1", PanelFormulae: "2", PanelCasks: "3", PanelOutdated: "4",
		PanelTaps: "5", PanelServices: "6", PanelSearch: "7",
	}
	for id := PanelStatus; id <= PanelSearch; id++ {
		for _, w := range []int{80, 100, 140} {
			m := sendKey(updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: 30}), digits[id])
			fit, more := fitFooter(m.footerHints(), w)
			line := renderFooter(fit, more)
			if got := lipgloss.Width(line); got > w {
				t.Errorf("%s at %d columns: the row is %d cells, so the frame cuts it", id, w, got)
			}
			want := footerMore.key + " " + footerMore.desc
			if !more {
				last := fit.all()[len(fit.all())-1]
				want = last.key + " " + last.desc
			}
			if plain := strings.TrimRight(plainLine(line), " "); !strings.HasSuffix(plain, want) {
				t.Errorf("%s at %d columns: the row %q does not end with %q", id, w, plain, want)
			}
		}
	}
}

// TestFooterNeverRepeatsAKey (lb29 r2 F-r1-04): at 140 columns no key shows twice
// in the row, on any panel, with or without a finished operation over it.
func TestFooterNeverRepeatsAKey(t *testing.T) {
	for id := PanelStatus; id <= PanelSearch; id++ {
		for _, done := range []bool{false, true} {
			m := newTestModel()
			m.activePanel = id
			if done {
				m.opState = &Operation{Status: opSuccess}
			}
			row, _ := fitFooter(m.footerHints(), 140)
			seen := map[string]bool{}
			for _, h := range row.all() {
				if seen[h.key] {
					t.Errorf("%s (operation done %v): %q shows twice in the row at 140 columns", id, done, h.key)
				}
				seen[h.key] = true
			}
		}
	}
}
