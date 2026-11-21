package gui

import (
	"slices"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

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
