package gui

import (
	"fmt"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// copyWidth is the width the main content is rendered at for a copy. It is
// wide enough that no line is cut to the view: a copy holds the text.
const copyWidth = 400

// copyScreen copies the text on screen to the terminal clipboard (y) and says
// how many lines it copied. A screen with no text says so instead.
func (m Model) copyScreen() (tea.Model, tea.Cmd) {
	text := m.copyText()
	if text == "" {
		cmd := m.setToast("Nothing to copy", modal.ToastInfo)
		return m, cmd
	}
	toast := m.setToast(fmt.Sprintf("Copied %d lines", strings.Count(text, "\n")+1), modal.ToastInfo)
	return m, tea.Batch(m.setClipboard(text), toast)
}

// copyText is the text y copies, ANSI-free and untruncated: the open
// Operation's title, its status row, the command a retry runs, and every output
// line, or else the main panel's content for the selected item. It is "" when
// there is nothing to copy.
func (m Model) copyText() string {
	if op := m.opState; op != nil {
		// The status row is not cut for a copy, and the y hint on it is the
		// key just pressed, so it is not copied.
		status := ansi.Strip(m.operationStatus(op, math.MaxInt))
		lines := []string{op.Title, strings.TrimSuffix(status, operationCopyHint)}
		if needs, ok := m.terminalRetry(); ok {
			lines = append(lines, "brew "+needs.Command)
		}
		for _, line := range op.Lines {
			lines = append(lines, ansi.Strip(line))
		}
		return plainText(strings.Join(lines, "\n"))
	}
	m.copying = true
	return plainText(ansi.Strip(m.renderContent(copyWidth, m.height)))
}

// plainText ends each line of s at its last visible cell and drops the blank
// lines at the end, so a copy holds the text and no padding.
func plainText(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
