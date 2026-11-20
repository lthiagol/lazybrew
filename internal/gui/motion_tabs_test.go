package gui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/config"
)

// motionModel is a sized model with animations on, on the Formulae panel:
// five tabs, the underline on Info.
func motionModel(t *testing.T, w, h int) *Model {
	t.Helper()
	t.Setenv(config.AnimationsEnv, "1")
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
	return sendKey(m, "2")
}

// frames delivers n frame ticks, one 60 Hz frame apart, and returns the model.
// The frames carry on from the last one the model saw, so a run of calls is a
// steady clock, not a clock that restarts at each call.
func frames(m *Model, n int) *Model {
	start := m.motionAt
	if start.IsZero() {
		start = time.Now()
	}
	for i := 1; i <= n; i++ {
		m = updateModel(m, frameAt(start, i))
	}
	return m
}

// settle runs frames until nothing moves, and returns the model.
func settle(t *testing.T, m *Model) *Model {
	t.Helper()
	for i := 0; i < 120; i++ {
		if m.motion.Settled() {
			return m
		}
		m = frames(m, 1)
	}
	t.Fatal("motion did not settle")
	return nil
}

// underlineRun is the first and one-past-last cell of the underline in a tab
// bar's rule row, counted in cells.
func underlineRun(t *testing.T, m *Model) (start, end int) {
	t.Helper()
	rows := strings.Split(ansi.Strip(m.renderTabBar(m.width)), "\n")
	if len(rows) != 2 {
		t.Fatalf("tab bar has %d rows; want 2", len(rows))
	}
	start, end = -1, -1
	for i, r := range []rune(rows[1]) {
		if r == '━' {
			if start < 0 {
				start = i
			}
			end = i + 1
		}
	}
	if start < 0 {
		t.Fatalf("no underline in the tab bar: %q", rows[1])
	}
	return start, end
}

func TestTabUnderlineRestsOnTheActiveTab(t *testing.T) {
	m := settle(t, sendKey(motionModel(t, 140, 40), "]"))
	x, w := m.activeTabSpan()
	if start, end := underlineRun(t, m); start != x || end != x+w {
		t.Fatalf("settled underline = [%d,%d); want the active tab's [%d,%d)", start, end, x, x+w)
	}
}

func TestTabUnderlineSlidesBetweenTabs(t *testing.T) {
	m := settle(t, motionModel(t, 140, 40))
	x0, w0 := m.activeTabSpan()
	m = sendKey(m, "]")
	m = sendKey(m, "]")
	m = sendKey(m, "]")
	x1, w1 := m.activeTabSpan()
	if x1 <= x0 || w1 == w0 {
		t.Fatalf("test tabs %d/%d are not apart: want a wider move", x0, x1)
	}

	// Before any frame the underline is still on the old tab.
	if start, _ := underlineRun(t, m); start != x0 {
		t.Fatalf("first frame underline starts at %d; want the old tab's %d", start, x0)
	}
	// Mid-flight the underline is between the two tabs.
	mid := frames(m, 6)
	start, end := underlineRun(t, mid)
	if start <= x0 || start >= x1 {
		t.Fatalf("mid-flight underline starts at %d; want strictly between %d and %d", start, x0, x1)
	}
	if end <= start {
		t.Fatalf("mid-flight underline is empty: [%d,%d)", start, end)
	}

	// Settled, it is exactly the active tab's underline.
	m = settle(t, mid)
	if start, end := underlineRun(t, m); start != x1 || end != x1+w1 {
		t.Fatalf("settled underline = [%d,%d); want [%d,%d)", start, end, x1, x1+w1)
	}
}

func TestSlidingTabBarAtRestIsTheRestingBar(t *testing.T) {
	// The sliding renderer and the resting renderer must draw the same cells
	// for the same geometry, so the frame does not change when the slide ends.
	m := settle(t, motionModel(t, 140, 40))
	for i := range m.tabs {
		if i > 0 {
			m = sendKey(m, "]")
		}
		m = settle(t, m)
		x, w := m.activeTabSpan()
		if got, want := m.renderTabBarSliding(float64(x), float64(w)), m.renderTabBar(m.width); got != want {
			t.Fatalf("tab %d: sliding bar differs from the resting bar:\n%q\n%q", i, got, want)
		}
	}
}

func TestAnimationsOffDrawsTheUnderlineAtOnce(t *testing.T) {
	t.Setenv(config.AnimationsEnv, "0")
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = sendKey(sendKey(m, "2"), "]")
	if _, _, moving := m.tabUnderlineSpan(); moving {
		t.Fatal("the underline moves with animations off")
	}
	if _, cmd := m.Update(frameAt(time.Now(), 1)); cmd != nil {
		t.Fatal("a frame ticked with animations off")
	}
}

func TestPanelSwitchSnapsTheUnderline(t *testing.T) {
	m := settle(t, motionModel(t, 140, 40))
	m = sendKey(m, "1")
	if _, _, moving := m.tabUnderlineSpan(); moving {
		t.Fatal("the underline slid across to another panel's tabs")
	}
	x, w := m.activeTabSpan()
	if start, end := underlineRun(t, m); start != x || end != x+w {
		t.Fatalf("underline after a panel switch = [%d,%d); want [%d,%d)", start, end, x, x+w)
	}
}

func TestKeysWorkWhileTheUnderlineSlides(t *testing.T) {
	m := settle(t, motionModel(t, 140, 40))
	m = sendKey(m, "]")
	m = frames(m, 2)
	m = sendKey(m, "]")
	if m.activeTab != 2 {
		t.Fatalf("activeTab = %d mid-slide; want 2", m.activeTab)
	}
	m = sendKey(m, "[")
	if m.activeTab != 1 {
		t.Fatalf("activeTab = %d after [; want 1", m.activeTab)
	}
}

func TestSlidingFramesAreTerminalSize(t *testing.T) {
	for _, sz := range [][2]int{{140, 40}, {80, 24}} {
		m := settle(t, motionModel(t, sz[0], sz[1]))
		m = sendKey(m, "]")
		for i := 0; i < 8; i++ {
			m = frames(m, 1)
			v := m.View().Content
			if w, h := lipgloss.Width(v), lipgloss.Height(v); w != sz[0] || h != sz[1] {
				t.Fatalf("%dx%d frame %d is %dx%d", sz[0], sz[1], i, w, h)
			}
		}
	}
}
