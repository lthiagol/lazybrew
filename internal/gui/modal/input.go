package modal

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
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
	ti.SetWidth(40)

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
	keys := fitLine(keyLine(
		keyHint(style.AccentColor, "enter", "submit"),
		keyHint(style.AccentColor, "esc", "cancel"),
	), formWidth)
	prompt := style.NormalItem.Render(wrapProse(m.prompt, formWidth))
	return tea.NewView(dialog(style.AccentColor, m.action, formWidth, []string{prompt, "", m.input.View()}, keys))
}

func (m *InputModal) Done() bool      { return m.done }
func (m *InputModal) Cancelled() bool { return m.cancelled }
func (m *InputModal) Result() interface{} {
	return &InputResult{Value: m.input.Value(), Cancelled: m.cancelled}
}
