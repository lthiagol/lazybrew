package brew

import (
	"testing"
)

func TestTrustStatusString(t *testing.T) {
	tests := []struct {
		status TrustStatus
		want   string
	}{
		{TrustOfficial, "official"},
		{TrustTrusted, "trusted"},
		{TrustUntrusted, "untrusted"},
		{TrustStatus(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("TrustStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestParseTrustStatus(t *testing.T) {
	tests := []struct {
		s    string
		want TrustStatus
	}{
		{"official", TrustOfficial},
		{"trusted", TrustTrusted},
		{"untrusted", TrustUntrusted},
		{"unknown", TrustUnknown},
		{"random", TrustUnknown},
	}
	for _, tt := range tests {
		if got := ParseTrustStatus(tt.s); got != tt.want {
			t.Errorf("ParseTrustStatus(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestServiceStatusString(t *testing.T) {
	tests := []struct {
		status ServiceStatus
		want   string
	}{
		{ServiceStarted, "started"},
		{ServiceStopped, "stopped"},
		{ServiceError, "error"},
		{ServiceNone, "none"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("ServiceStatus(%q).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestParseServiceStatus(t *testing.T) {
	tests := []struct {
		s    string
		want ServiceStatus
	}{
		{"started", ServiceStarted},
		{"stopped", ServiceStopped},
		{"error", ServiceError},
		{"none", ServiceNone},
		{"running", ServiceNone},
	}
	for _, tt := range tests {
		if got := ParseServiceStatus(tt.s); got != tt.want {
			t.Errorf("ParseServiceStatus(%q) = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestFormulaHas6Point0Fields(t *testing.T) {
	f := Formula{
		Aliases:             []string{"rg"},
		Binaries:            []string{"rg"},
		InstalledDependents: []string{"other-pkg"},
		ListVersions:        []string{"14.1.0", "14.1.1"},
		Revision:            "1",
	}
	if len(f.Aliases) != 1 {
		t.Error("Formula.Aliases should be populated")
	}
	if len(f.Binaries) != 1 {
		t.Error("Formula.Binaries should be populated")
	}
	if len(f.InstalledDependents) != 1 {
		t.Error("Formula.InstalledDependents should be populated")
	}
	if len(f.ListVersions) != 2 {
		t.Error("Formula.ListVersions should have 2 entries")
	}
}

func TestCaskHas6Point0Fields(t *testing.T) {
	c := Cask{
		Pinned:        true,
		Sha256:        "abc123",
		URL:           "https://example.com/app.dmg",
		DependsOn:     []string{"macos"},
		ConflictsWith: []string{"other-app"},
	}
	if !c.Pinned {
		t.Error("Cask.Pinned should be true")
	}
	if c.Sha256 != "abc123" {
		t.Error("Cask.Sha256 should be set")
	}
	if len(c.DependsOn) != 1 {
		t.Error("Cask.DependsOn should be populated")
	}
}

func TestTapHas6Point0Fields(t *testing.T) {
	tap := Tap{
		Trusted:      true,
		FormulaNames: []string{"foo"},
		CaskNames:    []string{"bar"},
	}
	if !tap.Trusted {
		t.Error("Tap.Trusted should be true")
	}
	if len(tap.FormulaNames) != 1 {
		t.Error("Tap.FormulaNames should be populated")
	}
}
