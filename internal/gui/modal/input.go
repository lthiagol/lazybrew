package modal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
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
	ti.Width = 40

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
	case tea.KeyMsg:
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

func (m *InputModal) View() string {
	width := 50
	title := style.AccentText.Render(m.prompt)
	input := m.input.View()

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		input,
	)

	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.AccentColor).
		Render(content)
}

func (m *InputModal) Done() bool        { return m.done }
func (m *InputModal) Cancelled() bool   { return m.cancelled }
func (m *InputModal) Result() interface{} {
	return &InputResult{Value: m.input.Value(), Cancelled: m.cancelled}
}
