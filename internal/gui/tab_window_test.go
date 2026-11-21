package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// windowRow is the names row of the tab bar at width, without colour.
func windowRow(m *Model, width int) string {
	return strings.Split(ansi.Strip(m.renderTabBar(width)), "\n")[0]
}

// checkTabWindow holds the tab bar at width to the window rules: the names row
// fits, the active tab is in it, the tabs it shows are whole and next to each
// other, and a marker stands on each side that has tabs cut off. When every tab
// fits, the row is the whole row with no marker.
func checkTabWindow(t *testing.T, m *Model, width int) {
	t.Helper()
	row := windowRow(m, width)
	if w := lipgloss.Width(row); w > width {
		t.Errorf("tab %d: names row is %d cells, want at most %d: %q", m.activeTab, w, width, row)
	}
	n := len(m.tabs)
	visible := make([]bool, n)
	total := 0
	for i, tab := range m.tabs {
		visible[i] = strings.Contains(row, tab.name)
		total += tabWidth(tab.name, tabGap(m.activePanel))
	}
	if !visible[m.activeTab] {
		t.Fatalf("tab %d is active but not in the row: %q", m.activeTab, row)
	}
	first, last := -1, -1
	for i := range visible {
		if visible[i] {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	for i := first; i <= last; i++ {
		if !visible[i] {
			t.Errorf("tab %d is hidden between shown tabs %d and %d: %q", i, first, last, row)
		}
	}
	if total <= width {
		if first != 0 || last != n-1 || strings.ContainsAny(row, "‹›") {
			t.Errorf("the tabs fit in %d cells, so the row is whole with no marker: %q", width, row)
		}
		return
	}
	if got, want := strings.Contains(row, "‹"), first > 0; got != want {
		t.Errorf("left marker present = %v, want %v (first shown tab %d): %q", got, want, first, row)
	}
	if got, want := strings.Contains(row, "›"), last < n-1; got != want {
		t.Errorf("right marker present = %v, want %v (last shown tab %d): %q", got, want, last, row)
	}
}

// statusOnVulns is a loaded Status panel at 80x24 with the Vulns tab active.
func statusOnVulns(t *testing.T) *Model {
	t.Helper()
	m := loadedStatus(80, 24)
	for range 4 {
		m = pressStatus(m, "l")
	}
	if m.activeTab != statusVulnsTab {
		t.Fatalf("four l presses should reach the Vulns tab, got tab %d", m.activeTab)
	}
	return m
}

// TestStatusTabWindowEveryTabAt80x24: with each of the seven Status tabs active
// in turn, the row at 80x24 fits the main box, keeps the active tab, and marks
// each side that has tabs cut off, in both themes.
func TestStatusTabWindowEveryTabAt80x24(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for theme, th := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
		style.ApplyTheme(th)
		t.Run(theme, func(t *testing.T) {
			m := loadedStatus(80, 24)
			cw := contentWidthIn(m.screenLayout())
			for i := range m.tabs {
				if i > 0 {
					m = settle(t, pressStatus(m, "l"))
				}
				checkTabWindow(t, m, cw)
			}
		})
	}
}

// TestStatusTabWindowRows: the exact names row at 80x24 for the first tab and
// for Vulns. Dashboard shows the first five tabs and a marker on the right;
// Vulns shows Doctor to Unused, with a marker on the left.
func TestStatusTabWindowRows(t *testing.T) {
	m := loadedStatus(80, 24)
	if got, want := windowRow(m, contentWidthIn(m.screenLayout())), " Dashboard  Config  Doctor  Missing  Vulns ›"; got != want {
		t.Errorf("Dashboard active: row %q, want %q", got, want)
	}
	m = statusOnVulns(t)
	if got, want := windowRow(m, contentWidthIn(m.screenLayout())), "‹ Doctor  Missing  Vulns  Cleanup  Unused "; got != want {
		t.Errorf("Vulns active: row %q, want %q", got, want)
	}
}

// TestStatusTabWindowUnderlineAtRest: the underline under the active tab sits
// under its name in the window, at 80x24, and spans the tab with its padding.
func TestStatusTabWindowUnderlineAtRest(t *testing.T) {
	m := statusOnVulns(t)
	cw := contentWidthIn(m.screenLayout())
	rows := strings.Split(ansi.Strip(m.renderTabBar(cw)), "\n")
	if len(rows) != 2 {
		t.Fatalf("tab bar has %d rows; want 2", len(rows))
	}
	nameAt := lipgloss.Width(rows[0][:strings.Index(rows[0], "Vulns")])
	start, end := -1, -1
	for i, r := range []rune(rows[1]) {
		if r == '━' {
			if start < 0 {
				start = i
			}
			end = i + 1
		}
	}
	if start != nameAt-1 || end-start != len("Vulns")+2 {
		t.Errorf("underline is cells [%d,%d) under Vulns at %d: rule %q, row %q", start, end, nameAt, rows[1], rows[0])
	}
}

// TestStatusTabWindowWhileSliding: while the underline is on its way to the
// active tab, the window still holds the active tab and fits the row.
func TestStatusTabWindowWhileSliding(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "1")
	m := loadedStatus(80, 24)
	cw := contentWidthIn(m.screenLayout())
	for range 4 {
		m = sendKey(m, "l")
		for i := 0; i < 8; i++ {
			m = frames(m, 1)
			checkTabWindow(t, m, cw)
		}
	}
}

// TestStatusTabWindowClicks: at 80x24 with Vulns active, a click on a shown tab
// selects it, a click on the left marker moves one tab left, and a click on the
// right marker moves one tab right.
func TestStatusTabWindowClicks(t *testing.T) {
	m := statusOnVulns(t)
	row := windowRow(m, contentWidthIn(m.screenLayout()))
	for _, name := range []string{"Doctor", "Missing", "Vulns", "Cleanup", "Unused"} {
		fresh := statusOnVulns(t)
		col := lipgloss.Width(row[:strings.Index(row, name)])
		nm, _ := fresh.handleTabBarClick(col)
		want := tabIndex(t, fresh, name)
		if got := modelOf(nm).activeTab; got != want {
			t.Errorf("click on %s (cell %d) selected tab %d, want %d", name, col, got, want)
		}
	}
	nm, _ := statusOnVulns(t).handleTabBarClick(0)
	if got, want := modelOf(nm).activeTab, tabIndex(t, m, "Missing"); got != want {
		t.Errorf("click on the left marker selected tab %d, want %d (Missing)", got, want)
	}

	m = loadedStatus(80, 24)
	row = windowRow(m, contentWidthIn(m.screenLayout()))
	if !strings.HasSuffix(strings.TrimRight(row, " "), "›") {
		t.Fatalf("Dashboard active should end in the right marker: %q", row)
	}
	nm, _ = m.handleTabBarClick(lipgloss.Width(row) - 1)
	if got, want := modelOf(nm).activeTab, tabIndex(t, m, "Config"); got != want {
		t.Errorf("click on the right marker selected tab %d, want %d (Config)", got, want)
	}
}

// tabIndex is the index of the tab named name on m's panel.
func tabIndex(t *testing.T, m *Model, name string) int {
	t.Helper()
	for i, tab := range m.tabs {
		if tab.name == name {
			return i
		}
	}
	t.Fatalf("no tab named %q", name)
	return -1
}

// TestTabWindowAt60CellContent: a bar given 60 cells of content width fits whole
// (Status is 60 cells, Formulae 48), so every tab is drawn with no marker. A 60
// column terminal is below minTerminalWidth, so this is a width, not a screen.
func TestTabWindowAt60CellContent(t *testing.T) {
	m := loadedStatus(80, 24)
	m = pressStatus(m, "2")
	for _, panel := range []PanelID{PanelStatus, PanelFormulae} {
		m.activePanel = panel
		m.tabs = panelTabs[panel]
		for i := range m.tabs {
			m.activeTab = i
			checkTabWindow(t, m, 60)
		}
	}
}

// TestTabWindowNarrow: at 40 columns both bars are windowed on every tab, so the
// rule holds on every panel with tabs, not only Status at 80x24.
func TestTabWindowNarrow(t *testing.T) {
	m := loadedStatus(80, 24)
	for _, panel := range []PanelID{PanelStatus, PanelFormulae} {
		m.activePanel = panel
		m.tabs = panelTabs[panel]
		for i := range m.tabs {
			m.activeTab = i
			checkTabWindow(t, m, 40)
		}
	}
}

// TestTabRowUnchangedWhenItFits: on every panel with tabs, a bar that fits the
// width is drawn as the unwindowed bar is, with no marker.
func TestTabRowUnchangedWhenItFits(t *testing.T) {
	m := newTestModel()
	m = updateModel(m, tea.WindowSizeMsg{Width: 140, Height: 40})
	cw := contentWidthIn(m.screenLayout())
	for p := range m.panels {
		m.activePanel = PanelID(p)
		m.tabs = panelTabs[m.activePanel]
		for i := range m.tabs {
			m.activeTab = i
			if got, want := m.renderTabBar(cw), m.renderTabBar(0); got != want {
				t.Errorf("panel %v tab %d: bar at 140x40 differs from the unwindowed bar", m.activePanel, i)
			}
		}
	}
}
