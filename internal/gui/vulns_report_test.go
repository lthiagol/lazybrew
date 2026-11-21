package gui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// fixtureVulns is brew's real JSON report, saved in the brew package's testdata:
// 7 findings, 24 vulnerabilities, 14 patched (libssh2), 90 checked, 37 skipped.
func fixtureVulns(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../brew/testdata/vulns_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// vulnsExit1 is brew's exit when it found vulnerabilities and printed a report.
var vulnsExit1 = &brew.BrewExitError{Command: "vulns", ExitCode: 1}

// checkedVulns runs the Vulns check on a loaded Status panel of size w x h, with
// brew answering out and err, and returns the model after the report lands.
func checkedVulns(t *testing.T, w, h int, out string, err error) *Model {
	t.Helper()
	calls := 0
	m := loadedStatus(w, h)
	m.client = brew.NewClient(vulnsRunner(out, err, &calls))
	return pressStatus(m, "v")
}

// tallVulns is the saved report on a terminal tall enough to show every block
// without scrolling.
func tallVulns(t *testing.T, w int) *Model {
	t.Helper()
	return checkedVulns(t, w, 120, fixtureVulns(t), vulnsExit1)
}

// vulnLine is the first line of view that holds s, or fails the test.
func vulnLine(t *testing.T, view, s string) string {
	t.Helper()
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, s) {
			return line
		}
	}
	t.Fatalf("no line holds %q:\n%s", s, view)
	return ""
}

// TestVulnsReportSummaryAndChips: the summary names the counts, the not-scanned
// count sits beside it, and the chips list only the severities that occur, worst
// first.
func TestVulnsReportSummaryAndChips(t *testing.T) {
	m := tallVulns(t, 140)
	out := ansi.Strip(m.View().Content)

	if line := vulnLine(t, out, "24 vulnerabilities"); !strings.Contains(line, "in 7 of 90 packages") {
		t.Errorf("summary line %q", line)
	}
	if line := vulnLine(t, out, "37 not scanned"); strings.Contains(line, "vulnerabilities") {
		t.Errorf("the not-scanned count should be its own muted line, got %q", line)
	}
	if line := vulnLine(t, out, "HIGH 8"); !strings.Contains(line, "HIGH 8  MEDIUM 5  UNKNOWN 11") {
		t.Errorf("chips line %q, want only the severities present, worst first", line)
	}
	if strings.Contains(out, "CRITICAL") || strings.Contains(out, "LOW ") {
		t.Errorf("a severity that is not in the report has a chip:\n%s", out)
	}
}

// TestVulnsReportPackageOrder: packages sort by worst severity, then count, then
// name: libsndfile and openssl@4 (HIGH, 6 each, by name), libxml2 (HIGH, 2), then
// the UNKNOWN packages by count and name.
func TestVulnsReportPackageOrder(t *testing.T) {
	m := tallVulns(t, 140)
	out := ansi.Strip(m.View().Content)

	order := []string{"libsndfile 1.2.2", "openssl@4 4.1.0", "libxml2 2.15.4", "ffmpeg 9.0.2", "expat 2.10.0", "gitleaks 8.30.1", "libssh2 1.11.1"}
	last := -1
	for _, name := range order {
		at := strings.Index(out, name)
		if at < 0 {
			t.Fatalf("no block for %q:\n%s", name, out)
		}
		if at < last {
			t.Errorf("%q is out of order", name)
		}
		last = at
	}
}

// TestVulnsReportTableRows: a block is an ID, a severity and a summary per row,
// worst first and then by ID.
func TestVulnsReportTableRows(t *testing.T) {
	m := tallVulns(t, 140)
	out := ansi.Strip(m.View().Content)

	row := vulnLine(t, out, "OSV-2021-777")
	if !strings.Contains(row, "HIGH") || !strings.Contains(row, "Heap-use-after-free in xmlAddNextSibling") {
		t.Errorf("row %q should hold the severity and the whole summary", row)
	}
	first := strings.Index(out, "CVE-2026-124709")
	mid := strings.Index(out, "CVE-2026-35190")
	last := strings.Index(out, "CVE-2026-88822")
	if !(first < mid && mid < last) {
		t.Errorf("openssl@4 rows out of order: HIGH %d, MEDIUM %d, MEDIUM %d", first, mid, last)
	}
}

// TestVulnsReportWrapsSummaries: at 80 columns the whole report draws no line
// wider than the width, no summary is cut, and every word of every summary is
// whole.
func TestVulnsReportWrapsSummaries(t *testing.T) {
	m := checkedVulns(t, 80, 24, fixtureVulns(t), vulnsExit1)
	report := m.vulnsReportView(fixtureReport(t), 80)
	for _, line := range strings.Split(report, "\n") {
		if w := lipgloss.Width(line); w > 80 {
			t.Errorf("a line is %d cells wide: %q", w, ansi.Strip(line))
		}
	}
	plain := ansi.Strip(report)
	if strings.Contains(plain, "…") {
		t.Errorf("a summary or path is cut with …:\n%s", plain)
	}
	tokens := map[string]bool{}
	for _, tok := range strings.Fields(plain) {
		tokens[tok] = true
	}
	for _, f := range fixtureReport(t).Findings {
		for _, v := range f.Vulnerabilities {
			for _, word := range strings.Fields(v.Summary) {
				if !tokens[word] {
					t.Errorf("%s: the word %q is not whole in the report", v.ID, word)
				}
			}
		}
	}
}

// fixtureReport is the saved report as the brew package reads it.
func fixtureReport(t *testing.T) brew.VulnReport {
	t.Helper()
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(fixtureVulns(t)), vulnsExit1
	}
	report, err := brew.NewDiagnosticsReader(r, brew.NewCache(0)).VulnsReport(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return report
}

// TestVulnsReportFrames: the report fits the terminal in both sizes and both
// themes, and so does the no-findings report.
func TestVulnsReportFrames(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	empty := fmt.Sprintf(`{"findings":[],"skipped_formulae":%s,"checked":90,"complete":false,"outdated_without_sbom":[],"unscanned_formulae":{}}`, skippedJSON(37))
	for _, size := range []struct{ w, h int }{{80, 24}, {140, 40}} {
		for theme, th := range map[string]*style.Theme{"dark": style.DarkTheme(), "light": style.LightTheme()} {
			style.ApplyTheme(th)
			for name, out := range map[string]string{"report": fixtureVulns(t), "no findings": empty} {
				t.Run(fmt.Sprintf("%s/%s/%dx%d", name, theme, size.w, size.h), func(t *testing.T) {
					m := checkedVulns(t, size.w, size.h, out, vulnsExit1)
					assertFrameSize(t, m.View().Content, size.w, size.h)
				})
			}
		}
	}
}

// skippedJSON is a JSON array of n formula names.
func skippedJSON(n int) string {
	names := make([]string, n)
	for i := range names {
		names[i] = fmt.Sprintf("%q", fmt.Sprintf("pkg%02d", i))
	}
	return "[" + strings.Join(names, ",") + "]"
}

// TestVulnsReportNoFindings: a report with nothing found says so with the number
// of packages checked, and the not-scanned count under it.
func TestVulnsReportNoFindings(t *testing.T) {
	out := fmt.Sprintf(`{"findings":[],"skipped_formulae":%s,"checked":90,"complete":false}`, skippedJSON(37))
	m := checkedVulns(t, 140, 40, out, nil)
	plain := ansi.Strip(m.View().Content)

	if !strings.Contains(plain, "✔ No known vulnerabilities in 90 packages checked") {
		t.Errorf("no-findings line missing:\n%s", plain)
	}
	if !strings.Contains(plain, "37 packages not scanned (no supported source URL)") {
		t.Errorf("not-scanned footer missing:\n%s", plain)
	}
	if strings.Contains(plain, "of 90 packages") {
		t.Errorf("a clean report should not have a summary:\n%s", plain)
	}
}

// TestVulnsReportFooter: the footer counts the patched findings per package, and
// the not-scanned packages are counted but not named.
func TestVulnsReportFooter(t *testing.T) {
	m := tallVulns(t, 140)
	plain := ansi.Strip(m.View().Content)

	if !strings.Contains(plain, "14 resolved by formula patches (libssh2 14)") {
		t.Errorf("patched footer missing:\n%s", plain)
	}
	if !strings.Contains(plain, "37 packages not scanned (no supported source URL)") {
		t.Errorf("not-scanned footer missing:\n%s", plain)
	}
	// The fixture's skipped names are for the y copy only.
	for _, name := range []string{"glib", "libedit", "mpg123"} {
		if strings.Contains(plain, name) {
			t.Errorf("the skipped package %q is named in the view; it belongs in the copy only", name)
		}
	}
}

// TestVulnsReportColours: each severity is drawn in its role: CRITICAL bold error,
// HIGH error, MEDIUM warning, LOW info, UNKNOWN muted, in the chips and the blocks.
func TestVulnsReportColours(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	out := `{"findings":[{"formula":"a","version":"1","vulnerabilities":[
		{"id":"C-1","severity":"CRITICAL","summary":"s"},
		{"id":"L-1","severity":"LOW","summary":"s"},
		{"id":"M-1","severity":"MEDIUM","summary":"s"},
		{"id":"U-1","severity":"UNKNOWN","summary":"s"}],"patched":[]}],
		"skipped_formulae":[],"checked":1,"complete":true}`
	m := checkedVulns(t, 140, 40, out, nil)
	view := m.View().Content

	want := map[string]lipgloss.Style{
		"CRITICAL 1": style.ErrorText.Bold(true),
		"MEDIUM 1":   style.WarningText,
		"LOW 1":      lipgloss.NewStyle().Foreground(style.InfoColor),
		"UNKNOWN 1":  style.MutedText,
	}
	for text, st := range want {
		if !strings.Contains(view, st.Render(text)) {
			t.Errorf("%q is not drawn in its role colour", text)
		}
	}
	if !strings.Contains(view, style.ErrorText.Bold(true).Render("CRITICAL")) {
		t.Error("the block's worst severity is not drawn in the CRITICAL role")
	}
}

// TestVulnsReportFallbackNote: output that is not brew's JSON is shown as its
// text, wrapped, under a note that brew's JSON could not be read.
func TestVulnsReportFallbackNote(t *testing.T) {
	text := "CVE-2026-1: openssl@3 " + strings.Repeat("affected-advisory ", 5) + "end\n"
	m := checkedVulns(t, 80, 24, text, nil)
	view := m.View().Content
	assertFrameSize(t, view, 80, 24)
	plain := ansi.Strip(view)

	if !strings.Contains(plain, "Could not read brew's JSON; showing its text.") {
		t.Errorf("the fallback note is missing:\n%s", plain)
	}
	if !strings.Contains(plain, "CVE-2026-1: openssl@3") {
		t.Errorf("brew's text is missing:\n%s", plain)
	}
}

// TestVulnsReportScrolls: a report longer than the tab can be scrolled to its
// footer with G, as the other tabs scroll.
func TestVulnsReportScrolls(t *testing.T) {
	m := checkedVulns(t, 80, 24, fixtureVulns(t), vulnsExit1)
	if strings.Contains(ansi.Strip(m.View().Content), "resolved by formula patches") {
		t.Fatal("the footer should start below the fold at 80x24")
	}
	m = pressStatus(m, "G")
	if !strings.Contains(ansi.Strip(m.View().Content), "resolved by formula patches") {
		t.Errorf("G did not reach the footer:\n%s", ansi.Strip(m.View().Content))
	}
}
