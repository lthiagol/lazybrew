package gui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// tabRowAndUnderline finds the Status tab bar in a frame: the row that names
// the Dashboard tab, and the rule row under it.
func tabRowAndUnderline(t *testing.T, view string) (row, rule string) {
	t.Helper()
	rows := strings.Split(ansi.Strip(view), "\n")
	for i, r := range rows {
		if strings.Contains(r, "Dashboard") && i+1 < len(rows) {
			return r, rows[i+1]
		}
	}
	t.Fatalf("no tab row naming Dashboard:\n%s", ansi.Strip(view))
	return "", ""
}

// TestStatusTabRowShowsEveryName: at 80x24 all five Status tab names are whole
// in the tab row, in both themes, and the underline spans the active one
// (F-r1-02).
func TestStatusTabRowShowsEveryName(t *testing.T) {
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
			row, rule := tabRowAndUnderline(t, m.View().Content)
			for _, name := range []string{"Dashboard", "Config", "Doctor", "Missing", "Vulns"} {
				if !strings.Contains(row, name) {
					t.Errorf("tab row lacks %q at 80x24: %q", name, row)
				}
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
