package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// v1 delivered a paste as a KeyMsg, so it reached the focused search box,
// started the search box from its panel, or went to the open modal. These
// tests pin the same three routes for the v2 PasteMsg.

func TestPasteIntoFocusedSearchInserts(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.activePanel = PanelSearch
	m.searchInput.Focus()

	m = updateModel(m, tea.PasteMsg{Content: "lazy"})
	if got := m.searchInput.Value(); got != "lazy" {
		t.Fatalf("search value = %q, want %q", got, "lazy")
	}
}

func TestPasteOnSearchPanelStartsSearchInput(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.activePanel = PanelSearch
	m.searchInput.Blur()

	m = updateModel(m, tea.PasteMsg{Content: "brew"})
	if !m.searchInput.Focused() {
		t.Fatal("paste on the search panel should focus the search input")
	}
	if got := m.searchInput.Value(); got != "brew" {
		t.Fatalf("search value = %q, want %q", got, "brew")
	}
}

func TestPasteReachesOpenInputModal(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.activeModal = modal.NewInputModal("Tap repository (user/repo):")

	m = updateModel(m, tea.PasteMsg{Content: "homebrew/cask"})
	if m.activeModal == nil {
		t.Fatal("paste should not close the input modal")
	}
	res := m.activeModal.Result().(*modal.InputResult)
	if res.Value != "homebrew/cask" {
		t.Fatalf("modal value = %q, want %q", res.Value, "homebrew/cask")
	}
}
