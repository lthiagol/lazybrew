package gui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestTapInfoSplitsWhenWide (v2-09 AC-03.1): at 100 cells the Tap Info fields
// sit left and the formula and cask counts sit right, the formula count on the
// Tap row; narrower, the counts follow the fields in one column.
func TestTapInfoSplitsWhenWide(t *testing.T) {
	p := &panelData{taps: []brew.Tap{{
		Name: "lthiagol/tap", Remote: "https://github.com/lthiagol/homebrew-tap",
		FormulaCount: 3, CaskCount: 1,
	}}}

	wide := ansi.Strip(renderTapInfo(p, 100))
	var row string
	for _, line := range strings.Split(wide, "\n") {
		if strings.Contains(line, "Tap:") {
			row = line
		}
	}
	if !strings.Contains(row, "Formulae:") || !strings.Contains(wide, "Casks:") {
		t.Fatalf("at 100 cells the counts should sit in the right column, got %q in:\n%s", row, wide)
	}

	narrow := ansi.Strip(renderTapInfo(p, 97))
	for _, line := range strings.Split(narrow, "\n") {
		if strings.Contains(line, "Tap:") && strings.Contains(line, "Formulae:") {
			t.Fatalf("at 97 cells the view is one column, got %q", line)
		}
	}
	if !strings.Contains(narrow, "Formulae:") {
		t.Fatalf("the one column should still show the counts:\n%s", narrow)
	}
}
