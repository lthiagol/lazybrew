package modal

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// Widths of the modal frames. A confirm dialog is min(frame-8, 72) wide; the
// app draws no frame narrower than 80 columns, so it is 72 in every frame.
const (
	confirmWidth = 72
	formWidth    = 50
)

// dialog frames a modal the change-gate way: a thick border in accent, the
// chip (its label) on top, then the body, centred. Width counts the border and
// padding (v2), so the box is exactly width cells. Body text longer than the
// box wraps inside it; a wrapped line never runs past the border.
func dialog(accent color.Color, label string, width int, body ...string) string {
	inner := width - 2 - 6 // thick border, then padding of 3 on each side
	chip := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.BgColor).
		Background(accent).
		Render(" " + label + " ")
	parts := []string{chip, ""}
	for _, b := range body {
		parts = append(parts, lipgloss.Wrap(b, inner, ""))
	}
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Padding(1, 3).
		Border(lipgloss.ThickBorder()).
		BorderForeground(accent).
		Render(content)
}

// keyHint is one key and what it does: the key in accent, the words subtle.
func keyHint(accent color.Color, key, desc string) string {
	return lipgloss.NewStyle().Foreground(accent).Bold(true).Render(key) +
		style.SubtleText.Render(" "+desc)
}

// keyLine joins key hints with the gap the footer uses.
func keyLine(hints ...string) string {
	return strings.Join(hints, "   ")
}
