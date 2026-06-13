package gui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
)

func renderBox(content string, width, height int, active bool) string {
	borderColor := style.SubtleColor
	if active {
		borderColor = style.AccentColor
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Render(content)
}
