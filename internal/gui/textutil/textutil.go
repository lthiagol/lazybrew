// Package textutil cuts and pads single lines of styled text to a width in
// terminal cells. ANSI styles survive a cut, and wide runes count as two cells.
package textutil

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Clip cuts s to at most w cells, ending a cut line with an ellipsis. A width of
// zero or less gives "", and a width of one gives the ellipsis alone. Nothing is
// padded.
func Clip(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return ansi.Truncate(s, w, "…")
}

// Fit makes s exactly w cells: Clip when it is wider, padded with spaces when it
// is narrower.
func Fit(s string, w int) string {
	s = Clip(s, w)
	return s + strings.Repeat(" ", max(0, w-ansi.StringWidth(s)))
}
