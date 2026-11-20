package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestServiceStatusOmitsEmptyValues (AC-01.1): a service that never ran has
// no user or file, so those rows are left out, not printed empty.
func TestServiceStatusOmitsEmptyValues(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "6")
	m = updateModel(m, DataLoadedMsg{PanelID: PanelServices, Items: []string{"postgresql"},
		Services: []brew.Service{{Name: "postgresql", Status: brew.ServiceStopped}}})
	out := ansi.Strip(m.renderContent(70, 20))
	if !strings.Contains(out, "Status:") || !strings.Contains(out, "stopped") {
		t.Fatalf("the Status tab should show the stopped service, got:\n%s", out)
	}
	for _, gone := range []string{"User:", "File:"} {
		if strings.Contains(out, gone) {
			t.Fatalf("an empty %q row should be left out, got:\n%s", gone, out)
		}
	}
}

// TestConfigTabCutsRowsToTheWidth (AC-01.1): config rows are built uncut and
// cut where drawn, so a long prefix ends with … inside the main width.
func TestConfigTabCutsRowsToTheWidth(t *testing.T) {
	cut := cutLines(formatConfig(&brew.BrewConfig{Prefix: strings.Repeat("p", 90)}), 40)
	for _, line := range strings.Split(ansi.Strip(cut), "\n") {
		if w := len([]rune(line)); w > 40 {
			t.Fatalf("row is %d cells, want at most 40: %q", w, line)
		}
	}
	if !strings.Contains(ansi.Strip(cut), "…") {
		t.Fatalf("the long prefix should be cut with …, got:\n%s", ansi.Strip(cut))
	}
}
