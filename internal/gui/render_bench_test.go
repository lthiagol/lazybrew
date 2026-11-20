package gui

import (
	"fmt"
	"testing"
	"time"

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

// BenchmarkOperationRenderPerLine is a long batch run: 10 000 lines of brew
// output, the operation rendered after each line, at 140x40 (cr3 AC-02). Run it
// with -benchtime=1x: one op is the whole run.
func BenchmarkOperationRenderPerLine(b *testing.B) {
	lines := []string{"==> Upgrading 5000 outdated packages:"}
	for i := 0; len(lines) < 10000; i++ {
		lines = append(lines, fmt.Sprintf("==> Upgrading pkg%04d", i), "  1.0 -> 1.1")
	}
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.opState = &Operation{Title: "upgrade all", Status: opRunning, Upgrades: true, Started: time.Now()}
		for _, line := range lines {
			m.opState.AppendLine(line)
			_ = m.renderOperation(140, 40)
		}
	}
}
