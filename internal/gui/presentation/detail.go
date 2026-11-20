package presentation

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// Field is one row of a detail view. Role colours the value; nil draws it in
// the Text role. Prose marks long text (a description, caveats), which wraps
// under the value column instead of being cut.
type Field struct {
	Label string
	Value string
	Role  color.Color
	Prose bool
	// Right moves the field to the right column of a wide Info view (see Info).
	Right bool
}

// splitWidth is the content width from which an Info view splits into two
// columns. The main box's inner width is 100 at this content width: its
// padding takes one cell either side.
const splitWidth = 98

// infoGap is the space between the two columns of a split Info view.
const infoGap = 4

// Info draws an Info view within width cells. Narrower than splitWidth it is
// one column, the fields in order. Wider, the fields not marked Right sit in a
// left column and the Right ones (prose and lists, which wrap) in a right
// column, each keeping its order.
func Info(fields []Field, width int) string {
	if width < splitWidth {
		return Detail(fields, width)
	}
	var left, right []Field
	for _, f := range fields {
		if f.Right {
			// The right column is prose and lists: its values wrap, never cut.
			f.Prose = true
			right = append(right, f)
		} else {
			left = append(left, f)
		}
	}
	leftW := width * 2 / 5
	rightW := width - leftW - infoGap
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(Detail(left, leftW)),
		lipgloss.NewStyle().Width(infoGap).Render(""),
		lipgloss.NewStyle().Width(rightW).Render(Detail(right, rightW)))
}

// Detail draws the fields of one detail view within width cells, one row per
// field. A field with an empty value is left out, so no label is ever left
// without a value. The labels share one column (the widest label plus a gap)
// drawn in the Muted role. A value too long for its row is cut with …; a
// prose value wraps with a hanging indent. A width of 0 or less neither cuts
// nor wraps, for a caller that cuts the finished text itself.
func Detail(fields []Field, width int) string {
	var shown []Field
	labelW := 0
	for _, f := range fields {
		if f.Value == "" {
			continue
		}
		shown = append(shown, f)
		labelW = max(labelW, lipgloss.Width(f.Label+":"))
	}
	if len(shown) == 0 {
		return ""
	}
	labelW += 2
	valueW := max(1, width-labelW)
	indent := strings.Repeat(" ", labelW)
	muted := lipgloss.NewStyle().Foreground(style.MutedColor)

	var rows []string
	for _, f := range shown {
		role := f.Role
		if role == nil {
			role = style.TextColor
		}
		value := lipgloss.NewStyle().Foreground(role)

		var lines []string
		switch {
		case width <= 0:
			lines = []string{f.Value}
		case f.Prose:
			lines = strings.Split(ansi.Wrap(f.Value, valueW, ""), "\n")
		default:
			lines = []string{ansi.Truncate(f.Value, valueW, "…")}
		}
		for i, line := range lines {
			lead := indent
			if i == 0 {
				lead = muted.Render(fmt.Sprintf("%-*s", labelW, f.Label+":"))
			}
			rows = append(rows, lead+value.Render(line))
		}
	}
	return strings.Join(rows, "\n")
}
