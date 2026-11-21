package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/logtail"
)

// stepMsg is updateModel that also returns the command the update started.
func stepMsg(m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	nm, cmd := m.Update(msg)
	switch v := nm.(type) {
	case Model:
		return &v, cmd
	case *Model:
		return v, cmd
	}
	return m, cmd
}

// landLogReads runs the command an update started and feeds its log read back
// into the model, as the program would. Other commands are not run: a tick
// would wait a second.
func landLogReads(m *Model, cmd tea.Cmd) *Model {
	if cmd == nil {
		return m
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			m = landLogReads(m, c)
		}
	case logsReadMsg:
		m, _ = stepMsg(m, msg)
	}
	return m
}

// pressKey is a key press that lands the read it starts.
func pressKey(m *Model, key string) *Model {
	m, cmd := stepMsg(m, tea.KeyPressMsg{Code: []rune(key)[0], Text: key})
	return landLogReads(m, cmd)
}

// writeLog puts lines in a new file under the test's temp directory.
func writeLog(t *testing.T, name string, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// numberedLines is n lines, "line 001" to "line 0nn".
func numberedLines(n int) []string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %03d", i+1)
	}
	return lines
}

// logsModel is a sized model with the Services panel loaded and its Logs tab
// open on moshi-hook, whose stdout and stderr logs are the given paths.
func logsModel(t *testing.T, w, h int, stdout, stderr string) *Model {
	t.Helper()
	info := servicesInfo()
	for i := range info {
		if info[i].Name == "moshi-hook" {
			info[i].LogPath = stdout
			info[i].ErrorLogPath = stderr
		}
	}
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: w, Height: h})
	m = loadServices(m, servicesList(), info)
	m = selectService(m, "moshi-hook")
	return pressKey(m, "l")
}

// plainView is the frame as plain text, one string per row.
func plainView(m *Model) string {
	return ansi.Strip(m.View().Content)
}

func TestServicesTabsAreInfoAndLogs(t *testing.T) {
	var names []string
	for _, tab := range panelTabs[PanelServices] {
		names = append(names, tab.name)
	}
	if got := strings.Join(names, ","); got != "Info,Logs" {
		t.Errorf("Services tabs = %q, want Info,Logs", got)
	}
}

func TestLogsTabShowsLastLinesOfStdout(t *testing.T) {
	stdout := writeLog(t, "moshi-hook.log", numberedLines(250)...)
	m := logsModel(t, 140, 40, stdout, writeLog(t, "err.log", "unused"))
	view := plainView(m)

	if !strings.Contains(view, "stdout · "+stdout) {
		t.Errorf("the header does not name the stdout stream and its path:\n%s", view)
	}
	if got := len(m.logs.lines); got != logsTailLines {
		t.Errorf("the view holds %d lines, want the last %d", got, logsTailLines)
	}
	if first, last := m.logs.lines[0], m.logs.lines[len(m.logs.lines)-1]; first != "line 051" || last != "line 250" {
		t.Errorf("the view holds %q to %q, want line 051 to line 250", first, last)
	}
	// The view opens on the newest line.
	if !strings.Contains(view, "line 250") {
		t.Errorf("the view does not show the newest line:\n%s", view)
	}
}

func TestLogsEToggleStreamAndBack(t *testing.T) {
	stdout := writeLog(t, "out.log", "out line")
	stderr := writeLog(t, "err.log", "err line")
	m := logsModel(t, 140, 40, stdout, stderr)
	if view := plainView(m); !strings.Contains(view, "out line") || strings.Contains(view, "err line") {
		t.Fatalf("the Logs tab starts on stdout:\n%s", view)
	}

	m = pressKey(m, "e")
	view := plainView(m)
	if !strings.Contains(view, "stderr · "+stderr) || !strings.Contains(view, "err line") {
		t.Errorf("e did not show stderr with its path:\n%s", view)
	}
	if strings.Contains(view, "out line") {
		t.Errorf("stderr still shows stdout's line:\n%s", view)
	}

	m = pressKey(m, "e")
	if view := plainView(m); !strings.Contains(view, "out line") {
		t.Errorf("a second e did not go back to stdout:\n%s", view)
	}
}

// When brew points stderr at the stdout file, the lines show with a note.
func TestLogsStderrSameFileNote(t *testing.T) {
	shared := writeLog(t, "moshi-hook.log", "shared line")
	m := pressKey(logsModel(t, 140, 40, shared, shared), "e")
	view := plainView(m)
	if !strings.Contains(view, "stderr goes to the same file as stdout") {
		t.Errorf("no note for a shared file:\n%s", view)
	}
	if !strings.Contains(view, "shared line") {
		t.Errorf("a shared file does not show its lines:\n%s", view)
	}
}

// The note is a header row, so it stays on screen while the lines scroll.
func TestLogsStderrSameFileNoteStaysWhileFollowing(t *testing.T) {
	shared := writeLog(t, "moshi-hook.log", numberedLines(250)...)
	m := pressKey(logsModel(t, 80, 24, shared, shared), "e")
	view := plainView(m)
	if !strings.Contains(view, "stderr goes to the same file as stdout") || !strings.Contains(view, "line 250") {
		t.Errorf("the note or the newest line is off screen:\n%s", view)
	}
	assertFrameSize(t, m.View().Content, 80, 24)
}

// Without the lines there is no note: a missing file says so instead.
func TestLogsNoNoteWithoutLines(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.log")
	m := pressKey(logsModel(t, 140, 40, missing, missing), "e")
	if view := plainView(m); strings.Contains(view, "same file as stdout") {
		t.Errorf("a missing shared file shows the note:\n%s", view)
	}
}

// The stream goes back to stdout when the service changes, and no line of the
// old service's log is kept.
func TestLogsStreamResetsOnServiceChange(t *testing.T) {
	stdout := writeLog(t, "out.log", "out line")
	stderr := writeLog(t, "err.log", "err line")
	m := pressKey(logsModel(t, 140, 40, stdout, stderr), "e")
	if !strings.Contains(plainView(m), "err line") {
		t.Fatal("the test did not reach stderr")
	}

	// j moves from moshi-hook to backup, which has no log path.
	m = pressKey(m, "j")
	view := plainView(m)
	if strings.Contains(view, "err line") || strings.Contains(view, "out line") {
		t.Errorf("another service's log lines are still shown:\n%s", view)
	}
	if !strings.Contains(view, "no stdout log for this service") {
		t.Errorf("backup shows no stdout empty state:\n%s", view)
	}
	if m.logs.stream != stdoutStream {
		t.Errorf("stream = %v after a service change, want stdout", m.logs.stream)
	}
}

func TestLogsEmptyStates(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.log")
	stdout := writeLog(t, "out.log", "out line")

	t.Run("path unset", func(t *testing.T) {
		m := logsModel(t, 140, 40, "", stdout)
		if view := plainView(m); !strings.Contains(view, "no stdout log for this service") {
			t.Errorf("an unset path shows:\n%s", view)
		}
	})
	t.Run("path unset on stderr", func(t *testing.T) {
		m := pressKey(logsModel(t, 140, 40, stdout, ""), "e")
		if view := plainView(m); !strings.Contains(view, "no stderr log for this service") {
			t.Errorf("an unset stderr path shows:\n%s", view)
		}
	})
	t.Run("file missing", func(t *testing.T) {
		m := logsModel(t, 140, 40, missing, missing+".err")
		if view := plainView(m); !strings.Contains(view, missing+" does not exist yet") {
			t.Errorf("a missing file shows:\n%s", view)
		}
	})
	t.Run("read error", func(t *testing.T) {
		dir := t.TempDir()
		m := logsModel(t, 140, 40, dir, stdout)
		view := plainView(m)
		if !strings.Contains(view, "is a directory") {
			t.Errorf("a read error shows:\n%s", view)
		}
		if len(m.toasts) != 0 {
			t.Errorf("a log read error raised %d toasts, want none", len(m.toasts))
		}
	})
}

// A line longer than the main area is cut, not wrapped: the frame keeps its
// size and the rest of the line is not on a row of its own.
func TestLogsCutsLongLinesAtEveryFrame(t *testing.T) {
	long := "START" + strings.Repeat("x", 300) + "END"
	stdout := writeLog(t, "out.log", long)
	for _, size := range [][2]int{{80, 24}, {140, 40}} {
		m := logsModel(t, size[0], size[1], stdout, stdout)
		view := m.View().Content
		assertFrameSize(t, view, size[0], size[1])
		plain := ansi.Strip(view)
		if strings.Contains(plain, "END") {
			t.Errorf("%dx%d: the cut line wrapped onto a row:\n%s", size[0], size[1], plain)
		}
		if !strings.Contains(plain, "START") {
			t.Errorf("%dx%d: the line's start is gone:\n%s", size[0], size[1], plain)
		}
	}
}

func TestLogsHeaderCutsPathToFit(t *testing.T) {
	path := "/home/linuxbrew/.linuxbrew/var/log/moshi-hook.log"
	if got := logHeader(stdoutStream, path, 80); got != "stdout · "+path {
		t.Errorf("a path that fits = %q", got)
	}
	got := logHeader(stdoutStream, path, 30)
	if want := "stdout · /…/log/moshi-hook.log"; got != want {
		t.Errorf("logHeader(30) = %q, want %q", got, want)
	}
	if w := lipgloss.Width(got); w > 30 {
		t.Errorf("header is %d cells, want at most 30", w)
	}
	if got := logHeader(stdoutStream, "", 30); got != "stdout" {
		t.Errorf("a header with no path = %q, want stdout", got)
	}
}

// The Logs keys: e and the tab keys are in the footer on the Logs tab only,
// and e is in help.
func TestLogsFooterListsStreamOnLogsOnly(t *testing.T) {
	m := selectService(loadServices(newTestModel(), servicesList(), servicesInfo()), "moshi-hook")
	hasHint := func(hints []keyHint, key, desc string) bool {
		for _, h := range hints {
			if h.key == key && h.desc == desc {
				return true
			}
		}
		return false
	}
	if hints := m.footerHints(); !hasHint(hints, "h/l", "tabs") || hasHint(hints, "e", "stream") {
		t.Errorf("Info footer = %v, want h/l tabs and no e", hints)
	}
	m = pressKey(m, "l")
	if hints := m.footerHints(); !hasHint(hints, "e", "stream") || !hasHint(hints, "h/l", "tabs") {
		t.Errorf("Logs footer = %v, want e stream and h/l tabs", hints)
	}
}

func TestLogsScrollKeys(t *testing.T) {
	stdout := writeLog(t, "out.log", numberedLines(250)...)
	m := logsModel(t, 80, 24, stdout, stdout)
	if view := plainView(m); strings.Contains(view, "line 001") {
		t.Fatalf("the view starts at the bottom, yet shows the first line:\n%s", view)
	}
	m = pressKey(m, "g")
	if view := plainView(m); !strings.Contains(view, "line 051") {
		t.Errorf("g did not go to the top of the last 200 lines:\n%s", view)
	}
	m = pressKey(m, "G")
	if view := plainView(m); !strings.Contains(view, "line 250") {
		t.Errorf("G did not go to the bottom:\n%s", view)
	}
}

func TestLogsReadStartsOnEachVisit(t *testing.T) {
	stdout := writeLog(t, "out.log", "out line")
	m := logsModel(t, 140, 40, stdout, stdout)
	var reads []string
	m.readLog = func(path string, n int) (logtail.Tail, error) {
		reads = append(reads, path)
		return logtail.Read(path, n)
	}
	// Leaving the tab starts no read; coming back starts one.
	m = pressKey(m, "h")
	m = pressKey(m, "l")
	if len(reads) != 1 {
		t.Errorf("reads after leaving and returning = %d, want 1", len(reads))
	}
	if !strings.Contains(plainView(m), "out line") {
		t.Errorf("the Logs tab lost its lines:\n%s", plainView(m))
	}
}

// The help lists e under Services, where it acts on the Logs tab.
func TestLogsHelpListsStreamKeyUnderServices(t *testing.T) {
	for _, page := range buildHelp() {
		if page.title != "Services" {
			continue
		}
		for _, sec := range page.sections {
			for _, item := range sec.items {
				if item.key == "e" {
					return
				}
			}
		}
		t.Fatal("the Services help page has no e entry")
	}
	t.Fatal("there is no Services help page")
}

// Two services, each with its own stdout log: after j the view shows the new
// service's lines and none of the old one's, and a read of the old view that
// lands after the switch is dropped.
func TestLogsSwitchServiceShowsNewLogAndDropsOld(t *testing.T) {
	alpha := writeLog(t, "alpha.log", "alpha line 1", "alpha line 2")
	beta := writeLog(t, "beta.log", "beta line 1")
	info := servicesInfo()
	for i := range info {
		switch info[i].Name {
		case "moshi-hook":
			info[i].LogPath, info[i].ErrorLogPath = alpha, alpha
		case "backup":
			info[i].LogPath, info[i].ErrorLogPath = beta, beta
		}
	}
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = loadServices(m, servicesList(), info)
	m = selectService(m, "moshi-hook")
	m = pressKey(m, "l")
	if view := plainView(m); !strings.Contains(view, "alpha line 2") {
		t.Fatalf("moshi-hook's log does not show before the switch:\n%s", view)
	}
	oldGen := m.logs.gen

	m = pressKey(m, "j")
	view := plainView(m)
	if !strings.Contains(view, "beta line 1") {
		t.Errorf("backup's log does not show after j:\n%s", view)
	}
	if strings.Contains(view, "alpha line") {
		t.Errorf("moshi-hook's lines are still shown after j:\n%s", view)
	}

	// A read of the old view lands late, carrying moshi-hook's lines.
	before := plainView(m)
	m, _ = stepMsg(m, logsReadMsg{gen: oldGen, path: alpha, tail: logtail.Tail{
		Lines: []string{"alpha line 1", "alpha line 2"},
	}})
	if after := plainView(m); after != before {
		t.Errorf("a stale read of the old service changed the view:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
