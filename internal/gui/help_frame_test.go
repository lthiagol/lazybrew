package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
