package gui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// tabRowAndUnderline finds the Status tab bar in a frame: the row that names
// the tab, and the rule row under it.
func tabRowAndUnderline(t *testing.T, view, name string) (row, rule string) {
	t.Helper()
	rows := strings.Split(ansi.Strip(view), "\n")
	for i, r := range rows {
		if strings.Contains(r, name) && i+1 < len(rows) && strings.Contains(rows[i+1], "━") {
			return r, rows[i+1]
		}
	}
	t.Fatalf("no tab row naming %s with an underline:\n%s", name, ansi.Strip(view))
	return "", ""
}

// TestStatusTabRowShowsTheActiveTab (F-r1-02, Q-lb31-tabs-80): at 80x24 with
// Vulns active the bar is a window onto the seven tabs: Vulns is in it, the
// tabs cut on the left are marked, and the underline spans Vulns, in both
// themes.
func TestStatusTabRowShowsTheActiveTab(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for theme, th := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
		style.ApplyTheme(th)
		t.Run(theme, func(t *testing.T) {
			m := loadedStatus(80, 24)
			for range 4 {
				m = pressStatus(m, "l")
			}
			if m.activeTab != statusVulnsTab {
				t.Fatalf("four l presses should reach the Vulns tab, got tab %d", m.activeTab)
			}
			row, rule := tabRowAndUnderline(t, m.View().Content, "Vulns")
			if want := "‹ Doctor  Missing  Vulns  Cleanup  Unused"; !strings.Contains(row, want) {
				t.Errorf("tab row at 80x24 is not the window %q: %q", want, row)
			}
			if strings.Contains(row, "Dashboard") {
				t.Errorf("Dashboard is in the window with Vulns active: %q", row)
			}
			if n := strings.Count(rule, "━"); n != len("Vulns")+2 {
				t.Errorf("underline under Vulns is %d cells, want %d: %q", n, len("Vulns")+2, rule)
			}
		})
	}
}

// TestOtherTabBarsKeepTheirGap: the Formulae tab bar keeps its one-cell gap
// either side of each name; only the Status bar is tightened.
func TestOtherTabBarsKeepTheirGap(t *testing.T) {
	m := loadedStatus(80, 24)
	m = pressStatus(m, "2")
	row := strings.Split(ansi.Strip(m.renderTabBar(80)), "\n")[0]
	if !strings.Contains(row, "Info    Deps    Used By") {
		t.Fatalf("the Formulae tab bar changed its gap: %q", row)
	}
}
