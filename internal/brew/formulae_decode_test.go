package brew

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// formulaeFixture is brew's `info --json=v2 --installed` output for n formulae.
// Each one has runtime dependencies, as brew writes them: objects with
// full_name, or with name, plus a bottle with three files.
func formulaeFixture(n int) []byte {
	entries := make([]string, 0, n)
	for i := 0; i < n; i++ {
		entries = append(entries, fmt.Sprintf(`{
  "name": "pkg%[1]d", "full_name": "homebrew/core/pkg%[1]d", "tap": "homebrew/core",
  "versions": {"stable": "1.%[1]d", "head": null},
  "desc": "package %[1]d", "homepage": "https://example.com/%[1]d", "license": "MIT",
  "installed": [{"version": "1.%[1]d", "installed_on_request": true, "installed_as_dependency": false, "time": 1700000000,
    "runtime_dependencies": [
      {"full_name": "homebrew/core/lib%[1]da", "version": "1.0"},
      {"full_name": "homebrew/core/lib%[1]db", "version": "2.0"},
      {"name": "lib%[1]dc", "version": "3.0"}]}],
  "dependencies": ["lib%[1]da", "lib%[1]db"],
  "build_dependencies": ["cmake"],
  "bottle": {"stable": {"files": {
    "arm64_sequoia": {"url": "https://example.com/arm64_sequoia/%[1]d", "sha256": "aa"},
    "arm64_sonoma": {"url": "https://example.com/arm64_sonoma/%[1]d", "sha256": "bb"},
    "x86_64_linux": {"url": "https://example.com/x86_64_linux/%[1]d", "sha256": "cc"}}}},
  "pinned": false, "keg_only": false}`, i))
	}
	return []byte(`{"formulae": [` + strings.Join(entries, ",") + `]}`)
}

// TestFormulaeFixtureDecodes: the fixture decodes to the formulae it describes,
// with both shapes of runtime dependency read as names.
func TestFormulaeFixtureDecodes(t *testing.T) {
	var decoded formulaeJSON
	if err := json.Unmarshal(formulaeFixture(300), &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Formulae) != 300 {
		t.Fatalf("decoded %d formulae, want 300", len(decoded.Formulae))
	}
	f := parseFormula(decoded.Formulae[7])
	wantDeps := []string{"homebrew/core/lib7a", "homebrew/core/lib7b", "lib7c"}
	if fmt.Sprint(f.Dependencies) != fmt.Sprint(wantDeps) {
		t.Errorf("runtime dependencies = %v, want %v", f.Dependencies, wantDeps)
	}
	if !f.Bottled {
		t.Error("expected the formula to read as bottled")
	}
}

// BenchmarkDecodeFormulaeList decodes and parses the 300-formula list, as
// FormulaeReader.List does on every cache miss.
func BenchmarkDecodeFormulaeList(b *testing.B) {
	data := formulaeFixture(300)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		var decoded formulaeJSON
		if err := json.Unmarshal(data, &decoded); err != nil {
			b.Fatal(err)
		}
		formulae := make([]Formula, 0, len(decoded.Formulae))
		for _, f := range decoded.Formulae {
			formulae = append(formulae, parseFormula(f))
		}
	}
}
