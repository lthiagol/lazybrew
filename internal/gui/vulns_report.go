package gui

import (
	"cmp"
	"fmt"
	"sort"
	"strconv"
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
// A copy (y) draws the same report with nothing wrapped and the full lists in it.
func (m Model) vulnsReportView(r brew.VulnReport, width int) string {
	if !r.Parsed {
		return vulnsTextView(r.Text, width)
	}
	full := m.copying
	affected := affectedPackages(r.Findings)
	var out []string
	if len(affected) == 0 {
		line := presentation.WithIcon(m.icons.OK, fmt.Sprintf("No known vulnerabilities in %d %s checked", r.Checked, plural(r.Checked, "package", "packages")))
		out = append(out, style.SuccessText.Render(line))
	} else {
		out = append(out, vulnsSummary(r, affected, width)...)
		for _, f := range affected {
			out = append(out, m.vulnBlock(f, width, full)...)
		}
	}
	if footer := vulnsFooter(r, width, full); len(footer) > 0 {
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
// how many vulnerabilities it has, the next step, then a table of them. full is a
// copy, where the header is one line and nothing is wrapped.
func (m Model) vulnBlock(f brew.VulnFinding, width int, full bool) []string {
	name := f.Formula
	if f.Version != "" {
		name += " " + f.Version
	}
	worst := severityWord(worstSeverity(f))
	left := style.TextStyle.Bold(true).Render(name) + "  " +
		severityStyle(worst).Render(worst) + style.MutedText.Render(fmt.Sprintf(" ×%d", len(f.Vulnerabilities)))
	step, role := m.vulnNextStep(f)
	head := vulnHeader(left, role.Render(step), width, full)
	return append(head, append(vulnTable(sortedVulns(f.Vulnerabilities), width, full), "")...)
}

// vulnHeader is a block's header: its left part, with the next step at the right
// when the width has room for it. A copy keeps the step on the same line, two
// cells after the name, since the copy has no width to right-align to.
func vulnHeader(left, step string, width int, full bool) []string {
	switch {
	case full:
		return []string{left + "  " + step}
	case lipgloss.Width(left)+2+lipgloss.Width(step) <= width:
		return []string{left + strings.Repeat(" ", width-lipgloss.Width(left)-lipgloss.Width(step)) + step}
	}
	return []string{left, step}
}

// vulnNextStep is what to do about a package, as a short phrase and its role. An
// Outdated formula, when the Outdated list has loaded, is upgraded; else a released
// fix is named (the highest); else the fix is said to be upstream only, or missing.
// A package whose vulnerabilities are all UNKNOWN draws its phrase muted.
func (m Model) vulnNextStep(f brew.VulnFinding) (string, lipgloss.Style) {
	text, role := "no fix yet", style.WarningText
	if fix, ok := highestReleasedFix(f); ok {
		text = "fix in " + fix
	} else if hasFixedVersion(f) {
		text, role = "fixed upstream, not released", style.MutedText
	}
	if old, ok := m.outdatedFormula(f.Formula); ok {
		text, role = "upgrade", style.SuccessText
		if old.NewVersion != "" {
			text += " → " + old.NewVersion
		}
	}
	if allUnknown(f) {
		role = style.MutedText
	}
	return text, role
}

// outdatedFormula is the formula of that name on the Outdated panel. It is not
// found while the panel has not loaded: then there is no claim to make.
func (m Model) outdatedFormula(name string) (brew.Formula, bool) {
	p := m.panels[PanelOutdated]
	if !p.loaded {
		return brew.Formula{}, false
	}
	for _, f := range p.formulae {
		if f.Name == name {
			return f, true
		}
	}
	return brew.Formula{}, false
}

// highestReleasedFix is the highest released fix version among a finding's
// advisories. A commit SHA is not a release and is never named.
func highestReleasedFix(f brew.VulnFinding) (string, bool) {
	best, found := "", false
	for _, v := range f.Vulnerabilities {
		for _, fix := range v.FixedVersions {
			if brew.ReleasedFix(fix) && (!found || versionAfter(fix, best)) {
				best, found = fix, true
			}
		}
	}
	return best, found
}

// hasFixedVersion is whether any advisory of a finding names a fix, released or not.
func hasFixedVersion(f brew.VulnFinding) bool {
	for _, v := range f.Vulnerabilities {
		if len(v.FixedVersions) > 0 {
			return true
		}
	}
	return false
}

// allUnknown is whether every advisory of a finding has an UNKNOWN severity.
func allUnknown(f brew.VulnFinding) bool {
	for _, v := range f.Vulnerabilities {
		if severityWord(v.Severity) != "UNKNOWN" {
			return false
		}
	}
	return true
}

// versionAfter is whether version a is later than b. The part before a dash is
// the release and the part after it a pre-release. Releases compare first; a
// release is later than its own pre-releases. A leading v is ignored.
func versionAfter(a, b string) bool {
	relA, preA, _ := strings.Cut(a, "-")
	relB, preB, _ := strings.Cut(b, "-")
	if c := compareParts(versionParts(relA), versionParts(relB)); c != 0 {
		return c > 0
	}
	if preA == "" || preB == "" {
		return preA == "" && preB != ""
	}
	return compareParts(versionParts(preA), versionParts(preB)) > 0
}

// compareParts orders two split versions: a part missing from the shorter one is
// 0, parts compare as numbers where both are numbers and else as text.
func compareParts(pa, pb []string) int {
	for i := range max(len(pa), len(pb)) {
		a, b := partAt(pa, i), partAt(pb, i)
		if a == b {
			continue
		}
		na, errA := strconv.Atoi(a)
		nb, errB := strconv.Atoi(b)
		if errA == nil && errB == nil {
			return cmp.Compare(na, nb)
		}
		return strings.Compare(a, b)
	}
	return 0
}

// partAt is part i of a split version, or "0" past its end.
func partAt(parts []string, i int) string {
	if i < len(parts) {
		return parts[i]
	}
	return "0"
}

// versionParts splits a version at its dots, dashes, underscores and pluses.
func versionParts(v string) []string {
	return strings.FieldsFunc(strings.TrimPrefix(v, "v"), func(r rune) bool {
		return r == '.' || r == '-' || r == '_' || r == '+'
	})
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
// line only. A copy (full) keeps each summary on one line and adds the fixed
// versions under it.
func vulnTable(vulns []brew.Vulnerability, width int, full bool) []string {
	idW := lipgloss.Width("ID")
	for _, v := range vulns {
		idW = max(idW, lipgloss.Width(v.ID))
	}
	sumW := max(vulnSummaryMin, width-idW-vulnSevW-2*tableGap)
	if full {
		sumW = 0
	}
	gap := strings.Repeat(" ", tableGap)
	blank := strings.Repeat(" ", idW) + gap + strings.Repeat(" ", vulnSevW) + gap

	rows := []string{style.MutedText.Render(padRight("ID", idW) + gap + padRight("Severity", vulnSevW) + gap + "Summary")}
	for _, v := range vulns {
		sev := severityWord(v.Severity)
		for i, line := range wrapWords(v.Summary, sumW) {
			lead := blank
			if i == 0 {
				lead = padRight(v.ID, idW) + gap + severityStyle(sev).Render(padRight(sev, vulnSevW)) + gap
			}
			rows = append(rows, lead+line)
		}
		if full && len(v.FixedVersions) > 0 {
			rows = append(rows, blank+"fixed versions: "+strings.Join(v.FixedVersions, ", "))
		}
	}
	return rows
}

// wrapWords breaks s into lines at most w cells wide, at spaces only. A word is
// kept whole unless it is wider than w, when it is cut across lines. Hyphens do
// not break a word, so "-d" and "AES-GCM-SIV" stay as they are written. A width
// of 0 or less keeps s on one line.
func wrapWords(s string, w int) []string {
	if w <= 0 {
		return []string{strings.Join(strings.Fields(s), " ")}
	}
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
// package, and the packages brew did not scan. A copy names the packages not
// scanned, and the patched IDs under each package that has them; the view counts.
func vulnsFooter(r brew.VulnReport, width int, full bool) []string {
	if full {
		width = 0
	}
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
		if full {
			for _, f := range r.Findings {
				if len(f.Patched) == 0 {
					continue
				}
				ids := make([]string, len(f.Patched))
				for i, v := range f.Patched {
					ids[i] = v.ID
				}
				lines = append(lines, style.MutedText.Render(f.Formula+" patched: "+strings.Join(ids, ", ")))
			}
		}
	}
	if n := len(r.SkippedFormulae); n > 0 {
		text := fmt.Sprintf("%d %s not scanned (no supported source URL)", n, plural(n, "package", "packages"))
		if full {
			text += ": " + strings.Join(r.SkippedFormulae, ", ")
		}
		lines = append(lines, mutedLines(text, width)...)
	}
	return lines
}

// mutedLines is text in the muted role, wrapped to width (one line at 0).
func mutedLines(text string, width int) []string {
	var out []string
	for _, line := range wrapWords(text, width) {
		out = append(out, style.MutedText.Render(line))
	}
	return out
}
