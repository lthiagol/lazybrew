package gui

import (
	"os"
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
