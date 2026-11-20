package gui

import (
	"math"
	"os"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/anim"
)

// animationsOn reports whether the screen moves. gui.animations is the default;
// LAZYBREW_ANIMATIONS overrides it, as LAZYBREW_ICONS overrides gui.icons. Any
// other value of the variable is ignored, as iconSet ignores an unknown name.
func animationsOn(cfg *config.Config) bool {
	switch os.Getenv(config.AnimationsEnv) {
	case "0":
		return false
	case "1":
		return true
	}
	return cfg.GUI.Animations
}

// setMotionTarget sets where a named value is going. With animations off the
// value is put there at once, so no tick is ever needed.
func (m *Model) setMotionTarget(name string, to float64) {
	if m.motionOn {
		m.motion.Target(name, to)
		return
	}
	m.motion.Snap(name, to)
}

// settleMotion makes the animated values agree with the state the update just
// produced, and starts the frame tick when something has to move. It runs at
// the end of every Update, so no caller has to remember to schedule motion.
func (m Model) settleMotion(cmd tea.Cmd) (Model, tea.Cmd) {
	m.motionTargets()
	if m.motionOn && !m.motionTicking && !m.motion.Settled() {
		m.motionTicking = true
		m.motionAt = time.Time{}
		return m, tea.Batch(cmd, anim.Tick())
	}
	return m, cmd
}

// stepMotion advances every value by the time since the last frame, and
// schedules the next frame only while something still moves. A frame that
// settles everything schedules nothing, so an idle screen never ticks.
func (m Model) stepMotion(at time.Time) (Model, tea.Cmd) {
	m.motionTicking = false
	m.motion.Step(m.motionStep(at).Seconds())
	m.motionAt = at
	m.motionTargets()
	m.followSidebar()
	if m.motion.Settled() {
		m.motionAt = time.Time{}
		return m, nil
	}
	m.motionTicking = true
	return m, anim.Tick()
}

// motionStep is the time one frame advances: the gap since the last frame,
// capped at anim.MaxStep. The first frame of a run has no last frame, so it
// advances one frame interval.
func (m Model) motionStep(at time.Time) time.Duration {
	if m.motionAt.IsZero() {
		return anim.FrameInterval
	}
	return min(max(at.Sub(m.motionAt), 0), anim.MaxStep)
}

// motionTargets sets where every animated value is going, from the model's
// state. Each animated feature adds its targets here.
func (m *Model) motionTargets() {
	m.motionTabs()
	m.motionSidebar()
}

// The names of the animated sidebar boxes: one value per panel, by index.
func sidebarName(i int) string {
	return "sidebar:" + strconv.Itoa(i)
}

// motionSidebar aims each sidebar box at its resting height. A resize snaps
// the boxes, since the total height changes with the window; a focus change
// springs them. The list windows follow the boxes (followSidebar).
func (m *Model) motionSidebar() {
	if !m.ready {
		return
	}
	snap := m.motionSize != [2]int{m.width, m.height}
	m.motionSize = [2]int{m.width, m.height}
	for i, h := range m.restingContentHeights() {
		if snap {
			m.motion.Snap(sidebarName(i), float64(h))
		} else {
			m.setMotionTarget(sidebarName(i), float64(h))
		}
	}
	if snap {
		m.followSidebar()
	}
}

// followSidebar sets each list's rows to its box's current height, and keeps
// the selected row in the window when the rows change. It runs on each frame
// that moves the boxes, and after a resize snaps them.
func (m *Model) followSidebar() {
	heights := m.computeContentHeights()
	for i, p := range m.panels {
		if p.visibleRows != heights[i] {
			p.visibleRows = heights[i]
			p.normalize()
		}
	}
}

// computeContentHeights is restingContentHeights while the sidebar moves: each
// box's height is its animated value, rounded to whole rows. The rounding
// keeps the total exact, so the boxes always fill the sidebar.
func (m Model) computeContentHeights() []int {
	rest := m.restingContentHeights()
	if !m.motionOn {
		return rest
	}
	vals := make([]float64, len(rest))
	for i := range rest {
		v, ok := m.motion.Get(sidebarName(i))
		if !ok {
			return rest
		}
		vals[i] = v
	}
	total := 0
	for _, h := range rest {
		total += h
	}
	return roundHeights(vals, total)
}

// roundHeights rounds animated heights to whole rows that sum to total. Each
// box ends where the running sum, rounded, says it ends, and the last box ends
// at total: the rows are exact, and each is within a row of its value.
func roundHeights(vals []float64, total int) []int {
	heights := make([]int, len(vals))
	sum, prev := 0.0, 0
	for i, v := range vals {
		sum += v
		end := int(math.Round(sum))
		if i == len(vals)-1 {
			end = total
		}
		heights[i] = end - prev
		prev = end
	}
	return heights
}

// The names of the animated values of the tab underline: its left column and
// its width, in cells from the left of the tab bar.
const (
	tabUnderlineX = "tab:x"
	tabUnderlineW = "tab:w"
)

// motionTabs aims the tab underline at the active tab. A switch to another
// panel snaps it instead: that panel's tabs are not the ones it slid from.
func (m *Model) motionTabs() {
	if len(m.tabs) == 0 {
		return
	}
	x, w := m.activeTabSpan()
	if m.motionPanel != m.activePanel {
		m.motion.Snap(tabUnderlineX, float64(x))
		m.motion.Snap(tabUnderlineW, float64(w))
	} else {
		m.setMotionTarget(tabUnderlineX, float64(x))
		m.setMotionTarget(tabUnderlineW, float64(w))
	}
	m.motionPanel = m.activePanel
}

// tabUnderlineSpan is where the underline is drawn: its column and width at
// this frame, and whether it is still on the way to the active tab. Before a
// first target exists, and when it is at rest, the resting bar is drawn.
func (m Model) tabUnderlineSpan() (x, w float64, moving bool) {
	x, okX := m.motion.Get(tabUnderlineX)
	w, okW := m.motion.Get(tabUnderlineW)
	if !okX || !okW {
		return 0, 0, false
	}
	tx, tw := m.activeTabSpan()
	return x, w, x != float64(tx) || w != float64(tw)
}
