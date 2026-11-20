package gui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// tableCol is one column of a main-area table. drop orders the columns that
// go when the table is too wide: the highest drop goes first, and a column with
// drop 0 is never dropped.
type tableCol struct {
	head string
	drop int
}

// tableRow is one line of a table: its cells, or a heading that spans the table
// in place of cells. A row with neither is a blank line.
type tableRow struct {
	heading string
	cells   []string
}

// tableGap is the space between two columns.
const tableGap = 2

// tableMinW is the narrowest a non-last column is cut to before the table is
// given up to the last column's cut.
const tableMinW = 4

// renderTable draws cols and rows as an aligned table at most w cells wide. The
// header row is Muted. Each column is as wide as its widest cell, and the last
// kept column takes the slack. Cells are cut with … and rows never wrap.
func renderTable(cols []tableCol, rows []tableRow, w int) string {
	widths := tableWidths(cols, rows, w)
	muted := lipgloss.NewStyle().Foreground(style.MutedColor)

	heads := make([]string, len(cols))
	for i, c := range cols {
		heads[i] = c.head
	}
	out := []string{tableLine(heads, widths, func(s string) string { return muted.Render(s) })}
	for _, r := range rows {
		switch {
		case r.heading != "":
			out = append(out, style.SubtleText.Render(cutCell(r.heading, w)))
		case len(r.cells) == 0:
			out = append(out, "")
		default:
			out = append(out, tableLine(r.cells, widths, func(s string) string { return s }))
		}
	}
	return strings.Join(out, "\n")
}

// tableWidths is the width of each column for a table w cells wide. A dropped
// column is 0. The columns keep their natural width; when that is too wide, the
// highest-drop columns go, and the wide non-last columns are cut to make room.
// The last kept column then takes what is left.
func tableWidths(cols []tableCol, rows []tableRow, w int) []int {
	nat := make([]int, len(cols))
	kept := make([]bool, len(cols))
	for i, c := range cols {
		nat[i] = lipgloss.Width(c.head)
		for _, r := range rows {
			if i < len(r.cells) {
				nat[i] = max(nat[i], lipgloss.Width(r.cells[i]))
			}
		}
		kept[i] = true
	}

	for natural(nat, kept) > w {
		drop := -1
		for i, c := range cols {
			if kept[i] && c.drop > 0 && (drop < 0 || c.drop > cols[drop].drop) {
				drop = i
			}
		}
		if drop < 0 {
			break
		}
		kept[drop] = false
	}

	last := -1
	widths := make([]int, len(cols))
	for i := range cols {
		if kept[i] {
			widths[i] = nat[i]
			last = i
		}
	}
	if last < 0 {
		return widths
	}

	// Everything but the last column must leave it one cell, so cut the widest
	// of the others until it does, never below tableMinW.
	for others(widths, kept, last)+gaps(kept) > w-1 {
		widest := -1
		for i := range widths {
			if kept[i] && i != last && widths[i] > tableMinW && (widest < 0 || widths[i] > widths[widest]) {
				widest = i
			}
		}
		if widest < 0 {
			break
		}
		widths[widest]--
	}
	widths[last] = max(1, w-others(widths, kept, last)-gaps(kept))
	return widths
}

// natural is the width of the kept columns at their natural widths, gaps
// included.
func natural(nat []int, kept []bool) int {
	total := 0
	for i, n := range nat {
		if kept[i] {
			total += n
		}
	}
	return total + gaps(kept)
}

// others is the width of the kept columns but last.
func others(widths []int, kept []bool, last int) int {
	total := 0
	for i, n := range widths {
		if kept[i] && i != last {
			total += n
		}
	}
	return total
}

// gaps is the space between the kept columns.
func gaps(kept []bool) int {
	n := 0
	for _, k := range kept {
		if k {
			n++
		}
	}
	return tableGap * max(0, n-1)
}

// tableLine is one row of cells, each cut or padded to its column, with the
// gap between the kept columns. paint styles each cell's text.
func tableLine(cells []string, widths []int, paint func(string) string) string {
	var parts []string
	for i, w := range widths {
		if w == 0 {
			continue
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		cut := cutCell(cell, w)
		parts = append(parts, paint(cut)+strings.Repeat(" ", w-lipgloss.Width(cut)))
	}
	return strings.Join(parts, strings.Repeat(" ", tableGap))
}

// cutCell cuts s to w cells with …. Styles survive the cut.
func cutCell(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return ansi.Truncate(s, w, "…")
}
