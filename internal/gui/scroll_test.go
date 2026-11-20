package gui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

var caveatLine = regexp.MustCompile(`line-(\d\d)`)

// scrollFixture is the Formulae Caveats tab with ninety caveat lines, scrolled to
// the top, so the main content is longer than its area at 140x40.
func scrollFixture(t *testing.T) *Model {
	t.Helper()
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "2")
	var lines []string
	for i := 0; i < 90; i++ {
		lines = append(lines, fmt.Sprintf("line-%02d", i))
	}
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"foo  1.0"},
		Formulae: []brew.Formula{{Name: "foo", Version: "1.0", Caveats: strings.Join(lines, "\n")}}})
	for range formulaCaveatsTab {
		m = sendKey(m, "]")
	}
	if m.activeTab != formulaCaveatsTab {
		t.Fatalf("fixture: want the Caveats tab, got %d", m.activeTab)
	}
	return m
}

// topCaveat is the first caveat line the main panel draws.
func topCaveat(m *Model) string {
	plain := sgr.ReplaceAllString(m.renderMainPanel(), "")
	if mm := caveatLine.FindString(plain); mm != "" {
		return mm
	}
	return ""
}

// TestScrollKeysMoveAndClamp (F-orch-03): PgUp/PgDn move the content by a page,
// Home/End and g/G go to the top and the bottom, and every key stops at both
// ends of the content.
func TestScrollKeysMoveAndClamp(t *testing.T) {
	m := scrollFixture(t)
	page := m.pageRows()
	last := m.contentLast()
	if last <= page {
		t.Fatalf("fixture: the content must be longer than a page, last=%d page=%d", last, page)
	}

	down := func(m *Model, code rune) *Model {
		return updateModel(m, tea.KeyPressMsg{Code: code})
	}

	// PgDn moves one page, and the view shows the row the offset names.
	m = down(m, tea.KeyPgDown)
	if got := m.contentScroll[m.contentKey()]; got != page {
		t.Fatalf("PgDn: offset %d, want one page %d", got, page)
	}
	if got := topCaveat(m); got != fmt.Sprintf("line-%02d", page) {
		t.Fatalf("PgDn: the top row should be line-%02d, got %q", page, got)
	}

	// PgUp moves back, and stops at the top.
	m = down(m, tea.KeyPgUp)
	m = down(m, tea.KeyPgUp)
	if got := m.contentScroll[m.contentKey()]; got != 0 {
		t.Fatalf("PgUp past the top: offset %d, want 0", got)
	}

	// Home and End go to the top and the bottom; g and G do the same.
	m = down(m, tea.KeyEnd)
	if got := m.contentScroll[m.contentKey()]; got != last {
		t.Fatalf("End: offset %d, want the last offset %d", got, last)
	}
	if got := topCaveat(m); got != fmt.Sprintf("line-%02d", last) {
		t.Fatalf("End: the top row should be line-%02d, got %q", last, got)
	}
	if plain := sgr.ReplaceAllString(m.renderMainPanel(), ""); !strings.Contains(plain, "line-59") {
		t.Fatal("End: the last caveat line should be on screen")
	}
	m = down(m, tea.KeyPgDown)
	if got := m.contentScroll[m.contentKey()]; got != last {
		t.Fatalf("PgDn past the bottom: offset %d, want the last offset %d", got, last)
	}
	m = down(m, tea.KeyHome)
	if got := m.contentScroll[m.contentKey()]; got != 0 {
		t.Fatalf("Home: offset %d, want 0", got)
	}
	m = updateModel(m, tea.KeyPressMsg{Code: 'G', Text: "G"})
	if got := m.contentScroll[m.contentKey()]; got != last {
		t.Fatalf("G: offset %d, want the last offset %d", got, last)
	}
	m = updateModel(m, tea.KeyPressMsg{Code: 'g', Text: "g"})
	if got := m.contentScroll[m.contentKey()]; got != 0 {
		t.Fatalf("g: offset %d, want 0", got)
	}
}

// TestScrollIsPerItem: an item starts at the top. Moving the selection to
// another formula does not carry the scroll position with it.
func TestScrollIsPerItem(t *testing.T) {
	m := scrollFixture(t)
	m = updateModel(m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	if m.contentScroll[m.contentKey()] == 0 {
		t.Fatal("fixture: PgDn should have scrolled")
	}
	before := m.contentKey()
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"foo  1.0", "bar  2.0"},
		Formulae: []brew.Formula{{Name: "foo", Version: "1.0"}, {Name: "bar", Version: "2.0"}}})
	m = sendKey(m, "j")
	if m.contentKey() == before {
		t.Fatal("fixture: the selection should have moved to another item")
	}
	if got := m.contentScroll[m.contentKey()]; got != 0 {
		t.Fatalf("another item must start at the top, offset %d", got)
	}
}

// TestScrollKeysAreInTheHelpAndFooter: the keys are documented where a user
// looks for them.
func TestScrollKeysAreInTheHelpAndFooter(t *testing.T) {
	var help strings.Builder
	for _, page := range buildHelp() {
		for _, sec := range page.sections {
			for _, item := range sec.items {
				help.WriteString(item.helpKeyText() + " " + item.help + "\n")
			}
		}
	}
	for _, want := range []string{"PgUp/PgDn", "Home/End", "Top/bottom (g/G outside Search)"} {
		if !strings.Contains(help.String(), want) {
			t.Errorf("the help should document %q", want)
		}
	}
	found := false
	for _, h := range globalHints {
		if h.key == "PgUp/Dn" {
			found = true
		}
	}
	if !found {
		t.Error("the footer's global hints should name the page keys")
	}
}

// TestLetterKeysOnUnfocusedSearchTypeToSearch (F-r2-01): on the Search panel with
// the input unfocused, g and G start a search with that letter, as they did
// before the scroll keys. The scroll keys do not take them there.
func TestLetterKeysOnUnfocusedSearchTypeToSearch(t *testing.T) {
	for _, key := range []string{"g", "G"} {
		m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
		m = sendKey(m, "7")
		m = sendSpecial(m, tea.KeyEscape)
		if m.searchInput.Focused() {
			t.Fatalf("fixture: Esc should leave the search input unfocused")
		}
		m = sendKey(m, key)
		if !m.searchInput.Focused() || m.searchInput.Value() != key {
			t.Fatalf("%s on unfocused Search: focused=%v value=%q, want focused with %q",
				key, m.searchInput.Focused(), m.searchInput.Value(), key)
		}
	}
}

// TestScrollLettersOutsideSearch: g and G still scroll the content on another
// panel, where they are the top and bottom keys.
func TestScrollLettersOutsideSearch(t *testing.T) {
	m := scrollFixture(t)
	last := m.contentLast()
	m = sendKey(m, "G")
	if got := m.contentScroll[m.contentKey()]; got != last {
		t.Fatalf("G outside Search: offset %d, want %d", got, last)
	}
	m = sendKey(m, "g")
	if got := m.contentScroll[m.contentKey()]; got != 0 {
		t.Fatalf("g outside Search: offset %d, want 0", got)
	}
}
