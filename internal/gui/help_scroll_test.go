package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// helpFrame is the help screen as plain rows at the model's size.
func helpFrame(m *Model) []string {
	return strings.Split(ansi.Strip(m.View().Content), "\n")
}

// TestHelpScrollReachesEveryLineAt80x24 (v2-05 AC-04.1): at 80x24 the help does
// not fit, and paging through it shows every key.
func TestHelpScrollReachesEveryLineAt80x24(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sendKey(m, "?")
	seen := strings.Join(helpFrame(m), "\n")
	for i := 0; i < 20 && m.helpScroll < m.helpLast(); i++ {
		m = sendSpecial(m, tea.KeyPgDown)
		seen += "\n" + strings.Join(helpFrame(m), "\n")
	}
	for _, page := range buildHelp() {
		for _, sec := range page.sections {
			for _, item := range sec.items {
				if !strings.Contains(seen, item.helpKeyText()) || !strings.Contains(seen, item.help) {
					t.Fatalf("paging the help never showed %q %q", item.helpKeyText(), item.help)
				}
			}
		}
	}
}

// TestHelpPositionHint (AC-04.1): the box's right label gives the rows on show.
func TestHelpPositionHint(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sendKey(m, "?")
	total := len(helpRows(m.helpTextWidth()))
	if first := helpFrame(m)[0]; !strings.Contains(first, fmt.Sprintf("1–%d of %d", m.helpViewRows(), total)) {
		t.Fatalf("the first screen should read 1–%d of %d, got %q", m.helpViewRows(), total, first)
	}
	m = sendSpecial(m, tea.KeyPgDown)
	from := m.helpViewRows() + 1
	if first := helpFrame(m)[0]; !strings.Contains(first, fmt.Sprintf("%d–", from)) {
		t.Fatalf("after a page the hint should start at %d, got %q", from, first)
	}
}

// TestHelpKeysScrollHelpNotThePanels (AC-04.1): with help open the scroll keys
// move the text; the panel behind it keeps its cursor.
func TestHelpKeysScrollHelpNotThePanels(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sendKey(m, "?")
	m = sendKey(m, "j")
	if m.helpScroll != 1 {
		t.Fatalf("j should scroll help by one row, got %d", m.helpScroll)
	}
	m = sendSpecial(m, tea.KeyEnd)
	if m.helpScroll != m.helpLast() {
		t.Fatalf("End should reach the last full page, got %d of %d", m.helpScroll, m.helpLast())
	}
	m = sendSpecial(m, tea.KeyHome)
	if m.helpScroll != 0 {
		t.Fatalf("Home should return to the top, got %d", m.helpScroll)
	}
	if m.panels[m.activePanel].selected != 0 {
		t.Fatalf("the panel behind help must not move, selected %d", m.panels[m.activePanel].selected)
	}
	m = sendKey(m, "?")
	if m.showHelp {
		t.Fatal("? should close help")
	}
}

// TestHelpOpensAtTheTop: each time help opens it starts at the top.
func TestHelpOpensAtTheTop(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sendKey(m, "?")
	m = sendSpecial(m, tea.KeyPgDown)
	m = sendKey(m, "?")
	m = sendKey(m, "?")
	if m.helpScroll != 0 {
		t.Fatalf("reopened help should start at the top, got %d", m.helpScroll)
	}
}

// TestHelpWheelScrollsTheText: with the mouse on, a wheel notch moves the help
// three rows, and a click still closes it.
func TestHelpWheelScrollsTheText(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 80, Height: 24})
	m.cfg.GUI.Mouse = true
	m = sendKey(m, "?")
	m = updateModel(m, tea.MouseWheelMsg{X: 5, Y: 5, Button: tea.MouseWheelDown})
	if m.helpScroll != 3 {
		t.Fatalf("a wheel notch should move help three rows, got %d", m.helpScroll)
	}
	m = updateModel(m, tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	if m.showHelp {
		t.Fatal("a click should still close help")
	}
}

// TestHelpSectionsLineUp (AC-04.2): every item of a section has its description
// in the same column, and the key column is as wide as the widest key.
func TestHelpSectionsLineUp(t *testing.T) {
	for _, page := range buildHelp() {
		for _, sec := range page.sections {
			var descAt []int
			for _, it := range sec.items {
				row := ansi.Strip(helpUnitRow(page, it.help))
				descAt = append(descAt, strings.Index(row, it.help))
			}
			for _, at := range descAt[1:] {
				if at != descAt[0] {
					t.Fatalf("section %q of %q: descriptions start at %v, want one column", sec.heading, page.title, descAt)
				}
			}
		}
	}
}

// helpUnitRow finds the help row that carries desc, for the alignment test.
func helpUnitRow(page helpPage, desc string) string {
	for _, row := range helpUnit(page) {
		if strings.Contains(ansi.Strip(row), desc) {
			return row
		}
	}
	return ""
}

// TestHelpTwoColumnsWhenWide (AC-04.2): a wide help is shorter, because whole
// pages flow into a second column; the description of the scroll keys is not
// wrapped at 80 columns.
func TestHelpTwoColumnsWhenWide(t *testing.T) {
	if wide, narrow := len(helpRows(134)), len(helpRows(74)); wide >= narrow {
		t.Fatalf("two columns should be shorter than one: %d rows wide, %d narrow", wide, narrow)
	}
	found := false
	for _, row := range helpRows(74) {
		if strings.Contains(ansi.Strip(row), "Top/bottom (g/G outside Search)") {
			found = true
		}
	}
	if !found {
		t.Fatal("the g/G note should sit on one row at 80 columns")
	}
}

// TestHelpColumnsKeepAPageWhole (AC-04.2): a page never splits across columns,
// so no heading is left at the foot of a column without its rows.
func TestHelpColumnsKeepAPageWhole(t *testing.T) {
	units := [][]string{
		{"A", "a1", "a2", "a3"},
		{"B", "b1", "b2", "b3", "b4", "b5"},
		{"C", "c1"},
		{"D", "d1", "d2"},
	}
	cols := helpColumns(units, 2)
	for _, u := range units {
		placed := 0
		for _, col := range cols {
			for _, row := range col {
				if row == u[0] {
					placed++
				}
			}
		}
		if placed != 1 {
			t.Fatalf("page %s should start in exactly one column, found %d", u[0], placed)
		}
		for _, col := range cols {
			for i, row := range col {
				if row != u[0] {
					continue
				}
				if i+len(u) > len(col) {
					t.Fatalf("page %s is cut at the foot of its column: %v", u[0], col)
				}
				for j, want := range u {
					if col[i+j] != want {
						t.Fatalf("page %s is split: %v", u[0], col)
					}
				}
			}
		}
	}
}
