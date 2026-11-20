package modal

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

type InputModal struct {
	prompt    string
	input     textinput.Model
	done      bool
	cancelled bool
}

func NewInputModal(prompt string) *InputModal {
	ti := textinput.New()
	ti.Placeholder = "type here..."
	ti.Focus()
	ti.SetWidth(40)

	return &InputModal{
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
	width := 50
	title := style.AccentText.Render(m.prompt)
	input := m.input.View()

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		input,
	)

	// Width counts the border in v2; add it back so the box is the v1 size.
	return tea.NewView(lipgloss.NewStyle().
		Width(width+2).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.AccentColor).
		Render(content))
}

func (m *InputModal) Done() bool      { return m.done }
func (m *InputModal) Cancelled() bool { return m.cancelled }
func (m *InputModal) Result() interface{} {
	return &InputResult{Value: m.input.Value(), Cancelled: m.cancelled}
}
