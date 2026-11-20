package gui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

type helpPage struct {
	title    string
	sections []helpSection
}

type helpSection struct {
	heading string
	items   []helpItem
}

type helpItem struct {
	key  string
	desc string
}

// buildHelp documents the LIVE keys — every entry must have a handler in
// gui.go's Update, and every handled key should appear here (r1 finding 24).
func buildHelp() []helpPage {
	return []helpPage{
		{
			title: "Global",
			sections: []helpSection{
				{
					heading: "Navigation",
					items: []helpItem{
						{"Tab", "Next panel"},
						{"Shift+Tab", "Previous panel"},
						{"1-7", "Jump to panel"},
						{"j/k", "Scroll list"},
						{"PgUp/PgDn", "Scroll content by a page"},
						{"Home/End", "Top/bottom (g/G outside Search)"},
						{"[/]", "Switch tabs"},
					},
				},
				{
					heading: "Actions",
					items: []helpItem{
						{"R", "Refresh all data"},
						{"/", "Search packages"},
						{"C", "Toggle command log"},
						{"q", "Quit lazybrew"},
						{"?", "This help"},
					},
				},
			},
		},
		{
			title: "Formulae",
			sections: []helpSection{
				{
					items: []helpItem{
						{"i", "Install selected"},
						{"x", "Uninstall selected"},
						{"r", "Reinstall selected"},
						{"u", "Upgrade selected"},
						{"U", "Upgrade all"},
						{"p", "Pin/Unpin selected"},
						{"L", "Toggle leaves filter"},
						{"F", "Fetch selected"},
					},
				},
			},
		},
		{
			title: "Casks",
			sections: []helpSection{
				{
					items: []helpItem{
						{"i", "Install"},
						{"x", "Uninstall"},
						{"X", "Zap uninstall"},
						{"r", "Reinstall"},
						{"u", "Upgrade"},
						{"U", "Upgrade all"},
						{"F", "Fetch"},
					},
				},
			},
		},
		{
			title: "Taps",
			sections: []helpSection{
				{
					items: []helpItem{
						{"a", "Add tap"},
						{"x", "Remove tap"},
						{"t", "Trust config"},
						{"r", "Repair tap"},
					},
				},
			},
		},
		{
			title: "Services",
			sections: []helpSection{
				{
					items: []helpItem{
						{"s", "Start"},
						{"S", "Stop"},
						{"r", "Restart"},
						{"f", "Run (foreground)"},
						{"c", "Cleanup stale services"},
					},
				},
			},
		},
		{
			title: "Outdated",
			sections: []helpSection{
				{
					items: []helpItem{
						{"u", "Upgrade selected"},
						{"U", "Upgrade all"},
						{"Space", "Toggle selection"},
						{"a", "Select/deselect all"},
					},
				},
			},
		},
		{
			title: "Status",
			sections: []helpSection{
				{
					items: []helpItem{
						{"R", "Refresh all data"},
						{"c", "Run brew cleanup"},
						{"d", "Run brew doctor"},
						{"A", "Run brew autoremove"},
						{"B", "Brewfile actions"},
						{"v", "Vulnerability check"},
						{"m", "Missing dependencies"},
						{"[ ]", "Switch tabs"},
					},
				},
			},
		},
		{
			title: "Search",
			sections: []helpSection{
				{
					heading: "Actions",
					items: []helpItem{
						{"/", "Type search query"},
						{"Enter", "Execute search"},
						{"i", "Install selected result"},
						{"F", "Fetch selected result"},
					},
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
	keyStyle := lipgloss.NewStyle().Foreground(style.AccentColor).Bold(true)
	rows := []string{style.AccentText.Render(p.title)}
	for _, sec := range p.sections {
		rows = append(rows, "")
		if sec.heading != "" {
			rows = append(rows, style.AccentText.Render(sec.heading))
		}
		keyW := 0
		for _, it := range sec.items {
			keyW = max(keyW, lipgloss.Width(it.key))
		}
		for _, it := range sec.items {
			rows = append(rows, keyStyle.Render(fmt.Sprintf("%-*s", keyW, it.key))+
				"  "+style.SubtleText.Render(it.desc))
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

// helpRows lays the whole help out at width cells: every page, in one column
// or two when the width allows, one row per screen row.
func helpRows(width int) []string {
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
			cells = append(cells, fitLine(cell, colW))
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
	page := m.helpViewRows()
	switch key {
	case "j", "down":
		return m.scrollHelpTo(m.helpScroll + 1), true
	case "k", "up":
		return m.scrollHelpTo(m.helpScroll - 1), true
	case "pgdown":
		return m.scrollHelpTo(m.helpScroll + page), true
	case "pgup":
		return m.scrollHelpTo(m.helpScroll - page), true
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
