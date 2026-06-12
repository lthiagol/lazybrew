package presentation

import (
	"strings"
	"testing"

	"github.com/thiago/lazybrew/internal/brew"
)

func TestSnapshotFormulaFormats(t *testing.T) {
	tests := []struct {
		name string
		f    brew.Formula
	}{
		{"normal", brew.Formula{Name: "ripgrep", Version: "14.1.1", Bottled: true}},
		{"pinned", brew.Formula{Name: "python@3.12", Version: "3.12.8", Pinned: true}},
		{"outdated", brew.Formula{Name: "node", Version: "22.5.0", Bottled: true, Outdated: true, NewVersion: "22.6.1"}},
		{"keg_only", brew.Formula{Name: "openssl", Version: "3.2.1", KegOnly: true, Bottled: true}},
		{"no_bottle", brew.Formula{Name: "fish", Version: "3.7.1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFormula(tt.f, 0)
			if got == "" {
				t.Error("expected non-empty output")
			}
			if !strings.Contains(got, tt.f.Name) {
				t.Errorf("expected name %q in output, got: %s", tt.f.Name, got)
			}
		})
	}
}

func TestSnapshotCaskFormats(t *testing.T) {
	tests := []struct {
		name string
		c    brew.Cask
	}{
		{"normal", brew.Cask{Name: "iterm2", Version: "3.5.10"}},
		{"outdated", brew.Cask{Name: "firefox", Version: "134.0", Outdated: true, NewVersion: "135.0"}},
		{"auto_update", brew.Cask{Name: "google-chrome", Version: "132.0", AutoUpdates: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCask(tt.c, 0)
			if got == "" {
				t.Error("expected non-empty output")
			}
			if !strings.Contains(got, tt.c.Name) {
				t.Errorf("expected name %q in output, got: %s", tt.c.Name, got)
			}
		})
	}
}

func TestSnapshotTapFormats(t *testing.T) {
	tests := []struct {
		name string
		tap  brew.Tap
	}{
		{"official", brew.Tap{Name: "homebrew/core", IsOfficial: true, Trusted: true, IsAPI: true}},
		{"trusted", brew.Tap{Name: "nicknisi/tap", Trusted: true}},
		{"untrusted", brew.Tap{Name: "some-org/formulas"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTap(tt.tap, 0)
			if got == "" {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestSnapshotServiceFormats(t *testing.T) {
	tests := []struct {
		name string
		s    brew.Service
	}{
		{"started", brew.Service{Name: "postgresql@16", Status: brew.ServiceStarted, User: "thiago"}},
		{"stopped", brew.Service{Name: "redis", Status: brew.ServiceStopped}},
		{"error", brew.Service{Name: "nginx", Status: brew.ServiceError, ExitCode: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatService(tt.s, 0)
			if got == "" {
				t.Error("expected non-empty output")
			}
			if !strings.Contains(got, tt.s.Name) {
				t.Errorf("expected name %q in output", tt.s.Name)
			}
		})
	}
}

func TestSnapshotDashboard(t *testing.T) {
	got := FormatStatusDashboard(124, 31, 7, 4, 2, 2, 3, 2, "6.0.0", "/opt/homebrew")
	if len(got) < 5 {
		t.Errorf("expected at least 5 dashboard lines, got %d", len(got))
	}
}
