package gui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// brewVulnsTimeout bounds `brew vulns`. It queries OSV.dev over the network for
// every installed formula, so it is far slower than the local reads and gets its
// own budget rather than brewReadTimeout.
var brewVulnsTimeout = 2 * time.Minute

// vulnsHint is the line the Vulns tab shows until the check has run, and again
// after a failed one.
const vulnsHint = "Press v to check installed formulae for known vulnerabilities (queries OSV.dev)"

// vulnsStatus is where the Vulns tab's check stands.
type vulnsStatus int

const (
	vulnsNotRun vulnsStatus = iota
	vulnsRunning
	vulnsDone
	vulnsFailed
)

// vulnsState is the Vulns tab's check: its status, the report of the last run
// that finished, and the error of the last run that failed.
type vulnsState struct {
	status vulnsStatus
	report brew.VulnReport
	err    error
}

// checkVulns opens the Vulns tab and starts the check. A check already running
// is left to finish: a second v only shows the tab.
func (m Model) checkVulns() (tea.Model, tea.Cmd) {
	m.activeTab = statusVulnsTab
	if m.vulns.status == vulnsRunning {
		return m, nil
	}
	m.vulns = vulnsState{status: vulnsRunning}
	return m, vulnsCmd(m.client)
}

// vulnsCmd runs brew vulns as a bounded read and returns its VulnsMsg. It is a
// read, so it does not go through the task manager.
func vulnsCmd(client *brew.Client) tea.Cmd {
	return func() tea.Msg {
		report, err := boundedRead(brewVulnsTimeout, "vulns", client.Diagnostics.VulnsReport)
		return VulnsMsg{Report: report, Err: err}
	}
}

// vulnsTabView is the Vulns tab: the hint until a check runs, the spinner while
// it runs, then the report or the empty line, or brew's error line and the hint.
func (m Model) vulnsTabView() string {
	switch m.vulns.status {
	case vulnsRunning:
		return m.spinner.View() + style.SubtleText.Render(" Checking…")
	case vulnsDone:
		return vulnsReportView(m.vulns.report)
	case vulnsFailed:
		return style.ErrorText.Render(vulnsErrorLine(m.vulns.err)) + "\n" + style.SubtleText.Render(vulnsHint)
	}
	return style.SubtleText.Render(vulnsHint)
}

// vulnsReportView is a finished check: brew's text when its JSON could not be
// read, the empty line when brew found nothing, else the summary line.
func vulnsReportView(r brew.VulnReport) string {
	if !r.Parsed {
		if strings.TrimSpace(r.Text) == "" {
			return style.SuccessText.Render("No known vulnerabilities")
		}
		return strings.TrimRight(r.Text, "\n")
	}
	n := 0
	for _, f := range r.Findings {
		n += len(f.Vulnerabilities)
	}
	return fmt.Sprintf("%d vulnerabilities in %d of %d packages", n, len(r.Findings), r.Checked)
}

// vulnsErrorLine is brew's own first error line, or the error's text when brew
// wrote no stderr.
func vulnsErrorLine(err error) string {
	var exit *brew.BrewExitError
	if errors.As(err, &exit) {
		if line := brewMessage(exit.Stderr); line != "" {
			return line
		}
	}
	return err.Error()
}
