package gui

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
	"github.com/lthiagol/lazybrew/internal/gui/textutil"
	"github.com/lthiagol/lazybrew/internal/logtail"
)

// logsTailLines is how many lines of a service's log the Logs tab shows: the
// last ones.
const logsTailLines = 200

// logsPollInterval is how often the Logs tab reads its log again while it is
// shown. A test may set it.
var logsPollInterval = time.Second

// logStream is the log the Logs tab reads: brew's stdout log (LogPath) or its
// stderr log (ErrorLogPath). The zero value is stdout.
type logStream int

const (
	stdoutStream logStream = iota
	stderrStream
)

func (s logStream) String() string {
	if s == stderrStream {
		return "stderr"
	}
	return "stdout"
}

// other is the stream the e key switches to.
func (s logStream) other() logStream {
	if s == stderrStream {
		return stdoutStream
	}
	return stderrStream
}

// logsView is the Logs tab's state: which service and stream it shows, and the
// last read of that log. A new service or stream starts a new view, so no line
// of another log is ever kept.
type logsView struct {
	service string
	stream  logStream
	// gen numbers the reads this view has started. A read that lands with an
	// older gen is stale and is dropped.
	gen int
	// started is set while the view's loop runs: a read or the tick before the
	// next one is in flight, and the view is shown.
	started bool
	loaded  bool
	path    string
	lines   []string
	size    int64
	modTime time.Time
	err     error
}

// logsReadMsg is the result of one read of the view's log file, from path.
type logsReadMsg struct {
	gen  int
	path string
	tail logtail.Tail
	err  error
}

// logsTickMsg is the wake-up that starts the next read of the view's log.
type logsTickMsg struct {
	gen int
}

// logsTick schedules the next read of the view for gen, one poll interval on.
func logsTick(gen int) tea.Cmd {
	return tea.Tick(logsPollInterval, func(time.Time) tea.Msg {
		return logsTickMsg{gen: gen}
	})
}

// selectedServiceName is the name of the service the Services cursor is on, or
// "" when there is none.
func (m *Model) selectedServiceName() string {
	if s := m.panels[PanelServices].selectedService(); s != nil {
		return s.Name
	}
	return ""
}

// onLogsTab reports whether the Logs tab of Services is the one on screen.
func (m Model) onLogsTab() bool {
	return m.activePanel == PanelServices && m.activeTab == servicesLogsTab
}

// logsShown reports whether the Logs tab is on screen for a service.
func (m *Model) logsShown() bool {
	return m.onLogsTab() && m.logs.service != ""
}

// logPath is the file the view's stream reads for its service: "" when brew
// reports no such log, or has not reported the service's info.
func (m Model) logPath() string {
	info := m.panels[PanelServices].serviceInfoFor(m.logs.service)
	if info == nil {
		return ""
	}
	if m.logs.stream == stderrStream {
		return info.ErrorLogPath
	}
	return info.LogPath
}

// logsSameFile is whether the stderr stream of the view's service is the stdout
// file: brew points both at one log.
func (m Model) logsSameFile() bool {
	info := m.panels[PanelServices].serviceInfoFor(m.logs.service)
	return info != nil && info.LogPath != "" && info.LogPath == info.ErrorLogPath
}

// syncLogs brings the Logs view in step with the model after each update. A
// new service resets the view to stdout; a visit to the Logs tab starts the
// loop, which reads at once and then every poll interval. It returns the
// command that first read starts, if any.
func (m *Model) syncLogs() tea.Cmd {
	if name := m.selectedServiceName(); name != m.logs.service {
		m.logs = logsView{service: name, gen: m.logs.gen}
	}
	if !m.logsShown() {
		m.logs.started = false
		return nil
	}
	if m.logs.started {
		return nil
	}
	m.logs.started = true
	m.logs.gen++
	return m.readLogsCmd()
}

// toggleLogStream switches the Logs view between stdout and stderr. The lines
// of the other stream are dropped, and syncLogs starts the read.
func (m *Model) toggleLogStream() {
	m.logs = logsView{service: m.logs.service, stream: m.logs.stream.other(), gen: m.logs.gen}
}

// readLogsCmd reads the tail of the view's log, as a command. It captures the
// path and the gen now, so the result belongs to the view it was started for.
func (m *Model) readLogsCmd() tea.Cmd {
	gen, path, read := m.logs.gen, m.logPath(), m.readLog
	return func() tea.Msg {
		if path == "" {
			return logsReadMsg{gen: gen, path: path, err: logtail.ErrNoPath}
		}
		tail, err := read(path, logsTailLines)
		return logsReadMsg{gen: gen, path: path, tail: tail, err: err}
	}
}

// shiftedBy is how many lines left the top of the window between two reads of
// one log: the smallest k for which the old lines from k on are the first lines
// of the new read. Lines are only appended, so the new read starts where the
// old one now starts, shifted by k. A read that keeps no old line shifts by all
// of them.
func shiftedBy(old, next []string) int {
	for k := range len(old) + 1 {
		kept := old[k:]
		if len(kept) <= len(next) && slices.Equal(kept, next[:len(kept)]) {
			return k
		}
	}
	return len(old)
}

// applyLogs stores a read's result. The view follows the newest line: when it
// was at the bottom, it stays there as the lines change. When the user had
// scrolled up, the same line stays at the top: the offset drops by the lines
// that left the top of the window, and clamps to the top when its own line
// left. A read of a file that has not changed since the last one (same path,
// size, modification time and error) keeps the lines as they are and rebuilds
// nothing.
func (m *Model) applyLogs(msg logsReadMsg) {
	if m.logs.loaded && msg.path == m.logs.path && sameError(msg.err, m.logs.err) &&
		msg.tail.Size == m.logs.size && msg.tail.ModTime.Equal(m.logs.modTime) {
		return
	}
	key := m.contentKey()
	offset := m.contentScroll[key]
	atBottom := offset >= m.contentLast()
	// The lines of the old read are the same log's only when the path matches
	// and neither read failed; otherwise nothing carried over, so nothing shifts.
	shift := 0
	if m.logs.loaded && msg.path == m.logs.path && m.logs.err == nil && msg.err == nil {
		shift = shiftedBy(m.logs.lines, msg.tail.Lines)
	}
	m.logs.loaded = true
	m.logs.path = msg.path
	m.logs.lines = msg.tail.Lines
	m.logs.size = msg.tail.Size
	m.logs.modTime = msg.tail.ModTime
	m.logs.err = msg.err
	if atBottom {
		m.contentScroll[key] = m.contentLast()
	} else {
		m.contentScroll[key] = max(0, offset-shift)
	}
}

// sameError is whether two read errors are the same error text, or both nil.
func sameError(a, b error) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Error() == b.Error()
}

// logsNote is whether the Logs tab says that stderr and stdout share a file. It
// says so only while it shows that file's lines.
func (m Model) logsNote() bool {
	return m.logs.stream == stderrStream && m.logsSameFile() && m.logs.loaded && m.logs.err == nil
}

// contentHead is the header of the main content that stays on screen while the
// content scrolls: the Logs tab's stream and file, and the shared-file note when
// it applies. It is "" on other tabs.
func (m Model) contentHead(width int) string {
	if !m.onLogsTab() {
		return ""
	}
	head := style.SubtleText.Render(logHeader(m.logs.stream, m.logPath(), width))
	if m.logsNote() {
		head += "\n" + style.SubtleText.Render("stderr goes to the same file as stdout")
	}
	return cutLines(head, width)
}

// renderServiceLogs is the Logs tab of Services, below its header: the last
// lines of the file, or the one muted line that says why there are none. Every
// line is cut to the width, never wrapped.
func (m Model) renderServiceLogs(width int) string {
	path := m.logPath()
	var rows []string
	switch {
	case path == "":
		rows = append(rows, style.SubtleText.Render(fmt.Sprintf("no %s log for this service", m.logs.stream)))
	case !m.logs.loaded:
		rows = append(rows, m.spinner.View()+style.SubtleText.Render(" Loading..."))
	case errors.Is(m.logs.err, fs.ErrNotExist):
		rows = append(rows, style.SubtleText.Render(path+" does not exist yet"))
	case m.logs.err != nil:
		rows = append(rows, style.SubtleText.Render(m.logs.err.Error()))
	default:
		rows = append(rows, m.logs.lines...)
	}
	return cutLines(strings.Join(rows, "\n"), width)
}

// logHeader names the stream and the file it reads, cut to fit: the
// directories of a path too long for the width give way to "/…", and the file
// name stays.
func logHeader(stream logStream, path string, width int) string {
	head := stream.String()
	if path == "" {
		return head
	}
	prefix := head + " · "
	return prefix + elidePath(path, width-lipgloss.Width(prefix))
}

// elidePath shortens path to w cells: it keeps the longest run of trailing
// components that fits, behind "/…", and cuts the name when even the file name
// does not fit.
func elidePath(path string, w int) string {
	if lipgloss.Width(path) <= w {
		return path
	}
	parts := strings.Split(path, "/")
	for i := 2; i < len(parts); i++ {
		if cand := "/…/" + strings.Join(parts[i:], "/"); lipgloss.Width(cand) <= w {
			return cand
		}
	}
	return textutil.Clip(path, w)
}
