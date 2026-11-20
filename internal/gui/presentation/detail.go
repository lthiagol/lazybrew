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
// columns (Q-v2-09-a). The main box's inner width is 84 at this content width:
// its padding takes one cell either side.
const splitWidth = 82

// infoGap is the space between the two columns of a split Info view.
const infoGap = 4

// infoMinRight is the narrowest the right column of a split Info view is kept
// at, so the prose beside a long field list stays readable.
const infoMinRight = 26

// Info draws an Info view within width cells. Narrower than splitWidth it is
// one column, the fields in order. Wider, the fields not marked Right sit in a
// left column and the Right ones in a right column, each keeping its order. A
// view splits only when every non-prose value in the right column fits that
// column whole (F-r4-01); otherwise it keeps the single column.
func Info(fields []Field, width int) string {
	if width < splitWidth {
		return Detail(fields, width)
	}
	var left, right []Field
	for _, f := range fields {
		if f.Right {
			right = append(right, f)
		} else {
			left = append(left, f)
		}
	}
	// The left column is as wide as its fields need, but at least two fifths of
	// the width, and it leaves the right column infoMinRight cells.
	leftW := min(max(width*2/5, naturalWidth(left)), width-infoGap-infoMinRight)
	rightW := width - leftW - infoGap
	if !fitsWhole(right, rightW) {
		return Detail(fields, width)
	}
	// The right column is prose and lists: its values wrap, never cut.
	for i := range right {
		right[i].Prose = true
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(Detail(left, leftW)),
		lipgloss.NewStyle().Width(infoGap).Render(""),
		lipgloss.NewStyle().Width(rightW).Render(Detail(right, rightW)))
}

// fitsWhole is whether every non-prose value of fields fits a column of width
// cells whole: no word of it is wider than the value column the labels leave.
// A path, a URL or a command is one such word, and a word that does not fit
// would break mid-token when drawn. Prose (Field.Prose) wraps at spaces, so it
// is not checked.
func fitsWhole(fields []Field, width int) bool {
	labelW := 0
	for _, f := range fields {
		if f.Value != "" {
			labelW = max(labelW, lipgloss.Width(f.Label+":"))
		}
	}
	valueW := width - (labelW + 2)
	for _, f := range fields {
		if f.Prose || f.Value == "" {
			continue
		}
		for _, word := range strings.Fields(f.Value) {
			if lipgloss.Width(word) > valueW {
				return false
			}
		}
	}
	return true
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
	muted := style.MutedText

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

// naturalWidth is the width Detail(fields, 0) draws, measured without drawing
// it: the widest row of the unwrapped fields. A value is drawn as one block, its
// lines padded to the widest and the first one after the label. lipgloss
// draws a tab as tabStop spaces.
func naturalWidth(fields []Field) int {
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
		return 0
	}
	labelW += 2
	w := 0
	for _, f := range shown {
		block := 0
		for _, line := range strings.Split(f.Value, "\n") {
			block = max(block, lipgloss.Width(strings.ReplaceAll(line, "\t", strings.Repeat(" ", tabStop))))
		}
		w = max(w, labelW+block)
	}
	return w
}

// tabStop is the width lipgloss draws a tab at.
const tabStop = 4
