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
	// The right column starts where the description label is on the first row.
	lines := strings.Split(ansi.Strip(FormatFormulaInfo(f, 100, style.UnicodeIcons)), "\n")
	start := ansi.StringWidth(lines[0][:strings.Index(lines[0], "Description:")])
	var right []string
	for _, line := range lines {
		right = append(right, ansi.Cut(line, start, 100))
	}
	if flat := strings.Join(strings.Fields(strings.Join(right, " ")), " "); !strings.Contains(flat, strings.Join(strings.Fields(desc), " ")) {
		t.Fatalf("every word of the description should be shown across its rows, got right column:\n%s", strings.Join(right, "\n"))
	}
}

// TestInfoKeepsAHomepageWholeAtTheSplitWidth (LB-51 FC4): a homepage the left
// column cannot hold whole drops the view to one column, so the URL is not cut.
func TestInfoKeepsAHomepageWholeAtTheSplitWidth(t *testing.T) {
	const homepage = "https://github.com/example-org/some-long-project-name-here"
	f := brew.Formula{Name: "some-project", Version: "1.0", Tap: "homebrew/core",
		Homepage: homepage, Description: "a formula with a long homepage"}
	out := ansi.Strip(FormatFormulaInfo(f, splitWidth, style.UnicodeIcons))
	if !strings.Contains(out, homepage) {
		t.Fatalf("the homepage should be whole at %d cells, got:\n%s", splitWidth, out)
	}
	if first := strings.Split(out, "\n")[0]; strings.Contains(first, "Description:") {
		t.Fatalf("the view should drop to one column, first row %q", first)
	}
}

// TestInfoKeepsAMultiWordLicenseWholeAtTheSplitWidth (LB-51 D2): a left value
// with spaces that is wider than its column is not cut with … when the view
// splits; each word fits, but the whole value does not, so the view drops to one
// column where the license is whole.
func TestInfoKeepsAMultiWordLicenseWholeAtTheSplitWidth(t *testing.T) {
	const license = "Apache-2.0 WITH LLVM-exception OR GPL-2.0-or-later"
	f := brew.Formula{Name: "llvm", Version: "19.1.7", Tap: "homebrew/core", License: license,
		Description: "compiler infrastructure"}
	out := ansi.Strip(FormatFormulaInfo(f, splitWidth, style.UnicodeIcons))
	if !strings.Contains(out, license) {
		t.Fatalf("the license should be whole at %d cells, got:\n%s", splitWidth, out)
	}
	if first := strings.Split(out, "\n")[0]; strings.Contains(first, "Description:") {
		t.Fatalf("the view should drop to one column, first row %q", first)
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

// TestFormulaInfoStillSplitsAt87 (v2-09 F-r4-01): the Description is prose, so a
// Formulae Info at 87 cells still splits, with the description beside the name.
func TestFormulaInfoStillSplitsAt87(t *testing.T) {
	f := brew.Formula{Name: "bash", Version: "5.3.20", Tap: "homebrew/core", Description: "GNU Bourne-Again SHell"}
	first := strings.Split(ansi.Strip(FormatFormulaInfo(f, 87, style.UnicodeIcons)), "\n")[0]
	if !strings.Contains(first, "Name:") || !strings.Contains(first, "Description:") {
		t.Fatalf("at 87 cells the Formulae Info should split, first row %q", first)
	}
}

// TestInfoFallsBackWhenAPathWouldBreak (v2-09 F-r4-01): a non-prose value whose
// word is wider than the right column would break mid-token, so the view keeps
// its single column and the path is whole on one row.
func TestInfoFallsBackWhenAPathWouldBreak(t *testing.T) {
	const path = "/home/linuxbrew/.linuxbrew/opt/some-long-formula/bin/some-long-formula"
	fields := []Field{
		{Label: "Name", Value: "a"},
		{Label: "Path", Value: path, Right: true},
	}
	out := ansi.Strip(Info(fields, 100))
	if !strings.Contains(out, path) {
		t.Fatalf("the path should be whole on one row, got:\n%s", out)
	}
	if first := strings.Split(out, "\n")[0]; strings.Contains(first, "Name:") && strings.Contains(first, "Path:") {
		t.Fatalf("the view should keep its single column, first row %q", first)
	}
}
