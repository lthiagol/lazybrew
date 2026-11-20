package gui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// TestToastFloatsAboveBottomBar locks AC-04: a toast is a bordered box drawn
// flush right, its last row on the last body row above the bottom bar. No cell
// outside the box changes.
func TestToastFloatsAboveBottomBar(t *testing.T) {
	const message = "Install completed"
	for _, sz := range [][2]int{{140, 40}, {80, 24}, {200, 60}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			w, h := sz[0], sz[1]
			base := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
			withToast := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
			withToast.toasts = []*modal.Toast{modal.NewToast(message, modal.ToastSuccess)}

			withToastView := withToast.View().Content
			assertFrameSize(t, withToastView, w, h)

			box := withToast.toastBoxes()[0]
			bw, bh := lipgloss.Width(box), lipgloss.Height(box)
			x, y := w-bw, h-bottomBarRows-bh
			if y+bh-1 != h-bottomBarRows-1 {
				t.Fatalf("toast box ends at row %d, want the last body row %d", y+bh-1, h-bottomBarRows-1)
			}

			before := plainRows(base.View().Content)
			after := plainRows(withToastView)
			boxRows := plainRows(box)
			for r := range before {
				for c := range before[r] {
					inBox := r >= y && r < y+bh && c >= x
					if !inBox && before[r][c] != after[r][c] {
						t.Fatalf("cell (%d,%d) changed with a toast showing: %q -> %q", c, r, before[r][c], after[r][c])
					}
				}
			}
			for i := range boxRows {
				if got, want := string(after[y+i][x:]), string(boxRows[i]); got != want {
					t.Fatalf("toast row %d is %q, want %q", i, got, want)
				}
			}
		})
	}
}
