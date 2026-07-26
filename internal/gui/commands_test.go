package gui

import (
	"testing"
)

func TestParseUpdateSummary(t *testing.T) {
	tests := []struct {
		name   string
		lines  []string
		expect string
	}{
		{
			name:   "empty output",
			lines:  nil,
			expect: "",
		},
		{
			name:   "empty lines only",
			lines:  []string{"", "  ", "\t"},
			expect: "",
		},
		{
			name:   "already up to date",
			lines:  []string{"Already up-to-date."},
			expect: "Already up to date",
		},
		{
			name:   "already up to date with surrounding output",
			lines:  []string{"", "HOMEBREW_BREW_GIT_REMOTE set: using https://github.com/Homebrew/brew", "Already up-to-date.", ""},
			expect: "Already up to date",
		},
		{
			name:   "updated single formula",
			lines:  []string{"Updated 1 tap (homebrew/core)."},
			expect: "Updated 1 tap (homebrew/core)",
		},
		{
			name:   "updated multiple taps",
			lines:  []string{"Updated 3 taps (homebrew/core, homebrew/cask, homebrew/services)."},
			expect: "Updated 3 taps (homebrew/core, homebrew/cask, homebrew/services)",
		},
		{
			name:   "updated formulae with count",
			lines:  []string{"==> Updating Homebrew...", "Updated 5 formulae."},
			expect: "Updated 5 formulae",
		},
		{
			name:   "error prefix",
			lines:  []string{"Error: Failed to connect."},
			expect: "",
		},
		{
			name:   "no matching lines",
			lines:  []string{"Some other output", "Nothing useful here"},
			expect: "",
		},
		{
			name: "multiple updated lines returns first",
			lines: []string{
				"Updated 1 tap (homebrew/core).",
				"Updated 2 taps (homebrew/cask).",
			},
			expect: "Updated 1 tap (homebrew/core)",
		},
		{
			name: "already up to date has priority over updated",
			lines: []string{
				"Already up-to-date.",
				"Updated 1 tap (homebrew/core).",
			},
			expect: "Already up to date",
		},
		{
			name:   "already up to date without dot",
			lines:  []string{"Already up-to-date"},
			expect: "Already up to date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUpdateSummary(tt.lines)
			if got != tt.expect {
				t.Errorf("parseUpdateSummary(%v) = %q, want %q", tt.lines, got, tt.expect)
			}
		})
	}
}
