package gui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/logtail"
)

// countReads makes m count the log reads it starts, through the injectable
// reader, and returns the counter.
func countReads(m *Model) *int {
	n := new(int)
	m.readLog = func(path string, lines int) (logtail.Tail, error) {
		*n++
		return logtail.Read(path, lines)
	}
	return n
}

// appendLog adds lines to the end of a log file.
func appendLog(t *testing.T, path string, lines ...string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		t.Fatal(err)
	}
}

// tick delivers the view's next wake-up and lands the read it starts.
func tick(m *Model) *Model {
	m, cmd := stepMsg(m, logsTickMsg{gen: m.logs.gen})
	return landLogReads(m, cmd)
}

// A line appended to the file shows on the next tick, and the view stays
// pinned to the newest line.
func TestLogsFollowShowsAppendedLines(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", "first", "second")
	m := logsModel(t, 80, 24, path, path)
	if !strings.Contains(plainView(m), "second") {
		t.Fatalf("the initial read is missing:\n%s", plainView(m))
	}

	appendLog(t, path, "third")
	m = tick(m)
	view := plainView(m)
	if !strings.Contains(view, "third") {
		t.Errorf("the appended line did not show on the tick:\n%s", view)
	}
	assertFrameSize(t, m.View().Content, 80, 24)
}

// A view at the bottom stays on the newest line as lines arrive, over many
// lines; the frame keeps its size.
func TestLogsFollowStaysPinnedAtBottom(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", numberedLines(250)...)
	m := logsModel(t, 80, 24, path, path)
	for i := 251; i <= 260; i++ {
		appendLog(t, path, fmt.Sprintf("line %03d", i))
		m = tick(m)
	}
	view := plainView(m)
	if !strings.Contains(view, "line 260") {
		t.Errorf("the view left the newest line:\n%s", view)
	}
	if got := len(m.logs.lines); got != logsTailLines {
		t.Errorf("the view holds %d lines, want %d", got, logsTailLines)
	}
	assertFrameSize(t, m.View().Content, 80, 24)
}

// A view the user scrolled up does not move when lines arrive: its scroll
// offset is the same, and the row it shows at the top is the same line.
func TestLogsFollowScrolledUpDoesNotMove(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", numberedLines(50)...)
	m := logsModel(t, 80, 24, path, path)
	m = pressKey(m, "g")
	before := m.contentScroll[m.contentKey()]
	top := plainView(m)

	appendLog(t, path, "line 051", "line 052")
	m = tick(m)
	if got := m.contentScroll[m.contentKey()]; got != before {
		t.Errorf("scroll offset moved from %d to %d on a new line", before, got)
	}
	if view := plainView(m); !strings.Contains(view, "line 001") || strings.Contains(view, "line 052") {
		t.Errorf("a scrolled-up view moved to the new lines:\n%s\nwas:\n%s", view, top)
	}

	// end brings the view back to the newest line.
	m = pressKey(m, "G")
	if view := plainView(m); !strings.Contains(view, "line 052") {
		t.Errorf("end did not show the newest line:\n%s", view)
	}
}

// A file that shrinks (truncated or rotated) is read again from its new end.
func TestLogsFollowRereadsShrunkFile(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", numberedLines(40)...)
	m := logsModel(t, 80, 24, path, path)

	if err := os.WriteFile(path, []byte("fresh one\nfresh two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m = tick(m)
	view := plainView(m)
	if !strings.Contains(view, "fresh two") || strings.Contains(view, "line 040") {
		t.Errorf("a rotated file was not read from its new end:\n%s", view)
	}
}

// A tick for a view that is no longer shown ends the loop: no read starts, and
// no next tick is scheduled.
func TestLogsLeavingTabStopsReads(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", "one")
	m := logsModel(t, 140, 40, path, path)
	reads := countReads(m)

	m = pressKey(m, "[")
	m, cmd := stepMsg(m, logsTickMsg{gen: m.logs.gen})
	if cmd != nil {
		t.Errorf("a tick on another tab started a command")
	}
	if *reads != 0 {
		t.Errorf("reads after leaving the tab = %d, want 0", *reads)
	}
}

// A read that lands after the view was left is not applied and schedules no
// next tick.
func TestLogsLateReadAfterLeavingSchedulesNothing(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", "one")
	m := logsModel(t, 140, 40, path, path)
	gen := m.logs.gen
	m = pressKey(m, "[")

	m, cmd := stepMsg(m, logsReadMsg{gen: gen, path: path, tail: logtail.Tail{Lines: []string{"late"}}})
	if cmd != nil {
		t.Errorf("a read that landed after leaving scheduled a tick")
	}
	if strings.Contains(strings.Join(m.logs.lines, "\n"), "late") {
		t.Errorf("a read that landed after leaving was applied")
	}
}

// Switching service or stream drops the old view's tick and read: the new view
// has its own loop, and no line of the old log is shown.
func TestLogsStaleTickAfterStreamToggleIsDropped(t *testing.T) {
	stdout := writeLog(t, "out.log", "out line")
	stderr := writeLog(t, "err.log", "err line")
	m := logsModel(t, 140, 40, stdout, stderr)
	oldGen := m.logs.gen

	m = pressKey(m, "e")
	if m.logs.gen == oldGen {
		t.Fatal("toggling the stream did not start a new view")
	}
	m, cmd := stepMsg(m, logsTickMsg{gen: oldGen})
	if cmd != nil {
		t.Errorf("a stale tick started a command")
	}
	m, cmd = stepMsg(m, logsReadMsg{gen: oldGen, path: stdout, tail: logtail.Tail{Lines: []string{"out line"}}})
	if cmd != nil || strings.Contains(plainView(m), "out line") {
		t.Errorf("a stale read was applied to the stderr view:\n%s", plainView(m))
	}
}

// A read of a file that has not changed keeps the lines and rebuilds nothing.
func TestLogsUnchangedFileIsNotReapplied(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", "one", "two")
	m := logsModel(t, 80, 24, path, path)
	before := m.logs
	m.logs.lines = []string{"kept"}

	// The same path, size, modification time and error: nothing is applied.
	m, _ = stepMsg(m, logsReadMsg{gen: m.logs.gen, path: path, tail: logtail.Tail{
		Lines:   []string{"replaced"},
		Size:    before.size,
		ModTime: before.modTime,
	}})
	if len(m.logs.lines) != 1 || m.logs.lines[0] != "kept" {
		t.Errorf("an unchanged file's read replaced the lines: %q", m.logs.lines)
	}

	// A changed size is applied.
	m, _ = stepMsg(m, logsReadMsg{gen: m.logs.gen, path: path, tail: logtail.Tail{
		Lines:   []string{"replaced"},
		Size:    before.size + 1,
		ModTime: before.modTime,
	}})
	if len(m.logs.lines) != 1 || m.logs.lines[0] != "replaced" {
		t.Errorf("a changed file's read was not applied: %q", m.logs.lines)
	}
}

// Each tick reads once and schedules the next wake-up, so the view keeps
// following while it is shown.
func TestLogsTickReadsAndSchedulesNext(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", "one")
	m := logsModel(t, 140, 40, path, path)
	reads := countReads(m)

	m, cmd := stepMsg(m, logsTickMsg{gen: m.logs.gen})
	if cmd == nil {
		t.Fatal("a tick on the Logs tab started no read")
	}
	// Run the read the tick started; its result schedules the next tick.
	m, cmd = stepMsg(m, cmd())
	if *reads != 1 {
		t.Errorf("reads = %d, want 1", *reads)
	}
	if cmd == nil {
		t.Error("the read's result scheduled no next tick while the view is shown")
	}
	if !m.logs.started {
		t.Error("the loop stopped while the view is shown")
	}
}

// The frame stays the terminal's size while lines arrive at 80x24 and 140x40.
func TestLogsFollowFrameSizeAtEachSize(t *testing.T) {
	long := "START" + strings.Repeat("y", 200) + "END"
	path := writeLog(t, "moshi-hook.log", "one")
	for _, size := range [][2]int{{80, 24}, {140, 40}} {
		m := logsModel(t, size[0], size[1], path, path)
		appendLog(t, path, long)
		m = tick(m)
		view := m.View().Content
		assertFrameSize(t, view, size[0], size[1])
		if strings.Contains(ansi.Strip(view), "END") {
			t.Errorf("%dx%d: a followed line wrapped", size[0], size[1])
		}
	}
}

// Past 200 lines the window slides: a view scrolled up keeps the same line at
// its top, so the offset drops by the lines that left the top.
func TestLogsFollowScrolledUpReanchorsOnSlide(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", numberedLines(250)...)
	m := logsModel(t, 80, 24, path, path)
	m.scrollContentTo(10)
	key := m.contentKey()
	if got := m.logs.lines[m.contentScroll[key]]; got != "line 061" {
		t.Fatalf("the test did not scroll to line 061: top is %q", got)
	}

	// Three lines arrive: the window holds lines 054 to 253, and line 061 has
	// moved from row 10 to row 7.
	appendLog(t, path, "line 251", "line 252", "line 253")
	m = tick(m)
	if got := m.contentScroll[key]; got != 7 {
		t.Errorf("offset = %d after the slide, want 7", got)
	}
	if got := m.logs.lines[m.contentScroll[key]]; got != "line 061" {
		t.Errorf("the top line is %q after the slide, want line 061", got)
	}
	if view := plainView(m); !strings.Contains(view, "line 061") {
		t.Errorf("line 061 is off the view after the slide:\n%s", view)
	}
	assertFrameSize(t, m.View().Content, 80, 24)
}

// When the line at the top of a scrolled-up view leaves the window, the view
// clamps to the top of the window.
func TestLogsFollowScrolledUpClampsWhenTopLineLeaves(t *testing.T) {
	path := writeLog(t, "moshi-hook.log", numberedLines(250)...)
	m := logsModel(t, 80, 24, path, path)
	m.scrollContentTo(2)
	key := m.contentKey()
	if got := m.logs.lines[m.contentScroll[key]]; got != "line 053" {
		t.Fatalf("the test did not scroll to line 053: top is %q", got)
	}

	// Five lines arrive: the window holds lines 056 to 255, and line 053 is gone.
	var more []string
	for i := 251; i <= 255; i++ {
		more = append(more, fmt.Sprintf("line %03d", i))
	}
	appendLog(t, path, more...)
	m = tick(m)
	if got := m.contentScroll[key]; got != 0 {
		t.Errorf("offset = %d after its line left, want 0", got)
	}
	if got := m.logs.lines[0]; got != "line 056" {
		t.Errorf("the window starts at %q, want line 056", got)
	}
}

// shiftedBy counts the lines that left the top of a window between two reads.
func TestShiftedBy(t *testing.T) {
	cases := []struct {
		name      string
		old, next []string
		want      int
	}{
		{"nothing read yet", nil, []string{"a", "b"}, 0},
		{"same lines", []string{"a", "b"}, []string{"a", "b"}, 0},
		{"appended, nothing left", []string{"a", "b", "c"}, []string{"a", "b", "c", "d", "e"}, 0},
		{"two left the top", []string{"a", "b", "c", "d"}, []string{"c", "d", "e", "f"}, 2},
		{"no line kept", []string{"a", "b"}, []string{"x", "y"}, 2},
		{"file shrank", []string{"a", "b", "c"}, []string{"a"}, 3},
	}
	for _, c := range cases {
		if got := shiftedBy(c.old, c.next); got != c.want {
			t.Errorf("%s: shiftedBy = %d, want %d", c.name, got, c.want)
		}
	}
}
