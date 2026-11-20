package gui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/anim"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// operationTailLines is how many output lines a step shows under its title.
const operationTailLines = 2

// operationBarWidth is the widest the batch progress bar draws.
const operationBarWidth = 40

// operationAutoClose is how long a successful operation stays on screen before
// it closes itself. The toast has already said it finished.
const operationAutoClose = 2 * time.Second

// autoCloseCmd starts the timer that closes a successful operation. A batch
// with tasks still queued or running gets none: the last task's success starts
// the timer.
func (m Model) autoCloseCmd() tea.Cmd {
	if m.opState == nil || m.opState.Status != opSuccess || m.batchCount > 0 {
		return nil
	}
	op := m.opState
	return tea.Tick(operationAutoClose, func(time.Time) tea.Msg {
		return opAutoCloseMsg{op: op}
	})
}

// opStep is one `==> ` step of a brew run and the output lines after it.
type opStep struct {
	title  string
	output []string
}

// parseSteps splits brew's output at its `==> ` lines. Blank lines are not
// output. Lines before the first step belong to no step.
func parseSteps(lines []string) []opStep {
	var steps []opStep
	for _, line := range lines {
		if rest, ok := strings.CutPrefix(line, "==> "); ok {
			steps = append(steps, opStep{title: strings.TrimSpace(rest)})
			continue
		}
		if len(steps) == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		last := &steps[len(steps)-1]
		last.output = append(last.output, line)
	}
	return steps
}

// formatElapsed reads a duration as the Operation view shows it: 12s, 1m04s.
func formatElapsed(d time.Duration) string {
	secs := int(max(0, d).Truncate(time.Second) / time.Second)
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	return fmt.Sprintf("%dm%02ds", secs/60, secs%60)
}

// actionIcon is the glyph of the action a task runs, read from its ID.
func (m Model) actionIcon(id string) string {
	switch {
	case strings.HasPrefix(id, "upgrade"), strings.HasPrefix(id, "batch-upgrade:"):
		return m.icons.Upgradable
	case strings.HasPrefix(id, "uninstall"), strings.HasPrefix(id, "zap:"):
		return m.icons.Stopped
	default:
		return m.icons.Refresh
	}
}

// renderOperation is the Operation view: the action and the package, the
// elapsed time or the final state, the batch's progress when there is one, then
// the steps of the run with the last lines of each. When the room runs out the
// steps are cut from the top, so the latest output stays on screen.
func (m Model) renderOperation(width, height int) string {
	op := m.opState
	if op == nil {
		return ""
	}
	icon := op.Icon
	if icon == "" {
		icon = m.icons.Refresh
	}
	head := []string{
		style.AccentText.Render(withIcon(icon, op.Title)),
		m.operationStatus(op),
	}
	if _, total := operationProgress(op); total > 0 {
		head = append(head, m.operationBar(op, width))
	}
	head = append(head, "")

	body := m.operationBody(op)
	if room := max(0, height-len(head)); len(body) > room {
		body = body[len(body)-room:]
	}
	return cutLines(strings.Join(append(head, body...), "\n"), width)
}

// operationStatus is the row under the title: the elapsed time while the run
// goes on, then the final state.
func (m Model) operationStatus(op *Operation) string {
	switch op.Status {
	case opSuccess:
		return style.InstalledBadge.Render(withIcon(m.icons.OK,
			"Completed in "+formatElapsed(op.Ended.Sub(op.Started))))
	case opError:
		return style.ErrorBadge.Render(withIcon(m.icons.Error, "Failed: "+op.Err.Error()))
	case opCancelled:
		return style.SubtleText.Render("Cancelled")
	default:
		return style.SubtleText.Render("Running " + formatElapsed(time.Since(op.Started)) + " · Esc to cancel")
	}
}

// newProgressBar is the batch's progress bar with no width or position yet.
// The Model keeps one, so its fill can spring from where it was.
func newProgressBar() progress.Model {
	return progress.New(
		progress.WithColors(style.AccentColor),
		progress.WithFillCharacters('━', '╌'),
		progress.WithoutPercentage(),
		progress.WithSpringOptions(anim.Frequency, anim.Damping),
	)
}

// operationBar is the batch's progress: a bar, then n/total. The bar draws
// the fill it is at: while it is moving that is its spring's position, and
// otherwise the batch's share, exactly. The count is always exact.
func (m Model) operationBar(op *Operation, width int) string {
	done, total := operationProgress(op)
	bar := m.opBar
	bar.SetWidth(max(10, min(operationBarWidth, width-12)))
	view := bar.ViewAs(float64(done) / float64(total))
	if m.motionOn && m.opBarOp == op && bar.IsAnimating() {
		view = bar.View()
	}
	return view + " " + style.SubtleText.Render(fmt.Sprintf("%d/%d", done, total))
}

// operationProgress is how many of a batch's packages are done, and how many
// there are. A batch reads its fixed counts. An upgrade of everything reads
// them from brew's output: the total is the N of every header that names the
// packages brew will upgrade (one per formula or cask upgrade, and one for the
// dependents of upgraded formulae; pinned packages are skipped and not named).
// A package is done once its own step has run. While the run goes on, the
// latest package step is taken as in progress, so the bar is never full before
// the run has finished. Until a header comes, the total is 0 and no bar shows.
func operationProgress(op *Operation) (done, total int) {
	if !op.Upgrades {
		return op.BatchDone, op.BatchTotal
	}
	begun := 0
	for _, s := range parseSteps(op.Lines) {
		if n, ok := upgradeHeaderCount(s.title); ok {
			total += n
		} else if isPackageUpgrade(s.title) {
			begun++
		}
	}
	switch {
	case op.Status == opSuccess:
		done = total
	case begun > 0:
		// The package in progress, or the one that failed, is not done.
		done = begun - 1
	}
	return min(max(0, done), total), total
}

// upgradeHeaderCount reads a header of brew's upgrade output that names packages
// to upgrade, and its N: "Upgrading N outdated package(s):" (formulae, then casks)
// and "Upgrading N dependent(s) of upgraded formula(e):".
func upgradeHeaderCount(title string) (int, bool) {
	if !strings.HasPrefix(title, "Upgrading ") || !isUpgradeHeader(title) {
		return 0, false
	}
	fields := strings.Fields(title)
	if len(fields) < 2 {
		return 0, false
	}
	n, err := strconv.Atoi(fields[1])
	return n, err == nil
}

// isUpgradeHeader is one of the headers upgradeHeaderCount reads.
func isUpgradeHeader(title string) bool {
	return strings.Contains(title, "outdated package") ||
		(strings.Contains(title, " dependent") && strings.Contains(title, "of upgraded"))
}

// isPackageUpgrade is a step that upgrades one package, "Upgrading bash" or
// "Upgrading firefox". A header is not one, and neither is a dependency that
// brew installs on the way, "Upgrading zsh dependency: ncurses (6.5)": that
// package is not in any header's count.
func isPackageUpgrade(title string) bool {
	return strings.HasPrefix(title, "Upgrading ") && !isUpgradeHeader(title) &&
		!strings.Contains(title, " dependency: ")
}

// operationBody is the step list of the run: each step with its mark, and the
// last lines of its output in the Muted role under it. With no steps, the raw
// output shows as it came.
func (m Model) operationBody(op *Operation) []string {
	steps := parseSteps(op.Lines)
	if len(steps) == 0 {
		var rows []string
		for _, line := range op.Lines {
			if strings.TrimSpace(line) != "" {
				rows = append(rows, style.SubtleText.Render(line))
			}
		}
		return rows
	}
	muted := lipgloss.NewStyle().Foreground(style.MutedColor)
	var rows []string
	for i, s := range steps {
		rows = append(rows, m.stepMark(op, i == len(steps)-1)+" "+style.NormalItem.Render(s.title))
		out := s.output
		if len(out) > operationTailLines {
			out = out[len(out)-operationTailLines:]
		}
		for _, line := range out {
			rows = append(rows, "    "+muted.Render(line))
		}
	}
	return rows
}

// stepMark is the mark of a step: a finished step is ✓; the last step takes
// the state of the run, the spinner while it runs, ✗ on failure, ■ when
// cancelled.
func (m Model) stepMark(op *Operation, last bool) string {
	switch {
	case !last:
		return lipgloss.NewStyle().Foreground(style.SuccessColor).Render(m.icons.Installed)
	case op.Status == opRunning:
		return m.spinner.View()
	case op.Status == opError:
		return lipgloss.NewStyle().Foreground(style.ErrorColor).Render(m.icons.Error)
	case op.Status == opCancelled:
		return style.SubtleText.Render(m.icons.Stopped)
	default:
		return lipgloss.NewStyle().Foreground(style.SuccessColor).Render(m.icons.Installed)
	}
}
