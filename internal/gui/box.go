package gui

import (
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

func renderBox(content string, width, height int, active bool) string {
	borderColor := style.SubtleColor
	if active {
		borderColor = style.AccentColor
	}
	// Lip Gloss v2 counts the border inside Width and Height; v1 added it on
	// top. The +2 keeps the box the size it was on v1.
	return lipgloss.NewStyle().
		Width(width + 2).
		Height(height + 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Render(content)
}
