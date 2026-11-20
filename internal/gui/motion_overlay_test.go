package gui

import (
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

// TestBackdropIsDimAtOnce: a new modal's backdrop is the resting backdrop on
// its first frame. It is not faded in: see dimBackdrop.
func TestBackdropIsDimAtOnce(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for name, theme := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
		style.ApplyTheme(theme)
		off := openTapPrompt(overlayModel(t, 140, 40, "0"))
		on := openTapPrompt(overlayModel(t, 140, 40, "1"))
		first := strings.Split(on.View().Content, "\n")
		rest := strings.Split(off.View().Content, "\n")
		if first[0] != rest[0] {
			t.Fatalf("%s theme: the backdrop on a new modal's first frame is not the resting backdrop", name)
		}
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
