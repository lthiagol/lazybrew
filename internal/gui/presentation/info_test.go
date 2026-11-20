package presentation

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// TestInfoSplitsFromTheSplitWidth (v2-09 AC-03.1): at the split width the
// fields sit left and the description sits right, on the same first row; below
// it the view is one column and the description follows the fields.
func TestInfoSplitsFromTheSplitWidth(t *testing.T) {
	f := brew.Formula{Name: "bash", Version: "5.3", Tap: "homebrew/core", Description: "GNU Bourne-Again SHell"}

	wide := ansi.Strip(FormatFormulaInfo(f, splitWidth, style.UnicodeIcons))
	first := strings.Split(wide, "\n")[0]
	if !strings.Contains(first, "Name:") || !strings.Contains(first, "Description:") {
		t.Fatalf("at %d cells the fields and the description should share the first row, got %q", splitWidth, first)
	}
	for i, line := range strings.Split(wide, "\n") {
		if w := lipgloss.Width(line); w != splitWidth {
			t.Fatalf("split line %d is %d cells, want %d: %q", i, w, splitWidth, line)
		}
	}

	narrow := ansi.Strip(FormatFormulaInfo(f, splitWidth-1, style.UnicodeIcons))
	first = strings.Split(narrow, "\n")[0]
	if strings.Contains(first, "Description:") {
		t.Fatalf("below %d cells the view is one column, got first row %q", splitWidth, first)
	}
	if !strings.Contains(narrow, "Description:") {
		t.Fatalf("the single column should still show the description:\n%s", narrow)
	}
}

// TestInfoRightColumnWraps (v2-09 AC-03.2): a long description in the right
// column wraps with a hanging indent and is never cut.
func TestInfoRightColumnWraps(t *testing.T) {
	desc := strings.Repeat("settles the terminal behaviour of every command ", 3)
	f := brew.Formula{Name: "bash", Description: desc}
	out := ansi.Strip(FormatFormulaInfo(f, 100, style.UnicodeIcons))
	if strings.Contains(out, "…") {
		t.Fatalf("the description should wrap, not be cut, got:\n%s", out)
	}
	// The right column starts after the left one and the gap: 100*2/5 + 4 cells.
	var right []string
	for _, line := range strings.Split(ansi.Strip(FormatFormulaInfo(f, 100, style.UnicodeIcons)), "\n") {
		right = append(right, ansi.Cut(line, 44, 100))
	}
	if flat := strings.Join(strings.Fields(strings.Join(right, " ")), " "); !strings.Contains(flat, strings.Join(strings.Fields(desc), " ")) {
		t.Fatalf("every word of the description should be shown across its rows, got right column:\n%s", strings.Join(right, "\n"))
	}
}

// TestInfoKeepsTheFieldOrderInOneColumn: the single column keeps the order the
// fields are listed in, Right or not.
func TestInfoKeepsTheFieldOrderInOneColumn(t *testing.T) {
	fields := []Field{
		{Label: "Name", Value: "a"},
		{Label: "Right", Value: "b", Right: true},
		{Label: "Last", Value: "c"},
	}
	out := ansi.Strip(Info(fields, 60))
	if strings.Index(out, "Name:") > strings.Index(out, "Right:") || strings.Index(out, "Right:") > strings.Index(out, "Last:") {
		t.Fatalf("the single column should keep the listed order, got:\n%s", out)
	}
}
