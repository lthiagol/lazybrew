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

// tableMinUseful is the narrowest a column is kept at. A column that cannot get
// this many cells goes, rather than being kept as a sliver; a column already
// narrower than this keeps its natural width. So a long description is cut with
// … to the slack, and is dropped only when even its minimum does not fit.
const tableMinUseful = 12

// renderTable draws cols and rows as an aligned table at most w cells wide. The
// header row is Muted. Each column is as wide as its widest cell, and the last
// kept column takes the slack. Cells are cut with … and rows never wrap.
func renderTable(cols []tableCol, rows []tableRow, w int) string {
	widths := tableWidths(cols, rows, w)
	muted := style.MutedText

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
// column is 0. The columns keep their natural width. When that is too wide, the
// highest-drop columns go while the rest cannot keep their floors; then the wide
// non-last columns are cut down to their floors, so the last column is never
// narrower than its own floor. The last kept column takes what is left.
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

	for minimal(nat, kept) > w {
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

	// Every kept column keeps at least its floor, min(natural, tableMinUseful).
	// Cut the widest of the others down to their floors until the last column
	// has its own floor; a drop above has already made that possible.
	floor := make([]int, len(cols))
	for i := range cols {
		if kept[i] {
			floor[i] = min(nat[i], tableMinUseful)
		}
	}
	cutOthers(widths, kept, last, w-floor[last], floor)
	// The floors cannot all be met (no column left to drop): cut the others
	// further, to tableMinW, rather than let the last column go below its floor.
	lowest := make([]int, len(cols))
	for i := range lowest {
		lowest[i] = tableMinW
	}
	cutOthers(widths, kept, last, w-floor[last], lowest)
	widths[last] = max(1, w-others(widths, kept, last)-gaps(kept))
	return widths
}

// cutOthers cuts the widest kept column other than last by one cell at a time,
// until the others and their gaps fit in room. A column is not cut below its
// entry in low.
func cutOthers(widths []int, kept []bool, last, room int, low []int) {
	for others(widths, kept, last)+gaps(kept) > room {
		widest := -1
		for i := range widths {
			if kept[i] && i != last && widths[i] > low[i] && (widest < 0 || widths[i] > widths[widest]) {
				widest = i
			}
		}
		if widest < 0 {
			return
		}
		widths[widest]--
	}
}

// minimal is the width of the kept columns, each at the least it may be kept at
// (its natural width, or tableMinUseful when that is narrower), gaps included.
func minimal(nat []int, kept []bool) int {
	total := 0
	for i, n := range nat {
		if kept[i] {
			total += min(n, tableMinUseful)
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
