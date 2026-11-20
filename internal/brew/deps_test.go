package brew

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// TestBashShowsRealRuntimeDependencies locks AC-07.3: brew reports an
// installed formula's runtime dependencies as objects with full_name,
// version and more. testdata/info_bash_fish.json is a trimmed copy of real
// `brew info --json=v2 bash fish` output.
func TestBashShowsRealRuntimeDependencies(t *testing.T) {
	raw, err := os.ReadFile("testdata/info_bash_fish.json")
	if err != nil {
		t.Fatal(err)
	}
	var data formulaeJSON
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	formulae := make([]Formula, 0, len(data.Formulae))
	for _, fj := range data.Formulae {
		formulae = append(formulae, parseFormula(fj))
	}

	bash := formulae[0]
	if want := []string{"ncurses", "readline"}; !reflect.DeepEqual(bash.Dependencies, want) {
		t.Fatalf("bash Dependencies = %v, want %v", bash.Dependencies, want)
	}
	fish := formulae[1]
	if want := []string{"pcre2"}; !reflect.DeepEqual(fish.Dependencies, want) {
		t.Fatalf("uninstalled fish Dependencies = %v, want %v", fish.Dependencies, want)
	}
}

func TestFlexStringsReadsDependencyObjects(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"strings", `["ncurses", "readline"]`, []string{"ncurses", "readline"}},
		{"objects with full_name", `[{"full_name":"ncurses","version":"6.6"}]`, []string{"ncurses"}},
		{"objects with name", `[{"name":"pcre2"}]`, []string{"pcre2"}},
		{"full_name wins over name", `[{"full_name":"homebrew/core/tool","name":"tool"}]`, []string{"homebrew/core/tool"}},
		{"mixed", `["a", {"full_name":"b"}, {"name":"c"}]`, []string{"a", "b", "c"}},
		{"single object", `{"full_name":"solo"}`, []string{"solo"}},
		{"cask depends_on object", `{"formula":["x"],"cask":["y"]}`, []string{}},
		{"null element skipped", `[null, "a", 3]`, []string{"a"}},
		{"empty array", `[]`, []string{}},
		{"null reads as none", `null`, []string{}},
		{"object with no name", `[{"version":"1"}]`, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var f flexStrings
			if err := json.Unmarshal([]byte(c.raw), &f); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual([]string(f), c.want) {
				t.Fatalf("got %v, want %v", []string(f), c.want)
			}
		})
	}
}
