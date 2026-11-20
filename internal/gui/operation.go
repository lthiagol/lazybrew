package gui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// operationTailLines is how many output lines a step shows under its title.
const operationTailLines = 2

// operationBarWidth is the widest the batch progress bar draws.
const operationBarWidth = 40

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
	if op.BatchTotal > 0 {
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

// operationBar is the batch's progress: a bar, then n/total.
func (m Model) operationBar(op *Operation, width int) string {
	done := min(max(0, m.operationDone(op)), op.BatchTotal)
	bar := progress.New(
		progress.WithColors(style.AccentColor),
		progress.WithFillCharacters('━', '╌'),
		progress.WithoutPercentage(),
		progress.WithWidth(max(10, min(operationBarWidth, width-12))),
	)
	return bar.ViewAs(float64(done)/float64(op.BatchTotal)) + " " +
		style.SubtleText.Render(fmt.Sprintf("%d/%d", done, op.BatchTotal))
}

// operationDone is how many of a batch's packages are done. An upgrade of
// everything counts the packages brew has begun upgrading, less the one still
// in progress.
func (m Model) operationDone(op *Operation) int {
	if !op.Upgrades {
		return op.BatchDone
	}
	begun := 0
	for _, s := range parseSteps(op.Lines) {
		if strings.HasPrefix(s.title, "Upgrading ") {
			begun++
		}
	}
	if op.Status == opRunning {
		begun--
	}
	return begun
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
