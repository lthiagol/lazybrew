package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestMainListFillsMainPanel: a list in the main panel shows its rows up to
// the panel's own height. The sidebar box's window used to cap it, which
// left the Status dashboard one row tall at 80x24.
func TestMainListFillsMainPanel(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 40}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
			var rows []string
			for i := 0; i < 9; i++ {
				rows = append(rows, fmt.Sprintf("row-%d", i))
			}
			m.panels[PanelStatus].items = rows

			main := m.renderMainPanel()
			for _, row := range rows {
				if !strings.Contains(main, row) {
					t.Fatalf("main panel does not show %q; the list is capped by the sidebar", row)
				}
			}
		})
	}
}

// TestMainListScrollsWithCursor locks F-r1-01: the main list keeps its own
// window on the cursor, so rows above the cursor show while they fit, and
// the cursor row shows once it moves past the panel's height. The sidebar
// keeps its own scrolling.
func TestMainListScrollsWithCursor(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {140, 40}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
			var rows []string
			for i := 0; i < 30; i++ {
				rows = append(rows, fmt.Sprintf("row-%02d", i))
			}
			m.panels[PanelStatus].items = rows

			for i := 0; i < 6; i++ {
				m = sendKey(m, "j")
			}
			main := m.renderMainPanel()
			for i := 0; i <= 6; i++ {
				if want := fmt.Sprintf("row-%02d", i); !strings.Contains(main, want) {
					t.Fatalf("cursor on row-06: main panel should show %q above the cursor", want)
				}
			}
			if !strings.Contains(main, "▸ row-06") {
				t.Fatal("main panel should draw the cursor row")
			}

			target := m.screenLayout().contentH + 2
			for m.panels[PanelStatus].selected < target {
				m = sendKey(m, "j")
			}
			want := fmt.Sprintf("▸ row-%02d", target)
			if main := m.renderMainPanel(); !strings.Contains(main, want) {
				t.Fatalf("cursor past the panel height: main panel should draw %q", want)
			}
			if side := sgr.ReplaceAllString(m.renderSidebar(), ""); !strings.Contains(side, want) {
				t.Fatalf("sidebar should still scroll to %q", want)
			}
		})
	}
}
