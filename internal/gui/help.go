package gui

import (
	"fmt"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
)

type helpPage struct {
	title    string
	sections []helpSection
}

type helpSection struct {
	heading string
	items   []keyHint
}

// buildHelp documents the LIVE keys. Each page lists its entries of the key
// tables (keybindings.go), so a key that is answered is listed, and
// a listed key is answered (keymap_test.go).
func buildHelp() []helpPage {
	return []helpPage{
		{
			title: "Global",
			sections: []helpSection{
				{
					heading: "Navigation",
					items:   []keyHint{keyNext, keyPrev, keyJump, keyScroll, keyEnds, keyTabs},
				},
				{
					heading: "Actions",
					items:   []keyHint{keyRefresh, keySearch, keyLog, keyQuit, keyHelp},
				},
			},
		},
		{
			title: "Formulae",
			sections: []helpSection{
				{
					items: []keyHint{keyFilter, fInstall, fRemove, fReinstall, keyUpgradeSelected, keyUpgradeAll, fPin, fLeaves, fFetch, keyCopy},
				},
			},
		},
		{
			title: "Casks",
			sections: []helpSection{
				{
					items: []keyHint{keyFilter, cInstall, cRemove, cZap, cReinstall, cUpgrade, keyUpgradeAll, cFetch, keyCopy},
				},
			},
		},
		{
			title: "Taps",
			sections: []helpSection{
				{
					items: []keyHint{tAdd, tUntap, tTrust, tTrustItem, tRepair, keyCopy},
				},
			},
		},
		{
			title: "Services",
			sections: []helpSection{
				{
					items: []keyHint{svStart, svStop, svRestart, svRun, svCleanup, svLogStream, keyCopy},
				},
			},
		},
		{
			title: "Outdated",
			sections: []helpSection{
				{
					items: []keyHint{keyUpgradeSelected, keyUpgradeAll, oSelect, oSelectAll, keyCopy},
				},
			},
		},
		{
			title: "Status",
			sections: []helpSection{
				{
					items: []keyHint{keyRefresh, stCleanup, stDoctor, stAutoremove, stBrewfile, stVulns, stMissing, stTabs, keyCopy},
				},
			},
		},
		{
			title: "Search",
			sections: []helpSection{
				{
					heading: "Actions",
					items:   []keyHint{sType, sExecute, sInstall, sUninstall, sZap, sFetch},
				},
			},
		},
	}
}

// helpIndent is the margin the help text keeps inside its box, on each side.
const helpIndent = 2

// helpColumnsMin is the text width from which help lays out in two columns.
const helpColumnsMin = 80

// helpColumnGap is the space between two help columns.
const helpColumnGap = 4

// helpUnit is one page as rows: its title, then each section with its heading
// and items. A section's key column is as wide as its widest key.
func helpUnit(p helpPage) []string {
	keyStyle := style.AccentText
	rows := []string{style.AccentText.Render(p.title)}
	for _, sec := range p.sections {
		rows = append(rows, "")
		if sec.heading != "" {
			rows = append(rows, style.AccentText.Render(sec.heading))
		}
		keyW := 0
		for _, it := range sec.items {
			keyW = max(keyW, lipgloss.Width(it.helpKeyText()))
		}
		for _, it := range sec.items {
			rows = append(rows, keyStyle.Render(fmt.Sprintf("%-*s", keyW, it.helpKeyText()))+
				"  "+style.SubtleText.Render(it.help))
		}
	}
	return rows
}

// helpColumns flows whole units into cols columns, balanced by height. A unit
// is never split, so a heading always sits with its rows.
func helpColumns(units [][]string, cols int) [][]string {
	total := 0
	for _, u := range units {
		total += len(u) + 1
	}
	target := (total + cols - 1) / cols
	columns := make([][]string, cols)
	c := 0
	for _, u := range units {
		if c < cols-1 && len(columns[c]) > 0 && len(columns[c])+len(u) > target {
			c++
		}
		if len(columns[c]) > 0 {
			columns[c] = append(columns[c], "")
		}
		columns[c] = append(columns[c], u...)
	}
	return columns
}

// helpMemo is the last help layout. The help is laid out again only when the
// width or the theme changes, not on every frame and scroll key (cr3 AC-07).
var helpMemo struct {
	mu    sync.Mutex
	width int
	theme *style.Theme
	rows  []string
}

// helpRows is the whole help laid out at width cells: every page, in one column
// or two when the width allows, one row per screen row. The rows are shared:
// a caller reads them and does not change them.
func helpRows(width int) []string {
	helpMemo.mu.Lock()
	defer helpMemo.mu.Unlock()
	if helpMemo.rows != nil && helpMemo.width == width && helpMemo.theme == style.CurrentTheme {
		return helpMemo.rows
	}
	rows := layoutHelp(width)
	helpMemo.width, helpMemo.theme, helpMemo.rows = width, style.CurrentTheme, rows
	return rows
}

// layoutHelp lays the help out at width: see helpRows.
func layoutHelp(width int) []string {
	var units [][]string
	for _, p := range buildHelp() {
		units = append(units, helpUnit(p))
	}
	cols, colW := 1, width
	if width >= helpColumnsMin {
		cols, colW = 2, (width-helpColumnGap)/2
	}
	columns := helpColumns(units, cols)

	height := 0
	for _, col := range columns {
		height = max(height, len(col))
	}
	rows := make([]string, height)
	for i := range rows {
		var cells []string
		for _, col := range columns {
			cell := ""
			if i < len(col) {
				cell = col[i]
			}
			cells = append(cells, textutil.Fit(cell, colW))
		}
		rows[i] = strings.Join(cells, strings.Repeat(" ", helpColumnGap))
	}
	return rows
}

// helpTextWidth is the width the help text is laid out at: the box less its
// border and its margins.
func (m Model) helpTextWidth() int {
	return max(1, m.width-2-2*helpIndent)
}

// helpViewRows is how many help rows the box shows: its height less the two
// border rows. The footer is the row below the box.
func (m Model) helpViewRows() int {
	return max(1, m.height-1-2)
}

// helpLast is the largest scroll offset that still fills the help box.
func (m Model) helpLast() int {
	return max(0, len(helpRows(m.helpTextWidth()))-m.helpViewRows())
}

// scrollHelpTo sets the help's scroll offset, held between the first and the
// last full page.
func (m Model) scrollHelpTo(row int) Model {
	m.helpScroll = min(max(0, row), m.helpLast())
	return m
}

// helpKey is the help screen's own keys, taken while it is open: the scroll
// keys move through the text, and ? or esc close it. Any other key passes
// through unchanged. It reports whether it took the key.
func (m Model) helpKey(key string) (Model, bool) {
	switch key {
	case "j", "down":
		return m.scrollHelpTo(m.helpScroll + 1), true
	case "k", "up":
		return m.scrollHelpTo(m.helpScroll - 1), true
	case "home", "g":
		return m.scrollHelpTo(0), true
	case "end", "G":
		return m.scrollHelpTo(m.helpLast()), true
	case "?", "esc":
		m.showHelp = false
		return m, true
	}
	return m, false
}

// renderHelpScreen is the help screen: the scrollable help in a box that fills
// the screen above the footer. The box's right label gives the rows on show,
// as "12–34 of 61".
func (m Model) renderHelpScreen() string {
	rows := helpRows(m.helpTextWidth())
	view := m.helpViewRows()
	off := min(m.helpScroll, max(0, len(rows)-view))
	shown := make([]string, 0, view)
	for _, row := range rows[off:min(off+view, len(rows))] {
		shown = append(shown, strings.Repeat(" ", helpIndent)+row)
	}
	hint := fmt.Sprintf("%d–%d of %d", off+1, off+len(shown), len(rows))
	box := panel(titled("Help"), borderLabel{full: style.SubtleText.Render(hint), count: true},
		strings.Join(shown, "\n"), m.width, m.height-1, true)
	return fitFrame(box+"\n"+m.renderBottomBar(), m.width, m.height)
}
