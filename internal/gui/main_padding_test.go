package gui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// The main area keeps one blank column either side of its content, inside its
// border, so text never touches the frame (v2-09 AC-01). Overlays cover the
// main box, so they are not measured.
func TestMainAreaHasInnerPadding(t *testing.T) {
	overlays := map[string]bool{
		"help": true, "confirm modal": true, "input modal": true,
		"menu modal": true, "toast": true, "toast stack": true,
	}
	for _, sz := range [][2]int{{140, 40}, {80, 24}} {
		for _, st := range frameStates {
			if overlays[st.name] {
				continue
			}
			t.Run(fmt.Sprintf("%s/%dx%d", st.name, sz[0], sz[1]), func(t *testing.T) {
				m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
				m = st.setup(m)
				l := m.screenLayout()
				rows := strings.Split(ansi.Strip(m.View().Content), "\n")
				// The rows between the main box's top and bottom borders.
				left, right := l.sideW+1, l.sideW+l.mainW-2
				for y := headerRows + 1; y < headerRows+l.mainH-1; y++ {
					if got := ansi.Strip(ansi.Cut(rows[y], left, left+1)); got != " " {
						t.Errorf("row %d: column %d is %q, want the blank padding", y, left, got)
					}
					if got := ansi.Strip(ansi.Cut(rows[y], right, right+1)); got != " " {
						t.Errorf("row %d: column %d is %q, want the blank padding", y, right, got)
					}
				}
			})
		}
	}
}

// The Log box keeps the same one column of padding either side as the main box
// (v2-09 r3). Its rows are those between the Log's own borders.
func TestLogBoxHasInnerPadding(t *testing.T) {
	for _, sz := range [][2]int{{140, 40}, {80, 24}} {
		for _, lines := range []bool{false, true} {
			name := fmt.Sprintf("empty/%dx%d", sz[0], sz[1])
			if lines {
				name = fmt.Sprintf("lines/%dx%d", sz[0], sz[1])
			}
			t.Run(name, func(t *testing.T) {
				m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
				if lines {
					m.commandLog.Append("brew outdated")
				}
				l := m.screenLayout()
				if l.logH == 0 {
					t.Fatal("fixture: the Log should be shown")
				}
				rows := strings.Split(ansi.Strip(m.View().Content), "\n")
				left, right := l.sideW+1, l.sideW+l.mainW-2
				logTop := headerRows + l.mainH
				for y := logTop + 1; y < logTop+l.logH-1; y++ {
					if got := ansi.Strip(ansi.Cut(rows[y], left, left+1)); got != " " {
						t.Errorf("Log row %d: column %d is %q, want the blank padding", y, left, got)
					}
					if got := ansi.Strip(ansi.Cut(rows[y], right, right+1)); got != " " {
						t.Errorf("Log row %d: column %d is %q, want the blank padding", y, right, got)
					}
				}
			})
		}
	}
}
