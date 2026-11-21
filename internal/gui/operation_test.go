package gui

import (
	"errors"
	"fmt"
	"reflect"
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

// TestOperationFinalStates (v2-05 AC-05.1): the run ends as ✔ Completed in,
// ✗ Failed: or Cancelled.
func TestOperationFinalStates(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	cases := []struct {
		name string
		op   Operation
		want string
	}{
		{"completed", Operation{Title: "upgrade bash", Status: opSuccess, Started: start, Ended: start.Add(34 * time.Second)},
			"✔ Completed in 34s"},
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
	m.opState = operationFrom(Operation{Title: "upgrade bash", Status: opRunning, Started: time.Now()}, upgradeTranscript)
	out := ansi.Strip(m.renderOperation(90, 30))
	if !strings.Contains(out, "✔ Fetching bash") {
		t.Fatalf("a finished step should be ticked, got:\n%s", out)
	}
	if !strings.Contains(out, "Removing: /opt/homebrew/Cellar/bash/5.2.37") {
		t.Fatalf("the current step's latest output should show, got:\n%s", out)
	}
	if strings.Contains(out, "✔ Running `brew cleanup") {
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
	m.opState = operationFrom(Operation{Title: "upgrade everything", Status: opRunning, Started: time.Now()}, lines)
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

// TestUpgradeAllCountsPackagesBegun (v2-05 AC-05.2, fix F-r1-01): the total is
// the N in brew's header, which skips pinned packages, and a package is done
// once its own step has run. The header is not a package.
func TestUpgradeAllCountsPackagesBegun(t *testing.T) {
	cases := []struct {
		name        string
		status      opStatus
		lines       []string
		done, total int
	}{
		{"no header yet", opRunning, []string{"==> Fetching bash"}, 0, 0},
		{"one package, running", opRunning, []string{
			"==> Upgrading 1 outdated package:",
			"bash 5.2 -> 5.3",
			"==> Upgrading bash 5.2 -> 5.3",
			"==> Pouring bash",
		}, 0, 1},
		{"one package, finished", opSuccess, []string{
			"==> Upgrading 1 outdated package:",
			"bash 5.2 -> 5.3",
			"==> Upgrading bash 5.2 -> 5.3",
			"==> Pouring bash",
		}, 1, 1},
		// Four outdated, one pinned: brew's header says three.
		{"three of three, mid-run", opRunning, []string{
			"==> Upgrading 3 outdated packages:",
			"==> Upgrading bash 5.2 -> 5.3",
			"==> Pouring bash",
			"==> Upgrading curl 8.1 -> 8.2",
			"==> Pouring curl",
			"==> Upgrading git 2.1 -> 2.2",
		}, 2, 3},
		{"three of three, failed on the second", opError, []string{
			"==> Upgrading 3 outdated packages:",
			"==> Upgrading bash 5.2 -> 5.3",
			"==> Upgrading curl 8.1 -> 8.2",
		}, 1, 3},
		{"three of three, finished", opSuccess, []string{
			"==> Upgrading 3 outdated packages:",
			"==> Upgrading bash 5.2 -> 5.3",
			"==> Upgrading curl 8.1 -> 8.2",
			"==> Upgrading git 2.1 -> 2.2",
		}, 3, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			op := operationFrom(Operation{Title: "Upgrade All", Status: c.status, Upgrades: true}, c.lines)
			done, total := operationProgress(op)
			if done != c.done || total != c.total {
				t.Fatalf("progress = %d/%d, want %d/%d", done, total, c.done, c.total)
			}
		})
	}
}

// TestUpgradeAllBarShowsOnlyOnceTheTotalIsKnown: no bar before brew's header.
func TestUpgradeAllBarShowsOnlyOnceTheTotalIsKnown(t *testing.T) {
	m := newTestModel()
	m.opState = operationFrom(Operation{Title: "Upgrade All", Status: opRunning, Started: time.Now(),
		Upgrades: true}, []string{"==> Fetching brew"})
	if out := ansi.Strip(m.renderOperation(60, 12)); strings.Contains(out, "/") {
		t.Fatalf("no bar before the header, got:\n%s", out)
	}
	m.opState.AppendLine("==> Upgrading 1 outdated package:")
	m.opState.AppendLine("==> Upgrading bash 5.2 -> 5.3")
	if out := ansi.Strip(m.renderOperation(60, 12)); !strings.Contains(out, "0/1") {
		t.Fatalf("the bar should read 0/1 mid-run, got:\n%s", out)
	}
}

// upgradeFormulaeThenCask is `brew upgrade` over two formulae and one cask. The
// strings are Homebrew's own: cmd/upgrade.rb:458 (the formulae header, with its
// version lines), upgrade.rb:246 (each formula's step, the version on the next
// line) and cask/upgrade.rb:102,434 (the cask header and step). The packages
// and their versions are invented.
var upgradeFormulaeThenCask = []string{
	"==> Upgrading 2 outdated packages:",
	"bash 5.2 -> 5.3",
	"curl 8.1 -> 8.2",
	"==> Upgrading bash",
	"  5.2 -> 5.3",
	"==> Pouring bash--5.3.arm64_sonoma.bottle.tar.gz",
	"==> Upgrading curl",
	"  8.1 -> 8.2",
	"==> Pouring curl--8.2.arm64_sonoma.bottle.tar.gz",
	"==> Upgrading 1 outdated package:",
	"firefox 135.0 -> 136.0",
	"==> Upgrading firefox",
	"==> Downloading https://example.test/firefox-136.0.dmg",
}

// upgradeWithDependent is `brew upgrade` over one formula and one of its
// dependents. The strings are Homebrew's own: upgrade.rb:424 (the check that
// comes before the dependents), upgrade.rb:400 (the dependents header, singular
// here: "1 dependent ... of upgraded formula:") and upgrade.rb:246 (the steps).
var upgradeWithDependent = []string{
	"==> Upgrading 1 outdated package:",
	"zlib 1.3 -> 1.3.1",
	"==> Upgrading zlib",
	"  -> 1.3.1",
	"==> Pouring zlib--1.3.1.arm64_sonoma.bottle.tar.gz",
	"==> Checking for dependents of upgraded formulae...",
	"==> Upgrading 1 dependent of upgraded formula:",
	"zsh 5.9 -> 5.9.1",
	"==> Upgrading zsh",
	"  5.9 -> 5.9.1",
	"==> Pouring zsh--5.9.1.arm64_sonoma.bottle.tar.gz",
}

// progressAt is the progress after the first n lines of a run with the given
// status.
func progressAt(lines []string, n int, status opStatus) (int, int) {
	return operationProgress(operationFrom(Operation{Status: status, Upgrades: true}, lines[:n]))
}

// TestUpgradeAllCountsFormulaeThenCask (v2-05 fix F-r2-01, case a): the total is
// the formulae header's 2 plus the cask header's 1.
func TestUpgradeAllCountsFormulaeThenCask(t *testing.T) {
	n := len(upgradeFormulaeThenCask)
	// Mid-run on the cask: two formulae done, firefox in progress.
	if done, total := progressAt(upgradeFormulaeThenCask, n, opRunning); done != 2 || total != 3 {
		t.Fatalf("mid-run on the cask: %d/%d, want 2/3", done, total)
	}
	if done, total := progressAt(upgradeFormulaeThenCask, n, opSuccess); done != 3 || total != 3 {
		t.Fatalf("finished: %d/%d, want 3/3", done, total)
	}
}

// TestUpgradeAllCountsDependents (fix F-r2-01, case b): the dependent's header
// adds to the total, and the dependent's step is a package.
func TestUpgradeAllCountsDependents(t *testing.T) {
	n := len(upgradeWithDependent)
	// Mid-run on the dependent: the formula is done, zsh in progress.
	if done, total := progressAt(upgradeWithDependent, n, opRunning); done != 1 || total != 2 {
		t.Fatalf("mid-run on the dependent: %d/%d, want 1/2", done, total)
	}
	if done, total := progressAt(upgradeWithDependent, n, opSuccess); done != 2 || total != 2 {
		t.Fatalf("finished: %d/%d, want 2/2", done, total)
	}
}

// TestUpgradeAllBarNeverFullWhileRunning (fix F-r2-01): at every point of either
// transcript where the run is still going, the bar is not full.
func TestUpgradeAllBarNeverFullWhileRunning(t *testing.T) {
	for name, lines := range map[string][]string{
		"formulae then cask": upgradeFormulaeThenCask,
		"with dependent":     upgradeWithDependent,
	} {
		for n := 1; n < len(lines); n++ {
			done, total := progressAt(lines, n, opRunning)
			if total > 0 && done >= total {
				t.Fatalf("%s: after %d lines the bar reads %d/%d while the run goes on", name, n, done, total)
			}
		}
	}
}

// TestUpgradeAllIgnoresDependencyInstalls: brew installs a dependency on the way
// ("Upgrading zlib dependency: ncurses (6.5)"), which no header names, so it is
// not counted.
func TestUpgradeAllIgnoresDependencyInstalls(t *testing.T) {
	lines := []string{
		"==> Upgrading 1 outdated package:",
		"zlib 1.3 -> 1.3.1",
		"==> Upgrading zlib",
		"==> Upgrading zlib dependency: ncurses (6.5)",
		"==> Pouring ncurses--6.5.arm64_sonoma.bottle.tar.gz",
	}
	if done, total := progressAt(lines, len(lines), opRunning); done != 0 || total != 1 {
		t.Fatalf("a dependency install is not a package: %d/%d, want 0/1", done, total)
	}
}

// operationFrom is an Operation whose lines arrived one at a time, through
// AppendLine, as a run's do. Render and progress read the step state that
// AppendLine keeps, so a fixture built from a Lines literal would have none.
func operationFrom(op Operation, lines []string) *Operation {
	for _, line := range lines {
		op.AppendLine(line)
	}
	return &op
}

// TestOperationStepsMatchParseSteps (cr3 AC-02): the step state kept as lines
// arrive equals parseSteps over the whole run, after every line.
func TestOperationStepsMatchParseSteps(t *testing.T) {
	fixtures := map[string][]string{
		"upgrade transcript":    append(append([]string{}, upgradeTranscript...), ""),
		"formulae then cask":    upgradeFormulaeThenCask,
		"with dependent":        upgradeWithDependent,
		"before the first step": {"Warning: something early", "==> Fetching bash", "  output", "", "==> Pouring bash"},
		"long step output":      append([]string{"==> Pouring bash"}, strings.Split(strings.Repeat("line\n", 9), "\n")...),
		"step with no output":   {"==> Fetching bash", "==> Fetching curl"},
	}
	for name, lines := range fixtures {
		op := &Operation{}
		for n, line := range lines {
			op.AppendLine(line)
			if want := parseSteps(lines[:n+1]); !reflect.DeepEqual(op.steps, want) {
				t.Fatalf("%s: after %d lines the step state is\n%+v\nwant\n%+v", name, n+1, op.steps, want)
			}
		}
	}
}

// TestOperationKeepsOnlyTheLastLinesOfAStep: a step's output is read for its
// last operationTailLines lines only, so the state keeps no more than that.
func TestOperationKeepsOnlyTheLastLinesOfAStep(t *testing.T) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("out %d", i))
	}
	steps := parseSteps(append([]string{"==> Pouring bash"}, lines...))
	if got := steps[0].output; !reflect.DeepEqual(got, []string{"out 98", "out 99"}) {
		t.Fatalf("the step's output = %q, want the last %d lines", got, operationTailLines)
	}
}

// TestOperationLinesAreCapped (cr3 AC-02): the output keeps its latest 5000
// lines. Older lines are dropped from Lines and the progress does not change:
// the header that named the total is dropped, and the total still reads 2000.
func TestOperationLinesAreCapped(t *testing.T) {
	op := &Operation{Status: opRunning, Upgrades: true}
	op.AppendLine("==> Upgrading 2000 outdated packages:")
	for i := 0; i < 2000; i++ {
		op.AppendLine(fmt.Sprintf("==> Upgrading pkg%04d", i))
		op.AppendLine("  1.0 -> 1.1")
		op.AppendLine("  fetched")
	}
	if got := len(op.Lines); got != 5000 {
		t.Fatalf("Lines holds %d lines, want the cap of 5000", got)
	}
	// The run is 6001 lines; the last 5000 start at the first output line of
	// package 333, which is line 1001 of the run.
	if op.Lines[0] != "  1.0 -> 1.1" || op.Lines[len(op.Lines)-1] != "  fetched" {
		t.Fatalf("kept lines run from %q to %q, want the last 5000 lines of the run", op.Lines[0], op.Lines[len(op.Lines)-1])
	}
	if done, total := operationProgress(op); done != 1999 || total != 2000 {
		t.Fatalf("progress after the cap = %d/%d, want 1999/2000", done, total)
	}
}

// parseSteps splits brew's output at its `==> ` lines. Blank lines are not
// output. Lines before the first step belong to no step. It is the step state
// of a whole run; an Operation keeps the same state as its lines arrive.
func parseSteps(lines []string) []opStep {
	var steps []opStep
	for _, line := range lines {
		steps = addStepLine(steps, line)
	}
	return steps
}
