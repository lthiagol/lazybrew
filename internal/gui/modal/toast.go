package modal

import (
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
	ToastWarning
)

type Toast struct {
	Message   string
	Type      ToastType
	created   time.Time
	dismissed bool
}

type ToastTickMsg struct{}

func NewToast(message string, t ToastType) *Toast {
	return &Toast{Message: message, Type: t, created: time.Now()}
}

func (t *Toast) Init() tea.Cmd {
	return nil
}

func (t *Toast) Update(msg tea.Msg) (*Toast, tea.Cmd) {
	switch msg.(type) {
	case ToastTickMsg:
		if time.Since(t.created) > 3*time.Second {
			t.dismissed = true
		}
		if !t.dismissed {
			return t, tick()
		}
	}
	return t, nil
}

func (t *Toast) Dismissed() bool {
	return t.dismissed
}

func (t *Toast) View() string {
	if t.dismissed {
		return ""
	}
	var fg color.Color
	switch t.Type {
	case ToastSuccess:
		fg = style.SuccessColor
	case ToastError:
		fg = style.ErrorColor
	case ToastInfo:
		fg = style.AccentColor
	case ToastWarning:
		fg = style.WarningColor
	}

	return lipgloss.NewStyle().
		Foreground(fg).
		Bold(true).
		Render("  " + t.Message + "  ")
}

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return ToastTickMsg{}
	})
}
