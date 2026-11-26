package gui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestConfigCoreTapRowKeepsItsKind: the core tap's value is a cache time or a
// commit, so its row is labelled with the key brew printed it under.
func TestConfigCoreTapRowKeepsItsKind(t *testing.T) {
	cases := []struct {
		name  string
		cfg   brew.BrewConfig
		label string
		value string
	}{
		{
			name:  "JSON cache time",
			cfg:   brew.BrewConfig{CoreTap: "26 Nov 11:31 UTC", CoreTapLabel: "Core tap JSON"},
			label: "Core tap JSON:",
			value: "26 Nov 11:31 UTC",
		},
		{
			name:  "git clone commit",
			cfg:   brew.BrewConfig{CoreTap: "1423910", CoreTapLabel: "Core tap HEAD"},
			label: "Core tap HEAD:",
			value: "1423910",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ansi.Strip(formatConfig(&tc.cfg))
			row := ""
			for _, line := range strings.Split(out, "\n") {
				if strings.Contains(line, tc.label) {
					row = line
				}
			}
			if !strings.Contains(row, tc.value) {
				t.Fatalf("no %q row with value %q in:\n%s", tc.label, tc.value, out)
			}
			if strings.Contains(out, "Core Tap") {
				t.Fatalf("the core tap row still says Core Tap:\n%s", out)
			}
		})
	}
}
