package modal

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type ConfirmModal struct {
	title     string
	message   string
	selected  int
	done      bool
	cancelled bool
	result    bool
	viewport  viewport.Model
	ready     bool
}

func NewConfirmModal(title, message string) *ConfirmModal {
	return &ConfirmModal{
		title:    title,
		message:  message,
		selected: 1,
	}
}

func (m *ConfirmModal) Init() tea.Cmd {
	return nil
}

func (m *ConfirmModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-8, 8)
			m.ready = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			m.cancelled = true
		case "h", "left":
			m.selected = 0
		case "l", "right":
			m.selected = 1
		case "y":
			m.result = true
			m.done = true
		case "enter":
			m.result = m.selected == 0
			m.done = true
		case "n":
			m.done = true
		}
	}
	return m, nil
}

func (m *ConfirmModal) View() string {
	if !m.ready {
		return "Loading..."
	}

	width := 50
	title := style.AccentText.Render(m.title)
	message := style.NormalItem.Render(m.message)

	yesStyle := style.NormalItem
	noStyle := style.NormalItem
	if m.selected == 0 {
		yesStyle = style.SelectedItem
	} else {
		noStyle = style.SelectedItem
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		yesStyle.Render(" [Yes] "),
		style.SubtleText.Render("  "),
		noStyle.Render(" [No] "),
	)

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		message,
		"",
		buttons,
	)

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.AccentColor).
		Render(content)
}

func (m *ConfirmModal) Done() bool      { return m.done }
func (m *ConfirmModal) Cancelled() bool { return m.cancelled }
func (m *ConfirmModal) Result() interface{} {
	return &ConfirmResult{Confirmed: m.result}
}
