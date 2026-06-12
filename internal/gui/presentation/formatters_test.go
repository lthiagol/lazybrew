package presentation

import (
	"strings"
	"testing"

	"github.com/thiago/lazybrew/internal/brew"
)

func TestFormatFormula(t *testing.T) {
	f := brew.Formula{
		Name:    "ripgrep",
		Version: "14.1.1",
		Bottled: true,
		Outdated: true,
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
		Name:    "firefox",
		Version: "135.0",
		Outdated: true,
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
	clean := FormatDoctorStatus([]brew.DoctorWarning{})
	if !strings.Contains(clean, "No issues") {
		t.Errorf("expected No issues, got: %s", clean)
	}

	warnings := FormatDoctorStatus([]brew.DoctorWarning{{Title: "test"}})
	if !strings.Contains(warnings, "1 warning") {
		t.Errorf("expected 1 warning, got: %s", warnings)
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
