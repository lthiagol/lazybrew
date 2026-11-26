package modal

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
)

// Widths of the modal frames. A confirm dialog is min(frame-8, 72) wide; the
// app draws no frame narrower than 80 columns, so it is 72 in every frame.
const (
	confirmWidth = 72
	formWidth    = 50
)

// dialogMaxRows is the tallest a dialog may be. The app draws no frame shorter
// than 24 rows, and the header and footer take two, so 20 always fits.
const dialogMaxRows = 20

// dialogBodyRows is how many rows of the box the body may take. The chip, two
// blank rows and the key line take 4 rows; with the border and padding that is
// 8 rows of the box that are not body.
const dialogBodyRows = dialogMaxRows - 8

// innerWidth is the text width of a dialog box: the width less the thick border
// (2) and the padding of 3 on each side.
func innerWidth(width int) int {
	return width - 2 - 6
}

// wrapProse wraps prose to the text width of a box of the given width, so a
// long question or consequence runs onto more lines but never past the border.
func wrapProse(s string, width int) string {
	return lipgloss.Wrap(s, innerWidth(width), "")
}

// proseLines wraps prose and styles each line on its own. Styling the whole
// paragraph would pad every line to the longest, leaving the short ones
// stranded at the left; a line styled alone keeps its own width to centre by.
func proseLines(s string, st lipgloss.Style, width int) []string {
	var out []string
	for _, line := range strings.Split(wrapProse(s, width), "\n") {
		out = append(out, st.Render(line))
	}
	return out
}

// leftBlock pads each line to the widest of them. Centred row by row, the
// padded lines keep their starts in one column, so the block reads as a
// left-aligned block centred in the box, with its indents kept.
func leftBlock(lines []string) []string {
	widest := 0
	for _, line := range lines {
		widest = max(widest, lipgloss.Width(line))
	}
	return padBlock(lines, widest)
}

// padBlock pads each line to width cells, which must be no narrower than any
// of them. A caller whose lines change under a fixed width (a scrolling list)
// passes that width, so the block keeps its column.
func padBlock(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, lipgloss.NewStyle().Width(width).Render(line))
	}
	return out
}

// cutRow is the ellipsis that replaces the last body row a dialog cannot hold.
// It keeps that row's indent and is padded to its width, so when each row is
// centred on its own it lands in the column of the row it replaces, not alone
// in the middle of the box.
func cutRow(line string) string {
	plain := ansi.Strip(line)
	row := strings.Repeat(" ", len(plain)-len(strings.TrimLeft(plain, " "))) + "…"
	return row + strings.Repeat(" ", max(lipgloss.Width(line)-lipgloss.Width(row), 0))
}

// dialog frames a modal the change-gate way: a thick border in accent, the
// action as a chip on top, then the body, then the key line. Width counts the
// border and padding (v2), so the box is exactly width cells. The body is the
// modal's prose, already wrapped; it is trimmed to fit dialogMaxRows, its last
// line becoming an ellipsis when it is.
func dialog(accent color.Color, label string, width int, body []string, keys string) string {
	chip := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.BgColor).
		Background(accent).
		Render(textutil.Clip(" "+strings.ToUpper(label)+" ", innerWidth(width)))

	lines := strings.Split(strings.Join(body, "\n"), "\n")
	if len(lines) > dialogBodyRows {
		lines = lines[:dialogBodyRows]
		lines[dialogBodyRows-1] = cutRow(lines[dialogBodyRows-1])
	}

	// Each row is centred on its own, so a wrapped line and a short one both
	// sit in the middle of the box.
	inner := innerWidth(width)
	parts := append([]string{chip, ""}, lines...)
	rows := make([]string, 0, len(parts)+2)
	for _, row := range append(parts, "", keys) {
		rows = append(rows, lipgloss.PlaceHorizontal(inner, lipgloss.Center, row))
	}
	content := strings.Join(rows, "\n")
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Padding(1, 3).
		Border(lipgloss.ThickBorder()).
		BorderForeground(accent).
		Render(content)
}

// keyHint is one key and what it does: the key in accent, the words subtle.
func keyHint(accent color.Color, key, desc string) string {
	return lipgloss.NewStyle().Foreground(accent).Bold(true).Render(key) +
		style.SubtleText.Render(" "+desc)
}

// hintViews draws each hint as a key hint, for a key line.
func hintViews(accent color.Color, hints []Hint) []string {
	out := make([]string, 0, len(hints))
	for _, h := range hints {
		out = append(out, keyHint(accent, h.Key, h.Desc))
	}
	return out
}

// keyLine joins key hints with the muted dot the footer uses.
func keyLine(hints ...string) string {
	return strings.Join(hints, " "+style.MutedText.Render("·")+" ")
}
