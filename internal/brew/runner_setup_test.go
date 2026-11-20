package brew

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// setHomebrewPrefix sets HOMEBREW_PREFIX for the test, or unsets it when prefix
// is "". The original value comes back when the test ends.
func setHomebrewPrefix(t *testing.T, prefix string) {
	t.Helper()
	t.Setenv("HOMEBREW_PREFIX", prefix)
	if prefix == "" {
		os.Unsetenv("HOMEBREW_PREFIX")
	}
}

// TestFindBrewPathUsesPrefixWhenSet: a HOMEBREW_PREFIX that holds a brew is
// the one found.
func TestFindBrewPathUsesPrefixWhenSet(t *testing.T) {
	prefix := t.TempDir()
	if err := os.MkdirAll(filepath.Join(prefix, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	brew := filepath.Join(prefix, "bin", "brew")
	if err := os.WriteFile(brew, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	setHomebrewPrefix(t, prefix)

	got, err := findBrewPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != brew {
		t.Errorf("findBrewPath() = %q, want %q", got, brew)
	}
}

// TestBrewCandidatesNamePrefixOnlyWhenSet: the HOMEBREW_PREFIX candidate is
// there only when the variable is set. An unset prefix must not add a bare
// "/bin/brew" to the list, nor to the list BrewNotFoundError reports.
func TestBrewCandidatesNamePrefixOnlyWhenSet(t *testing.T) {
	fixed := []string{
		"/opt/homebrew/bin/brew",
		"/usr/local/bin/brew",
		"/home/linuxbrew/.linuxbrew/bin/brew",
	}
	prefix := t.TempDir()

	cases := []struct {
		name   string
		prefix string
		want   []string
	}{
		{"prefix unset", "", fixed},
		{"prefix set", prefix, append([]string{prefix + "/bin/brew"}, fixed...)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setHomebrewPrefix(t, tc.prefix)
			if got := brewCandidates(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("brewCandidates() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestExecuteAndExecuteStreamShareCommandSetup: both paths run brew with
// auto-update and prompts off, and with an empty stdin.
func TestExecuteAndExecuteStreamShareCommandSetup(t *testing.T) {
	r := writeFakeBrew(t, `echo "$HOMEBREW_NO_AUTO_UPDATE $HOMEBREW_NO_ASK"
read -r line
echo "stdin:$line"
`)
	want := "1 1\nstdin:\n"

	out, err := r.Execute(context.Background(), "info")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != want {
		t.Errorf("Execute output = %q, want %q", out, want)
	}

	lines, errs := r.ExecuteStream(context.Background(), "info")
	var got []string
	for line := range lines {
		got = append(got, line)
	}
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"1 1", "stdin:"}) {
		t.Errorf("ExecuteStream lines = %q, want [1 1 stdin:]", got)
	}
}
