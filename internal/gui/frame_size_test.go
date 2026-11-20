package gui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// frameState is one screen the frame size rule must hold for. To guard a
// new screen, add a row here: setup receives a model that has already been
// sized, and returns it in the state to measure.
type frameState struct {
	name  string
	setup func(m *Model) *Model
}

var frameStates = []frameState{
	{"status", func(m *Model) *Model { return m }},
	{"formulae", func(m *Model) *Model { return sendKey(m, "2") }},
	{"formulae info tab", func(m *Model) *Model { return sendKey(sendKey(m, "2"), "]") }},
	{"casks", func(m *Model) *Model { return sendKey(m, "3") }},
	{"outdated", func(m *Model) *Model { return sendKey(m, "4") }},
	{"taps", func(m *Model) *Model { return sendKey(m, "5") }},
	{"services", func(m *Model) *Model { return sendKey(m, "6") }},
	{"search", func(m *Model) *Model { return sendKey(m, "7") }},
	{"panel loading", func(m *Model) *Model {
		m.panels[PanelFormulae].loading = true
		return sendKey(m, "2")
	}},
	{"log collapsed", func(m *Model) *Model { return sendKey(m, "C") }},
	{"help", func(m *Model) *Model { return sendKey(m, "?") }},
	{"confirm modal", func(m *Model) *Model {
		m.activeModal = modal.NewConfirmModal(modal.ActionDanger, "Uninstall bash", "Remove bash? Other packages may depend on it.")
		return m
	}},
	{"input modal", func(m *Model) *Model {
		m.activeModal = modal.NewInputModal("Tap name")
		return m
	}},
	{"menu modal", func(m *Model) *Model {
		m.activeModal = modal.NewMenuModal("Trust", []string{"Trust tap", "Untrust tap", "Cancel"})
		return m
	}},
	{"toast", func(m *Model) *Model {
		m.toast = modal.NewToast("Install completed", modal.ToastSuccess)
		return m
	}},
	{"operation running", func(m *Model) *Model {
		m.opState = &Operation{Title: "install bash", Status: opRunning}
		return m
	}},
	{"operation completed", func(m *Model) *Model {
		m.opState = &Operation{Title: "install bash", Status: opSuccess}
		return m
	}},
	{"operation failed", func(m *Model) *Model {
		m.opState = &Operation{Title: "install bash", Status: opError, Err: errors.New("exit 1")}
		return m
	}},
}

// assertFrameSize fails unless view is exactly h rows and every row is
// exactly w cells wide (lipgloss.Width counts cells, not bytes).
func assertFrameSize(t *testing.T, view string, w, h int) {
	t.Helper()
	rows := strings.Split(view, "\n")
	if len(rows) != h {
		t.Fatalf("frame has %d rows, want %d", len(rows), h)
	}
	for i, row := range rows {
		if got := lipgloss.Width(row); got != w {
			t.Fatalf("row %d is %d cells wide, want %d: %q", i, got, w, row)
		}
	}
}

func TestEveryFrameIsTerminalSize(t *testing.T) {
	sizes := [][2]int{{140, 40}, {80, 24}, {200, 60}, {70, 20}}
	for _, sz := range sizes {
		for _, st := range frameStates {
			t.Run(fmt.Sprintf("%s/%dx%d", st.name, sz[0], sz[1]), func(t *testing.T) {
				m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
				m = st.setup(m)
				assertFrameSize(t, m.View().Content, sz[0], sz[1])
			})
		}
	}
}
