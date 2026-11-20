package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// notAModal is an open modal whose Update hands back a model that is not a
// modal.Modal. The key and paste routes must not read that as a finished modal.
type notAModal struct{}

func (notAModal) Init() tea.Cmd                       { return nil }
func (notAModal) Update(tea.Msg) (tea.Model, tea.Cmd) { return plainModel{}, nil }
func (notAModal) View() tea.View                      { return tea.NewView("") }
func (notAModal) Done() bool                          { return false }
func (notAModal) Cancelled() bool                     { return false }
func (notAModal) Result() interface{}                 { return nil }

// plainModel is a tea.Model that is not a modal.Modal.
type plainModel struct{}

func (plainModel) Init() tea.Cmd                       { return nil }
func (plainModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return plainModel{}, nil }
func (plainModel) View() tea.View                      { return tea.NewView("") }

// A non-modal result must clear the modal and must not panic: the key route
// used to call Done on the nil it had just stored.
func TestModalNonModalResultClearsTheModal(t *testing.T) {
	routes := []struct {
		name string
		msg  tea.Msg
	}{
		{"key", tea.KeyPressMsg{Code: 'x', Text: "x"}},
		{"paste", tea.PasteMsg{Content: "lazy"}},
	}
	for _, r := range routes {
		t.Run(r.name, func(t *testing.T) {
			m := newTestModel()
			m = updateModel(m, tea.WindowSizeMsg{Width: 120, Height: 40})
			m.activeModal = notAModal{}

			var got *Model
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Fatalf("%s route panicked: %v", r.name, p)
					}
				}()
				got = updateModel(m, r.msg)
			}()
			if got.activeModal != nil {
				t.Errorf("activeModal = %T after a non-modal result, want nil", got.activeModal)
			}
		})
	}
}
