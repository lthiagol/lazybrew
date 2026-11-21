package presentation

import (
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

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
	got := FormatFormulaInfo(f, 60, style.UnicodeIcons)
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
	got := FormatFormulaInfo(f, 60, style.UnicodeIcons)
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
	got := FormatFormulaInfo(f, 60, style.UnicodeIcons)
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
	got := FormatCaskInfo(c, 60, style.UnicodeIcons)
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
	got := FormatCaskInfo(c, 60, style.UnicodeIcons)
	if !strings.Contains(got, "outdated") {
		t.Error("expected outdated status")
	}
	if !strings.Contains(got, "132.0") {
		t.Error("expected current version")
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
