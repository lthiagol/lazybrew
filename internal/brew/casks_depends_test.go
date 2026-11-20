package brew

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// TestCaskDependsOnRealBrewShape (v2-05 AC-02.2, ruling Q-v2-cask-deps):
// testdata/casks_installed.json is real `brew info --json=v2 --cask
// --installed` output. gcloud-cli's depends_on is brew's object shape, which
// the parser used to read as empty.
func TestCaskDependsOnRealBrewShape(t *testing.T) {
	raw, err := os.ReadFile("testdata/casks_installed.json")
	if err != nil {
		t.Fatal(err)
	}
	var data casksJSON
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	casks := make(map[string]Cask, len(data.Casks))
	for _, cj := range data.Casks {
		c := parseCask(cj)
		casks[c.Name] = c
	}

	gcloud := casks["Google Cloud CLI"]
	if want := []string{"python@3.14"}; !reflect.DeepEqual(gcloud.DependsOnFormulae, want) {
		t.Fatalf("Google Cloud CLI DependsOnFormulae = %v, want %v", gcloud.DependsOnFormulae, want)
	}
	if want := []string{"python@3.14"}; !reflect.DeepEqual(gcloud.DependsOn, want) {
		t.Fatalf("Google Cloud CLI DependsOn = %v, want %v", gcloud.DependsOn, want)
	}
	if len(gcloud.DependsOnCasks) != 0 || len(gcloud.Requirements) != 0 {
		t.Fatalf("Google Cloud CLI has no cask or requirement deps, got casks %v, requirements %v",
			gcloud.DependsOnCasks, gcloud.Requirements)
	}

	// An empty object is a cask with no dependencies, not an error.
	claude, ok := casks["Claude Code"]
	if !ok {
		t.Fatal("Claude Code missing from the fixture")
	}
	if len(claude.DependsOn) != 0 || len(claude.DependsOnFormulae) != 0 {
		t.Fatalf("Claude Code depends_on {} should read as none, got %v", claude.DependsOn)
	}
}

// TestCaskUnmarshalReadsTheObjectShape: a Cask decoded straight from JSON
// (as parsePackageInfo does) takes the object shape too, and does not fail.
func TestCaskUnmarshalReadsTheObjectShape(t *testing.T) {
	raw := `{"name":"tool","depends_on":{"formula":["git"],"cask":["docker"],` +
		`"macos":{">=":["13"]},"arch":["arm64"]}}`
	var c Cask
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if want := []string{"git"}; !reflect.DeepEqual(c.DependsOnFormulae, want) {
		t.Errorf("DependsOnFormulae = %v, want %v", c.DependsOnFormulae, want)
	}
	if want := []string{"docker"}; !reflect.DeepEqual(c.DependsOnCasks, want) {
		t.Errorf("DependsOnCasks = %v, want %v", c.DependsOnCasks, want)
	}
	if want := []string{"git", "docker"}; !reflect.DeepEqual(c.DependsOn, want) {
		t.Errorf("DependsOn = %v, want %v", c.DependsOn, want)
	}
	if want := []string{"macOS ≥ 13", "arch arm64"}; !reflect.DeepEqual(c.Requirements, want) {
		t.Errorf("Requirements = %v, want %v", c.Requirements, want)
	}
}

func TestParseCaskDependsShapes(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want caskDepends
	}{
		{"absent", ``, caskDepends{}},
		{"empty object", `{}`, caskDepends{}},
		{"legacy list is formulae", `["a","b"]`, caskDepends{Formulae: []string{"a", "b"}}},
		{"macos list", `{"macos":["monterey","ventura"]}`,
			caskDepends{Requirements: []string{"macOS monterey or ventura"}}},
		{"macos comparison, sorted", `{"macos":{"<=":["14"],">=":["12"]}}`,
			caskDepends{Requirements: []string{"macOS ≤ 14", "macOS ≥ 12"}}},
		{"arch objects", `{"arch":[{"type":"arm","bits":64},{"type":"intel","bits":64}]}`,
			caskDepends{Requirements: []string{"arch arm64 or x86_64"}}},
		{"unknown shape reads as none", `"just a string"`, caskDepends{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parseCaskDepends(json.RawMessage(c.raw))
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}
