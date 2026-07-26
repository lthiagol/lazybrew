package brew

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDiagnosticsDoctorClean(t *testing.T) {
	r := NewMockRunner()
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
	r := NewMockRunner()
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
	r := NewMockRunner()
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
	r := NewMockRunner()
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

func TestDiagnosticsVersion(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("Homebrew 6.0.0"), nil
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	version, err := diag.Version(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if version != "6.0.0" {
		t.Errorf("expected 6.0.0, got %s", version)
	}
}

func TestDiagnosticsConfig(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("HOMEBREW_VERSION: 6.0.0\nHOMEBREW_PREFIX: /opt/homebrew\nHOMEBREW_CELLAR: /opt/homebrew/Cellar\nHOMEBREW_REPOSITORY: /opt/homebrew\nHOMEBREW_CORE_TAP: homebrew/core\nHOMEBREW_SYSTEM: macOS\n"), nil
	}
	cache := NewCache(time.Minute)
	diag := NewDiagnosticsReader(r, cache)

	cfg, err := diag.Config(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HomebrewVersion != "6.0.0" {
		t.Errorf("HomebrewVersion = %q, want 6.0.0", cfg.HomebrewVersion)
	}
	if cfg.Prefix != "/opt/homebrew" {
		t.Errorf("Prefix = %q, want /opt/homebrew", cfg.Prefix)
	}
	if cfg.CoreTap != "homebrew/core" {
		t.Errorf("CoreTap = %q, want homebrew/core", cfg.CoreTap)
	}
}

func TestDiagnosticsMissing(t *testing.T) {
	r := NewMockRunner()
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

func TestParseDoctorWarnings(t *testing.T) {
	tests := []struct {
		input string
		count int
	}{
		{"", 0},
		{"Your system is ready to brew.", 0},
		{"Warning: First warning.\nDetail line.\n\nWarning: Second warning.\n", 2},
		{"Warning: Only one.", 1},
	}
	for _, tt := range tests {
		warnings := parseDoctorWarnings(tt.input)
		if len(warnings) != tt.count {
			t.Errorf("parseDoctorWarnings(%q) = %d warnings, want %d", tt.input[:min(len(tt.input), 30)], len(warnings), tt.count)
		}
	}
}
