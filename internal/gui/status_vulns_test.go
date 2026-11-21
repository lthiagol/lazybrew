package gui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// vulnsRunner answers `brew vulns` with out and err, and counts the calls.
func vulnsRunner(out string, err error, calls *int) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "vulns" {
			*calls++
			return []byte(out), err
		}
		return []byte{}, nil
	}
	return r
}

// vKey is the v press on the Status panel.
var vKey = tea.KeyPressMsg{Code: 'v', Text: "v"}

// TestVulnsTabOpenRunsNothing: opening the Vulns tab shows the hint and runs no
// check, so the network is not touched until v is pressed.
func TestVulnsTabOpenRunsNothing(t *testing.T) {
	calls := 0
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(vulnsRunner("CVE-2026-1: openssl@3\n", nil, &calls))

	for range 4 {
		m = pressStatus(m, "l")
	}
	if m.activeTab != statusVulnsTab {
		t.Fatalf("four l presses should reach the Vulns tab, got tab %d", m.activeTab)
	}
	if calls != 0 {
		t.Fatalf("opening the Vulns tab ran brew vulns %d times, want 0", calls)
	}
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, vulnsHint) {
		t.Fatalf("the tab should show the hint:\n%s", out)
	}
}

// TestVulnsKeyRunsOnceWhileRunning: v starts the check and shows the spinner
// line; a second v while it runs starts nothing; the result lands once.
func TestVulnsKeyRunsOnceWhileRunning(t *testing.T) {
	calls := 0
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(vulnsRunner("CVE-2026-1: openssl@3\n", nil, &calls))

	nm, cmd := m.Update(vKey)
	m = modelOf(nm)
	if m.activeTab != statusVulnsTab {
		t.Fatalf("v should switch to the Vulns tab, got tab %d", m.activeTab)
	}
	if m.vulns.status != vulnsRunning {
		t.Fatalf("v should start the check, status %v", m.vulns.status)
	}
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "Checking…") {
		t.Fatalf("a running check should say so:\n%s", out)
	}

	nm, again := m.Update(vKey)
	m = modelOf(nm)
	if again != nil {
		t.Fatal("a second v while the check runs started another run")
	}

	m = landCmd(m, cmd)
	if calls != 1 {
		t.Fatalf("brew vulns ran %d times, want 1", calls)
	}
	if m.vulns.status != vulnsDone || m.vulns.report != "CVE-2026-1: openssl@3\n" {
		t.Fatalf("the report should be kept, got status %v report %q", m.vulns.status, m.vulns.report)
	}
}

// TestVulnsResultShapes: a report, an empty report, a report under exit 1 (the
// warning case), and a real error each show what the tab says they show.
func TestVulnsResultShapes(t *testing.T) {
	cases := []struct {
		name   string
		out    string
		err    error
		want   string
		failed bool
	}{
		{"report", "CVE-2026-1: openssl@3\n", nil, "CVE-2026-1: openssl@3", false},
		{"empty", "", nil, "No known vulnerabilities", false},
		{"exit 1 with report", "CVE-2026-2: node\n", &brew.BrewExitError{Command: "vulns", ExitCode: 1}, "CVE-2026-2: node", false},
		{"error", "", &brew.BrewExitError{Command: "vulns", ExitCode: 2, Stderr: "Error: Failed to reach OSV.dev\n"}, "Failed to reach OSV.dev", true},
		// brew without the vulns command exits 1 with no report: a failure, not
		// an empty report (F-r1-01).
		{"exit 1 no report", "", &brew.BrewExitError{Command: "vulns", ExitCode: 1, Stderr: "Error: Invalid usage: Unknown command: brew vulns\n"}, "Invalid usage: Unknown command: brew vulns", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			calls := 0
			m := loadedStatus(140, 40)
			m.client = brew.NewClient(vulnsRunner(c.out, c.err, &calls))
			m = pressStatus(m, "v")

			out := ansi.Strip(m.View().Content)
			if !strings.Contains(out, c.want) {
				t.Fatalf("want %q in the tab:\n%s", c.want, out)
			}
			if c.failed {
				if m.vulns.status != vulnsFailed {
					t.Fatalf("status %v, want failed", m.vulns.status)
				}
				if !strings.Contains(out, vulnsHint) {
					t.Errorf("an error should keep the hint to retry with v:\n%s", out)
				}
			} else if m.vulns.status != vulnsDone {
				t.Fatalf("status %v, want done", m.vulns.status)
			}
		})
	}
}

// TestVulnsReportSurvivesRefreshAndPanelSwitch: R does not run the check and
// does not clear its report; neither does leaving the panel and coming back.
func TestVulnsReportSurvivesRefreshAndPanelSwitch(t *testing.T) {
	calls := 0
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(vulnsRunner("CVE-2026-1: openssl@3\n", nil, &calls))
	m = pressStatus(m, "v")

	nm, cmd := m.Update(RefreshMsg{LoadOutdated: true, ReloadDiagnostics: true})
	m = landCmd(modelOf(nm), cmd)
	if calls != 1 {
		t.Fatalf("R ran brew vulns: %d calls, want the 1 from v", calls)
	}
	if m.vulns.report == "" {
		t.Fatal("R cleared the Vulns report")
	}

	m = pressStatus(m, "2")
	m = pressStatus(m, "1")
	for range 4 {
		m = pressStatus(m, "l")
	}
	if calls != 1 {
		t.Fatalf("coming back to the tab ran brew vulns: %d calls, want 1", calls)
	}
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "CVE-2026-1") {
		t.Fatalf("the report should still show:\n%s", out)
	}
}

// TestVulnsTabFrames: each state of the tab draws a frame of the terminal's
// size in both themes.
func TestVulnsTabFrames(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	cases := []struct {
		name  string
		out   string
		err   error
		press []string
		want  string
	}{
		{"not run", "", nil, []string{"l", "l", "l", "l"}, "Press v"},
		{"report", "CVE-2026-1: openssl@3\n", nil, []string{"v"}, "CVE-2026-1"},
		{"empty", "", nil, []string{"v"}, "No known vulnerabilities"},
		{"error", "", &brew.BrewExitError{Command: "vulns", ExitCode: 2, Stderr: "Error: Failed to reach OSV.dev\n"}, []string{"v"}, "Failed to reach OSV.dev"},
	}
	for _, size := range []struct{ w, h int }{{80, 24}, {140, 40}} {
		for theme, th := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
			style.ApplyTheme(th)
			for _, c := range cases {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", c.name, theme, size.w, size.h), func(t *testing.T) {
					calls := 0
					m := loadedStatus(size.w, size.h)
					m.client = brew.NewClient(vulnsRunner(c.out, c.err, &calls))
					for _, key := range c.press {
						m = pressStatus(m, key)
					}
					view := m.View().Content
					assertFrameSize(t, view, size.w, size.h)
					if !strings.Contains(ansi.Strip(view), c.want) {
						t.Errorf("want %q in the frame:\n%s", c.want, ansi.Strip(view))
					}
				})
			}
		}
	}
}

// TestVulnsTimeoutShowsErrorAndHint: a check that outlasts its budget shows the
// timeout as the error line, with the hint to retry with v (F-r1-03).
func TestVulnsTimeoutShowsErrorAndHint(t *testing.T) {
	defer func(d time.Duration) { brewVulnsTimeout = d }(brewVulnsTimeout)
	brewVulnsTimeout = 50 * time.Millisecond

	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(r)
	m = pressStatus(m, "v")

	if m.vulns.status != vulnsFailed {
		t.Fatalf("status %v, want failed", m.vulns.status)
	}
	out := ansi.Strip(m.View().Content)
	if !strings.Contains(out, "brew vulns timed out after 50ms") {
		t.Errorf("the tab should name the timeout:\n%s", out)
	}
	if !strings.Contains(out, vulnsHint) {
		t.Errorf("a timeout should keep the hint to retry with v:\n%s", out)
	}
}

// TestVulnsRunningFrameFitsAt80x24: while the check runs, the tab shows the
// spinner and Checking… in a frame of the terminal's size, in both themes.
func TestVulnsRunningFrameFitsAt80x24(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	for theme, th := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
		style.ApplyTheme(th)
		t.Run(theme, func(t *testing.T) {
			calls := 0
			m := loadedStatus(80, 24)
			m.client = brew.NewClient(vulnsRunner("CVE-2026-1: openssl@3\n", nil, &calls))
			nm, _ := m.Update(vKey) // the check is started, not landed: it is running
			m = modelOf(nm)

			view := m.View().Content
			assertFrameSize(t, view, 80, 24)
			if !strings.Contains(ansi.Strip(view), "Checking…") {
				t.Errorf("a running check should say so:\n%s", ansi.Strip(view))
			}
		})
	}
}

// TestYCopiesTheVulnsReportWhole: y on the Vulns tab copies the whole report,
// every line uncut and ANSI-free, although the pane at 80 columns wraps it.
func TestYCopiesTheVulnsReportWhole(t *testing.T) {
	report := "CVE-2026-1: openssl@3 " + strings.Repeat("affected-advisory ", 6) + "end\n" +
		"CVE-2026-2: node " + strings.Repeat("unfixed-version ", 8) + "end\n"
	calls := 0
	m := loadedStatus(80, 24)
	copied := captureClipboard(m)
	m.client = brew.NewClient(vulnsRunner(report, nil, &calls))
	m = pressStatus(m, "v")

	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "CVE-2026-1") {
		t.Fatalf("the report should be on screen:\n%s", out)
	}
	m = sendKey(m, "y")

	if len(*copied) != 1 {
		t.Fatalf("y copied %d times, want 1", len(*copied))
	}
	got := (*copied)[0]
	if ansi.Strip(got) != got {
		t.Errorf("the copy holds escape codes: %q", got)
	}
	if strings.Contains(got, "…") {
		t.Errorf("the copy is cut: %q", got)
	}
	for _, line := range strings.Split(strings.TrimRight(report, "\n"), "\n") {
		if !strings.Contains(got, line) {
			t.Errorf("the copy lacks %q whole:\n%s", line, got)
		}
	}
}
