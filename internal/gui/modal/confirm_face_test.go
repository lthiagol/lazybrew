package modal

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

var sgrRe = regexp.MustCompile("\x1b\\[[0-9;:]*m")

// plainLines is the view with colour codes removed, one string per row.
func plainLines(view string) []string {
	return strings.Split(sgrRe.ReplaceAllString(view, ""), "\n")
}

// reversed returns the text each reverse-video run in the view covers. Its
// SGR parameters are walked one by one, so a 7 inside a colour (38;2;7;...)
// is not taken for the reverse attribute.
func reversed(view string) []string {
	var out []string
	on := false
	var run strings.Builder
	flush := func() {
		if on && run.Len() > 0 {
			out = append(out, run.String())
		}
		run.Reset()
	}
	for len(view) > 0 {
		loc := sgrRe.FindStringIndex(view)
		if loc == nil {
			run.WriteString(view)
			break
		}
		run.WriteString(view[:loc[0]])
		params := strings.Split(view[loc[0]+2:loc[1]-1], ";")
		view = view[loc[1]:]
		for i := 0; i < len(params); i++ {
			switch params[i] {
			case "", "0", "27":
				flush()
				on = false
			case "7":
				flush()
				on = true
			case "38", "48":
				if i+1 < len(params) && params[i+1] == "2" {
					i += 4
				} else if i+1 < len(params) && params[i+1] == "5" {
					i += 2
				}
			}
		}
	}
	flush()
	return out
}

// TestConfirmFaceFrame locks AC-02.1-2: a thick border in the action's accent,
// the upper-cased action chip on top, the question in bold, the detail below.
func TestConfirmFaceFrame(t *testing.T) {
	m := NewConfirmModal(ActionDanger, "Uninstall bash", "Uninstall bash? This cannot be undone.").
		WithDetail("removes bash; 2 packages depend on it")
	view := m.View().Content

	lines := plainLines(view)
	if !strings.HasPrefix(lines[0], "┏") || !strings.HasSuffix(lines[0], "┓") {
		t.Fatalf("top border is not thick: %q", lines[0])
	}
	if !strings.HasPrefix(lines[len(lines)-1], "┗") {
		t.Fatalf("bottom border is not thick: %q", lines[len(lines)-1])
	}
	accent := lipgloss.NewStyle().Foreground(style.ErrorColor).Render("x")
	if !strings.Contains(view, accent[:strings.Index(accent, "x")]+"┏") {
		t.Errorf("border is not drawn in the danger accent")
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"UNINSTALL BASH", "Uninstall bash? This cannot be undone.", "removes bash; 2 packages depend on it"} {
		if !strings.Contains(joined, want) {
			t.Errorf("dialog does not show %q:\n%s", want, joined)
		}
	}
	if !strings.Contains(view, "\x1b[1m") {
		t.Errorf("the question is not bold")
	}
	if strings.Index(joined, "UNINSTALL BASH") > strings.Index(joined, "Uninstall bash?") ||
		strings.Index(joined, "Uninstall bash?") > strings.Index(joined, "removes bash") {
		t.Errorf("chip, question and detail are out of order:\n%s", joined)
	}
}

// TestConfirmFaceKeyLine locks AC-02.4: the key line names the keys, and the
// current Yes/No selection is the reverse-video one.
func TestConfirmFaceKeyLine(t *testing.T) {
	m := NewConfirmModal(ActionUpgrade, "Confirm Upgrade", "Upgrade bash?")
	view := m.View().Content
	joined := strings.Join(plainLines(view), "\n")
	for _, want := range []string{"[y]es", "[n]o", "esc cancel"} {
		if !strings.Contains(joined, want) {
			t.Errorf("key line lacks %q", want)
		}
	}
	if got := reversed(view); len(got) != 1 || got[0] != "o" {
		t.Fatalf("default selection (No) reversed = %q, want [\"o\"]", got)
	}
	m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	if got := reversed(m.View().Content); len(got) != 1 || got[0] != "es" {
		t.Fatalf("after h, reversed = %q, want [\"es\"]", got)
	}
}

// TestConfirmFaceWidth locks AC-02.5: the box is 72 cells wide, every row is
// exactly that wide, and long text wraps inside the box instead of past it.
func TestConfirmFaceWidth(t *testing.T) {
	long := strings.Repeat("a very long sentence that must wrap inside the box ", 6)
	m := NewConfirmModal(ActionNeutral, "Untap user/tap", "Remove tap user/tap?").WithDetail(long)
	lines := plainLines(m.View().Content)
	for i, line := range lines {
		if w := lipgloss.Width(line); w != confirmWidth {
			t.Fatalf("row %d is %d cells wide, want %d: %q", i, w, confirmWidth, line)
		}
	}
	if !strings.Contains(strings.Join(lines, " "), "wrap inside the box") {
		t.Fatalf("the long detail was cut, not wrapped:\n%s", strings.Join(lines, "\n"))
	}
}

// TestInputAndMenuShareTheFrame locks AC-02.6: the input and menu modals are
// framed the same way: a thick border, the action as the chip, their keys.
func TestInputAndMenuShareTheFrame(t *testing.T) {
	cases := []struct {
		name  string
		m     Modal
		chip  string
		texts []string
		keys  []string
	}{
		{"input", NewInputModal("Add tap", "Tap repository (user/repo):"), " ADD TAP ",
			[]string{"Tap repository (user/repo):"}, []string{"enter submit", "esc cancel"}},
		{"menu", NewMenuModal("Trust: homebrew/foo", []string{"Trust tap", "Cancel"}), " TRUST: HOMEBREW/FOO ",
			[]string{"Trust tap"}, []string{"↑↓ move", "enter choose", "esc cancel"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := plainLines(tc.m.View().Content)
			if !strings.HasPrefix(lines[0], "┏") || !strings.HasSuffix(lines[0], "┓") {
				t.Fatalf("border is not thick: %q", lines[0])
			}
			joined := strings.Join(lines, "\n")
			for _, want := range append(append([]string{tc.chip}, tc.texts...), tc.keys...) {
				if !strings.Contains(joined, want) {
					t.Errorf("view lacks %q:\n%s", want, joined)
				}
			}
		})
	}
}

// TestConfirmQuestionWrapsAndBoxFits80x24 locks Q-v2-04-wrap: at 80x24 a long
// question and consequence wrap inside the box, the box stays within 20 rows,
// a body that is still too long ends in an ellipsis, and the key line stays.
func TestConfirmQuestionWrapsAndBoxFits80x24(t *testing.T) {
	question := strings.Repeat("alpha beta gamma ", 20) + "ENDQ"
	detail := strings.Repeat("a consequence that runs on for a while ", 4) + "ENDD"
	var deps []string
	for i := 0; i < 30; i++ {
		deps = append(deps, "  dependent-package-"+strings.Repeat("x", i%5))
	}
	m := NewConfirmModal(ActionDanger, "Uninstall bash", question+"\n"+strings.Join(deps, "\n")).
		WithDetail(detail)
	view := m.View().Content
	lines := plainLines(view)
	if len(lines) > dialogMaxRows {
		t.Fatalf("box is %d rows, want at most %d", len(lines), dialogMaxRows)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != confirmWidth {
			t.Fatalf("row %d is %d cells wide, want %d: %q", i, w, confirmWidth, line)
		}
	}
	joined := strings.Join(strings.Fields(strings.ReplaceAll(strings.Join(lines, "\n"), "┃", " ")), " ")
	if !strings.Contains(joined, "ENDQ") || !strings.Contains(joined, "ENDD") {
		t.Fatalf("the question or consequence was cut, not wrapped:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(joined, "…") {
		t.Fatalf("the over-long body does not end in an ellipsis:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(joined, "[y]es") || !strings.Contains(joined, "esc cancel") {
		t.Fatalf("the key line was pushed out of the box:\n%s", strings.Join(lines, "\n"))
	}
}

// TestLongMenuItemAndChipAreCut locks Q-v2-04-wrap: a menu item and a chip are
// single lines, so one too long for the box is cut with an ellipsis.
func TestLongMenuItemAndChipAreCut(t *testing.T) {
	long := strings.Repeat("a menu item that is far too long for the box ", 3)
	m := NewMenuModal(strings.Repeat("title ", 20), []string{long, "Cancel"})
	lines := plainLines(m.View().Content)
	for i, line := range lines {
		if w := lipgloss.Width(line); w != formWidth {
			t.Fatalf("row %d is %d cells wide, want %d: %q", i, w, formWidth, line)
		}
	}
	// Row 0 is the top border, row 1 the padding, row 2 the chip.
	if chip := lines[2]; !strings.Contains(chip, "…") || !strings.Contains(chip, "TITLE") {
		t.Fatalf("the chip was not cut to its width: %q", chip)
	}
	for _, line := range lines {
		if strings.Contains(line, "0: ") && !strings.Contains(line, "…") {
			t.Fatalf("the menu item was not cut to its width: %q", line)
		}
	}
}

// contentRows is the text inside a dialog's box, one string per content row:
// the rows between the top and bottom padding, without the border and the
// three cells of padding on each side.
func contentRows(view string, width int) []string {
	lines := plainLines(view)
	var out []string
	for _, line := range lines[2 : len(lines)-2] {
		runes := []rune(line)
		out = append(out, string(runes[4:len(runes)-4]))
	}
	return out
}

// fieldCells is the input field's width: its prompt, 30 cells of text, cursor.
const fieldCells = 33

// TestDialogsFitAndCentreTheirRows locks F-r2-01 for all three dialogs: every
// content row is at most the inner width, and each non-empty row is centred
// (its left and right gaps differ by at most one cell).
func TestDialogsFitAndCentreTheirRows(t *testing.T) {
	long := strings.Repeat("a long input prompt that goes on ", 3)
	cases := []struct {
		name  string
		m     Modal
		width int
	}{
		{"confirm", NewConfirmModal(ActionDanger, "Uninstall", "Uninstall bash?").WithDetail("removes bash; 2 packages depend on it"), confirmWidth},
		{"confirm long", NewConfirmModal(ActionUpgrade, "Upgrade", strings.Repeat("word ", 40)).WithDetail(strings.Repeat("detail ", 30)), confirmWidth},
		{"input", NewInputModal("Add tap", "Tap repository (user/repo):"), formWidth},
		{"input long", NewInputModal("Add tap", long), formWidth},
		{"menu", NewMenuModal("Trust: homebrew/foo", []string{"Trust tap", "Cancel"}), formWidth},
		{"menu long", NewMenuModal("Brewfile", []string{strings.Repeat("an item ", 20), "Cancel"}), formWidth},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			view := tc.m.View().Content
			inner := innerWidth(tc.width)
			for i, row := range contentRows(view, tc.width) {
				if w := lipgloss.Width(row); w > inner {
					t.Fatalf("content row %d is %d cells, more than the inner width %d: %q", i, w, inner, row)
				}
			}
			// The block is centred: its least gap on the left and on the right
			// differ by at most one cell. A left-aligned list block that is
			// centred passes; a row centred on its own passes too.
			left, right := inner, inner
			for _, row := range contentRows(view, tc.width) {
				if strings.Trim(row, " ") == "" {
					continue
				}
				// The input field is a fixed 33 cells (prompt, 30 of text, cursor)
				// whose unused text is blank, so its trailing gap is the field's
				// padding, not the row's. It is checked on its own below.
				if strings.Contains(row, "> ") {
					if want := (inner - fieldCells) / 2; strings.Index(row, "> ") != want {
						t.Fatalf("the input field starts at column %d, want %d: %q", strings.Index(row, "> "), want, row)
					}
					continue
				}
				left = min(left, len(row)-len(strings.TrimLeft(row, " ")))
				right = min(right, len(row)-len(strings.TrimRight(row, " ")))
			}
			if d := left - right; d < -1 || d > 1 {
				t.Fatalf("the block is off-centre: least gap %d cells left, %d right", left, right)
			}
			// The chip is the first content row; it is centred on its own.
			chip := contentRows(view, tc.width)[0]
			chipLeft := len(chip) - len(strings.TrimLeft(chip, " "))
			chipRight := len(chip) - len(strings.TrimRight(chip, " "))
			if d := chipLeft - chipRight; d < -1 || d > 1 {
				t.Fatalf("the chip is off-centre: %d cells left, %d right: %q", chipLeft, chipRight, chip)
			}
		})
	}
}

// TestDependentsAlignAsOneBlock locks F-r3-01: the dependents list of an
// uninstall keeps its lines aligned, so every dependent name starts in the same
// column however long it is.
func TestDependentsAlignAsOneBlock(t *testing.T) {
	message := "Uninstall bash? This cannot be undone.\n\nThe following depend on bash:\n" +
		"  direnv\n  zsh-completions-long-name\n  git"
	m := NewConfirmModal(ActionDanger, "Uninstall", message).WithDetail("removes bash; 3 packages depend on it")
	rows := contentRows(m.View().Content, confirmWidth)
	starts := map[string]int{}
	for _, name := range []string{"direnv", "zsh-completions-long-name", "git"} {
		for _, row := range rows {
			if i := strings.Index(row, name); i >= 0 {
				starts[name] = len([]rune(row[:i]))
			}
		}
		if _, ok := starts[name]; !ok {
			t.Fatalf("dependent %q is not drawn:\n%s", name, strings.Join(rows, "\n"))
		}
	}
	if starts["direnv"] != starts["zsh-completions-long-name"] || starts["git"] != starts["direnv"] {
		t.Fatalf("dependent names start in different columns: %v", starts)
	}
	// The names keep the two-cell indent under the heading they belong to.
	for _, row := range rows {
		if i := strings.Index(row, "The following"); i >= 0 {
			if heading := len([]rune(row[:i])); starts["git"] != heading+2 {
				t.Fatalf("the names are at column %d, want two cells in from the heading at %d", starts["git"], heading)
			}
		}
	}
}

// firstTextColumn is the cell column where a box row's text starts, past the
// border and the padding.
func firstTextColumn(row string) int {
	rest := strings.TrimLeft(row, " ┃")
	return lipgloss.Width(row[:len(row)-len(rest)])
}

// TestCutBodyEllipsisSitsInListColumn locks LB-10: a body cut to fit the box
// ends in an ellipsis that starts in the column of the list it cuts, not
// centred alone in the middle of the box.
func TestCutBodyEllipsisSitsInListColumn(t *testing.T) {
	var deps []string
	for i := 0; i < 30; i++ {
		deps = append(deps, fmt.Sprintf("  dependent-package-%02d", i))
	}
	m := NewConfirmModal(ActionDanger, "Uninstall bash", "Uninstall bash?\n"+strings.Join(deps, "\n"))
	lines := plainLines(m.View().Content)
	if len(lines) > dialogMaxRows {
		t.Fatalf("box is %d rows, want at most %d", len(lines), dialogMaxRows)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != confirmWidth {
			t.Fatalf("row %d is %d cells wide, want %d: %q", i, w, confirmWidth, line)
		}
	}

	list, cut := -1, -1
	for _, line := range lines {
		switch {
		case strings.Contains(line, "dependent-package-"):
			col := firstTextColumn(line)
			if list >= 0 && col != list {
				t.Fatalf("list rows start at columns %d and %d:\n%s", list, col, strings.Join(lines, "\n"))
			}
			list = col
		case strings.Contains(line, "…"):
			cut = firstTextColumn(line)
		}
	}
	if list < 0 || cut < 0 {
		t.Fatalf("want a list row and an ellipsis row:\n%s", strings.Join(lines, "\n"))
	}
	if cut != list {
		t.Fatalf("the ellipsis starts at column %d, the list at column %d:\n%s",
			cut, list, strings.Join(lines, "\n"))
	}
}
