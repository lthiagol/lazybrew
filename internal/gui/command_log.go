package gui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type CommandStatus int

const (
	CommandRunning CommandStatus = iota
	CommandSuccess
	CommandError
)

type LogEntry struct {
	Command   string
	Status    CommandStatus
	Timestamp time.Time
}

type CommandLog struct {
	entries []LogEntry
	max     int
	next    int
	count   int
}

func NewCommandLog(max int) *CommandLog {
	return &CommandLog{
		entries: make([]LogEntry, max),
		max:     max,
	}
}

func (cl *CommandLog) Append(cmd string) {
	cl.entries[cl.next] = LogEntry{
		Command:   cmd,
		Status:    CommandRunning,
		Timestamp: time.Now(),
	}
	cl.next = (cl.next + 1) % cl.max
	if cl.count < cl.max {
		cl.count++
	}
}

func (cl *CommandLog) SetStatus(cmd string, status CommandStatus) {
	for i := 0; i < cl.count; i++ {
		idx := (cl.next - 1 - i + cl.max) % cl.max
		if cl.entries[idx].Command == cmd && cl.entries[idx].Status == CommandRunning {
			cl.entries[idx].Status = status
			return
		}
	}
}

func (cl *CommandLog) Entries() []LogEntry {
	n := cl.count
	if n == 0 {
		return nil
	}
	result := make([]LogEntry, n)
	for i := 0; i < n; i++ {
		idx := (cl.next - n + i + cl.max) % cl.max
		result[i] = cl.entries[idx]
	}
	return result
}

func (cl *CommandLog) View(width, height int) string {
	entries := cl.Entries()
	if len(entries) == 0 {
		return lipgloss.NewStyle().Width(width).Height(height).Render(
			style.SubtleText.Render("No commands executed yet"),
		)
	}

	var lines []string
	start := 0
	if len(entries) > height {
		start = len(entries) - height
	}
	for _, e := range entries[start:] {
		prefix := "⟳"
		cmdStyle := style.SubtleText
		switch e.Status {
		case CommandRunning:
			prefix = "⟳"
			cmdStyle = style.AccentText
		case CommandSuccess:
			prefix = "✓"
			cmdStyle = style.NormalItem
		case CommandError:
			prefix = "✗"
			cmdStyle = style.ErrorBadge
		}
		line := prefix + " brew " + e.Command
		lines = append(lines, cmdStyle.Render(line))
	}

	rendered := lipgloss.JoinVertical(lipgloss.Top, lines...)
	return lipgloss.NewStyle().Width(width).Height(height).Render(rendered)
}

func brewCommandString(args []string) string {
	return strings.Join(args, " ")
}
