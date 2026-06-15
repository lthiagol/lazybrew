package modal

import (
	"strconv"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type MenuModal struct {
	title     string
	items     []string
	selected  int
	done      bool
	cancelled bool
	viewport  viewport.Model
	ready     bool
}

func NewMenuModal(title string, items []string) *MenuModal {
	return &MenuModal{
		title: title,
		items: items,
	}
}

func (m *MenuModal) Init() tea.Cmd {
	return nil
}

func (m *MenuModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-8, min(len(m.items)+4, 12))
			m.ready = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			m.cancelled = true
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "enter":
			m.done = true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.Runes[0] - '0')
			if idx < len(m.items) {
				m.selected = idx
				m.done = true
			}
		}
	}
	return m, nil
}

func (m *MenuModal) View() string {
	width := 50
	title := style.AccentText.Render(m.title)

	var lines []string
	for i, item := range m.items {
		prefix := "  "
		itemStyle := style.NormalItem
		if i == m.selected {
			prefix = "▸ "
			itemStyle = style.SelectedItem
		}
		shortcut := style.SubtleText.Render(strconv.Itoa(i) + ": ")
		lines = append(lines, itemStyle.Render(shortcut+prefix+item))
	}

	list := lipgloss.JoinVertical(lipgloss.Top, lines...)
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		list,
	)

	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.AccentColor).
		Render(content)
}

func (m *MenuModal) Done() bool      { return m.done }
func (m *MenuModal) Cancelled() bool { return m.cancelled }
func (m *MenuModal) Result() interface{} {
	return &MenuResult{SelectedIndex: m.selected, Cancelled: m.cancelled}
}
