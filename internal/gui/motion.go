package gui

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/anim"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
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

// WithoutMotion turns the motion off for this model, whatever
// LAZYBREW_ANIMATIONS says. The test helpers use it, so the frames they build
// are the resting frames.
func (m *Model) WithoutMotion() *Model {
	m.motionOn = false
	return m
}

// MotionOn reports whether the model animates.
func (m *Model) MotionOn() bool {
	return m.motionOn
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
	cmd = tea.Batch(cmd, m.motionTargets())
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
	cmd := m.motionTargets()
	m.followSidebar()
	if m.motion.Settled() {
		m.motionAt = time.Time{}
		return m, cmd
	}
	m.motionTicking = true
	return m, tea.Batch(cmd, anim.Tick())
}

// The names of the animated modal values: how far above its resting row the
// modal sits, and how much of the muted backdrop has faded in (0 to 1).
const (
	modalDrop = "modal:drop"
	modalFade = "modal:fade"
	// modalDropRows is how many rows above its resting place a new modal
	// starts; it springs down onto it.
	modalDropRows = 3
)

// motionModal starts the entrance of a modal that has just opened: it springs
// down from modalDropRows above its place, and the backdrop fades in. Closing
// needs no motion: the modal is gone at once.
func (m *Model) motionModal() {
	if m.activeModal == nil {
		m.motionShown = nil
		return
	}
	if m.activeModal == m.motionShown {
		return
	}
	m.motionShown = m.activeModal
	m.motion.Snap(modalDrop, modalDropRows)
	m.motion.Snap(modalFade, 0)
	m.setMotionTarget(modalDrop, 0)
	m.setMotionTarget(modalFade, 1)
}

// modalDropAt is how many rows the open modal still sits above its place.
func (m Model) modalDropAt() int {
	if !m.motionOn {
		return 0
	}
	v, _ := m.motion.Get(modalDrop)
	return int(math.Round(v))
}

// backdropFadeAt is how far the backdrop has faded to the muted role: 0 is the
// frame's own text role, 1 is the muted role, the resting backdrop.
func (m Model) backdropFadeAt() float64 {
	if !m.motionOn {
		return 1
	}
	v, ok := m.motion.Get(modalFade)
	if !ok {
		return 1
	}
	return v
}

// The names of a toast's animated values. A toast is named by its pointer, so
// a toast the stack has dropped leaves names no live toast owns.
//
// dx and dy are how far the toast is drawn from its slot, in cells: a new toast
// starts at the right edge (dx), and a toast whose slot moved starts from the
// slot it left (dy). slot is the row the toast last rested on.
func toastKey(t *modal.Toast) string {
	return fmt.Sprintf("toast:%p", t)
}

func toastName(t *modal.Toast, part string) string {
	return toastKey(t) + ":" + part
}

// motionToasts springs the toast stack. A new toast slides in from the right
// edge into its slot. A toast whose slot moved, because the stack grew or
// lost one, springs from where it was. A resize snaps the stack, and a toast
// that has left the stack is forgotten.
func (m *Model) motionToasts(resized bool) {
	alive := make(map[string]bool, len(m.toasts))
	for i, p := range m.restingToastPlacements() {
		t := m.toasts[i]
		alive[toastKey(t)] = true
		slot := toastName(t, "slot")
		dy := toastName(t, "dy")
		last, seen := m.motion.Get(slot)
		switch {
		case !seen:
			m.motion.Snap(toastName(t, "dx"), float64(m.width-p.x))
			m.setMotionTarget(toastName(t, "dx"), 0)
		case resized:
			m.motion.Snap(dy, 0)
		case int(last) != p.y:
			// Keep the offset the toast is drawn at now, so the move is
			// continuous when it is retargeted mid-flight.
			now, _ := m.motion.Get(dy)
			m.motion.Snap(dy, now+last-float64(p.y))
			m.setMotionTarget(dy, 0)
		}
		m.motion.Snap(slot, float64(p.y))
	}
	for _, name := range m.motion.Names() {
		i := strings.LastIndex(name, ":")
		if i < 0 {
			continue
		}
		if key := name[:i]; strings.HasPrefix(key, "toast:") && !alive[key] {
			m.motion.Remove(name)
		}
	}
}

// toastOffset is how far a toast is drawn from its slot, in cells, along one
// axis: dx or dy. A toast with no offset is drawn in its slot.
func (m Model) toastOffset(t *modal.Toast, part string) int {
	v, _ := m.motion.Get(toastName(t, part))
	return int(math.Round(v))
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
// state. Each animated feature adds its targets here. The command it returns
// is the progress bar's own frame, when the bar has started to move.
func (m *Model) motionTargets() tea.Cmd {
	resized := m.motionSize != [2]int{m.width, m.height}
	m.motionSize = [2]int{m.width, m.height}
	m.motionTabs()
	m.motionSidebar(resized)
	m.motionModal()
	m.motionToasts(resized)
	return m.motionProgress()
}

// motionProgress aims the batch's progress bar at the share of the batch that
// is done. A new operation gets a fresh bar, so its fill starts from empty and
// does not slide down from the last batch's. With animations off the bar is
// never aimed: it is drawn at its share at once.
func (m *Model) motionProgress() tea.Cmd {
	if !m.motionOn || m.opState == nil {
		return nil
	}
	if m.opState != m.opBarOp {
		m.opBar = newProgressBar()
		m.opBarOp = m.opState
	}
	done, total := operationProgress(m.opState)
	if total <= 0 {
		return nil
	}
	share := float64(done) / float64(total)
	if share == m.opBar.Percent() {
		return nil
	}
	return m.opBar.SetPercent(share)
}

// The names of the animated sidebar boxes: one value per panel, by index.
func sidebarName(i int) string {
	return "sidebar:" + strconv.Itoa(i)
}

// motionSidebar aims each sidebar box at its resting height. A resize snaps
// the boxes, since the total height changes with the window; a focus change
// springs them. The list windows follow the boxes (followSidebar).
func (m *Model) motionSidebar(resized bool) {
	if !m.ready {
		return
	}
	for i, h := range m.restingContentHeights() {
		if resized {
			m.motion.Snap(sidebarName(i), float64(h))
		} else {
			m.setMotionTarget(sidebarName(i), float64(h))
		}
	}
	if resized {
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
