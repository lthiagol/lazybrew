package modal

import (
	"context"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type ProgressModal struct {
	title     string
	lines     []string
	cancel    context.CancelFunc
	done      bool
	cancelled bool
	err       error
	viewport  viewport.Model
	ready     bool
}

func NewProgressModal(title string, cancel context.CancelFunc) *ProgressModal {
	return &ProgressModal{
		title:  title,
		cancel: cancel,
	}
}

func (m *ProgressModal) Init() tea.Cmd {
	return nil
}

func (m *ProgressModal) AppendLine(line string) {
	m.lines = append(m.lines, line)
	if m.ready {
		m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, m.lines...))
		m.viewport.GotoBottom()
	}
}

func (m *ProgressModal) SetDone(err error) {
	m.done = true
	m.err = err
}

func (m *ProgressModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-8, 12)
			m.ready = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if !m.done {
				m.cancelled = true
				if m.cancel != nil {
					m.cancel()
				}
			} else {
				m.done = true
			}
		case "enter":
			if m.done {
				m.done = true
			}
		}
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *ProgressModal) View() string {
	width := 60
	title := style.AccentText.Render(m.title)

	output := ""
	if m.ready {
		output = m.viewport.View()
	} else {
		output = lipgloss.JoinVertical(lipgloss.Top, m.lines...)
	}

	status := ""
	if !m.done {
		status = style.SubtleText.Render("Running...  (Esc to cancel)")
	} else if m.cancelled {
		status = lipgloss.NewStyle().Foreground(style.WarningColor).Render("Cancelled")
	} else if m.err != nil {
		status = style.ErrorBadge.Render("Error: " + m.err.Error())
	} else {
		status = style.InstalledBadge.Render("Completed")
	}

	content := lipgloss.JoinVertical(lipgloss.Top,
		title,
		"",
		output,
		"",
		status,
	)

	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.AccentColor).
		Render(content)
}

func (m *ProgressModal) Done() bool      { return m.done && !m.cancelled }
func (m *ProgressModal) Cancelled() bool { return m.cancelled }
func (m *ProgressModal) Result() interface{} {
	return m.err
}
