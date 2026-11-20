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
