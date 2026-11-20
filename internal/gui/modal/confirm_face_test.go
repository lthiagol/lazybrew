package modal

import (
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
