package gui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// upgradeTranscript is the shape of a `brew upgrade bash` run: brew's `==> `
// steps with their output under them. It is a transcript written to brew's
// documented step lines; no upgrade was run to make it.
var upgradeTranscript = []string{
	"==> Upgrading 1 outdated package:",
	"bash 5.2.37 -> 5.3.0",
	"==> Upgrading bash 5.2.37 -> 5.3.0",
	"==> Fetching bash",
	"==> Downloading https://ghcr.io/v2/homebrew/core/bash/manifests/5.3.0",
	"Already downloaded: /home/user/Library/Caches/Homebrew/downloads/bash.tar.gz",
	"==> Pouring bash--5.3.0.arm64_sonoma.bottle.tar.gz",
	"🍺  /opt/homebrew/Cellar/bash/5.3.0: 213 files, 3.1MB",
	"==> Running `brew cleanup bash`...",
	"Removing: /opt/homebrew/Cellar/bash/5.2.37... (1,234 files, 3MB)",
}

// TestParseStepsFromBrewUpgradeTranscript (v2-05 AC-05.4): each `==> ` line
// opens a step, and the lines after it are that step's output. A blank line is
// not output.
func TestParseStepsFromBrewUpgradeTranscript(t *testing.T) {
	steps := parseSteps(append(upgradeTranscript, ""))
	titles := make([]string, 0, len(steps))
	for _, s := range steps {
		titles = append(titles, s.title)
	}
	want := []string{
		"Upgrading 1 outdated package:",
		"Upgrading bash 5.2.37 -> 5.3.0",
		"Fetching bash",
		"Downloading https://ghcr.io/v2/homebrew/core/bash/manifests/5.3.0",
		"Pouring bash--5.3.0.arm64_sonoma.bottle.tar.gz",
		"Running `brew cleanup bash`...",
	}
	if strings.Join(titles, "|") != strings.Join(want, "|") {
		t.Fatalf("steps:\n%q\nwant:\n%q", titles, want)
	}
	if got := steps[0].output; len(got) != 1 || got[0] != "bash 5.2.37 -> 5.3.0" {
		t.Fatalf("the first step's output = %q", got)
	}
	if got := steps[len(steps)-1].output; len(got) != 1 {
		t.Fatalf("the cleanup step's output = %q", got)
	}
}

func TestParseStepsIgnoresLinesBeforeTheFirstStep(t *testing.T) {
	steps := parseSteps([]string{"Warning: something early", "==> Fetching bash"})
	if len(steps) != 1 || len(steps[0].output) != 0 {
		t.Fatalf("lines before the first step belong to no step, got %+v", steps)
	}
}

func TestFormatElapsed(t *testing.T) {
	cases := map[time.Duration]string{
		0:                              "0s",
		12 * time.Second:               "12s",
		59*time.Second + 900e6:         "59s",
		64 * time.Second:               "1m04s",
		10*time.Minute + 5*time.Second: "10m05s",
		-3 * time.Second:               "0s",
	}
	for d, want := range cases {
		if got := formatElapsed(d); got != want {
			t.Errorf("formatElapsed(%v) = %q, want %q", d, got, want)
		}
	}
}

// TestOperationFinalStates (v2-05 AC-05.1): the run ends as ✓ Completed in,
// ✗ Failed: or Cancelled.
func TestOperationFinalStates(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	cases := []struct {
		name string
		op   Operation
		want string
	}{
		{"completed", Operation{Title: "upgrade bash", Status: opSuccess, Started: start, Ended: start.Add(34 * time.Second)},
			"✓ Completed in 34s"},
		{"failed", Operation{Title: "upgrade bash", Status: opError, Err: errors.New("boom"), Started: start, Ended: start.Add(2 * time.Second)},
			"✗ Failed: boom"},
		{"cancelled", Operation{Title: "upgrade bash", Status: opCancelled, Started: start, Ended: start.Add(time.Second)},
			"Cancelled"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := newTestModel()
			m.icons = style.UnicodeIcons
			op := c.op
			m.opState = &op
			out := ansi.Strip(m.renderOperation(60, 12))
			if !strings.Contains(out, c.want) {
				t.Fatalf("want %q in:\n%s", c.want, out)
			}
		})
	}
}

// TestOperationElapsedWhileRunning: the view counts up from the start.
func TestOperationElapsedWhileRunning(t *testing.T) {
	m := newTestModel()
	m.opState = &Operation{Title: "upgrade bash", Status: opRunning, Started: time.Now().Add(-12 * time.Second)}
	out := ansi.Strip(m.renderOperation(60, 12))
	if !strings.Contains(out, "Running 12s") {
		t.Fatalf("the running view should show the elapsed time, got:\n%s", out)
	}
}

// TestOperationStepMarks (v2-05 AC-05.1): earlier steps are ✓, the current step
// has the spinner, and its output is under it.
func TestOperationStepMarks(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
	m.opState = &Operation{Title: "upgrade bash", Status: opRunning, Started: time.Now(), Lines: upgradeTranscript}
	out := ansi.Strip(m.renderOperation(90, 30))
	if !strings.Contains(out, "✓ Fetching bash") {
		t.Fatalf("a finished step should be ticked, got:\n%s", out)
	}
	if !strings.Contains(out, "Removing: /opt/homebrew/Cellar/bash/5.2.37") {
		t.Fatalf("the current step's latest output should show, got:\n%s", out)
	}
	if strings.Contains(out, "✓ Running `brew cleanup") {
		t.Fatalf("the current step is not finished, got:\n%s", out)
	}
}

// TestOperationTailKeepsTheLatestOutput: when the steps do not fit, the top is
// cut and the latest output stays on screen.
func TestOperationTailKeepsTheLatestOutput(t *testing.T) {
	m := newTestModel()
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, "==> step "+strings.Repeat("x", i%5))
	}
	lines = append(lines, "==> the last step")
	m.opState = &Operation{Title: "upgrade everything", Status: opRunning, Started: time.Now(), Lines: lines}
	out := ansi.Strip(m.renderOperation(60, 8))
	if !strings.Contains(out, "the last step") {
		t.Fatalf("the latest step must stay visible, got:\n%s", out)
	}
	if got := strings.Count(out, "\n") + 1; got > 8 {
		t.Fatalf("the view is %d rows, want at most 8", got)
	}
}

// TestBatchBarShowsNOfTotal (v2-05 AC-05.2): a batch shows a progress bar and
// n/total beside it.
func TestBatchBarShowsNOfTotal(t *testing.T) {
	m := newTestModel()
	m.opState = &Operation{Title: "Upgrade foo (7/10 done)", Status: opRunning, Started: time.Now(),
		BatchDone: 7, BatchTotal: 10}
	out := ansi.Strip(m.renderOperation(60, 12))
	if !strings.Contains(out, "7/10") {
		t.Fatalf("the batch should show 7/10, got:\n%s", out)
	}
	if !strings.Contains(out, "━") {
		t.Fatalf("the batch should draw a bar, got:\n%s", out)
	}
}

// TestUpgradeAllCountsPackagesBegun (v2-05 AC-05.2): an upgrade of everything
// counts the packages brew has begun upgrading, less the one in progress.
func TestUpgradeAllCountsPackagesBegun(t *testing.T) {
	m := newTestModel()
	m.opState = &Operation{Title: "Upgrade All", Status: opRunning, Started: time.Now(),
		BatchTotal: 5, Upgrades: true, Lines: []string{
			"==> Upgrading bash 5.2 -> 5.3",
			"==> Pouring bash",
			"==> Upgrading curl 8.1 -> 8.2",
			"==> Pouring curl",
			"==> Upgrading git 2.1 -> 2.2",
		}}
	if got := m.operationDone(m.opState); got != 2 {
		t.Fatalf("3 begun with one in progress should be 2 done, got %d", got)
	}
	m.opState.Status = opSuccess
	if got := m.operationDone(m.opState); got != 3 {
		t.Fatalf("a finished run has every begun package done, got %d", got)
	}
}
