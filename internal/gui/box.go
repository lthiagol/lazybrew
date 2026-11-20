package gui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// borderTitle is a box's title in two forms: full, with the panel's icon, and
// bare, without it. When a count needs the room, the title drops to bare first.
type borderTitle struct {
	full, bare string
}

// titled is a title with no icon to drop.
func titled(s string) borderTitle {
	return borderTitle{full: s, bare: s}
}

// borderLabel is a box's right-hand label, both forms already styled. A label
// with a short form yields to the title: the full form when it fits, else the
// short form, else it is left out. A count is marked count: it is shown whole,
// and the title yields to it instead.
type borderLabel struct {
	full, short string
	count       bool
}

// panel draws one bordered box, w by h cells outer size, border included. The
// title sits inside the top border at the left, and the label at the right. A
// body line is cut to the inner width with an ellipsis, never wrapped; rows
// past the height are dropped. A focused panel has the accent border and
// title, an unfocused one the Border role and the subtext title.
func panel(title borderTitle, label borderLabel, body string, w, h int, focused bool) string {
	if w < 2 || h < 2 {
		return fitBox("", max(0, w), max(0, h))
	}
	edge := lipgloss.NewStyle().Foreground(style.BorderColor)
	titleStyle := style.SubtleText
	if focused {
		edge = lipgloss.NewStyle().Foreground(style.FocusColor)
		titleStyle = style.PanelTitleActive
	}

	name, tail := borderHead(title, label, w)
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

// borderHead picks the title and the label that share the top border of a
// w-cell box. The row reads ╭─ title ─── label ─╮, so the title and the label
// together take at most w-7 cells, and the dashes at least one.
//
// A count is never left out or cut: the title yields first its icon, then its
// name, which is cut with an ellipsis. Any other label keeps the title whole
// and shows its full form when that fits, else its short form, else nothing.
func borderHead(title borderTitle, label borderLabel, w int) (name, tail string) {
	if label.full == "" {
		return clipLine(title.full, w-7), ""
	}
	if label.count {
		labelW := lipgloss.Width(label.full) + 2
		for _, n := range []string{title.full, title.bare} {
			if lipgloss.Width(n)+labelW <= w-7 {
				return n, " " + label.full + " "
			}
		}
		return clipLine(title.bare, w-7-labelW), " " + label.full + " "
	}
	name = clipLine(title.full, w-7)
	room := w - 9 - lipgloss.Width(name)
	switch {
	case lipgloss.Width(label.full) <= room:
		return name, " " + label.full + " "
	case label.short != "" && lipgloss.Width(label.short) <= room:
		return name, " " + label.short + " "
	}
	return name, ""
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
