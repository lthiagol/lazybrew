package brew

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// cleanupDryRun is `brew cleanup -n` in brew's own shape (Library/Homebrew/
// cleanup.rb: "Would remove:", "Would prune", the broken-link and empty-directory
// lines, the autoremove section, and the closing total). The paths are made up.
const cleanupDryRun = `Would remove: /home/linuxbrew/.linuxbrew/Cellar/old-formula/1.0 (3 files, 230.3MB)
Would remove: /home/user/.cache/Homebrew/downloads/0a1b-foo-1.0.tar.gz (48.2MB)
Would prune 12 files from: /home/user/.cache/Homebrew/api
Would remove (broken link): /home/linuxbrew/.linuxbrew/bin/gone
Would remove (empty directory): /home/linuxbrew/.linuxbrew/var/empty
==> Would autoremove 2 unneeded formulae:
cmake
openssl@3
==> This operation would free approximately 950MB of disk space.
`

// cleanupRunner answers `brew cleanup -n` with out and err, and records the
// arguments of each call.
func cleanupRunner(out string, err error, calls *[][]string) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		*calls = append(*calls, args)
		return []byte(out), err
	}
	return r
}

func TestCleanupPreviewParsesItemsAndTotal(t *testing.T) {
	var calls [][]string
	diag := NewDiagnosticsReader(cleanupRunner(cleanupDryRun, nil, &calls), NewCache(time.Minute))

	p, err := diag.CleanupPreview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []CleanupItem{
		{Path: "/home/linuxbrew/.linuxbrew/Cellar/old-formula/1.0", Files: 3, Size: "230.3MB", Bytes: 230300000},
		{Path: "/home/user/.cache/Homebrew/downloads/0a1b-foo-1.0.tar.gz", Files: 1, Size: "48.2MB", Bytes: 48200000},
		{Path: "/home/user/.cache/Homebrew/api", Files: 12},
		{Path: "/home/linuxbrew/.linuxbrew/bin/gone", Kind: "broken link"},
		{Path: "/home/linuxbrew/.linuxbrew/var/empty", Kind: "empty directory"},
	}
	if !reflect.DeepEqual(p.Items, want) {
		t.Fatalf("items:\n got %+v\nwant %+v", p.Items, want)
	}
	if p.Total != "950MB" {
		t.Errorf("Total = %q, want 950MB", p.Total)
	}
	if wantUnused := []string{"cmake", "openssl@3"}; !reflect.DeepEqual(p.Unused, wantUnused) {
		t.Errorf("Unused = %v, want %v", p.Unused, wantUnused)
	}
	if len(p.Raw) != 0 {
		t.Errorf("Raw = %v, want none: every line of the fixture is known", p.Raw)
	}
}

func TestCleanupPreviewRunsOnlyTheDryRun(t *testing.T) {
	var calls [][]string
	diag := NewDiagnosticsReader(cleanupRunner("", nil, &calls), NewCache(time.Minute))

	if _, err := diag.CleanupPreview(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"cleanup", "-n"}}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("brew calls = %v, want %v", calls, want)
	}
}

// Nothing to clean: brew prints nothing and exits 0.
func TestCleanupPreviewEmpty(t *testing.T) {
	var calls [][]string
	diag := NewDiagnosticsReader(cleanupRunner("", nil, &calls), NewCache(time.Minute))

	p, err := diag.CleanupPreview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 0 || p.Total != "" || len(p.Unused) != 0 || len(p.Raw) != 0 {
		t.Fatalf("empty output should give the zero preview, got %+v", p)
	}
}

func TestCleanupPreviewKeepsUnknownLines(t *testing.T) {
	var calls [][]string
	out := "Would `brew cleanup`:\nSomething brew says in a new shape\n"
	diag := NewDiagnosticsReader(cleanupRunner(out, nil, &calls), NewCache(time.Minute))

	p, err := diag.CleanupPreview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 0 {
		t.Errorf("Items = %+v, want none", p.Items)
	}
	want := []string{"Would `brew cleanup`:", "Something brew says in a new shape"}
	if !reflect.DeepEqual(p.Raw, want) {
		t.Errorf("Raw = %q, want %q", p.Raw, want)
	}
}

// A failed dry run is an error. Exit 1 with a listing is a result: brew exits 1
// when a keg cannot be removed and still lists what it would do.
func TestCleanupPreviewExitCodes(t *testing.T) {
	cases := []struct {
		name    string
		out     string
		err     error
		wantErr bool
	}{
		{"exit 2 with no output is an error", "", &BrewExitError{Command: "cleanup", ExitCode: 2}, true},
		{"exit 1 with no output is an error", "", &BrewExitError{Command: "cleanup", ExitCode: 1}, true},
		{"exit 1 with a listing is a result", cleanupDryRun, &BrewExitError{Command: "cleanup", ExitCode: 1}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls [][]string
			diag := NewDiagnosticsReader(cleanupRunner(tc.out, tc.err, &calls), NewCache(time.Minute))
			p, err := diag.CleanupPreview(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && len(p.Items) != 5 {
				t.Errorf("Items = %d, want 5", len(p.Items))
			}
		})
	}
}

func TestAutoremovePreviewNames(t *testing.T) {
	var calls [][]string
	out := "==> Would autoremove 2 unneeded formulae:\nbar\nfoo\n"
	diag := NewDiagnosticsReader(cleanupRunner(out, nil, &calls), NewCache(time.Minute))

	p, err := diag.AutoremovePreview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"autoremove", "-n"}}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("brew calls = %v, want %v", calls, want)
	}
	if want := []string{"bar", "foo"}; !reflect.DeepEqual(p.Formulae, want) {
		t.Errorf("Formulae = %v, want %v", p.Formulae, want)
	}
}

func TestAutoremovePreviewSingularAndEmpty(t *testing.T) {
	var calls [][]string
	diag := NewDiagnosticsReader(cleanupRunner("==> Would autoremove 1 unneeded formula:\nfoo\n", nil, &calls), NewCache(time.Minute))
	p, err := diag.AutoremovePreview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"foo"}; !reflect.DeepEqual(p.Formulae, want) {
		t.Errorf("singular: Formulae = %v, want %v", p.Formulae, want)
	}

	// brew prints no heading when nothing is unused.
	diag = NewDiagnosticsReader(cleanupRunner("", nil, &calls), NewCache(time.Minute))
	p, err = diag.AutoremovePreview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Formulae) != 0 || len(p.Raw) != 0 {
		t.Errorf("empty: got %+v, want nothing", p)
	}
}

func TestAutoremovePreviewKeepsUnknownLines(t *testing.T) {
	var calls [][]string
	out := "==> Would autoremove 1 unneeded formula:\nfoo\nA line with spaces\n"
	diag := NewDiagnosticsReader(cleanupRunner(out, nil, &calls), NewCache(time.Minute))
	p, err := diag.AutoremovePreview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"A line with spaces"}; !reflect.DeepEqual(p.Raw, want) {
		t.Errorf("Raw = %q, want %q", p.Raw, want)
	}
}

func TestParseSize(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"230.3MB", 230300000, true},
		{"512B", 512, true},
		{"1.2GB", 1200000000, true},
		{"45KB", 45000, true},
		{"bad", 0, false},
		{"", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseSize(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("parseSize(%q) = %d, %v; want %d, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
