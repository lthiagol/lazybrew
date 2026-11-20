package gui

import (
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// overlayModel is a model sized w x h, with motion on or off, and nothing open.
func overlayModel(t *testing.T, w, h int, motion string) *Model {
	t.Helper()
	t.Setenv(config.AnimationsEnv, motion)
	return updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
}

// openTapPrompt opens the add-tap input modal, the way a key would, and runs
// the update that starts its entrance.
func openTapPrompt(m *Model) *Model {
	m.activeModal = modal.NewInputModal("Add tap", "Tap repository (user/repo):")
	return updateModel(m, probeMsg{})
}

// rowIndex is the first frame row that holds text, or -1.
func rowIndex(m *Model, text string) int {
	return rowOf(plainRows(m.View().Content), text)
}

// colIndex is the first column of text on its frame row, or -1 when the text
// is not on screen (a toast still off the right edge is not on screen).
func colIndex(m *Model, text string) int {
	for _, row := range plainRows(m.View().Content) {
		if i := strings.Index(string(row), text); i >= 0 {
			return len([]rune(string(row)[:i]))
		}
	}
	return -1
}

func TestModalSlidesDownIntoItsPlace(t *testing.T) {
	off := openTapPrompt(overlayModel(t, 140, 40, "0"))
	on := openTapPrompt(overlayModel(t, 140, 40, "1"))

	const prompt = "Tap repository"
	rest := rowIndex(off, prompt)
	if rest < 0 {
		t.Fatal("the tap prompt is not on screen")
	}
	if got := rowIndex(on, prompt); got != rest-modalDropRows {
		t.Fatalf("new modal's prompt is on row %d; want %d rows above its place (%d)", got, modalDropRows, rest-modalDropRows)
	}
	mid := frames(on, 3)
	if got := rowIndex(mid, prompt); got <= rest-modalDropRows || got >= rest {
		t.Fatalf("mid-entrance prompt on row %d; want between %d and %d", got, rest-modalDropRows, rest)
	}
	settled := settle(t, mid)
	if got := rowIndex(settled, prompt); got != rest {
		t.Fatalf("settled prompt on row %d; want its resting row %d", got, rest)
	}
	if settled.View().Content != off.View().Content {
		t.Fatal("the settled modal frame differs from the frame with animations off")
	}
}

func TestBackdropFadesIntoTheMutedRole(t *testing.T) {
	off := openTapPrompt(overlayModel(t, 140, 40, "0"))
	on := openTapPrompt(overlayModel(t, 140, 40, "1"))
	first := strings.Split(on.View().Content, "\n")[0]
	settled := strings.Split(settle(t, frames(on, 2)).View().Content, "\n")[0]
	if first == settled {
		t.Fatal("the backdrop is already at its resting colour on the first frame")
	}
	if want := strings.Split(off.View().Content, "\n")[0]; settled != want {
		t.Fatal("the settled backdrop differs from the resting backdrop")
	}
}

func TestClosingAModalIsInstant(t *testing.T) {
	on := settle(t, openTapPrompt(overlayModel(t, 140, 40, "1")))
	on.activeModal = nil
	if _, cmd := on.Update(probeMsg{}); cmd != nil {
		t.Fatal("closing a modal scheduled a frame")
	}
	closed := updateModel(overlayModel(t, 140, 40, "0"), probeMsg{})
	if got := updateModel(on, probeMsg{}).View().Content; got != closed.View().Content {
		t.Fatal("the frame after close is not the frame with no modal")
	}
}

func TestModalTakesKeysFromItsFirstFrame(t *testing.T) {
	m := openTapPrompt(overlayModel(t, 140, 40, "1"))
	m = sendSpecial(m, tea.KeyEscape)
	if m.activeModal != nil {
		t.Fatal("esc on the first frame of a modal did not reach it")
	}
}

func TestToastSlidesInFromTheRight(t *testing.T) {
	off := overlayModel(t, 140, 40, "0")
	off.setToast("first done", modal.ToastSuccess)
	on := overlayModel(t, 140, 40, "1")
	on.setToast("first done", modal.ToastSuccess)
	on = updateModel(on, probeMsg{})

	const text = "first done"
	restCol := colIndex(off, text)
	if restCol < 0 {
		t.Fatal("the toast is not on screen")
	}
	if first := colIndex(on, text); first >= 0 && first <= restCol {
		t.Fatalf("new toast starts at column %d; want it right of its place (%d)", first, restCol)
	}
	mid := frames(on, 6)
	if col := colIndex(mid, text); col >= 0 && (col <= restCol || col >= mid.width) {
		t.Fatalf("mid-entrance toast at column %d; want between %d and the right edge", col, restCol)
	}
	settled := settle(t, mid)
	if got := colIndex(settled, text); got != restCol {
		t.Fatalf("settled toast at column %d; want %d", got, restCol)
	}
	if settled.View().Content != off.View().Content {
		t.Fatal("the settled toast frame differs from the frame with animations off")
	}
}

func TestExistingToastsSpringWhenOneArrives(t *testing.T) {
	off := overlayModel(t, 140, 40, "0")
	off.setToast("first done", modal.ToastSuccess)
	off.setToast("second note", modal.ToastInfo)
	restRow := rowIndex(off, "first done")

	on := overlayModel(t, 140, 40, "1")
	on.setToast("first done", modal.ToastSuccess)
	on = settle(t, updateModel(on, probeMsg{}))
	before := rowIndex(on, "first done")
	on.setToast("second note", modal.ToastInfo)
	on = updateModel(on, probeMsg{})
	if got := rowIndex(on, "first done"); got != before {
		t.Fatalf("the older toast moved on the first frame: row %d, was %d", got, before)
	}
	mid := frames(on, 4)
	if got := rowIndex(mid, "first done"); got >= before || got <= restRow {
		t.Fatalf("mid-move older toast on row %d; want between %d and %d", got, restRow, before)
	}
	settled := settle(t, mid)
	if got := rowIndex(settled, "first done"); got != restRow {
		t.Fatalf("settled older toast on row %d; want %d", got, restRow)
	}
	if settled.View().Content != off.View().Content {
		t.Fatal("the settled two-toast frame differs from the frame with animations off")
	}
}

func TestDroppedToastsAreForgotten(t *testing.T) {
	m := overlayModel(t, 140, 40, "1")
	m.setToast("first done", modal.ToastSuccess)
	m = settle(t, updateModel(m, probeMsg{}))
	m.toasts = nil
	m = updateModel(m, probeMsg{})
	for _, name := range m.motion.Names() {
		if strings.HasPrefix(name, "toast:") {
			t.Fatalf("a dropped toast left %q behind", name)
		}
	}
}

func TestOverlayFramesAreTerminalSize(t *testing.T) {
	for _, sz := range [][2]int{{140, 40}, {80, 24}} {
		m := overlayModel(t, sz[0], sz[1], "1")
		m.setToast("first done", modal.ToastSuccess)
		m = openTapPrompt(m)
		for i := 0; i < 8; i++ {
			m = frames(m, 1)
			v := m.View().Content
			if w, h := lipgloss.Width(v), lipgloss.Height(v); w != sz[0] || h != sz[1] {
				t.Fatalf("%dx%d frame %d is %dx%d", sz[0], sz[1], i, w, h)
			}
		}
	}
}

// luma is the relative brightness of a colour, from its red, green and blue.
func luma(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
}

// TestBackdropFadeNeverBrightens holds every step of the fade to the one before
// it: the backdrop only ever dims, from the subtle role to the muted role.
func TestBackdropFadeNeverBrightens(t *testing.T) {
	if !(luma(style.MutedColor) < luma(style.SubtleColor) && luma(style.SubtleColor) < luma(style.TextColor)) {
		t.Fatal("the theme's roles are not ordered text > subtle > muted: the fade's premise fails")
	}
	prev := luma(backdropColor(0))
	if prev > luma(style.SubtleColor) {
		t.Fatalf("the fade starts at %v, brighter than the subtle role", prev)
	}
	for f := 0.0; f <= 1.0; f += 0.01 {
		l := luma(backdropColor(f))
		if l > prev {
			t.Fatalf("fade %.2f is brighter than the step before it: %v > %v", f, l, prev)
		}
		prev = l
	}
	if got, want := luma(backdropColor(1)), luma(style.MutedColor); got != want {
		t.Fatalf("the resting backdrop is not the muted role: %v, want %v", got, want)
	}
}

// TestModalFadeSpringIsMonotone runs a modal's fade frame by frame: it rises
// without overshoot, in steps of at most a quarter, and settles exactly on 1.
func TestModalFadeSpringIsMonotone(t *testing.T) {
	m := openTapPrompt(overlayModel(t, 140, 40, "1"))
	prev := m.backdropFadeAt()
	if prev != 0 {
		t.Fatalf("a new modal's backdrop starts at fade %v; want 0", prev)
	}
	for i := 0; i < 60 && !m.motion.Settled(); i++ {
		m = frames(m, 1)
		f := m.backdropFadeAt()
		if f < prev || f > 1 || f-prev > 0.25 {
			t.Fatalf("frame %d: fade went from %v to %v", i, prev, f)
		}
		prev = f
	}
	if got := m.backdropFadeAt(); got != 1 {
		t.Fatalf("the fade settled at %v; want exactly 1", got)
	}
}
