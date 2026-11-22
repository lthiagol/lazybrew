package gui

import (
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

type CommandStatus int

const (
	CommandRunning CommandStatus = iota
	CommandSuccess
	CommandError
	// CommandWarn is a command whose exit 1 printed a report (see
	// brew.ExitOneReport), not a failure.
	CommandWarn
)

type LogEntry struct {
	Command   string
	Status    CommandStatus
	Timestamp time.Time

	// id is the append-order sequence number. Status updates are keyed by
	// it, never by the command text: two concurrent identical commands
	// (e.g. two "list" reads) would otherwise mark each other's entry
	// (r1 finding 26).
	id int
}

// CommandLog is written from brew runner goroutines (LoggingRunner
// OnStart/OnExec callbacks) and read from the render thread, so every
// accessor takes the mutex (r1 finding 3).
type CommandLog struct {
	mu      sync.Mutex
	entries []LogEntry
	max     int
	next    int
	count   int
	seq     int
	icons   style.Icons
}

func NewCommandLog(max int) *CommandLog {
	return &CommandLog{
		entries: make([]LogEntry, max),
		max:     max,
		icons:   style.UnicodeIcons,
	}
}

// SetIcons sets the glyph set the status marks are drawn with.
func (cl *CommandLog) SetIcons(icons style.Icons) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.icons = icons
}

func (cl *CommandLog) currentIcons() style.Icons {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.icons
}

// Append records a new running entry and returns its unique id. Callers
// pass that id to SetStatus so the right entry is updated even when
// identical commands run concurrently.
func (cl *CommandLog) Append(cmd string) int {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	id := cl.seq
	cl.seq++
	cl.entries[cl.next] = LogEntry{
		Command:   cmd,
		Status:    CommandRunning,
		Timestamp: time.Now(),
		id:        id,
	}
	cl.next = (cl.next + 1) % cl.max
	if cl.count < cl.max {
		cl.count++
	}
	return id
}

func (cl *CommandLog) SetStatus(id int, status CommandStatus) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	for i := 0; i < cl.count; i++ {
		idx := (cl.next - 1 - i + cl.max) % cl.max
		if cl.entries[idx].id == id && cl.entries[idx].Status == CommandRunning {
			cl.entries[idx].Status = status
			return
		}
	}
}

// Entries is every entry the log holds, oldest first.
func (cl *CommandLog) Entries() []LogEntry {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.lastLocked(cl.count)
}

// Len is how many entries the log holds.
func (cl *CommandLog) Len() int {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.count
}

// Last is the newest n entries, oldest first: fewer when the log holds fewer.
// It copies those n, not the whole ring.
func (cl *CommandLog) Last(n int) []LogEntry {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.lastLocked(n)
}

// lastLocked is the newest n entries, oldest first. The caller holds the lock.
func (cl *CommandLog) lastLocked(n int) []LogEntry {
	n = min(n, cl.count)
	if n <= 0 {
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
	if cl.Len() == 0 {
		return lipgloss.NewStyle().Width(width).Height(height).Render(
			style.SubtleText.Render("No commands executed yet"),
		)
	}

	var lines []string
	icons := cl.currentIcons()
	for _, e := range cl.Last(height) {
		glyph := icons.Refresh
		cmdStyle := style.SubtleText
		switch e.Status {
		case CommandRunning:
			glyph = icons.Refresh
			cmdStyle = style.AccentText
		case CommandSuccess:
			glyph = icons.OK
			cmdStyle = style.NormalItem
		case CommandWarn:
			glyph = icons.Warn
			cmdStyle = style.WarningText
		case CommandError:
			glyph = icons.Error
			cmdStyle = style.ErrorBadge
		}
		line := presentation.WithIcon(glyph, "brew "+e.Command)
		lines = append(lines, cmdStyle.Render(line))
	}

	rendered := lipgloss.JoinVertical(lipgloss.Top, lines...)
	return lipgloss.NewStyle().Width(width).Height(height).Render(rendered)
}

func brewCommandString(args []string) string {
	return strings.Join(args, " ")
}
