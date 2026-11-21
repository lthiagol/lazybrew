package gui

import (
	"encoding/json"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// sha is a 40-hex commit id, the fixed version brew gives for an unreleased fix.
const sha = "ead311df1037e7994a6d86c7e862d71ad59c52c7"

// reportJSON is brew's JSON report with one finding: formula at version, with the
// given vulnerabilities.
func reportJSON(t *testing.T, formula, version string, vulns ...brew.Vulnerability) string {
	t.Helper()
	r := brew.VulnReport{
		Findings: []brew.VulnFinding{{Formula: formula, Version: version, Vulnerabilities: vulns}},
		Checked:  5,
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// vuln is one advisory of the given severity with the given fixed versions.
func vuln(id, severity string, fixed ...string) brew.Vulnerability {
	return brew.Vulnerability{ID: id, Severity: severity, Summary: "summary of " + id, FixedVersions: fixed}
}

// outdatedLoaded puts formula, outdated to version, on the Outdated panel, loaded
// or not.
func outdatedLoaded(m *Model, formula, version string, loaded bool) {
	m.panels[PanelOutdated].formulae = []brew.Formula{{Name: formula, Version: "1.0", NewVersion: version, Outdated: true}}
	m.panels[PanelOutdated].loaded = loaded
}

// TestVulnsNextStepUpgrade: an Outdated formula with a loaded Outdated list is
// upgraded, and the next step says so in place of the fix.
func TestVulnsNextStepUpgrade(t *testing.T) {
	out := reportJSON(t, "pkg", "1.0", vuln("CVE-1", "HIGH", "1.5"))
	calls := 0
	m := loadedStatus(140, 40)
	outdatedLoaded(m, "pkg", "2.0", true)
	m.client = brew.NewClient(vulnsRunner(out, nil, &calls))
	m = pressStatus(m, "v")

	line := vulnLine(t, ansi.Strip(m.View().Content), "pkg 1.0")
	if !strings.Contains(line, "upgrade → 2.0") {
		t.Errorf("the block should say upgrade → 2.0: %q", line)
	}
	if strings.Contains(line, "fix in") {
		t.Errorf("an outdated package should not also claim a fix: %q", line)
	}
}

// TestVulnsNextStepOutdatedNotLoaded: while the Outdated list has not loaded, the
// next step makes no upgrade claim; it names the fix.
func TestVulnsNextStepOutdatedNotLoaded(t *testing.T) {
	out := reportJSON(t, "pkg", "1.0", vuln("CVE-1", "HIGH", "1.5"))
	calls := 0
	m := loadedStatus(140, 40)
	outdatedLoaded(m, "pkg", "2.0", false)
	m.client = brew.NewClient(vulnsRunner(out, nil, &calls))
	m = pressStatus(m, "v")

	line := vulnLine(t, ansi.Strip(m.View().Content), "pkg 1.0")
	if strings.Contains(line, "upgrade") {
		t.Errorf("the Outdated list is not loaded, so no upgrade is claimed: %q", line)
	}
	if !strings.Contains(line, "fix in 1.5") {
		t.Errorf("the fix should be named: %q", line)
	}
}

// TestVulnsNextStepRoles: each phrase is in its role, and a package whose
// vulnerabilities are all UNKNOWN draws its phrase muted.
func TestVulnsNextStepRoles(t *testing.T) {
	defer style.ApplyTheme(style.DarkTheme())
	cases := []struct {
		name     string
		vulns    []brew.Vulnerability
		phrase   string
		role     lipgloss.Style
		outdated bool
	}{
		{"released fix", []brew.Vulnerability{vuln("A", "HIGH", "1.5")}, "fix in 1.5", style.WarningText, false},
		{"commit only", []brew.Vulnerability{vuln("A", "HIGH", sha)}, "fixed upstream, not released", style.MutedText, false},
		{"no fix", []brew.Vulnerability{vuln("A", "HIGH")}, "no fix yet", style.WarningText, false},
		{"upgrade", []brew.Vulnerability{vuln("A", "HIGH", "1.5")}, "upgrade → 2.0", style.SuccessText, true},
		{"all unknown", []brew.Vulnerability{vuln("A", "UNKNOWN", "1.5")}, "fix in 1.5", style.MutedText, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := loadedStatus(140, 40)
			outdatedLoaded(m, "pkg", "2.0", c.outdated)
			f := brew.VulnFinding{Formula: "pkg", Vulnerabilities: c.vulns}
			text, role := m.vulnNextStep(f)
			if text != c.phrase {
				t.Fatalf("phrase %q, want %q", text, c.phrase)
			}
			if got, want := role.Render(text), c.role.Render(text); got != want {
				t.Errorf("phrase drawn as %q, want the role's %q", got, want)
			}
		})
	}
}

// TestVulnsNextStepHighestRelease: the fix named is the highest released version,
// compared by its numbers, and a commit SHA is never named.
func TestVulnsNextStepHighestRelease(t *testing.T) {
	m := loadedStatus(140, 40)
	f := brew.VulnFinding{Formula: "pkg", Vulnerabilities: []brew.Vulnerability{
		vuln("A", "HIGH", "1.9.0", sha),
		vuln("B", "HIGH", "1.10.0"),
		vuln("C", "MEDIUM", "1.9.9"),
	}}
	if text, _ := m.vulnNextStep(f); text != "fix in 1.10.0" {
		t.Errorf("next step %q, want fix in 1.10.0", text)
	}
}

// TestVersionAfter: versions compare by their numbers, with a leading v ignored.
func TestVersionAfter(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.10.0", "1.9.0", true},
		{"1.9.0", "1.10.0", false},
		{"2.0", "1.99.1", true},
		{"v1.10", "v1.9", true},
		{"1.0.1", "1.0", true},
		{"1.0", "1.0", false},
		// Regression F-r1-02: a release is later than its pre-releases.
		{"2.0.0", "2.0.0-rc1", true},
		{"v2.0.0", "2.0.0-rc1", true},
		{"2.0.0-rc1", "2.0.0", false},
		{"2.0.0-rc2", "2.0.0-rc1", true},
		{"2.0.0-rc1", "1.9.9", true},
		{"2.0", "2.0.0-rc1", true},
		{"1.0", "1.0.0", false},
	}
	for _, c := range cases {
		if got := versionAfter(c.a, c.b); got != c.want {
			t.Errorf("versionAfter(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// TestVulnsNextStepRightAligned: on a wide terminal the next step sits at the
// right of the block's header line.
func TestVulnsNextStepRightAligned(t *testing.T) {
	out := reportJSON(t, "pkg", "1.0", vuln("CVE-1", "HIGH", "1.5"))
	calls := 0
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(vulnsRunner(out, nil, &calls))
	m = pressStatus(m, "v")

	line := vulnLine(t, ansi.Strip(m.View().Content), "pkg 1.0")
	if at := strings.Index(line, "fix in 1.5"); at < len(line)/2 {
		t.Errorf("the next step should be at the right of the header, at %d of %d: %q", at, len(line), line)
	}
}

// TestVulnsCopyHoldsTheWholeReport: y copies the summary, the full fixed versions
// of each advisory, the names of the packages not scanned, and the patched IDs,
// with every summary on one line.
func TestVulnsCopyHoldsTheWholeReport(t *testing.T) {
	calls := 0
	m := loadedStatus(80, 24)
	copied := captureClipboard(m)
	m.client = brew.NewClient(vulnsRunner(fixtureVulns(t), vulnsExit1, &calls))
	m = pressStatus(m, "v")
	sendKey(m, "y")

	if len(*copied) != 1 {
		t.Fatalf("y copied %d times, want 1", len(*copied))
	}
	got := (*copied)[0]
	if ansi.Strip(got) != got {
		t.Errorf("the copy holds escape codes")
	}
	for _, want := range []string{
		sha, // a full fixed version, not a shortened one
		"9a829113c88a51e57c1e46473e90609e4b7df151", // libsndfile's fix
		"glib", "mpg123", "libedit", // skipped packages, by name
		"CVE-2025-15661", "OSV-2025-433", // patched IDs of libssh2
		"not scanned",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the copy lacks %q", want)
		}
	}
	for _, f := range fixtureReport(t).Findings {
		for _, v := range f.Vulnerabilities {
			if !strings.Contains(got, v.Summary) {
				t.Errorf("the copy does not hold %s's summary whole: %q", v.ID, v.Summary)
			}
		}
	}
	if strings.Contains(got, "…") {
		t.Errorf("the copy is cut: %q", got)
	}
}
