package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// searchPanelData is a Search panel holding rows, its items the bare names
// the way the panel keeps them, with no row selected.
func searchPanelData(rows ...brew.SearchResult) *panelData {
	p := &panelData{id: PanelSearch, results: rows, selected: -1}
	for _, r := range rows {
		p.items = append(p.items, r.Name)
	}
	return p
}

// TestSearchRowsAlignOnTheCellSystem (lb25 AC-02): a Search row is drawn from
// its cells. The mark column is blank for a row that is not installed, so names
// start at one column; the type is flush right; the word installed is gone.
func TestSearchRowsAlignOnTheCellSystem(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	rows := []brew.SearchResult{
		{Name: "alpha-cask", IsCask: true, Installed: true},
		{Name: "beta-formula", IsFormula: true},
		{Name: "gamma-unknown"},
	}
	types := []string{"cask", "formula", "unknown"}
	for name, theme := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
		style.ApplyTheme(theme)
		for _, w := range []int{30, 49, 80} {
			p := searchPanelData(rows...)
			for i, r := range rows {
				row := plainLine(p.sidebarRow(i, w, nil, style.UnicodeIcons, false))
				if got := lipgloss.Width(row); got != w {
					t.Fatalf("%s w=%d: row is %d cells: %q", name, w, got, row)
				}
				wantLead := "    "
				if r.Installed {
					wantLead = "  " + style.UnicodeIcons.Installed + " "
				}
				if !strings.HasPrefix(row, wantLead+r.Name) {
					t.Errorf("%s w=%d: row should be %q then the name: %q", name, w, wantLead, row)
				}
				if !strings.HasSuffix(row, types[i]) {
					t.Errorf("%s w=%d: row should end with the type %q: %q", name, w, types[i], row)
				}
				if strings.Contains(row, "installed") {
					t.Errorf("%s w=%d: the word installed is still on the row: %q", name, w, row)
				}
			}
		}
	}
}

// TestSearchRowsWithIconsOffDrawNoMark: show_icons off draws no mark, so the
// name starts after the marker alone and the type still closes the row.
func TestSearchRowsWithIconsOffDrawNoMark(t *testing.T) {
	p := searchPanelData(brew.SearchResult{Name: "alpha-cask", IsCask: true, Installed: true})
	row := plainLine(p.sidebarRow(0, 40, nil, style.Icons{}, false))
	if !strings.HasPrefix(row, "  alpha-cask") || !strings.HasSuffix(row, "cask") {
		t.Errorf("icons off: want the name after the marker and the type at the end, got %q", row)
	}
	if strings.Contains(row, style.UnicodeIcons.Installed) {
		t.Errorf("icons off: the installed mark is still drawn: %q", row)
	}
}

// TestSearchLongNameKeepsType: a name too long for the row is cut with an
// ellipsis, and its type stays at the end of the row.
func TestSearchLongNameKeepsType(t *testing.T) {
	long := strings.Repeat("a-very-long-formula-name-", 3) + "end"
	p := searchPanelData(brew.SearchResult{Name: long, IsFormula: true})
	row := plainLine(p.sidebarRow(0, 40, nil, style.UnicodeIcons, false))
	if lipgloss.Width(row) != 40 {
		t.Fatalf("row is %d cells, want 40: %q", lipgloss.Width(row), row)
	}
	if !strings.Contains(row, "…") || !strings.HasSuffix(row, "formula") {
		t.Errorf("a long name should be cut with … and keep the type: %q", row)
	}
}

// TestSearchFrameAlignsRows (lb25 AC-02): on the drawn screen, every Search row
// starts its name in one column and ends its type in one column, in both themes.
// The Search box shows two rows at 80x24 and four at 140x40; the rows drawn are
// checked. The installed row alone carries the mark.
func TestSearchFrameAlignsRows(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	rows := []brew.SearchResult{
		{Name: "alpha-tool-x", IsCask: true, Installed: true},
		{Name: "delta-" + strings.Repeat("x", 60), IsFormula: true},
		{Name: "beta-tool", IsFormula: true},
		{Name: "gamma-none"},
	}
	types := []string{"cask", "formula", "formula", "unknown"}
	shown := map[int]int{80: 2, 140: 4}
	for theme, t2 := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
		style.ApplyTheme(t2)
		for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 140, Height: 40}} {
			m := updateModel(newTestModel(), size)
			m.searchResults = rows
			p := m.panels[PanelSearch]
			p.results, p.items = rows, searchItems(rows)
			// switchPanel re-derives the sidebar's visible rows, as a real switch does.
			m.switchPanel(PanelSearch)
			p.selected = 0
			lines := strings.Split(ansi.Strip(m.View().Content), "\n")

			nameCol, typeEnd := -1, -1
			for i, r := range rows[:shown[size.Width]] {
				// A long name is cut on screen, so the row is found by its start.
				key := r.Name[:min(len(r.Name), 8)]
				line := findLine(lines, key)
				if line == "" {
					t.Fatalf("%s %dx%d: no row for %q in:\n%s", theme, size.Width, size.Height, key, strings.Join(lines, "\n"))
				}
				at := strings.Index(line, key)
				col := lipgloss.Width(line[:at])
				if nameCol < 0 {
					nameCol = col
				}
				if col != nameCol {
					t.Errorf("%s %dx%d: name %q starts at %d, want %d: %q", theme, size.Width, size.Height, r.Name, col, nameCol, line)
				}
				rest := line[at+len(key):]
				ti := strings.Index(rest, types[i])
				if ti < 0 {
					t.Fatalf("%s %dx%d: no type %q after %q: %q", theme, size.Width, size.Height, types[i], key, line)
				}
				end := lipgloss.Width(line[:at+len(key)+ti+len(types[i])])
				if typeEnd < 0 {
					typeEnd = end
				}
				if end != typeEnd {
					t.Errorf("%s %dx%d: type %q ends at %d, want %d: %q", theme, size.Width, size.Height, types[i], end, typeEnd, line)
				}
				if marked := strings.Contains(line, "✔"); marked != r.Installed {
					t.Errorf("%s %dx%d: row %q installed mark = %v, want %v", theme, size.Width, size.Height, r.Name, marked, r.Installed)
				}
			}
		}
	}
}

// findLine is the first line that holds s, or "".
func findLine(lines []string, s string) string {
	for _, line := range lines {
		if strings.Contains(line, s) {
			return line
		}
	}
	return ""
}

// TestSearchSelectedRowHighlightsWhenFocused (lb25 AC-02): the selected Search
// row takes the highlight when its box is focused, and the text is the same.
func TestSearchSelectedRowHighlightsWhenFocused(t *testing.T) {
	p := searchPanelData(brew.SearchResult{Name: "alpha-tool-x", IsCask: true, Installed: true})
	p.selected = 0
	focused := p.sidebarRow(0, 40, nil, style.UnicodeIcons, true)
	unfocused := p.sidebarRow(0, 40, nil, style.UnicodeIcons, false)
	if !bgOpen.MatchString(focused) {
		t.Errorf("the focused selected row should carry the highlight background: %q", focused)
	}
	if bgOpen.MatchString(unfocused) {
		t.Errorf("an unfocused box must not highlight its selected row: %q", unfocused)
	}
	if plainLine(focused) != plainLine(unfocused) {
		t.Errorf("focus changes the colour only, not the text:\n%q\n%q", plainLine(focused), plainLine(unfocused))
	}
}
