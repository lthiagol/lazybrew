package gui

import (
	"fmt"
	"strings"
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

// BenchmarkFilesTabWheelDown is the Files tab of a formula with 5000 installed
// files, a wheel notch down and the view drawn after each notch, at 140x40 (cr3
// AC-04; the page keys went in lb19 AC-03).
func BenchmarkFilesTabWheelDown(b *testing.B) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelFormulae, Items: []string{"bash"},
		Formulae: []brew.Formula{{Name: "bash", Version: "5.2"}}})
	m = sendKey(m, "2")
	m.activeTab = formulaFilesTab
	var files strings.Builder
	for i := 0; i < 5000; i++ {
		fmt.Fprintf(&files, "/home/linuxbrew/.linuxbrew/Cellar/bash/5.2/share/file%04d.txt\n", i)
	}
	m.tabContent[tabKey(PanelFormulae, formulaFilesTab, "bash")] = files.String()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m = wheelDown(m)
		_ = m.View()
	}
}

// BenchmarkViewCommandLog is View() at 140x40 with the command log full, the
// frame every key and tick draws (cr3 AC-05).
func BenchmarkViewCommandLog(b *testing.B) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	for i := 0; i < 500; i++ {
		m.commandLog.SetStatus(m.commandLog.Append(fmt.Sprintf("info --json=v2 pkg%d", i)), CommandSuccess)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

// BenchmarkFormulaePanelView is View() at 140x40 with a 300-formula list, the
// panel and its sidebar rows drawn (cr3 AC-06).
func BenchmarkFormulaePanelView(b *testing.B) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	var msg DataLoadedMsg
	msg.PanelID = PanelFormulae
	for i := 0; i < 300; i++ {
		name := fmt.Sprintf("pkg%03d", i)
		msg.Items = append(msg.Items, name)
		msg.Formulae = append(msg.Formulae, brew.Formula{Name: name, Version: "1.0", Description: "a formula"})
	}
	m = updateModel(m, msg)
	m = sendKey(m, "2")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

// BenchmarkHelpScreenView is View() with the help open at 140x40, the frame a
// scroll key draws (cr3 AC-07).
func BenchmarkHelpScreenView(b *testing.B) {
	m := sendKey(updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40}), "?")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

// BenchmarkSortFormulae sorts a 1000-formula list, the copy each load makes
// before the sort (cr3 AC-07).
func BenchmarkSortFormulae(b *testing.B) {
	base := make([]brew.Formula, 1000)
	for i := range base {
		base[i] = brew.Formula{Name: fmt.Sprintf("Pkg%04d", (i*7919)%1000), Outdated: i%7 == 0, Pinned: i%11 == 0}
	}
	fs := make([]brew.Formula, len(base))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(fs, base)
		sortFormulae(fs)
	}
}
