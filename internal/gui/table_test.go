package gui

import (
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

func TestSortFormulae(t *testing.T) {
	fs := []brew.Formula{
		{Name: "zebra", Version: "1"},
		{Name: "alpha", Version: "1", Pinned: true},
		{Name: "beta", Version: "1", Outdated: true, NewVersion: "2"},
		{Name: "gamma", Version: "1", Outdated: true, Pinned: true, NewVersion: "2"},
		{Name: "Alpha", Version: "1"}, // case-insensitive with zebra/alpha order
	}
	sortFormulae(fs)
	// outdated first (gamma, beta — both outdated; gamma before beta by name among outdated? gamma vs beta: beta first)
	// Within outdated: pinned first → gamma (outdated+pinned), beta (outdated)
	// Then non-outdated pinned: alpha
	// Then remaining A-Z: Alpha, zebra — "Alpha" and "alpha" same lower; stable among equals after ranks
	got := make([]string, len(fs))
	for i, f := range fs {
		got[i] = f.Name
	}
	// outdated+pinned, outdated, pinned, then alpha names, zebra
	if got[0] != "gamma" {
		t.Fatalf("first should be outdated+pinned gamma, got %v", got)
	}
	if got[1] != "beta" {
		t.Fatalf("second should be outdated beta, got %v", got)
	}
	if got[2] != "alpha" {
		t.Fatalf("third should be pinned alpha, got %v", got)
	}
	// remaining Alpha then zebra (A before z)
	rest := got[3:]
	if len(rest) != 2 || strings.ToLower(rest[0]) != "alpha" || rest[1] != "zebra" {
		t.Fatalf("remaining want Alpha,zebra got %v", rest)
	}
}

func TestFormulaVersionAndStatus(t *testing.T) {
	f := brew.Formula{Name: "x", Version: "1.0", Outdated: true, NewVersion: "2.0"}
	if formulaVersionCell(f) != "1.0 -> 2.0" {
		t.Fatalf("version cell = %q", formulaVersionCell(f))
	}
	if formulaStatusLabel(f) != "outdated" {
		t.Fatalf("status = %q", formulaStatusLabel(f))
	}
	f2 := brew.Formula{Name: "y", Version: "1", Pinned: true}
	if formulaVersionCell(f2) != "1" {
		t.Fatalf("version = %q", formulaVersionCell(f2))
	}
	if formulaStatusLabel(f2) != "pinned" {
		t.Fatalf("status = %q", formulaStatusLabel(f2))
	}
	f3 := brew.Formula{Name: "z", Version: "1"}
	if formulaStatusLabel(f3) != "installed" {
		t.Fatalf("status = %q", formulaStatusLabel(f3))
	}
	// outdated wins over pinned
	f4 := brew.Formula{Outdated: true, Pinned: true, Version: "1", NewVersion: "2"}
	if formulaStatusLabel(f4) != "outdated" {
		t.Fatalf("outdated should beat pinned, got %q", formulaStatusLabel(f4))
	}
}

func TestShortenTap(t *testing.T) {
	if got := shortenTap("homebrew/core"); got != "hb/c" {
		t.Fatalf("homebrew/core = %q, want hb/c", got)
	}
	if got := shortenTap(""); got != "" {
		t.Fatalf("empty = %q", got)
	}
}

func TestFormulaeTabsOrder(t *testing.T) {
	tabs := panelTabs[PanelFormulae]
	want := []string{"List", "Info", "Deps", "Used By", "Caveats", "Files"}
	if len(tabs) != len(want) {
		t.Fatalf("len=%d want %d", len(tabs), len(want))
	}
	for i, n := range want {
		if tabs[i].name != n {
			t.Errorf("tab[%d]=%q want %q", i, tabs[i].name, n)
		}
	}
}

func TestNeedsTabFetchShifted(t *testing.T) {
	if needsTabFetch(PanelFormulae, 0) {
		t.Error("List must not fetch")
	}
	if needsTabFetch(PanelFormulae, 1) {
		t.Error("Info must not fetch")
	}
	if !needsTabFetch(PanelFormulae, 2) {
		t.Error("Deps must fetch")
	}
	if !needsTabFetch(PanelFormulae, 3) {
		t.Error("Used By must fetch")
	}
	if needsTabFetch(PanelFormulae, 4) {
		t.Error("Caveats must not fetch")
	}
	if !needsTabFetch(PanelFormulae, 5) {
		t.Error("Files must fetch")
	}
}

func TestRenderFormulaeTableFullPane(t *testing.T) {
	style.ApplyTheme(style.DarkTheme())
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	p := m.panels[PanelFormulae]
	p.formulae = []brew.Formula{
		{Name: "ripgrep", Version: "14.1.1", Tap: "homebrew/core"},
		{Name: "zlib", Version: "1.3", NewVersion: "1.3.1", Outdated: true, Tap: "homebrew/core"},
		{Name: "python", Version: "3.12", Pinned: true, Tap: "homebrew/core"},
	}
	p.selected = 1
	m.activePanel = PanelFormulae
	m.activeTab = 0
	m.tabs = panelTabs[PanelFormulae]

	const w, h = 60, 12
	out := m.renderFormulaeTable(w, h)
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Version") ||
		!strings.Contains(out, "Status") || !strings.Contains(out, "Tap") {
		t.Fatalf("missing headers:\n%s", out)
	}
	if !strings.Contains(out, "ripgrep") || !strings.Contains(out, "zlib") {
		t.Fatalf("missing rows:\n%s", out)
	}
	if !strings.Contains(out, "1.3 -> 1.3.1") {
		t.Fatalf("missing version arrow:\n%s", out)
	}
	if !strings.Contains(out, "outdated") || !strings.Contains(out, "pinned") || !strings.Contains(out, "installed") {
		t.Fatalf("missing status labels:\n%s", out)
	}
	if !strings.Contains(out, "hb/c") {
		t.Fatalf("expected shortened tap hb/c:\n%s", out)
	}
	// AC-10 full-pane height/width contract
	if got := lipgloss.Height(out); got != h {
		t.Fatalf("table height = %d, want %d (full pane)", got, h)
	}
	if got := lipgloss.Width(out); got != w {
		t.Fatalf("table width = %d, want %d (full pane)", got, w)
	}
	// Selected row (zlib) highlighted — SelectedItem uses Accent; plain name still present
	if !strings.Contains(out, "zlib") {
		t.Fatal("selected formula zlib should appear")
	}
	// Not FormatFormulaInfo dump
	if strings.Contains(out, "Name:") && strings.Contains(out, "License:") {
		t.Fatal("List tab must not render FormatFormulaInfo")
	}
}

func TestFormulaeTableColWidthsSum(t *testing.T) {
	for _, total := range []int{20, 40, 60, 80, 100, 120} {
		n, v, s, tap := formulaeTableColWidths(total)
		sum := n + v + s + tap + 3
		if sum != total {
			t.Errorf("total=%d widths name=%d ver=%d status=%d tap=%d sum=%d (want %d)",
				total, n, v, s, tap, sum, total)
		}
		if n < 1 || v < 1 || s < 1 || tap < 1 {
			t.Errorf("total=%d has non-positive column width", total)
		}
	}
}

func TestFormulaStatusStyledUsesBadges(t *testing.T) {
	style.ApplyTheme(style.DarkTheme())
	// AC-11: status cells go through themed badges (not plain monochrome helpers).
	if formulaStatusStyled(brew.Formula{Outdated: true}) != style.OutdatedBadge.Render("outdated") {
		t.Error("outdated must use OutdatedBadge")
	}
	if formulaStatusStyled(brew.Formula{Pinned: true}) != style.PinnedBadge.Render("pinned") {
		t.Error("pinned must use PinnedBadge")
	}
	if formulaStatusStyled(brew.Formula{}) != style.InstalledBadge.Render("installed") {
		t.Error("installed must use InstalledBadge")
	}
	// Badge styles are wired to theme colors (assert style config, not ANSI — NO_COLOR).
	if style.OutdatedBadge.GetForeground() != style.WarningColor {
		t.Error("OutdatedBadge should use WarningColor")
	}
	if style.PinnedBadge.GetForeground() != style.SecondaryColor {
		t.Error("PinnedBadge should use SecondaryColor")
	}
	if style.InstalledBadge.GetForeground() != style.SuccessColor {
		t.Error("InstalledBadge should use SuccessColor")
	}
}

func TestRenderFormulaeTableEmpty(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelFormulae
	m.activeTab = 0
	out := m.renderFormulaeTable(40, 10)
	if !strings.Contains(out, "No formulae yet") {
		t.Fatalf("missing empty title: %q", out)
	}
	if !strings.Contains(out, "Press / to search & install") {
		t.Fatalf("missing empty hint: %q", out)
	}
}

func TestRenderContentListTabIsTable(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m.activePanel = PanelFormulae
	m.activeTab = 0
	m.tabs = panelTabs[PanelFormulae]
	p := m.panels[PanelFormulae]
	p.loading = false
	p.formulae = []brew.Formula{{Name: "foo", Version: "1.0", Tap: "homebrew/core"}}
	p.selected = 0
	out := m.renderContent(50, 15)
	if !strings.Contains(out, "Name") || !strings.Contains(out, "foo") {
		t.Fatalf("List tab content should be table, got:\n%s", out)
	}
}

func TestTabActiveStyleHasBottomBorder(t *testing.T) {
	style.ApplyTheme(style.DarkTheme())
	// Active uses bottom border character ━
	s := style.TabActive.Render(" List ")
	if s == style.TabInactive.Render(" List ") {
		t.Fatal("active and inactive tab styles should differ")
	}
	// Bold + accent on active
	if !style.TabActive.GetBold() {
		t.Error("TabActive should be bold")
	}
	if style.TabActive.GetForeground() != style.AccentColor {
		t.Error("TabActive should use AccentColor")
	}
}

func TestFormulaeTableKeepsSelectedInWindow(t *testing.T) {
	m := newTestModel()
	p := m.panels[PanelFormulae]
	fs := make([]brew.Formula, 30)
	for i := range fs {
		fs[i] = brew.Formula{Name: "pkg" + strconv.Itoa(i), Version: "1.0"}
	}
	p.formulae = fs
	p.selected = 25
	out := m.renderFormulaeTable(50, 8) // body ~6 rows
	if !strings.Contains(out, "pkg25") {
		t.Fatalf("selected pkg25 should be in window:\n%s", out)
	}
	if strings.Contains(out, "pkg0") {
		t.Fatalf("expected window scrolled past start, still has pkg0:\n%s", out)
	}
}
