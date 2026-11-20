package gui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// TestToastFloatsAboveBottomBar locks AC-04: a toast is drawn flush right on
// the last row above the bottom bar, and no other row or cell changes.
func TestToastFloatsAboveBottomBar(t *testing.T) {
	const message = "Install completed"
	for _, sz := range [][2]int{{140, 40}, {80, 24}, {200, 60}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			w, h := sz[0], sz[1]
			base := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
			withToast := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
			withToast.toast = modal.NewToast(message, modal.ToastSuccess)

			withToastView := withToast.View().Content
			assertFrameSize(t, withToastView, w, h)

			before := plainRows(base.View().Content)
			after := plainRows(withToastView)
			toastRow := h - bottomBarRows - 1
			toastText := []rune("  " + message + "  ")
			x := w - len(toastText)

			for r := range before {
				for c := range before[r] {
					covered := r == toastRow && c >= x
					if !covered && before[r][c] != after[r][c] {
						t.Fatalf("cell (%d,%d) changed with a toast showing: %q -> %q", c, r, before[r][c], after[r][c])
					}
				}
			}
			if got := string(after[toastRow][x:]); got != string(toastText) {
				t.Fatalf("toast row %d is %q at x=%d, want %q", toastRow, got, x, string(toastText))
			}
		})
	}
}
