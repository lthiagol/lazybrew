package gui

import (
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// maxToasts is how many toasts the stack shows; a newer one pushes the oldest out.
const maxToasts = 3

// toastTickInterval is how often the dismissal tick re-checks the stack. The
// tests set it to a millisecond, so no test waits on it.
var toastTickInterval = 500 * time.Millisecond

// setToast pushes a toast onto the stack and arms the dismissal tick if none
// is running. Every toast producer must go through this: the tick is the only
// thing that removes a toast, so a toast with no running tick would stay on
// screen forever (r1 finding 4). It returns nil when a tick is already running.
func (m *Model) setToast(message string, t modal.ToastType) tea.Cmd {
	kept := m.toasts
	if len(kept) >= maxToasts {
		kept = kept[len(kept)-maxToasts+1:]
	}
	m.toasts = append(slices.Clone(kept), modal.NewToast(message, t))
	if m.toastTicking {
		return nil
	}
	m.toastTicking = true
	return modal.ToastTick(toastTickInterval)
}

// liveToasts keeps the toasts not yet past their lifetime at now.
func liveToasts(toasts []*modal.Toast, now time.Time) []*modal.Toast {
	var live []*modal.Toast
	for _, t := range toasts {
		if !t.Expired(now) {
			live = append(live, t)
		}
	}
	return live
}

// toastMessageWidth is how wide a toast message may run before it is cut with
// an ellipsis: half the frame, and never more than 60 cells.
func toastMessageWidth(frameWidth int) int {
	return min(frameWidth/2, 60)
}

// newestToast is the toast drawn lowest in the stack, or nil when there is none.
func (m Model) newestToast() *modal.Toast {
	if len(m.toasts) == 0 {
		return nil
	}
	return m.toasts[len(m.toasts)-1]
}

// toastIcon is the icon a toast of type t draws, from the icon set. The set is
// empty under show_icons: false, and then the icon is "".
func (m Model) toastIcon(t modal.ToastType) string {
	switch t {
	case modal.ToastError:
		return m.icons.Error
	case modal.ToastWarning:
		return m.icons.Warn
	case modal.ToastInfo:
		return m.icons.Info
	default:
		return m.icons.OK
	}
}

// toastBoxes draws each toast, oldest first, so the newest box is the last.
func (m Model) toastBoxes() []string {
	boxes := make([]string, 0, len(m.toasts))
	for _, t := range m.toasts {
		boxes = append(boxes, t.View(m.toastIcon(t.Type), toastMessageWidth(m.width)))
	}
	return boxes
}

// toastPlacement is where one toast's box goes in the frame.
type toastPlacement struct {
	box  string
	x, y int
}

// toastPlacements lays the stack out: the boxes stack upward from the last
// body row, flush right, the newest lowest. Each box keeps its own width, so
// a narrow toast does not pad out to the widest one.
func (m Model) toastPlacements() []toastPlacement {
	out := m.restingToastPlacements()
	for i := range out {
		t := m.toasts[i]
		out[i].x += m.toastOffset(t, "dx")
		out[i].y += m.toastOffset(t, "dy")
	}
	return out
}

// restingToastPlacements is toastPlacements with every toast in its slot.
func (m Model) restingToastPlacements() []toastPlacement {
	boxes := m.toastBoxes()
	stack := lipgloss.JoinVertical(lipgloss.Right, boxes...)
	left, top := bottomRightOrigin(stack, m.width, m.height)
	right := left + lipgloss.Width(stack)
	out := make([]toastPlacement, 0, len(boxes))
	for _, box := range boxes {
		out = append(out, toastPlacement{box: box, x: right - lipgloss.Width(box), y: top})
		top += lipgloss.Height(box)
	}
	return out
}

// toastLayers floats each toast as its own layer, so only its box covers the
// frame: the cells beside a narrower toast keep the base frame.
func (m Model) toastLayers() []*lipgloss.Layer {
	var layers []*lipgloss.Layer
	for _, p := range m.toastPlacements() {
		layers = append(layers, lipgloss.NewLayer(p.box).X(p.x).Y(p.y))
	}
	return layers
}

// toastAt is the index of the toast whose box covers the frame cell (x, y), or
// -1. It reads the same placements toastLayers draws.
func (m Model) toastAt(x, y int) int {
	for i, p := range m.toastPlacements() {
		bw, bh := lipgloss.Width(p.box), lipgloss.Height(p.box)
		if x >= p.x && x < p.x+bw && y >= p.y && y < p.y+bh {
			return i
		}
	}
	return -1
}
