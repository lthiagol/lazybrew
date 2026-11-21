package brew

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// realVulnsJSON is brew vulns --json --max-summary 0 on a real machine, saved
// as testdata/vulns_v1.json: 7 findings, 24 vulnerabilities, 14 patched, 90
// checked, 37 skipped.
func realVulnsJSON(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/vulns_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestVulnsReportRealFixture(t *testing.T) {
	// brew exits 1 when it finds vulnerabilities, and prints the JSON anyway.
	report, err := vulnsOnce(t, realVulnsJSON(t), &BrewExitError{Command: "vulns", ExitCode: 1})
	if err != nil {
		t.Fatalf("exit 1 with a JSON report is a result, got %v", err)
	}
	if !report.Parsed {
		t.Fatal("the saved report should parse")
	}
	if len(report.Findings) != 7 || report.Checked != 90 || report.Complete {
		t.Errorf("findings %d checked %d complete %v, want 7, 90, false", len(report.Findings), report.Checked, report.Complete)
	}
	if len(report.SkippedFormulae) != 37 {
		t.Errorf("skipped %d, want 37", len(report.SkippedFormulae))
	}
	vulns, patched := 0, 0
	severities := map[string]int{}
	for _, f := range report.Findings {
		vulns += len(f.Vulnerabilities)
		patched += len(f.Patched)
		for _, v := range f.Vulnerabilities {
			severities[v.Severity]++
		}
		if f.Formula == "libssh2" && len(f.Patched) != 14 {
			t.Errorf("libssh2 patched %d, want 14", len(f.Patched))
		}
	}
	if vulns != 24 || patched != 14 {
		t.Errorf("vulnerabilities %d patched %d, want 24 and 14", vulns, patched)
	}
	want := map[string]int{"HIGH": 8, "MEDIUM": 5, "UNKNOWN": 11}
	if !reflect.DeepEqual(severities, want) {
		t.Errorf("severities %v, want %v", severities, want)
	}
}

func TestVulnsReportAsksForJSON(t *testing.T) {
	var got []string
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		got = args
		return []byte(`{"findings":[]}`), nil
	}
	if _, err := NewDiagnosticsReader(r, NewCache(time.Minute)).VulnsReport(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := []string{"vulns", "--json", "--max-summary", "0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("brew ran %v, want %v", got, want)
	}
}

func TestVulnsReportSyntheticSeverities(t *testing.T) {
	out := `{"findings":[{"formula":"x","version":"1","vulnerabilities":[
		{"id":"A","severity":"CRITICAL","summary":"s","fixed_versions":["2.0"]},
		{"id":"B","severity":"LOW","summary":"s","fixed_versions":[]}],"patched":[]}],
		"skipped_formulae":[],"checked":1,"complete":true}`
	report, err := vulnsOnce(t, out, nil)
	if err != nil {
		t.Fatal(err)
	}
	vs := report.Findings[0].Vulnerabilities
	if vs[0].Severity != "CRITICAL" || vs[1].Severity != "LOW" {
		t.Errorf("severities %q %q, want CRITICAL and LOW", vs[0].Severity, vs[1].Severity)
	}
	if !report.Complete {
		t.Error("complete should read true")
	}
}

func TestVulnsReportEmptyFindings(t *testing.T) {
	out := `{"findings":[],"skipped_formulae":["m4"],"checked":3,"complete":true,"outdated_without_sbom":[],"unscanned_formulae":{}}`
	report, err := vulnsOnce(t, out, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Parsed || len(report.Findings) != 0 || report.Checked != 3 || len(report.SkippedFormulae) != 1 {
		t.Errorf("empty report read as %+v", report)
	}
}

func TestVulnsReportMalformedJSONFallsBackToText(t *testing.T) {
	out := `{"findings": [`
	report, err := vulnsOnce(t, out, nil)
	if err != nil {
		t.Fatalf("malformed JSON on exit 0 is not an error, got %v", err)
	}
	if report.Parsed || report.Text != out {
		t.Errorf("malformed JSON should fall back to brew's stdout, got %+v", report)
	}
}

func TestReleasedFix(t *testing.T) {
	cases := map[string]bool{
		"ead311df1037e7994a6d86c7e862d71ad59c52c7": false, // a commit SHA: not released
		"EAD311DF1037E7994A6D86C7E862D71AD59C52C7": false,
		"2.15.4":    true,
		"v1.2.3":    true,
		"":          false,
		"3.0.0-rc1": true,
	}
	for v, want := range cases {
		if got := ReleasedFix(v); got != want {
			t.Errorf("ReleasedFix(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestSeverityRank(t *testing.T) {
	order := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"}
	for i := 1; i < len(order); i++ {
		if SeverityRank(order[i-1]) <= SeverityRank(order[i]) {
			t.Errorf("%s should outrank %s", order[i-1], order[i])
		}
	}
	if SeverityRank("high") != SeverityRank("HIGH") {
		t.Error("severity is compared without case")
	}
	if SeverityRank("something new") != SeverityRank("UNKNOWN") {
		t.Error("an unrecognised severity ranks as UNKNOWN")
	}
}
