package brew

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
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

// TestOutdatedAndInstalledCaskNamesAgree (LB-12): the Outdated and Casks lists
// both name a cask by its token, so a row in one finds the same cask in the other.
func TestOutdatedAndInstalledCaskNamesAgree(t *testing.T) {
	raw, err := os.ReadFile("testdata/casks_installed_token.json")
	if err != nil {
		t.Fatal(err)
	}
	_, installed, err := ParseInfo(raw)
	if err != nil {
		t.Fatal(err)
	}

	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{"casks":[{"name":"claude-code","full_name":"claude-code","installed_versions":["2.1.310"],"current_version":"2.1.311"}]}`), nil
	}
	outdated, err := NewCasksReader(r, NewCache(time.Minute)).Outdated(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(outdated) != 1 || len(installed) != 1 {
		t.Fatalf("want one outdated and one installed cask, got %d and %d", len(outdated), len(installed))
	}
	if outdated[0].Name != installed[0].Name {
		t.Fatalf("Outdated names %q, Casks names %q: they must match", outdated[0].Name, installed[0].Name)
	}
}
