package modal

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
)

type InputModal struct {
	action    string
	prompt    string
	input     textinput.Model
	done      bool
	cancelled bool
}

// NewInputModal opens a text prompt for action (its name, the chip) with the
// question prompt, which wraps inside the box.
func NewInputModal(action, prompt string) *InputModal {
	ti := textinput.New()
	ti.Placeholder = "type here..."
	ti.Focus()
	// The field is a fixed 30 cells of text; with its prompt ("> ") and cursor it
	// draws 33, narrower than the dialog's 42, so the dialog centres it.
	ti.SetWidth(30)

	return &InputModal{
		action: action,
		prompt: prompt,
		input:  ti,
	}
}

func (m *InputModal) Init() tea.Cmd {
	return textinput.Blink
}

func (m *InputModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			m.cancelled = true
			return m, nil
		case "enter":
			m.done = true
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *InputModal) View() tea.View {
	keys := textutil.Clip(keyLine(hintViews(style.AccentColor, m.Hints())...), innerWidth(formWidth))
	prompt := proseLines(m.prompt, style.NormalItem, formWidth)
	return tea.NewView(dialog(style.AccentColor, m.action, formWidth, append(prompt, "", m.input.View()), keys))
}

// Hints are the keys the input answers to: enter, and esc.
func (m *InputModal) Hints() []Hint {
	return []Hint{{Key: "enter", Desc: "submit"}, hintCancel}
}

func (m *InputModal) Done() bool      { return m.done }
func (m *InputModal) Cancelled() bool { return m.cancelled }
func (m *InputModal) Result() interface{} {
	return &InputResult{Value: m.input.Value(), Cancelled: m.cancelled}
}
