package gui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

var sgr = regexp.MustCompile("\x1b\\[[0-9;:]*m")

// plainRows returns the view as one string per row with colour codes
// removed, so a test can compare cells by position.
func plainRows(view string) [][]rune {
	var rows [][]rune
	for _, row := range strings.Split(view, "\n") {
		rows = append(rows, []rune(sgr.ReplaceAllString(row, "")))
	}
	return rows
}

// TestModalFloatsOverFrame locks AC-02: an open modal is composited on top
// of the normal frame, centred, and changes no cell outside its own box.
func TestModalFloatsOverFrame(t *testing.T) {
	modals := []struct {
		name string
		new  func() modal.Modal
	}{
		{"confirm", func() modal.Modal { return modal.NewConfirmModal("Uninstall bash", "Remove bash?") }},
		{"input", func() modal.Modal { return modal.NewInputModal("Tap name") }},
		{"menu", func() modal.Modal { return modal.NewMenuModal("Trust", []string{"Trust tap", "Cancel"}) }},
	}
	for _, sz := range [][2]int{{140, 40}, {80, 24}, {200, 60}} {
		for _, md := range modals {
			t.Run(fmt.Sprintf("%s/%dx%d", md.name, sz[0], sz[1]), func(t *testing.T) {
				base := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
				withModal := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
				withModal.activeModal = md.new()

				overlay := plainRows(style.ActiveBorder.Render(md.new().View().Content))
				oh := len(overlay)
				ow := len(overlay[0])
				x := (sz[0] - ow) / 2
				y := (sz[1] - oh) / 2

				before := plainRows(base.View().Content)
				after := plainRows(withModal.View().Content)
				assertFrameSize(t, withModal.View().Content, sz[0], sz[1])

				for r := range before {
					for c := range before[r] {
						inBox := r >= y && r < y+oh && c >= x && c < x+ow
						if !inBox && before[r][c] != after[r][c] {
							t.Fatalf("cell (%d,%d) outside the modal changed: %q -> %q", c, r, before[r][c], after[r][c])
						}
					}
				}
				for i := 0; i < oh; i++ {
					got := string(after[y+i][x : x+ow])
					want := string(overlay[i])
					if got != want {
						t.Fatalf("modal row %d not centred at (%d,%d):\nwant %q\n got %q", i, x, y, want, got)
					}
				}
			})
		}
	}
}

// TestMouseIgnoredWhileModalOpen keeps the pre-v2 mouse rule: a click while
// a modal is open reaches neither the sidebar nor the main panel.
func TestMouseIgnoredWhileModalOpen(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.cfg.GUI.Mouse = true
	m.activeModal = modal.NewConfirmModal("Uninstall bash", "Remove bash?")

	heights := m.computeContentHeights()
	y := heights[0] + 2 + 1 // the second sidebar box, where a click would switch to Formulae
	nm, _ := m.Update(tea.MouseClickMsg{X: 2, Y: y, Button: tea.MouseLeft})
	got := nm.(Model)
	if got.activePanel != PanelStatus {
		t.Fatalf("click with a modal open changed the panel to %v", got.activePanel)
	}
	if got.activeModal == nil {
		t.Fatal("click with a modal open dismissed the modal")
	}
}
