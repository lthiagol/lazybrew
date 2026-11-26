package brew

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

func TestDiagnosticsDoctorClean(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("Your system is ready to brew."), nil
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	warnings, err := diag.Doctor(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(warnings))
	}
}

func TestDiagnosticsDoctorExitCode1(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("Warning: Your Homebrew is outdated.\nRun `brew update` to get the latest.\n"), &BrewExitError{Command: "doctor", ExitCode: 1}
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	warnings, err := diag.Doctor(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
}

func TestDiagnosticsDoctorRealError(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, &BrewExitError{Command: "doctor", ExitCode: 2}
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	_, err := diag.Doctor(context.Background())
	if err == nil {
		t.Fatal("expected error for exit code 2")
	}
}

func TestDiagnosticsDoctorWarnings(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("Warning: Your Homebrew is outdated.\nRun `brew update` to get the latest.\n\nWarning: Broken symlinks found.\n/opt/homebrew/bin/old-tool -> (missing)\n"), nil
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	warnings, err := diag.Doctor(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(warnings))
	}
	if !strings.Contains(warnings[0].Title, "outdated") {
		t.Errorf("first warning should mention outdated, got: %s", warnings[0].Title)
	}
}

func TestDiagnosticsConfig(t *testing.T) {
	// testdata/config_linux.txt is real `HOMEBREW_NO_AUTO_UPDATE=1 brew config`
	// output (brew 7.0.16, Debian 13).
	out, err := os.ReadFile("testdata/config_linux.txt")
	if err != nil {
		t.Fatal(err)
	}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return out, nil
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	cfg, err := diag.Config(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HomebrewVersion != "7.0.16-52-g1423910" {
		t.Errorf("HomebrewVersion = %q, want 7.0.16-52-g1423910", cfg.HomebrewVersion)
	}
	if cfg.Prefix != "/home/linuxbrew/.linuxbrew" {
		t.Errorf("Prefix = %q, want /home/linuxbrew/.linuxbrew", cfg.Prefix)
	}
	if cfg.CoreTap != "26 Nov 11:31 UTC" || cfg.CoreTapLabel != "Core tap JSON" {
		t.Errorf("CoreTap = %q under %q, want the Core tap JSON time", cfg.CoreTap, cfg.CoreTapLabel)
	}
	if want := "Debian GNU/Linux 13 (trixie) (trixie)"; cfg.OS != want {
		t.Errorf("OS = %q, want %q", cfg.OS, want)
	}
}

// TestDiagnosticsConfigMacOSAndGitTap covers the lines brew config prints on
// macOS (macOS:) and for a git core tap (Core tap HEAD:). No macOS capture is
// on this machine; the lines follow brew's source, extend/os/mac/system_config.rb
// and system_config.rb dump_tap_config.
func TestDiagnosticsConfigMacOSAndGitTap(t *testing.T) {
	cases := []struct {
		name, output, wantOS, wantCoreTap, wantLabel string
	}{
		{
			name:        "macOS with the JSON core tap",
			output:      "HOMEBREW_VERSION: 5.0.0\nHOMEBREW_PREFIX: /opt/homebrew\nCore tap JSON: 10 Nov 12:34 UTC\nmacOS: 15.4-arm64\n",
			wantOS:      "15.4-arm64",
			wantCoreTap: "10 Nov 12:34 UTC",
			wantLabel:   "Core tap JSON",
		},
		{
			name:        "git core tap reads its HEAD",
			output:      "HOMEBREW_VERSION: 5.0.0\nCore tap HEAD: 1423910\nOS: Ubuntu 24.04\n",
			wantOS:      "Ubuntu 24.04",
			wantCoreTap: "1423910",
			wantLabel:   "Core tap HEAD",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := brewtest.NewMockRunner()
			r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
				return []byte(tc.output), nil
			}
			cfg, err := NewDiagnosticsReader(r, NewCache(time.Minute)).Config(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if cfg.OS != tc.wantOS {
				t.Errorf("OS = %q, want %q", cfg.OS, tc.wantOS)
			}
			if cfg.CoreTap != tc.wantCoreTap {
				t.Errorf("CoreTap = %q, want %q", cfg.CoreTap, tc.wantCoreTap)
			}
			if cfg.CoreTapLabel != tc.wantLabel {
				t.Errorf("CoreTapLabel = %q, want %q", cfg.CoreTapLabel, tc.wantLabel)
			}
		})
	}
}

func TestDiagnosticsMissing(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("python@3.12: missing libsqlite\nnode: missing icu4c\n"), nil
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	missing, err := diag.Missing(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 2 {
		t.Fatalf("expected 2, got %d", len(missing))
	}
	if missing[0].Formula != "python@3.12" || missing[0].Missing != "missing libsqlite" {
		t.Errorf("got %+v, want python@3.12: libsqlite", missing[0])
	}
}

func TestDiagnosticsMissingExitCode1ParsesOutput(t *testing.T) {
	// `brew missing` exits 1 when dependencies are missing: the output
	// must still be parsed (r1 finding 18).
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("node: missing icu4c\n"),
			&BrewExitError{Command: "missing", ExitCode: 1}
	}
	diag := NewDiagnosticsReader(r, NewCache(time.Minute))

	missing, err := diag.Missing(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 1 || missing[0].Formula != "node" {
		t.Fatalf("expected parsed missing deps on exit 1, got %+v (err=%v)", missing, err)
	}
}

func TestDiagnosticsMissingExitCode2IsError(t *testing.T) {
	// Regression for r1 finding 18: the old substring match on the error
	// text treated ANY failure of `brew missing` (the command name
	// contains "missing") as success. A non-1 exit code must surface as
	// an error.
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, &BrewExitError{Command: "missing", ExitCode: 2, Stderr: "boom"}
	}
	diag := NewDiagnosticsReader(r, NewCache(time.Minute))

	missing, err := diag.Missing(context.Background())
	if err == nil {
		t.Fatalf("expected error on exit 2, got %+v", missing)
	}
	if !IsExitCode(err, 2) {
		t.Errorf("expected exit code 2 error, got %v", err)
	}
}

// vulnsOnce runs VulnsReport against one brew answer.
func vulnsOnce(t *testing.T, out string, err error) (VulnReport, error) {
	t.Helper()
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(out), err
	}
	return NewDiagnosticsReader(r, NewCache(time.Minute)).VulnsReport(context.Background())
}

func TestDiagnosticsVulnsExitCode1WithReportIsResult(t *testing.T) {
	// brew vulns exits 1 when it finds vulnerabilities and prints its JSON report.
	out := `{"findings":[{"formula":"openssl@3","version":"3.0","vulnerabilities":[{"id":"CVE-2026-1","severity":"HIGH","summary":"s"}],"patched":[]}],"skipped_formulae":[],"checked":1,"complete":true}`
	report, err := vulnsOnce(t, out, &BrewExitError{Command: "vulns", ExitCode: 1})
	if err != nil {
		t.Fatalf("exit 1 with a JSON report is a result, got error %v", err)
	}
	if !report.Parsed || len(report.Findings) != 1 {
		t.Fatalf("the JSON report should be read, got %+v", report)
	}
}

func TestDiagnosticsVulnsExitCode1WithTextIsError(t *testing.T) {
	// Regression F-r1-01: exit 1 with stdout that is not JSON, and no Error: line,
	// is still a failure. Only exit 0 may fall back to brew's text.
	_, err := vulnsOnce(t, "CVE-2026-1: openssl@3\n", &BrewExitError{Command: "vulns", ExitCode: 1})
	if !IsExitCode(err, 1) {
		t.Fatalf("expected the exit 1 error for text on exit 1, got %v", err)
	}
}

func TestDiagnosticsVulnsExitCode1WithoutReportIsError(t *testing.T) {
	// Regression F-r1-01: a brew without the vulns command exits 1 with an
	// empty stdout and an Error: line on stderr. That is a failure, not an
	// empty report.
	_, err := vulnsOnce(t, "", &BrewExitError{
		Command: "vulns", ExitCode: 1,
		Stderr: "Error: Invalid usage: Unknown command: brew vulns\n",
	})
	if !IsExitCode(err, 1) {
		t.Fatalf("expected the exit 1 error, got %v", err)
	}
}

func TestDiagnosticsVulnsErrorLineIsError(t *testing.T) {
	// An Error: line on stderr makes the run a failure even when stdout
	// holds lines.
	_, err := vulnsOnce(t, "CVE-2026-1: openssl@3\n", &BrewExitError{
		Command: "vulns", ExitCode: 1, Stderr: "Error: Failed to reach OSV.dev\n",
	})
	if !IsExitCode(err, 1) {
		t.Fatalf("expected the exit 1 error, got %v", err)
	}
}

func TestDiagnosticsDoctorExitCode1ReadsWarningsFromStderr(t *testing.T) {
	// Real brew 7.0.15: `brew doctor` exits 1, its stdout is one blank line,
	// and the Warning: blocks are on stderr. Execute trims stderr.
	stdout, err := os.ReadFile("testdata/doctor_exit1_stdout.txt")
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := os.ReadFile("testdata/doctor_exit1_stderr.txt")
	if err != nil {
		t.Fatal(err)
	}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return stdout, &BrewExitError{Command: "doctor", ExitCode: 1, Stderr: strings.TrimSpace(string(stderr))}
	}
	diag := NewDiagnosticsReader(r, NewCache(time.Minute))

	warnings, err := diag.Doctor(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning from stderr, got %d: %+v", len(warnings), warnings)
	}
	if want := "The following taps are not trusted:"; warnings[0].Title != want {
		t.Errorf("Title = %q, want %q", warnings[0].Title, want)
	}
	if !strings.Contains(warnings[0].Details, "danielfoehrkn/switch") {
		t.Errorf("Details should list the tap, got: %s", warnings[0].Details)
	}
}
