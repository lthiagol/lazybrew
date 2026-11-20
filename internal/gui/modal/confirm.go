package modal

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// ActionKind is what a confirm dialog is about. It picks the theme role the
// dialog is framed in, so a destructive action never looks like a routine one.
type ActionKind int

const (
	ActionNeutral ActionKind = iota
	ActionDanger
	ActionUpgrade
	ActionInstall
)

// accent is the theme role a kind is drawn in.
func (k ActionKind) accent() color.Color {
	switch k {
	case ActionDanger:
		return style.ErrorColor
	case ActionUpgrade:
		return style.WarningColor
	case ActionInstall:
		return style.InfoColor
	default:
		return style.AccentColor
	}
}

type ConfirmModal struct {
	kind      ActionKind
	title     string
	message   string
	selected  int
	done      bool
	cancelled bool
	result    bool
}

func NewConfirmModal(kind ActionKind, title, message string) *ConfirmModal {
	return &ConfirmModal{
		kind:     kind,
		title:    title,
		message:  message,
		selected: 1,
	}
}

// Kind reports the action the dialog confirms.
func (m *ConfirmModal) Kind() ActionKind { return m.kind }

func (m *ConfirmModal) Init() tea.Cmd {
	return nil
}

func (m *ConfirmModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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

func (m *ConfirmModal) View() tea.View {
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

	// Width counts the border in v2; add it back so the box is the v1 size.
	return tea.NewView(lipgloss.NewStyle().
		Width(width + 2).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.kind.accent()).
		Render(content))
}

func (m *ConfirmModal) Done() bool      { return m.done }
func (m *ConfirmModal) Cancelled() bool { return m.cancelled }
func (m *ConfirmModal) Result() interface{} {
	return &ConfirmResult{Confirmed: m.result}
}
