package modal

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
	ToastWarning
)

type Toast struct {
	Message string
	Type    ToastType
	created time.Time
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
	var color lipgloss.Color
	switch t.Type {
	case ToastSuccess:
		color = style.SuccessColor
	case ToastError:
		color = style.ErrorColor
	case ToastInfo:
		color = style.AccentColor
	case ToastWarning:
		color = style.WarningColor
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render("  " + t.Message + "  ")
}

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return ToastTickMsg{}
	})
}
