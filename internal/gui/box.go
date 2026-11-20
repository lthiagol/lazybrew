package gui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// panel draws one bordered box, w by h cells outer size, border included.
// The title sits inside the top border at the left, and label (already
// styled, possibly empty) at the right. A body line is cut to the inner width
// with an ellipsis, never wrapped; rows past the height are dropped. A
// focused panel has the accent border and title, an unfocused one the Border
// role and the subtext title.
func panel(title, label, body string, w, h int, focused bool) string {
	if w < 2 || h < 2 {
		return fitBox("", max(0, w), max(0, h))
	}
	edge := lipgloss.NewStyle().Foreground(style.BorderColor)
	titleStyle := style.SubtleText
	if focused {
		edge = lipgloss.NewStyle().Foreground(style.FocusColor)
		titleStyle = style.PanelTitleActive
	}

	// The top row reads ╭─ title ───── label ─╮. The title has priority and
	// keeps its whole name when it can; the label gets the room that is left
	// and is cut to fit. A box too narrow for even one label cell drops it.
	name := clipLine(title, w-7)
	tail := ""
	if room := w - 9 - lipgloss.Width(name); label != "" && room >= 1 {
		tail = " " + clipLine(label, room) + " "
	}
	dashes := max(0, w-6-lipgloss.Width(name)-lipgloss.Width(tail))
	top := edge.Render("╭─") + titleStyle.Render(" "+name+" ") +
		edge.Render(strings.Repeat("─", dashes)) + tail + edge.Render("─╮")

	inner := w - 2
	lines := strings.Split(body, "\n")
	rows := make([]string, 0, h)
	rows = append(rows, fitLine(top, w))
	for i := 0; i < h-2; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		rows = append(rows, edge.Render("│")+fitLine(line, inner)+edge.Render("│"))
	}
	rows = append(rows, edge.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return strings.Join(rows, "\n")
}

// clipLine cuts s to at most w cells, ending a cut line with an ellipsis.
// Styles survive the cut, and nothing is padded.
func clipLine(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	return lipgloss.NewStyle().MaxWidth(w-1).Render(s) + "…"
}

// fitLine makes s exactly w cells: cut with an ellipsis when it is wider,
// padded with spaces when it is narrower.
func fitLine(s string, w int) string {
	s = clipLine(s, w)
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
}
