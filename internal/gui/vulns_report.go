package gui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/presentation"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// vulnSeverities is the order the summary chips list the severities in, worst first.
var vulnSeverities = []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"}

// vulnSevW is the width of the severity column: the widest severity word.
const vulnSevW = 8

// vulnSummaryMin is the narrowest a summary column is kept at, however wide the
// IDs are, so a summary wraps rather than being squeezed out.
const vulnSummaryMin = 20

// vulnsReportView is a finished check: brew's text when its JSON could not be
// read, else the report. A report with nothing found says so; otherwise it is the
// summary, the severity chips, one block per affected package, and the footer.
func (m Model) vulnsReportView(r brew.VulnReport, width int) string {
	if !r.Parsed {
		return vulnsTextView(r.Text, width)
	}
	affected := affectedPackages(r.Findings)
	var out []string
	if len(affected) == 0 {
		line := presentation.WithIcon(m.icons.OK, fmt.Sprintf("No known vulnerabilities in %d %s checked", r.Checked, plural(r.Checked, "package", "packages")))
		out = append(out, style.SuccessText.Render(line))
	} else {
		out = append(out, vulnsSummary(r, affected, width)...)
		for _, f := range affected {
			out = append(out, vulnBlock(f, width)...)
		}
	}
	if footer := vulnsFooter(r, width); len(footer) > 0 {
		out = append(out, "")
		out = append(out, footer...)
	}
	return strings.Join(out, "\n")
}

// vulnsTextView is brew's output when it is not the JSON report: the text, wrapped,
// under a note that the JSON could not be read. Empty output is the empty line.
func vulnsTextView(text string, width int) string {
	if strings.TrimSpace(text) == "" {
		return style.SuccessText.Render("No known vulnerabilities")
	}
	note := style.SubtleText.Render("Could not read brew's JSON; showing its text.")
	return note + "\n" + strings.Join(wrapContent(strings.TrimRight(text, "\n"), width), "\n")
}

// affectedPackages are the findings with at least one vulnerability, worst first:
// by the worst severity they hold, then by how many vulnerabilities, then by name.
func affectedPackages(findings []brew.VulnFinding) []brew.VulnFinding {
	var out []brew.VulnFinding
	for _, f := range findings {
		if len(f.Vulnerabilities) > 0 {
			out = append(out, f)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if wa, wb := worstRank(a), worstRank(b); wa != wb {
			return wa > wb
		}
		if len(a.Vulnerabilities) != len(b.Vulnerabilities) {
			return len(a.Vulnerabilities) > len(b.Vulnerabilities)
		}
		return a.Formula < b.Formula
	})
	return out
}

// worstRank is the highest severity rank among a finding's vulnerabilities.
func worstRank(f brew.VulnFinding) int {
	worst := 0
	for _, v := range f.Vulnerabilities {
		worst = max(worst, brew.SeverityRank(v.Severity))
	}
	return worst
}

// severityWord is a brew severity as one of vulnSeverities; any other word is
// UNKNOWN.
func severityWord(s string) string {
	u := strings.ToUpper(s)
	for _, w := range vulnSeverities {
		if u == w {
			return w
		}
	}
	return "UNKNOWN"
}

// severityStyle is the role a severity is drawn in: CRITICAL is bold error, HIGH
// error, MEDIUM warning, LOW info and UNKNOWN muted.
func severityStyle(sev string) lipgloss.Style {
	switch sev {
	case "CRITICAL":
		return style.ErrorText.Bold(true)
	case "HIGH":
		return style.ErrorText
	case "MEDIUM":
		return style.WarningText
	case "LOW":
		return lipgloss.NewStyle().Foreground(style.InfoColor)
	}
	return style.MutedText
}

// vulnsSummary is the headline: the vulnerability count, the packages it is in
// out of those checked, and the muted counts of what was not scanned. The chips
// follow, one per severity that occurs.
func vulnsSummary(r brew.VulnReport, affected []brew.VulnFinding, width int) []string {
	total := 0
	counts := map[string]int{}
	for _, f := range affected {
		total += len(f.Vulnerabilities)
		for _, v := range f.Vulnerabilities {
			counts[severityWord(v.Severity)]++
		}
	}
	headline := style.TextStyle.Bold(true).Render(fmt.Sprintf("%d %s in %d of %d %s",
		total, plural(total, "vulnerability", "vulnerabilities"),
		len(affected), r.Checked, plural(r.Checked, "package", "packages")))
	notScanned := []string{fmt.Sprintf("%d not scanned", len(r.SkippedFormulae))}
	if n := len(r.OutdatedWithoutSBOM); n > 0 {
		notScanned = append(notScanned, fmt.Sprintf("%d outdated without SBOM", n))
	}
	if n := len(r.UnscannedFormulae); n > 0 {
		notScanned = append(notScanned, fmt.Sprintf("%d unscanned %s", n, plural(n, "formula", "formulae")))
	}
	lines := append(wrapContent(headline, width), wrapContent(style.MutedText.Render(strings.Join(notScanned, " · ")), width)...)

	var chips []string
	for _, sev := range vulnSeverities {
		if n := counts[sev]; n > 0 {
			chips = append(chips, severityStyle(sev).Render(fmt.Sprintf("%s %d", sev, n)))
		}
	}
	return append(lines, strings.Join(chips, "  "), "")
}

// vulnBlock is one affected package: its name and version, its worst severity and
// how many vulnerabilities it has, then a table of them.
func vulnBlock(f brew.VulnFinding, width int) []string {
	name := f.Formula
	if f.Version != "" {
		name += " " + f.Version
	}
	worst := severityWord(worstSeverity(f))
	head := style.TextStyle.Bold(true).Render(name) + "  " +
		severityStyle(worst).Render(worst) + style.MutedText.Render(fmt.Sprintf(" ×%d", len(f.Vulnerabilities)))
	return append([]string{head}, append(vulnTable(sortedVulns(f.Vulnerabilities), width), "")...)
}

// worstSeverity is the severity of the worst of a finding's vulnerabilities.
func worstSeverity(f brew.VulnFinding) string {
	worst := ""
	for _, v := range f.Vulnerabilities {
		if brew.SeverityRank(v.Severity) > brew.SeverityRank(worst) {
			worst = v.Severity
		}
	}
	return worst
}

// sortedVulns is a copy of vulns, worst first and then by ID.
func sortedVulns(vulns []brew.Vulnerability) []brew.Vulnerability {
	out := append([]brew.Vulnerability(nil), vulns...)
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := brew.SeverityRank(out[i].Severity), brew.SeverityRank(out[j].Severity)
		if ri != rj {
			return ri > rj
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// vulnTable is the ID, Severity and Summary of each vulnerability. The summary is
// wrapped to the width left, so no word is cut; the other cells sit on its first
// line only.
func vulnTable(vulns []brew.Vulnerability, width int) []string {
	idW := lipgloss.Width("ID")
	for _, v := range vulns {
		idW = max(idW, lipgloss.Width(v.ID))
	}
	sumW := max(vulnSummaryMin, width-idW-vulnSevW-2*tableGap)
	gap := strings.Repeat(" ", tableGap)

	rows := []string{style.MutedText.Render(padRight("ID", idW) + gap + padRight("Severity", vulnSevW) + gap + "Summary")}
	blank := strings.Repeat(" ", idW) + gap + strings.Repeat(" ", vulnSevW) + gap
	for _, v := range vulns {
		sev := severityWord(v.Severity)
		for i, line := range wrapWords(v.Summary, sumW) {
			lead := blank
			if i == 0 {
				lead = padRight(v.ID, idW) + gap + severityStyle(sev).Render(padRight(sev, vulnSevW)) + gap
			}
			rows = append(rows, lead+line)
		}
	}
	return rows
}

// wrapWords breaks s into lines at most w cells wide, at spaces only. A word is
// kept whole unless it is wider than w, when it is cut across lines. Hyphens do
// not break a word, so "-d" and "AES-GCM-SIV" stay as they are written.
func wrapWords(s string, w int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		for lipgloss.Width(word) > w {
			head := ansi.Truncate(word, w, "")
			if head == "" {
				break
			}
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			lines = append(lines, head)
			word = word[len(head):]
		}
		switch {
		case line == "":
			line = word
		case lipgloss.Width(line)+1+lipgloss.Width(word) <= w:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	return append(lines, line)
}

// vulnsFooter is the muted footer: the findings the formula patches resolve, per
// package, and the packages brew did not scan. The names of the packages not
// scanned are in the copy only.
func vulnsFooter(r brew.VulnReport, width int) []string {
	var lines []string
	var parts []string
	patched := 0
	for _, f := range r.Findings {
		if n := len(f.Patched); n > 0 {
			patched += n
			parts = append(parts, fmt.Sprintf("%s %d", f.Formula, n))
		}
	}
	if patched > 0 {
		text := fmt.Sprintf("%d resolved by formula %s (%s)", patched, plural(patched, "patch", "patches"), strings.Join(parts, ", "))
		lines = append(lines, mutedLines(text, width)...)
	}
	if n := len(r.SkippedFormulae); n > 0 {
		text := fmt.Sprintf("%d %s not scanned (no supported source URL)", n, plural(n, "package", "packages"))
		lines = append(lines, mutedLines(text, width)...)
	}
	return lines
}

// mutedLines is text in the muted role, wrapped to width.
func mutedLines(text string, width int) []string {
	var out []string
	for _, line := range wrapContent(text, width) {
		out = append(out, style.MutedText.Render(strings.TrimRight(line, " ")))
	}
	return out
}
