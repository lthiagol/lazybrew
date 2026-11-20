package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// mutedPrefix is the SGR sequence that opens the muted role, so a test can
// tell whether a row starts in it.
func mutedPrefix() string {
	sample := lipgloss.NewStyle().Foreground(style.MutedColor).Render("x")
	return sample[:strings.Index(sample, "x")]
}

// TestDimBackdropMutesEveryRowButTheFooter locks AC-03: while a modal is open
// every row behind it is drawn in the muted role, the footer is not, and the
// text of the frame does not change. Without a modal nothing is muted.
func TestDimBackdropMutesEveryRowButTheFooter(t *testing.T) {
	muted := lipgloss.NewStyle().Foreground(style.MutedColor)
	for _, sz := range [][2]int{{140, 40}, {80, 24}, {200, 60}} {
		t.Run(fmt.Sprintf("%dx%d", sz[0], sz[1]), func(t *testing.T) {
			w, h := sz[0], sz[1]
			base := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h}).View().Content
			withModal := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
			withModal.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Uninstall bash", "Remove bash?")
			with := withModal.View().Content

			overlay := withModal.activeModal.View().Content
			oh, ow := lipgloss.Height(overlay), lipgloss.Width(overlay)
			y0, x0 := (h-oh)/2, (w-ow)/2

			baseRows := strings.Split(base, "\n")
			withRows := strings.Split(with, "\n")
			assertFrameSize(t, with, w, h)
			for r := 0; r < h-1; r++ {
				if r < y0 || r >= y0+oh {
					if want := muted.Render(ansi.Strip(baseRows[r])); withRows[r] != want {
						t.Fatalf("row %d behind the modal is not the muted base row:\nwant %q\n got %q", r, want, withRows[r])
					}
					continue
				}
				// A row the modal crosses still starts in the muted role left of the box.
				if x0 > 0 && !strings.HasPrefix(withRows[r], mutedPrefix()) {
					t.Fatalf("row %d left of the modal is not muted: %q", r, withRows[r])
				}
			}
			// The footer shows the modal's keys, and is not muted.
			if strings.HasPrefix(withRows[h-1], mutedPrefix()) {
				t.Fatalf("the footer is muted with a modal open: %q", withRows[h-1])
			}
			if !strings.Contains(ansi.Strip(withRows[h-1]), "esc cancel") {
				t.Fatalf("the footer does not show the modal's keys: %q", ansi.Strip(withRows[h-1]))
			}
			if strings.HasPrefix(baseRows[0], mutedPrefix()) {
				t.Fatalf("the frame without a modal is already muted: %q", baseRows[0])
			}
		})
	}
}

// TestDimBackdropKeepsText checks the style-only rule on its own: the dimmed
// rows carry the same text, and the last row is returned untouched.
func TestDimBackdropKeepsText(t *testing.T) {
	in := "\x1b[31mred\x1b[m text\n\x1b[1mfooter\x1b[m"
	out := strings.Split(dimBackdrop(in), "\n")
	if got := ansi.Strip(out[0]); got != "red text" {
		t.Fatalf("dimmed row text = %q, want %q", got, "red text")
	}
	if !strings.HasPrefix(out[0], mutedPrefix()) {
		t.Fatalf("row is not in the muted role: %q", out[0])
	}
	if out[1] != "\x1b[1mfooter\x1b[m" {
		t.Fatalf("footer row changed: %q", out[1])
	}
}
