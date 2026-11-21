package gui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
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

// y on a failed Operation copies its title, its status row, and every output
// line, with the escape codes removed and nothing cut.
func TestYCopiesTheOverlayAsPlainText(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
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

	want := "install foo\n✗ Failed: boom · Esc to close\n==> Installing foo\nError: " + strings.Repeat("long ", 60) + "boom"
	if len(*copied) != 1 || (*copied)[0] != want {
		t.Fatalf("y copied %q, want one copy of %q", *copied, want)
	}
	if got := lastToastText(m); got != "Copied 4 lines" {
		t.Errorf("toast = %q, want %q", got, "Copied 4 lines")
	}
}

// The overlay copy holds the status row as shown, without the y hint that
// names the key just pressed.
func TestYCopiesTheStatusRowWithoutTheCopyHint(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
	copied := captureClipboard(m)
	m.opState = &Operation{Title: "install foo", Status: opError, Err: errors.New("boom"), Started: time.Now(), Ended: time.Now()}

	m = sendKey(m, "y")

	if len(*copied) != 1 || !strings.Contains((*copied)[0], "\n✗ Failed: boom · Esc to close") {
		t.Fatalf("y copied %q, want the status row", *copied)
	}
	if strings.Contains((*copied)[0], "y copy") {
		t.Errorf("the copy holds the y hint: %q", (*copied)[0])
	}
}

// y on Search copies a failed overlay: the overlay owns the keys while it is
// open, so the letter is not typed into the hidden search box.
func TestYOnSearchCopiesAFailedOverlay(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
	copied := captureClipboard(m)
	m.activePanel = PanelSearch
	m.opState = &Operation{
		Title:   "install foo",
		Status:  opError,
		Err:     errors.New("boom"),
		Started: time.Now(),
		Ended:   time.Now(),
		Lines:   []string{"==> Installing foo", "Error: boom"},
	}

	m = sendKey(m, "y")

	want := "install foo\n✗ Failed: boom · Esc to close\n==> Installing foo\nError: boom"
	if len(*copied) != 1 || (*copied)[0] != want {
		t.Fatalf("y copied %q, want one copy of %q", *copied, want)
	}
	if got := m.searchInput.Value(); got != "" {
		t.Errorf("y typed %q into the search box under an overlay", got)
	}
}

// The same overlay copy holds on Formulae, where y copies the overlay too.
func TestYOnFormulaeCopiesAFailedOverlay(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
	copied := captureClipboard(m)
	m.activePanel = PanelFormulae
	m.opState = &Operation{
		Title:   "install foo",
		Status:  opError,
		Err:     errors.New("boom"),
		Started: time.Now(),
		Ended:   time.Now(),
		Lines:   []string{"Error: boom"},
	}

	m = sendKey(m, "y")

	want := "install foo\n✗ Failed: boom · Esc to close\nError: boom"
	if len(*copied) != 1 || (*copied)[0] != want {
		t.Fatalf("y copied %q, want one copy of %q", *copied, want)
	}
}

// A sudo refusal that the retry runs copies the command under its status row,
// as the overlay shows it.
func TestYCopiesTheRetryCommandUnderTheStatusRow(t *testing.T) {
	m := sudoFailedModel("install", "--cask", "foo")
	m.icons = style.UnicodeIcons
	copied := captureClipboard(m)
	m.activePanel = PanelSearch

	m = sendKey(m, "y")

	want := "install foo\n✗ Failed: needs sudo · Enter  run in this terminal · Esc  close\nbrew install --cask foo"
	if len(*copied) != 1 || (*copied)[0] != want {
		t.Fatalf("y copied %q, want one copy of %q", *copied, want)
	}
}

// With an overlay open, a focused search box still takes y as a letter.
func TestYWithAnOverlayOpenAndTheBoxFocusedTypesIt(t *testing.T) {
	m := newTestModel()
	copied := captureClipboard(m)
	m.activePanel = PanelSearch
	m.searchInput.Focus()
	m.opState = &Operation{Title: "install foo", Status: opError, Err: errors.New("boom"), Started: time.Now(), Ended: time.Now()}

	m = sendKey(m, "y")

	if got := m.searchInput.Value(); got != "y" {
		t.Errorf("search box = %q, want %q", got, "y")
	}
	if len(*copied) != 0 {
		t.Errorf("y copied %q while the box had focus", *copied)
	}
}

// In the real frame the failed retry row offers y copy where the row holds it:
// at 140x40 after the keys, at 80x24 after "needs sudo", the keys being on the
// line under the row.
func TestFailedRetryRowOffersYCopyWhereItFits(t *testing.T) {
	cases := []struct {
		width, height int
		row           string
	}{
		{80, 24, "✗ Failed: needs sudo · y copy"},
		{140, 40, "Enter  run in this terminal · Esc  close · y copy"},
	}
	for _, c := range cases {
		m := updateModel(sudoFailedModel("install", "--cask", "foo"), tea.WindowSizeMsg{Width: c.width, Height: c.height})
		got := ansi.Strip(m.View().Content)
		if !strings.Contains(got, c.row) {
			t.Errorf("%dx%d: the failed row lacks %q:\n%s", c.width, c.height, c.row, got)
		}
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

// An overlay copies its status row even with no output: the row is on screen,
// so there is text to copy. "Nothing to copy" is left for an empty screen.
func TestYCopiesAnOverlayWithNoOutputByItsStatusRow(t *testing.T) {
	m := newTestModel()
	copied := captureClipboard(m)
	m.opState = &Operation{Title: "install foo", Status: opRunning, Started: time.Now()}

	m = sendKey(m, "y")

	want := "install foo\nRunning 0s · Esc to cancel"
	if len(*copied) != 1 || (*copied)[0] != want {
		t.Fatalf("y copied %q, want one copy of %q", *copied, want)
	}
	if got := lastToastText(m); got != "Copied 2 lines" {
		t.Errorf("toast = %q, want %q", got, "Copied 2 lines")
	}
}
