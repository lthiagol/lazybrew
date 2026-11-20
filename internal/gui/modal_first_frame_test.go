package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestModalsShowContentOnFirstFrame locks AC-03: a modal opened on a model
// that has already been sized shows its real content on the first frame,
// never the Loading... placeholder it used to wait a resize to replace.
func TestModalsShowContentOnFirstFrame(t *testing.T) {
	cases := map[string]struct {
		open func(m *Model) *Model
		want string
	}{
		"uninstall confirm": {
			open: func(m *Model) *Model {
				return updateModel(m, DepCheckMsg{MutType: mutUninstall, Name: "bash", Label: "Uninstall", Message: "Uninstall bash? This cannot be undone."})
			},
			want: "Uninstall bash? This cannot be undone.",
		},
		"tap add input": {
			open: func(m *Model) *Model { return sendKey(sendKey(m, "5"), "a") },
			want: "Tap repository (user/repo):",
		},
		"brewfile menu": {
			open: func(m *Model) *Model { return sendKey(m, "B") },
			want: "Export to Brewfile (brew bundle dump)",
		},
		"trust menu": {
			open: func(m *Model) *Model {
				m = sendKey(m, "5")
				m.panels[PanelTaps].items = []string{"homebrew/foo"}
				return sendKey(m, "t")
			},
			want: "Trust entire tap: homebrew/foo",
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
			m = c.open(m)
			if m.activeModal == nil {
				t.Fatal("modal did not open")
			}
			// Check the modal's own view: a panel that is still loading
			// behind it legitimately shows its own spinner. The want string
			// is one only the modal draws, so the modal itself must show it.
			if modalView := m.activeModal.View().Content; strings.Contains(modalView, "Loading...") || !strings.Contains(modalView, c.want) {
				t.Fatalf("modal does not show %q on its first frame: %q", c.want, modalView)
			}
			if view := m.View().Content; !strings.Contains(view, c.want) {
				t.Fatalf("first frame does not show the modal content %q", c.want)
			}
		})
	}
}

// TestModalSurvivesResizeWhileOpen: a resize after the modal opened keeps
// the modal on screen and re-centres it in the new frame.
func TestModalSurvivesResizeWhileOpen(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DepCheckMsg{MutType: mutUninstall, Name: "bash", Label: "Uninstall", Message: "Uninstall bash?"})
	m = updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 30})

	view := m.View().Content
	if strings.Contains(view, "Loading...") {
		t.Fatal("modal showed Loading... after a resize")
	}
	if !strings.Contains(view, "Uninstall bash") {
		t.Fatal("modal disappeared after a resize")
	}
	assertFrameSize(t, view, 100, 30)
}
