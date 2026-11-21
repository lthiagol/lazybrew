package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// TestHelpKeepsBorderAndHint: the help screen cuts its text to fit, so at
// any size its box still closes and the close hint is still the last row.
func TestHelpKeepsBorderAndHint(t *testing.T) {
	for _, sz := range [][2]int{{140, 40}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
			m = sendKey(m, "?")
			rows := plainRows(m.View().Content)
			h := sz[1]
			// The footer shows help's own keys, so the close hint is the last row.
			if got := string(rows[h-1]); !strings.Contains(got, "?/esc close") {
				t.Fatalf("last row should be the footer's close hint, got %q", got)
			}
			if got := rows[h-2][0]; got != '╰' {
				t.Fatalf("help box should close on row %d, got %q", h-2, string(rows[h-2]))
			}
		})
	}
}

// TestHelpNeverClipsAWord (lb29 r2 F-r1-01): no help line is cut at 80x24, 100x30,
// 140x40 or 200x50 on either theme. A cut line ends in the … that textutil.Fit
// leaves, and every page is checked, not only the rows on show.
func TestHelpNeverClipsAWord(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	themes := []struct {
		name  string
		theme *style.Theme
	}{{"dark", style.DarkTheme()}, {"light", style.LightTheme()}}
	for _, th := range themes {
		style.ApplyTheme(th.theme)
		for _, sz := range [][2]int{{80, 24}, {100, 30}, {140, 40}, {200, 50}} {
			t.Run(fmt.Sprintf("%s %dx%d", th.name, sz[0], sz[1]), func(t *testing.T) {
				m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
				for _, row := range helpRows(m.helpTextWidth()) {
					if plain := plainLine(row); strings.Contains(plain, "…") {
						t.Errorf("help line is clipped: %q", plain)
					}
				}
			})
		}
	}
}
