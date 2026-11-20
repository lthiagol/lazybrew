package modal

import (
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
	ToastWarning
)

// Toast is one notification. The GUI keeps a stack of them and owns their
// dismissal tick; a toast only knows its message, its type and when it began.
type Toast struct {
	Message string
	Type    ToastType
	created time.Time
}

// ToastTickMsg is the tick that re-checks the toast stack for expiry.
type ToastTickMsg struct{}

func NewToast(message string, t ToastType) *Toast {
	return NewToastAt(message, t, time.Now())
}

// NewToastAt is NewToast with its start time given, so a test can age a toast.
func NewToastAt(message string, t ToastType, at time.Time) *Toast {
	return &Toast{Message: message, Type: t, created: at}
}

// Lifetime is how long the toast stays up. Errors stay twice as long, so the
// message can be read.
func (t *Toast) Lifetime() time.Duration {
	if t.Type == ToastError {
		return 6 * time.Second
	}
	return 3 * time.Second
}

// Expired reports whether the toast has been up longer than its lifetime at now.
func (t *Toast) Expired(now time.Time) bool {
	return now.Sub(t.created) > t.Lifetime()
}

// Color is the theme role the toast is drawn in: its border and its icon.
func (t *Toast) Color() color.Color {
	switch t.Type {
	case ToastError:
		return style.ErrorColor
	case ToastInfo:
		return style.InfoColor
	case ToastWarning:
		return style.WarningColor
	default:
		return style.SuccessColor
	}
}

// View draws the toast as a small rounded box in its role colour: the icon
// (none when icon is empty), then the message cut with an ellipsis to
// msgWidth cells.
func (t *Toast) View(icon string, msgWidth int) string {
	fg := t.Color()
	body := style.NormalItem.Render(ansi.Truncate(t.Message, msgWidth, "…"))
	if icon != "" {
		body = lipgloss.NewStyle().Foreground(fg).Bold(true).Render(icon) + " " + body
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(fg).
		Padding(0, 1).
		Render(body)
}

// ToastTick is the command that delivers a ToastTickMsg after one interval.
func ToastTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return ToastTickMsg{}
	})
}
