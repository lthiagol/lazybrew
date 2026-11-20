package modal

import (
	"image/color"
	"strings"
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
// (none when icon is empty), then the message wrapped to msgWidth cells. The
// message is not cut while it fits in toastMaxRows rows, so a message that
// names a command keeps the command.
func (t *Toast) View(icon string, msgWidth int) string {
	fg := t.Color()
	rows := toastRows(t.Message, msgWidth)
	lead, indent := "", ""
	if icon != "" {
		lead = lipgloss.NewStyle().Foreground(fg).Bold(true).Render(icon) + " "
		indent = strings.Repeat(" ", lipgloss.Width(icon)+1)
	}
	lines := make([]string, len(rows))
	for i, row := range rows {
		prefix := indent
		if i == 0 {
			prefix = lead
		}
		lines[i] = prefix + style.NormalItem.Render(row)
	}
	body := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(fg).
		Padding(0, 1).
		Render(body)
}

// toastMaxRows is how many rows a toast message may take. A message still
// longer is cut with an ellipsis at the end of its last row.
const toastMaxRows = 3

// toastRows wraps msg to width cells, at spaces where it can, over at most
// toastMaxRows rows.
func toastRows(msg string, width int) []string {
	if width < 1 {
		return []string{ansi.Truncate(msg, width, "…")}
	}
	all := strings.Split(ansi.Wrap(msg, width, ""), "\n")
	if len(all) <= toastMaxRows {
		return all
	}
	last := ansi.Truncate(strings.Join(all[toastMaxRows-1:], " "), width, "…")
	return append(all[:toastMaxRows-1:toastMaxRows-1], last)
}

// ToastTick is the command that delivers a ToastTickMsg after one interval.
func ToastTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return ToastTickMsg{}
	})
}
