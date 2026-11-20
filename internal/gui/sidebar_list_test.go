package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestSidebarListFillsItsBox (replaces TestMainListFillsMainPanel): the list is
// in the sidebar now, so its rows fill the box the sidebar gives it, down to the
// last row the box has.
func TestSidebarListFillsItsBox(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 40}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
			var rows []string
			for i := 0; i < 30; i++ {
				rows = append(rows, fmt.Sprintf("row-%02d", i))
			}
			m.panels[PanelStatus].items = rows

			shown := strings.Count(m.renderSidebar(), "row-")
			if want := m.panels[PanelStatus].visibleRows; shown != want {
				t.Fatalf("the sidebar box shows %d rows, want its %d", shown, want)
			}
		})
	}
}

// TestSidebarListScrollsWithCursor (replaces TestMainListScrollsWithCursor): the
// sidebar keeps the cursor in its window, and the main panel never draws the
// list again: it shows the selected item's details.
func TestSidebarListScrollsWithCursor(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 40}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
			var rows []string
			for i := 0; i < 30; i++ {
				rows = append(rows, fmt.Sprintf("row-%02d", i))
			}
			m.panels[PanelStatus].items = rows

			for i := 0; i < 6; i++ {
				m = sendKey(m, "j")
			}
			if main := m.renderMainPanel(); strings.Contains(main, "row-0") {
				t.Fatal("the main panel must not redraw the sidebar's list")
			}

			target := m.panels[PanelStatus].visibleCount() + 2
			for m.panels[PanelStatus].selected < target {
				m = sendKey(m, "j")
			}
			want := fmt.Sprintf("row-%02d", m.panels[PanelStatus].selected)
			if side := sgr.ReplaceAllString(m.renderSidebar(), ""); !strings.Contains(side, want) {
				t.Fatalf("sidebar should scroll to %q", want)
			}
		})
	}
}

// TestMainAreaShowsDetailsNotTheList (AC-03.1): the Services, Taps and Outdated
// main tabs describe the selected item. The other items' rows are not drawn.
func TestMainAreaShowsDetailsNotTheList(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})

	m = sendKey(m, "6")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelServices, Items: []string{"alpha", "bravo", "charlie"},
		Services: []brew.Service{{Name: "alpha"}, {Name: "bravo"}, {Name: "charlie"}}})
	m = sendKey(m, "j")
	if out := m.renderContent(60, 20); !strings.Contains(out, "bravo") || strings.Contains(out, "charlie") {
		t.Fatalf("Services Status should show the selected bravo only, got:\n%s", out)
	}

	m = sendKey(m, "5")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelTaps, Items: []string{"aa/one", "bb/two", "cc/three"},
		Taps: []brew.Tap{{Name: "aa/one"}, {Name: "bb/two"}, {Name: "cc/three"}}})
	m = sendKey(m, "j")
	if out := m.renderContent(60, 20); !strings.Contains(out, "3 taps") || !strings.Contains(out, "bb/two") || strings.Contains(out, "cc/three") {
		t.Fatalf("Tap Info should show the summary and the selected tap only, got:\n%s", out)
	}

	m = sendKey(m, "4")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: []string{"alpha 1 → 2", "bravo 3 → 4"},
		Formulae: []brew.Formula{{Name: "alpha", Version: "1", NewVersion: "2"}, {Name: "bravo", Version: "3", NewVersion: "4"}}})
	m = sendKey(m, "j")
	m = sendKey(m, "]")
	out := m.renderContent(60, 20)
	if !strings.Contains(out, "bravo") || !strings.Contains(out, "Installed:") || strings.Contains(out, "alpha") {
		t.Fatalf("Versions should show the selected bravo's versions only, got:\n%s", out)
	}
	if !strings.Contains(out, "none selected") {
		t.Fatalf("Versions should show the batch summary, got:\n%s", out)
	}
}

// TestOutdatedVersionsShowsBatchSummary (AC-03.2): the batch selected from the
// sidebar is counted on the Versions tab, with the keys that act on it.
func TestOutdatedVersionsShowsBatchSummary(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "4")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: []string{"alpha", "bravo"},
		Formulae: []brew.Formula{{Name: "alpha", Version: "1", NewVersion: "2"}, {Name: "bravo", Version: "3", NewVersion: "4"}}})
	m = sendKey(m, "]")
	m = sendKey(m, " ")
	m = sendKey(m, "j")
	m = sendKey(m, " ")
	if out := m.renderContent(60, 20); !strings.Contains(out, "2 selected — u to upgrade, U all") {
		t.Fatalf("Versions should count the batch, got:\n%s", out)
	}
}

// sidebarBox is the rows of the sidebar box whose top border carries title, from
// its top border to its bottom one, with the styling stripped.
func sidebarBox(side, title string) string {
	var rows []string
	in := false
	for _, row := range strings.Split(side, "\n") {
		plain := sgr.ReplaceAllString(row, "")
		if strings.Contains(plain, title) {
			in = true
		}
		if in {
			rows = append(rows, plain)
			if strings.HasPrefix(plain, "╰") {
				break
			}
		}
	}
	return strings.Join(rows, "\n")
}

// TestBatchMarksOnlyTheOutdatedList (F-orch-01): a batch made on Outdated marks
// only the Outdated rows. The Formulae rows at the same indices stay unmarked.
func TestBatchMarksOnlyTheOutdatedList(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "4")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelOutdated, Items: []string{"alpha", "bravo", "charlie"},
		Formulae: []brew.Formula{{Name: "alpha", Version: "1", NewVersion: "2"}, {Name: "bravo", Version: "3", NewVersion: "4"}, {Name: "charlie", Version: "5", NewVersion: "6"}}})
	m.batch.selected = map[int]bool{0: true, 1: true}
	if box := sidebarBox(m.renderSidebar(), "Outdated"); !strings.Contains(box, "●") {
		t.Fatalf("the Outdated list should mark its batch, got:\n%s", box)
	}

	m = sendKey(m, "2")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"alpha", "bravo", "charlie"},
		Formulae: []brew.Formula{{Name: "alpha", Version: "1"}, {Name: "bravo", Version: "3"}, {Name: "charlie", Version: "5"}}})
	side := m.renderSidebar()
	if box := sidebarBox(side, "Formulae"); strings.Contains(box, "●") {
		t.Fatalf("the Formulae list must not show the Outdated batch, got:\n%s", box)
	}
}

// TestTapTrustTabShowsTheTrustStory (F-orch-02): the Trust tab says what the
// tap's state means for its formulae and casks, and names the key that changes
// it. Tap Info keeps the details.
func TestTapTrustTabShowsTheTrustStory(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "5")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelTaps, Items: []string{"aa/one", "bb/two"},
		Taps: []brew.Tap{
			{Name: "aa/one", FormulaCount: 2, CaskCount: 1},
			{Name: "bb/two", FormulaCount: 4, CaskCount: 0, Trusted: true, Remote: "https://example.test/bb/homebrew-two"},
		}})
	m = sendKey(m, "]")
	out := sgr.ReplaceAllString(m.renderContent(70, 20), "")
	for _, want := range []string{"Trust:", "untrusted", "may not load", "Its 2 formulae and 1 cask", "Press t"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the untrusted Trust tab should say %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Remote:") {
		t.Fatalf("the Trust tab should not repeat the tap details, got:\n%s", out)
	}

	m = sendKey(m, "j")
	if out := sgr.ReplaceAllString(m.renderContent(70, 20), ""); !strings.Contains(out, "may load") {
		t.Fatalf("a trusted tap should say Homebrew may load it, got:\n%s", out)
	}

	m = sendKey(m, "[")
	if out := sgr.ReplaceAllString(m.renderContent(70, 20), ""); !strings.Contains(out, "Remote:") {
		t.Fatalf("Tap Info keeps the details, got:\n%s", out)
	}
}

// TestTrustStoryPluralises (F-r2-02): a count of one takes the singular noun.
func TestTrustStoryPluralises(t *testing.T) {
	cases := []struct {
		tap  brew.Tap
		want string
	}{
		{brew.Tap{Name: "a/one", FormulaCount: 1, CaskCount: 0}, "Its 1 formula and 0 casks follow"},
		{brew.Tap{Name: "a/two", FormulaCount: 2, CaskCount: 1}, "Its 2 formulae and 1 cask follow"},
		{brew.Tap{Name: "a/none", FormulaCount: 0, CaskCount: 0}, "Its 0 formulae and 0 casks follow"},
	}
	for _, c := range cases {
		if got := tapTrustStory(&c.tap, 60); !strings.Contains(got, c.want) {
			t.Errorf("%s: want %q in:\n%s", c.tap.Name, c.want, got)
		}
	}
}
