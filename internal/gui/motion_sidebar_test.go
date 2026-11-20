package gui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/config"
)

// sidebarModel is a sized model with animations on, on the Status panel.
func sidebarModel(t *testing.T, w, h int) *Model {
	t.Helper()
	t.Setenv(config.AnimationsEnv, "1")
	return updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
}

// boxesFill reports whether the sidebar boxes add up to the sidebar's height,
// and that every box has at least one list row.
func boxesFill(t *testing.T, m *Model) []int {
	t.Helper()
	heights := m.computeContentHeights()
	total := 0
	for _, h := range heights {
		if h < 1 {
			t.Fatalf("a sidebar box has %d rows: %v", h, heights)
		}
		total += h + 2
	}
	if want := m.screenLayout().bodyH; total != want {
		t.Fatalf("sidebar boxes total %d rows; want the sidebar's %d: %v", total, want, heights)
	}
	return heights
}

func TestSidebarBoxesSpringWithFocus(t *testing.T) {
	m := settle(t, sidebarModel(t, 140, 40))
	old := m.computeContentHeights()
	m = sendKey(m, "2")
	rest := m.restingContentHeights()
	if fmt.Sprint(old) == fmt.Sprint(rest) {
		t.Fatal("test focus change does not move the boxes")
	}

	// Mid-flight the boxes are between their old and new heights, and fill the
	// sidebar exactly at every frame.
	m = frames(m, 3)
	mid := boxesFill(t, m)
	for i := range mid {
		lo, hi := min(old[i], rest[i]), max(old[i], rest[i])
		if mid[i] < lo || mid[i] > hi {
			t.Fatalf("box %d is %d rows mid-flight; want between %d and %d", i, mid[i], lo, hi)
		}
	}
	if fmt.Sprint(mid) == fmt.Sprint(old) {
		t.Fatalf("three frames did not move the boxes: %v", mid)
	}
	for i := 0; i < 60 && !m.motion.Settled(); i++ {
		m = frames(m, 1)
		boxesFill(t, m)
	}

	// Settled, the boxes are the resting layout.
	if got := boxesFill(t, m); fmt.Sprint(got) != fmt.Sprint(rest) {
		t.Fatalf("settled heights = %v; want the resting %v", got, rest)
	}
}

func TestSidebarFramesAreExactSizeWhileItMoves(t *testing.T) {
	for _, sz := range [][2]int{{140, 40}, {80, 24}} {
		m := settle(t, sidebarModel(t, sz[0], sz[1]))
		m = sendKey(m, "4")
		for i := 0; i < 12; i++ {
			m = frames(m, 1)
			side := m.renderSidebar()
			if w, h := lipgloss.Width(side), lipgloss.Height(side); h != m.screenLayout().bodyH || w != m.screenLayout().sideW {
				t.Fatalf("%dx%d frame %d: sidebar is %dx%d", sz[0], sz[1], i, w, h)
			}
			boxesFill(t, m)
		}
	}
}

func TestSelectedRowStaysVisibleWhileBoxesMove(t *testing.T) {
	m := sidebarModel(t, 140, 40)
	p := m.panels[PanelFormulae]
	p.items = make([]string, 80)
	for i := range p.items {
		p.items[i] = fmt.Sprintf("pkg%02d 1.0", i)
	}
	p.loaded = true
	p.selected = 60
	p.normalize()
	m = settle(t, m)

	m = sendKey(m, "2")
	for i := 0; i < 15; i++ {
		m = frames(m, 1)
		p := m.panels[PanelFormulae]
		if p.selected < p.offset || p.selected >= p.offset+p.visibleRows {
			t.Fatalf("frame %d: selected row %d is off the window [%d,%d)",
				i, p.selected, p.offset, p.offset+p.visibleRows)
		}
	}
}

func TestSidebarSnapsWhenMotionIsOff(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "0")
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(m, "2")
	if fmt.Sprint(m.computeContentHeights()) != fmt.Sprint(m.restingContentHeights()) {
		t.Fatal("the boxes move with animations off")
	}
	if _, cmd := m.Update(frameAt(m.motionAt, 1)); cmd != nil {
		t.Fatal("a frame ticked with animations off")
	}
}

func TestResizeSnapsTheBoxes(t *testing.T) {
	m := settle(t, sidebarModel(t, 140, 40))
	m = sendKey(m, "2")
	m = frames(m, 2)
	m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 30})
	if got, want := boxesFill(t, m), m.restingContentHeights(); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("after a resize the boxes are %v; want the resting %v", got, want)
	}
}

func TestRoundHeightsSumsToTheTotal(t *testing.T) {
	cases := []struct {
		vals  []float64
		total int
		want  []int
	}{
		{[]float64{10, 4, 4}, 18, []int{10, 4, 4}},
		{[]float64{9.6, 4.2, 4.2}, 18, []int{10, 4, 4}},
		{[]float64{2.5, 2.5, 2.5, 2.5}, 10, []int{3, 2, 3, 2}}, // halves round up
		{[]float64{17.9, 0.05}, 18, []int{18, 0}},
	}
	for _, c := range cases {
		got := roundHeights(c.vals, c.total)
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("roundHeights(%v, %d) = %v; want %v", c.vals, c.total, got, c.want)
		}
	}
}
