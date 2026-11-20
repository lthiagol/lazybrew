package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// BenchmarkSearchPanelView is View() of the search panel with an installed
// formula's info loaded, at 140x40 (cr3 AC-01). It is not part of the gate:
// run it with -bench SearchPanelView.
func BenchmarkSearchPanelView(b *testing.B) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "7")
	m = updateModel(m, SearchDoneMsg{Results: []string{"jq"}, Raw: []brew.SearchResult{{Name: "jq", IsFormula: true, Installed: true}}})
	m = updateModel(m, SearchInfoLoadedMsg{Content: searchFormulaInstalled})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}
