package gui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestTapInfoSplitsWhenWide (v2-09 AC-03.1, Q-v2-09-a): at the split width (82
// cells of content, a main box inner width of 84) the Tap Info fields sit left
// and the formula and cask counts sit right, the formula count on the Tap row;
// one cell narrower, the counts follow the fields in one column.
func TestTapInfoSplitsWhenWide(t *testing.T) {
	p := &panelData{taps: []brew.Tap{{
		Name: "lthiagol/tap", Remote: "https://github.com/lthiagol/homebrew-tap",
		FormulaCount: 3, CaskCount: 1,
	}}}

	wide := ansi.Strip(renderTapInfo(p, 82))
	// The Remote, which is long, sits in the right column, so the counts come
	// after it there: the formula count is found past the left column.
	for _, want := range []string{"Formulae:", "Casks:"} {
		at := strings.Index(wide, want)
		if at < 0 {
			t.Fatalf("at 82 cells the view should show %q, got:\n%s", want, wide)
		}
		if col := ansi.StringWidth(wide[:at]); col < 2*82/5 {
			t.Fatalf("at 82 cells %q should sit in the right column, at cell %d:\n%s", want, col, wide)
		}
	}

	narrow := ansi.Strip(renderTapInfo(p, 81))
	for _, line := range strings.Split(narrow, "\n") {
		if strings.Contains(line, "Tap:") && strings.Contains(line, "Formulae:") {
			t.Fatalf("at 81 cells the view is one column, got %q", line)
		}
	}
	if !strings.Contains(narrow, "Formulae:") {
		t.Fatalf("the one column should still show the counts:\n%s", narrow)
	}
}

// TestInfoSplitsAt140x40 (Q-v2-09-a): the default 140x40 layout is wide enough
// for the Info views to split, so the description sits beside the name.
func TestInfoSplitsAt140x40(t *testing.T) {
	m := formulaWithDeps(t)
	p := m.panels[PanelFormulae]
	p.formulae[0].Description = "GNU Bourne-Again SHell"
	out := ansi.Strip(m.renderContent(m.mainContentWidth(), 20))
	first := strings.Split(out, "\n")[0]
	if !strings.Contains(first, "Name:") || !strings.Contains(first, "Description:") {
		t.Fatalf("at 140x40 the Info should split, first row %q in:\n%s", first, out)
	}
}
