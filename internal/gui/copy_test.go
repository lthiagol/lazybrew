package gui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// captureClipboard swaps the clipboard of m for one that records what it is
// asked to copy, and sends nothing to the terminal.
func captureClipboard(m *Model) *[]string {
	var copied []string
	m.setClipboard = func(s string) tea.Cmd {
		copied = append(copied, s)
		return nil
	}
	return &copied
}

// lastToastText is the message of the newest toast, or "" when there is none.
func lastToastText(m *Model) string {
	if t := m.newestToast(); t != nil {
		return t.Message
	}
	return ""
}

// y on a failed Operation copies its title and every output line, with the
// escape codes removed and nothing cut.
func TestYCopiesTheOverlayAsPlainText(t *testing.T) {
	m := newTestModel()
	copied := captureClipboard(m)
	m.opState = &Operation{
		Title:  "install foo",
		Status: opError,
		Err:    errors.New("boom"),
		Lines: []string{
			"\x1b[1m==> Installing foo\x1b[0m",
			"\x1b[31mError: " + strings.Repeat("long ", 60) + "boom\x1b[0m",
		},
	}

	m = sendKey(m, "y")

	want := "install foo\n==> Installing foo\nError: " + strings.Repeat("long ", 60) + "boom"
	if len(*copied) != 1 || (*copied)[0] != want {
		t.Fatalf("y copied %q, want one copy of %q", *copied, want)
	}
	if got := lastToastText(m); got != "Copied 3 lines" {
		t.Errorf("toast = %q, want %q", got, "Copied 3 lines")
	}
}

// y with no overlay copies the main panel's content: the Formulae Info tab, as
// plain text, not cut to the view.
func TestYCopiesTheMainPanelAsPlainText(t *testing.T) {
	m := newTestModel()
	copied := captureClipboard(m)
	m.activePanel = PanelFormulae
	m.activeTab = formulaInfoTab
	panel := m.panels[PanelFormulae]
	panel.items = []string{"foo  1.0"}
	panel.formulae = []brew.Formula{{Name: "foo", Description: strings.Repeat("wide ", 40) + "end"}}
	panel.selected = 0

	m = sendKey(m, "y")

	if len(*copied) != 1 {
		t.Fatalf("y copied %d times, want 1", len(*copied))
	}
	got := (*copied)[0]
	if ansi.Strip(got) != got {
		t.Errorf("the copy holds escape codes: %q", got)
	}
	if strings.Contains(got, "…") {
		t.Errorf("the copy is cut: %q", got)
	}
	if !strings.Contains(got, "foo") || !strings.Contains(got, strings.Repeat("wide ", 40)+"end") {
		t.Errorf("the copy lacks the formula's description: %q", got)
	}
	if want := fmt.Sprintf("Copied %d lines", strings.Count(got, "\n")+1); lastToastText(m) != want {
		t.Errorf("toast = %q, want %q", lastToastText(m), want)
	}
}

// y typed into a focused search box is a letter, not a copy.
func TestYInTheFocusedSearchBoxTypesIt(t *testing.T) {
	m := newTestModel()
	copied := captureClipboard(m)
	m.activePanel = PanelSearch
	m.searchInput.Focus()

	m = sendKey(m, "y")

	if got := m.searchInput.Value(); got != "y" {
		t.Errorf("search box = %q, want %q", got, "y")
	}
	if len(*copied) != 0 {
		t.Errorf("y copied %q while the box had focus", *copied)
	}
}

// On Search with the box not focused, y is typed too, as the other letters
// that Search has no key for are.
func TestYOnSearchTypesWhenTheBoxIsNotFocused(t *testing.T) {
	m := newTestModel()
	copied := captureClipboard(m)
	m.activePanel = PanelSearch

	m = sendKey(m, "y")

	if got := m.searchInput.Value(); got != "y" || !m.searchInput.Focused() {
		t.Errorf("search box = %q focused=%v, want %q focused", got, m.searchInput.Focused(), "y")
	}
	if len(*copied) != 0 {
		t.Errorf("y copied %q on Search", *copied)
	}
}

// y with nothing on screen to copy says so and copies nothing.
func TestYWithNothingToCopySaysSo(t *testing.T) {
	m := newTestModel()
	copied := captureClipboard(m)
	m.opState = &Operation{}

	m = sendKey(m, "y")

	if len(*copied) != 0 {
		t.Errorf("y copied %q, want nothing", *copied)
	}
	if got := lastToastText(m); got != "Nothing to copy" {
		t.Errorf("toast = %q, want %q", got, "Nothing to copy")
	}
}
