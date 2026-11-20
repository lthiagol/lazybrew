package modal

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
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
	action    string
	message   string
	detail    string
	selected  int
	done      bool
	cancelled bool
	result    bool
}

// NewConfirmModal opens a confirm for action (its name, shown in the chip, such
// as "Uninstall") with the question message, which names the package or tap.
func NewConfirmModal(kind ActionKind, action, message string) *ConfirmModal {
	return &ConfirmModal{
		kind:     kind,
		action:   action,
		message:  message,
		selected: 1,
	}
}

// Kind reports the kind of action the dialog confirms.
func (m *ConfirmModal) Kind() ActionKind { return m.kind }

// Action is the action's name, as the chip shows it.
func (m *ConfirmModal) Action() string { return m.action }

// Detail is the subtle consequence line under the question; empty if none.
func (m *ConfirmModal) Detail() string { return m.detail }

// WithDetail sets the subtle line under the question: what the action will do.
func (m *ConfirmModal) WithDetail(detail string) *ConfirmModal {
	m.detail = detail
	return m
}

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

// View draws the dialog: the first line of the message is the bold question,
// the detail follows in subtle text, and the rest of the message is body text.
// Prose wraps inside the box; the key line is cut to fit it.
func (m *ConfirmModal) View() tea.View {
	accent := m.kind.accent()
	question, body, hasBody := strings.Cut(m.message, "\n")

	lines := proseLines(question, lipgloss.NewStyle().Bold(true), confirmWidth)
	if m.detail != "" {
		lines = append(lines, proseLines(m.detail, style.SubtleText, confirmWidth)...)
	}
	if hasBody {
		// The body is a list or a note, so it is one left-aligned block, not
		// centred line by line: its indents must line up.
		lines = append(lines, "")
		lines = append(lines, leftBlock(proseLines(body, style.NormalItem, confirmWidth))...)
	}
	return tea.NewView(dialog(accent, m.action, confirmWidth, lines, textutil.Clip(m.choices(accent), innerWidth(confirmWidth))))
}

// choices is the key line: [y]es and [n]o, the selected one reversed, then esc.
func (m *ConfirmModal) choices(accent color.Color) string {
	return keyLine(
		choice(accent, "[y]", "es", m.selected == 0),
		choice(accent, "[n]", "o", m.selected == 1),
		keyHint(accent, "esc", "cancel"),
	)
}

func choice(accent color.Color, key, rest string, selected bool) string {
	word := style.NormalItem
	if selected {
		word = style.SelectedItem.Reverse(true)
	}
	return lipgloss.NewStyle().Foreground(accent).Bold(true).Render(key) + word.Render(rest)
}

func (m *ConfirmModal) Done() bool      { return m.done }
func (m *ConfirmModal) Cancelled() bool { return m.cancelled }
func (m *ConfirmModal) Result() interface{} {
	return &ConfirmResult{Confirmed: m.result}
}
