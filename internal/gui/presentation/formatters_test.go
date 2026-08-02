package presentation

import (
	"errors"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
)

func TestFormatFormula(t *testing.T) {
	f := brew.Formula{
		Name:       "ripgrep",
		Version:    "14.1.1",
		Bottled:    true,
		Outdated:   true,
		NewVersion: "14.1.2",
	}
	got := FormatFormula(f, 0)
	if !strings.Contains(got, "ripgrep") {
		t.Errorf("expected ripgrep in output, got: %s", got)
	}
	if !strings.Contains(got, "outdated 14.1.2") {
		t.Errorf("expected outdated marker, got: %s", got)
	}
}

func TestFormatFormulaPinned(t *testing.T) {
	f := brew.Formula{
		Name:    "python@3.12",
		Version: "3.12.8",
		Pinned:  true,
	}
	got := FormatFormula(f, 0)
	if !strings.Contains(got, "pinned") {
		t.Errorf("expected pinned marker, got: %s", got)
	}
}

func TestFormatCask(t *testing.T) {
	c := brew.Cask{
		Name:       "firefox",
		Version:    "135.0",
		Outdated:   true,
		NewVersion: "136.0",
	}
	got := FormatCask(c, 0)
	if !strings.Contains(got, "firefox") {
		t.Errorf("expected firefox in output, got: %s", got)
	}
	if !strings.Contains(got, "outdated 136.0") {
		t.Errorf("expected outdated marker, got: %s", got)
	}
}

func TestFormatTap(t *testing.T) {
	tap := brew.Tap{
		Name:       "homebrew/core",
		IsOfficial: true,
		Trusted:    true,
		IsAPI:      true,
	}
	got := FormatTap(tap, 0)
	if !strings.Contains(got, "official") {
		t.Errorf("expected official, got: %s", got)
	}

	thirdParty := brew.Tap{
		Name:    "nicknisi/tap",
		Trusted: false,
	}
	got2 := FormatTap(thirdParty, 0)
	if !strings.Contains(got2, "untrusted") {
		t.Errorf("expected untrusted, got: %s", got2)
	}
}

func TestFormatService(t *testing.T) {
	s := brew.Service{
		Name:   "postgresql@16",
		Status: brew.ServiceStarted,
		User:   "thiago",
	}
	got := FormatService(s, 0)
	if !strings.Contains(got, "started") {
		t.Errorf("expected started, got: %s", got)
	}
	if !strings.Contains(got, "thiago") {
		t.Errorf("expected user, got: %s", got)
	}
}

func TestFormatStatusDashboard(t *testing.T) {
	items := FormatStatusDashboard(124, 31, 7, 4, 2, 2, 3, 2, "6.0.0", "/opt/homebrew")
	if len(items) == 0 {
		t.Fatal("expected non-empty dashboard")
	}
	full := strings.Join(items, "\n")
	if !strings.Contains(full, "6.0.0") {
		t.Errorf("expected version, got: %s", full)
	}
	if !strings.Contains(full, "124") {
		t.Errorf("expected count, got: %s", full)
	}
}

func TestFormatDoctorStatus(t *testing.T) {
	clean := FormatDoctorStatus([]brew.DoctorWarning{}, nil)
	if !strings.Contains(clean, "No issues") {
		t.Errorf("expected No issues, got: %s", clean)
	}

	warnings := FormatDoctorStatus([]brew.DoctorWarning{{Title: "test"}}, nil)
	if !strings.Contains(warnings, "1 warning") {
		t.Errorf("expected 1 warning, got: %s", warnings)
	}

	unavailable := FormatDoctorStatus(nil, errors.New("brew not found"))
	if !strings.Contains(unavailable, "unavailable") {
		t.Errorf("expected unavailable marker, got: %s", unavailable)
	}
	if !strings.Contains(unavailable, "brew not found") {
		t.Errorf("expected error reason in unavailable line, got: %s", unavailable)
	}

	errBeatsWarnings := FormatDoctorStatus([]brew.DoctorWarning{{Title: "stale"}}, errors.New("exec failed"))
	if !strings.Contains(errBeatsWarnings, "unavailable") {
		t.Errorf("expected unavailable marker when err present, got: %s", errBeatsWarnings)
	}
	if strings.Contains(errBeatsWarnings, "stale") || strings.Contains(errBeatsWarnings, "warning") {
		t.Errorf("err should win; warnings must NOT appear, got: %s", errBeatsWarnings)
	}
}

func TestFormatFormulaInfoSnapshot(t *testing.T) {
	f := brew.Formula{
		Name:        "ripgrep",
		Version:     "14.1.1",
		Tap:         "homebrew/core",
		Description: "A command-line search tool",
		Homepage:    "https://github.com/BurntSushi/ripgrep",
		License:     "MIT",
		Bottled:     true,
	}
	got := FormatFormulaInfo(f, 60)
	if !strings.Contains(got, "Name:") {
		t.Error("expected Name label")
	}
	if !strings.Contains(got, "ripgrep") {
		t.Error("expected formula name")
	}
	if !strings.Contains(got, "MIT") {
		t.Error("expected license")
	}
	if !strings.Contains(got, "installed") {
		t.Error("expected status")
	}
}

func TestFormatFormulaInfoOutdated(t *testing.T) {
	f := brew.Formula{
		Name:       "neovim",
		Version:    "0.10.4",
		Outdated:   true,
		NewVersion: "0.11.0",
		KegOnly:    false,
		Bottled:    true,
	}
	got := FormatFormulaInfo(f, 60)
	if !strings.Contains(got, "outdated") {
		t.Error("expected outdated status")
	}
	if !strings.Contains(got, "0.10.4") {
		t.Error("expected current version")
	}
	if !strings.Contains(got, "0.11.0") {
		t.Error("expected new version")
	}
}

func TestFormatFormulaInfoPinned(t *testing.T) {
	f := brew.Formula{
		Name:    "python@3.12",
		Version: "3.12.8",
		Pinned:  true,
	}
	got := FormatFormulaInfo(f, 60)
	if !strings.Contains(got, "pinned") {
		t.Error("expected pinned status")
	}
}

func TestFormatCaskInfoSnapshot(t *testing.T) {
	c := brew.Cask{
		Name:        "firefox",
		Version:     "135.0",
		Tap:         "homebrew/cask",
		Description: "Mozilla Firefox web browser",
		Homepage:    "https://www.mozilla.org/firefox/",
	}
	got := FormatCaskInfo(c, 60)
	if !strings.Contains(got, "Name:") {
		t.Error("expected Name label")
	}
	if !strings.Contains(got, "firefox") {
		t.Error("expected cask name")
	}
	if !strings.Contains(got, "installed") {
		t.Error("expected status")
	}
}

func TestFormatCaskInfoOutdated(t *testing.T) {
	c := brew.Cask{
		Name:       "google-chrome",
		Version:    "132.0",
		Outdated:   true,
		NewVersion: "133.0",
	}
	got := FormatCaskInfo(c, 60)
	if !strings.Contains(got, "outdated") {
		t.Error("expected outdated status")
	}
	if !strings.Contains(got, "132.0") {
		t.Error("expected current version")
	}
}

func TestPadRight(t *testing.T) {
	got := padRight("abc", 6)
	if len(got) != 6 {
		t.Errorf("expected length 6, got %d", len(got))
	}
	if got != "abc   " {
		t.Errorf("expected 'abc   ', got '%s'", got)
	}
}

func TestPadRightUTF8(t *testing.T) {
	got := padRight("café", 10)
	runes := []rune(got)
	if len(runes) != 10 {
		t.Errorf("expected 10 runes, got %d: %q", len(runes), got)
	}
	if runes[0] != 'c' || runes[3] != 'é' {
		t.Errorf("UTF-8 corrupted: %q", got)
	}
}

func TestPadRightTruncation(t *testing.T) {
	got := padRight("café", 3)
	if got != "caf" {
		t.Errorf("expected 'caf', got %q", got)
	}
}

func TestFormatOutdatedFormula(t *testing.T) {
	got := FormatOutdatedFormula(brew.Formula{Name: "ripgrep", Version: "13.0.0", NewVersion: "14.0.0"})
	want := "ripgrep                 13.0.0       -> 14.0.0"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatOutdatedCask(t *testing.T) {
	got := FormatOutdatedCask(brew.Cask{Name: "firefox", Version: "120.0", NewVersion: "125.0"})
	want := "firefox                 120.0        -> 125.0  cask"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestVersionDisplayNoList(t *testing.T) {
	got := versionDisplay(brew.Formula{Version: "1.2.3"})
	if got != "1.2.3" {
		t.Errorf("got %q, want %q", got, "1.2.3")
	}
}

func TestVersionDisplayWithList(t *testing.T) {
	got := versionDisplay(brew.Formula{
		Version:      "2.0.0",
		ListVersions: []string{"1.0.0", "2.0.0"},
	})
	if got != "2.0.0 (1.0.0)" {
		t.Errorf("got %q, want %q", got, "2.0.0 (1.0.0)")
	}
}

func TestTruncateShort(t *testing.T) {
	got := truncate("hello", 10)
	if got != "hello" {
		t.Errorf("got %q, want hello (no truncation)", got)
	}
}

func TestTruncateLong(t *testing.T) {
	got := truncate("hello world", 6)
	if got != "hello…" {
		t.Errorf("got %q, want %q", got, "hello…")
	}
}

func TestTruncateUTF8(t *testing.T) {
	got := truncate("café latte", 5)
	want := "café…"
	if got != want {
		t.Errorf("got %q, want %q (utf-8 safe)", got, want)
	}
}
