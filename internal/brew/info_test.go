package brew

import (
	"testing"
	"time"
)

// ParseInfo reads a formula and a cask out of one payload with the readers'
// parsers, so a search info decode gets the same fields the lists get.
func TestParseInfoReadsFormulaAndCask(t *testing.T) {
	raw := `{
	  "formulae": [{
	    "name": "jq", "full_name": "jq", "tap": "homebrew/core",
	    "versions": {"stable": "1.8.2"},
	    "bottle": {"stable": {"files": {"arm64_linux": {"url": "u"}}}},
	    "installed": [{"version": "1.8.2", "time": 1000, "installed_on_request": true}]
	  }],
	  "casks": [{"token": "tool", "version": "2.0", "installed": "1.9"}]
	}`
	formulae, casks, err := ParseInfo([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(formulae) != 1 || len(casks) != 1 {
		t.Fatalf("got %d formulae and %d casks, want 1 and 1", len(formulae), len(casks))
	}
	f := formulae[0]
	if f.Version != "1.8.2" || !f.Bottled || f.InstalledOn.Unix() != 1000 || !f.InstalledOnReq {
		t.Errorf("formula = %+v", f)
	}
	c := casks[0]
	if c.Version != "1.9" {
		t.Errorf("cask Version = %q, want the installed 1.9", c.Version)
	}
	if len(c.Installed) != 1 || c.Installed[0].Version != "1.9" {
		t.Errorf("cask Installed = %+v, want one entry for 1.9", c.Installed)
	}
}

func TestParseInfoFormulaOnlyPayloadHasNoCasks(t *testing.T) {
	formulae, casks, err := ParseInfo([]byte(`{"formulae": [{"name": "jq"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(formulae) != 1 || len(casks) != 0 {
		t.Fatalf("got %d formulae and %d casks, want 1 and 0", len(formulae), len(casks))
	}
	if formulae[0].InstalledOn != (time.Time{}) {
		t.Errorf("a formula with no installed entry must have no InstalledOn, got %v", formulae[0].InstalledOn)
	}
}

func TestParseInfoRejectsBadJSON(t *testing.T) {
	if _, _, err := ParseInfo([]byte(`{"formulae": [`)); err == nil {
		t.Fatal("bad JSON must be an error")
	}
}
