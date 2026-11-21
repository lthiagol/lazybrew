package brew

import (
	"encoding/json"
	"os"
	"testing"
)

// TestCaskNameIsTokenInRealBrewShape (LB-12, Q-cr1-cask-name): brew writes the
// token in "token" and the display name as a list in "name". Cask.Name is the
// token, because every cask action sends it to brew and brew rejects "Claude Code".
func TestCaskNameIsTokenInRealBrewShape(t *testing.T) {
	raw, err := os.ReadFile("testdata/casks_installed_token.json")
	if err != nil {
		t.Fatal(err)
	}
	var data casksJSON
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Casks) != 1 {
		t.Fatalf("fixture has %d casks, want 1", len(data.Casks))
	}
	c := parseCask(data.Casks[0])
	if c.Name != "claude-code" {
		t.Fatalf("Name = %q, want the token %q", c.Name, "claude-code")
	}
	if c.DisplayName != "Claude Code" {
		t.Fatalf("DisplayName = %q, want %q", c.DisplayName, "Claude Code")
	}
}

// TestCaskNameFallsBackToDisplayName: a payload with no token (an older brew
// shape) keeps the first display name as the Name, as it always did.
func TestCaskNameFallsBackToDisplayName(t *testing.T) {
	raw := `{"casks":[{"name":["Google Chrome"],"version":"120.0"}]}`
	var data casksJSON
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	c := parseCask(data.Casks[0])
	if c.Name != "Google Chrome" {
		t.Fatalf("Name = %q, want the display name fallback %q", c.Name, "Google Chrome")
	}
	if c.DisplayName != "Google Chrome" {
		t.Fatalf("DisplayName = %q, want %q", c.DisplayName, "Google Chrome")
	}
}
