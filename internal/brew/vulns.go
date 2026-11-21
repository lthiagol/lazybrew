package brew

import (
	"encoding/json"
	"regexp"
	"strings"
)

// VulnReport is brew vulns' report. Parsed is false when brew's output was not
// its JSON report; Text then holds that output, which is shown as it is.
type VulnReport struct {
	Findings            []VulnFinding              `json:"findings"`
	SkippedFormulae     []string                   `json:"skipped_formulae"`
	Checked             int                        `json:"checked"`
	Complete            bool                       `json:"complete"`
	OutdatedWithoutSBOM []json.RawMessage          `json:"outdated_without_sbom"`
	UnscannedFormulae   map[string]json.RawMessage `json:"unscanned_formulae"`
	Parsed              bool                       `json:"-"`
	Text                string                     `json:"-"`
}

// VulnFinding is one installed formula with known vulnerabilities. Patched holds
// the findings a formula patch already resolves; brew does not count them.
type VulnFinding struct {
	Formula         string          `json:"formula"`
	Version         string          `json:"version"`
	Tag             string          `json:"tag"`
	RepoURL         string          `json:"repo_url"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Patched         []Vulnerability `json:"patched"`
}

// Vulnerability is one advisory. FixedVersions holds commit SHAs for an unreleased
// fix and version strings for a released one (see ReleasedFix).
type Vulnerability struct {
	ID            string   `json:"id"`
	Severity      string   `json:"severity"`
	Summary       string   `json:"summary"`
	Aliases       []string `json:"aliases"`
	FixedVersions []string `json:"fixed_versions"`
}

// parseVulnReport reads brew's JSON report. Output that is not that JSON is kept
// as text, so a brew whose report changes still shows what it printed.
func parseVulnReport(output []byte) VulnReport {
	var report VulnReport
	if err := json.Unmarshal(output, &report); err != nil {
		return VulnReport{Text: string(output)}
	}
	report.Parsed = true
	return report
}

// commitSHA is a 40-hex commit id, which is what brew gives for a fix that is not
// in a release yet.
var commitSHA = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

// ReleasedFix is whether a fixed version is a release (a version string), not a
// commit SHA.
func ReleasedFix(v string) bool {
	return v != "" && !commitSHA.MatchString(v)
}

// SeverityRank orders severities: CRITICAL highest, then HIGH, MEDIUM, LOW, and
// UNKNOWN (or any other word) lowest. Case does not matter.
func SeverityRank(s string) int {
	switch strings.ToUpper(s) {
	case "CRITICAL":
		return 4
	case "HIGH":
		return 3
	case "MEDIUM":
		return 2
	case "LOW":
		return 1
	}
	return 0
}
