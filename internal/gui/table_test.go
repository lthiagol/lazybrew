package gui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// pkgTableCols is the column set of a package table with icons on. The
// description goes first when the table is narrow, then the version.
var pkgTableCols = []tableCol{{head: ""}, {head: "Name"}, {head: "Version", drop: 1}, {head: "Description", drop: 2}}

// TestTableAlignsColumnsAndTakesSlack: the rows' names start under the Name
// header, every line is exactly the table's width, and the last column takes
// the slack.
func TestTableAlignsColumnsAndTakesSlack(t *testing.T) {
	rows := []tableRow{
		{cells: []string{"✓", "ncurses", "6.5", "Terminal handling"}},
		{cells: []string{"○", "readline", "", ""}},
	}
	out := ansi.Strip(renderTable(pkgTableCols, rows, 60))
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("a header and two rows should be three lines, got %d:\n%s", len(lines), out)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != 60 {
			t.Fatalf("line %d is %d cells, want the table's 60: %q", i, got, line)
		}
	}
	nameCol := ansi.StringWidth(lines[0][:strings.Index(lines[0], "Name")])
	for _, row := range lines[1:] {
		name := "ncurses"
		if !strings.Contains(row, name) {
			name = "readline"
		}
		if got := ansi.StringWidth(row[:strings.Index(row, name)]); got != nameCol {
			t.Fatalf("%s starts at cell %d, want the Name header's %d:\n%s", name, got, nameCol, out)
		}
	}
	if !strings.HasSuffix(strings.TrimRight(lines[1], " "), "Terminal handling") {
		t.Fatalf("the last column should hold the description, got %q", lines[1])
	}
}

// TestTableCutsTheLastColumn: a cell wider than its column is cut with …, and
// a row never wraps onto a second line.
func TestTableCutsTheLastColumn(t *testing.T) {
	long := strings.Repeat("word ", 30)
	out := ansi.Strip(renderTable([]tableCol{{head: "Description"}}, []tableRow{{cells: []string{long}}}, 20))
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("a long cell should not wrap, got %d lines:\n%s", len(lines), out)
	}
	if !strings.HasSuffix(strings.TrimRight(lines[1], " "), "…") {
		t.Fatalf("a cut cell should end with …, got %q", lines[1])
	}
}

// TestTableDropsDescriptionThenVersion: as the width narrows the description
// goes first, then the version; the icon and the name stay.
func TestTableDropsDescriptionThenVersion(t *testing.T) {
	rows := []tableRow{{cells: []string{"✓", "ncurses", "6.5", "Terminal handling for every screen"}}}
	wide := ansi.Strip(renderTable(pkgTableCols, rows, 80))
	if !strings.Contains(wide, "Terminal handling") || !strings.Contains(wide, "6.5") {
		t.Fatalf("at 80 cells every column should show, got:\n%s", wide)
	}
	mid := ansi.Strip(renderTable(pkgTableCols, rows, 28))
	if strings.Contains(mid, "Terminal handling") || !strings.Contains(mid, "6.5") {
		t.Fatalf("at 28 cells the description should go and the version stay, got:\n%s", mid)
	}
	narrow := ansi.Strip(renderTable(pkgTableCols, rows, 14))
	if strings.Contains(narrow, "6.5") || !strings.Contains(narrow, "ncurses") || !strings.Contains(narrow, "✓") {
		t.Fatalf("at 14 cells only the icon and the name should stay, got:\n%s", narrow)
	}
}

// TestLongDescriptionIsCutNotDropped (v2-09 F-r1-01): a description longer than
// the slack is kept and cut with …; the name and the version stay with it, and
// every line is the table's width.
func TestLongDescriptionIsCutNotDropped(t *testing.T) {
	rows := []tableRow{
		{cells: []string{"✓", "ncurses", "6.6", "Text-based UI library"}},
		{cells: []string{"✓", "openssl@3", "3.4.1", strings.Repeat("Cryptography and SSL/TLS toolkit ", 4)}},
	}
	out := ansi.Strip(renderTable(pkgTableCols, rows, 87))
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("a header and two rows should be three lines, got %d:\n%s", len(lines), out)
	}
	if !strings.Contains(lines[0], "Description") || !strings.Contains(lines[0], "Version") {
		t.Fatalf("the Description and Version columns should both be kept, got header %q", lines[0])
	}
	if !strings.HasSuffix(strings.TrimRight(lines[2], " "), "…") {
		t.Fatalf("the long description should be cut with …, got %q", lines[2])
	}
	if !strings.Contains(lines[2], "3.4.1") || !strings.Contains(lines[2], "openssl@3") {
		t.Fatalf("the name and version should stay with the long description, got %q", lines[2])
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != 87 {
			t.Fatalf("line %d is %d cells, want the table's 87: %q", i, got, line)
		}
	}
}

// TestTableHeaderIsMuted: the header row is drawn in the Muted role.
func TestTableHeaderIsMuted(t *testing.T) {
	out := renderTable([]tableCol{{head: "Path"}}, []tableRow{{cells: []string{"/a/b"}}}, 20)
	want := lipgloss.NewStyle().Foreground(style.MutedColor).Render("Path")
	if !strings.HasPrefix(out, want) {
		t.Fatalf("the header should be drawn in the Muted role, got %q", out)
	}
}

// TestTableHeadingSpansTheTable: a heading is a row of its own, with no cells.
func TestTableHeadingSpansTheTable(t *testing.T) {
	rows := []tableRow{{heading: "Runtime"}, {cells: []string{"✓", "ncurses"}}}
	out := ansi.Strip(renderTable([]tableCol{{head: ""}, {head: "Name"}}, rows, 30))
	lines := strings.Split(out, "\n")
	if lines[1] != "Runtime" {
		t.Fatalf("the heading should be its own row, got %q", lines[1])
	}
}

// TestPackageRowsCarryLoadedDataOnly: a package row takes its version and
// description from the loaded installed list, and leaves them empty when the
// package is not installed or the list has not loaded.
func TestPackageRowsCarryLoadedDataOnly(t *testing.T) {
	m := formulaWithDeps(t)
	m.panels[PanelFormulae].formulae = []brew.Formula{
		{Name: "ncurses", Version: "6.5", Description: "Terminal handling"},
	}
	cells := func(name string) string {
		return ansi.Strip(strings.Join(m.pkgRow(name, m.installedList(depFormula)).cells, "|"))
	}
	if got := cells("ncurses"); got != "✓|ncurses|6.5|Terminal handling" {
		t.Fatalf("an installed package should carry its version and description, got %q", got)
	}
	if got := cells("cmake"); got != "○|cmake||" {
		t.Fatalf("a package that is not installed has no version or description, got %q", got)
	}
	m.panels[PanelFormulae].loaded = false
	if got := cells("ncurses"); got != "|ncurses||" {
		t.Fatalf("before the installed list loads no row claims an icon or data, got %q", got)
	}
}

// TestPackageColumnsWithIconsOff: with icons off there is no icon column, and
// the state is written after the name.
func TestPackageColumnsWithIconsOff(t *testing.T) {
	m := formulaWithDeps(t)
	m.icons = style.Icons{}
	if cols := m.pkgColumns(); len(cols) != 3 || cols[0].head != "Name" {
		t.Fatalf("with icons off the first column is the name, got %+v", cols)
	}
	row := m.pkgRow("cmake", m.installedList(depFormula))
	if got := ansi.Strip(row.cells[0]); got != "cmake (not installed)" {
		t.Fatalf("with icons off the state follows the name, got %q", got)
	}
}

// TestPkgRowsMatchPkgRow (v2-09 F-r1-03): one table read of the installed list
// gives each row what a row read on its own gives.
func TestPkgRowsMatchPkgRow(t *testing.T) {
	m := formulaWithDeps(t)
	m.panels[PanelFormulae].formulae = []brew.Formula{{Name: "ncurses", Version: "6.5", Description: "Terminal handling"}}
	names := []string{"ncurses", "cmake", "readline"}
	rows := m.pkgRows(names, depFormula)
	list := m.installedList(depFormula)
	for i, name := range names {
		if got, want := rows[i].cells, m.pkgRow(name, list).cells; strings.Join(got, "|") != strings.Join(want, "|") {
			t.Fatalf("row %d (%s) = %q, want %q", i, name, got, want)
		}
	}
}

// TestDescriptionKeepsItsFloorBesideALongName (v2-09 F-r2-01): at 46 cells a
// name of about 29 characters is cut so the description keeps at least its
// minimum useful width, and the description column is kept and cut with …,
// never left as a sliver.
func TestDescriptionKeepsItsFloorBesideALongName(t *testing.T) {
	rows := []tableRow{{cells: []string{"✓", "gobject-introspection-runtime", "1.86.0_1",
		"Generate introspection data for GObject libraries"}}}
	out := ansi.Strip(renderTable(pkgTableCols, rows, 46))
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("a header and one row should be two lines, got %d:\n%s", len(lines), out)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != 46 {
			t.Fatalf("line %d is %d cells, want 46: %q", i, got, line)
		}
	}
	head := ansi.StringWidth(lines[0][:strings.Index(lines[0], "Description")])
	at := strings.Index(lines[1], "Generate")
	if at < 0 {
		t.Fatalf("the description should be shown, got %q", lines[1])
	}
	if got := ansi.StringWidth(lines[1][:at]); got != head {
		t.Fatalf("the description starts at cell %d, want the header's %d:\n%s", got, head, out)
	}
	shown := strings.TrimRight(lines[1][at:], " ")
	if lipgloss.Width(shown) < tableMinUseful || !strings.HasSuffix(shown, "…") {
		t.Fatalf("the description should keep at least %d cells and end with …, got %q", tableMinUseful, shown)
	}
}
