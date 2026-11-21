package brew

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

func TestSearchServiceSearch(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		// Each typed read answers with its own rows, as brew does.
		if args[1] == "--formula" {
			return []byte("neovim\nneovim-qt\n"), nil
		}
		return []byte("neovide\n"), nil
	}
	svc := NewSearchService(r)

	results, err := svc.Search(context.Background(), "neovim")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	f := results[0]
	if f.Name != "neovim" {
		t.Errorf("Name = %q, want neovim", f.Name)
	}
	if !f.IsFormula {
		t.Error("neovim should be a formula")
	}

	cask := results[2]
	if cask.Name != "neovide" {
		t.Errorf("Name = %q, want neovide", cask.Name)
	}
	if !cask.IsCask {
		t.Error("neovide should be a cask")
	}
}

// TestSearchKeepsTheReadThatWorked (lb19 AC-02, Q-lb19-search-partial): when one
// typed read fails, the rows of the other read are kept. The search errors only
// when both reads fail.
func TestSearchKeepsTheReadThatWorked(t *testing.T) {
	for _, tc := range []struct{ failFlag, want string }{{"--cask", "jq"}, {"--formula", "neovide"}} {
		r := brewtest.NewMockRunner()
		r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
			switch {
			case args[1] == tc.failFlag:
				return nil, brewtest.ErrTest
			case args[1] == "--formula":
				return []byte("jq\n"), nil
			}
			return []byte("neovide\n"), nil
		}
		results, err := NewSearchService(r).Search(context.Background(), "x")
		if err != nil {
			t.Fatalf("%s failed: Search err = %v, want the other read's rows", tc.failFlag, err)
		}
		if len(results) != 1 || results[0].Name != tc.want {
			t.Errorf("%s failed: results = %+v, want only %s", tc.failFlag, results, tc.want)
		}
	}

	both := brewtest.NewMockRunner()
	both.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, brewtest.ErrTest
	}
	if _, err := NewSearchService(both).Search(context.Background(), "x"); err == nil {
		t.Error("both reads failed: Search err = nil, want an error")
	}
}

// TestSearchErrorsWhenNoRowsAndAReadFailed (LB-20 a): a read that finds nothing
// (exit 1) and a read that fails leave no rows. The failure must be the error,
// not an empty list that reads as "no results".
func TestSearchErrorsWhenNoRowsAndAReadFailed(t *testing.T) {
	for _, tc := range []struct{ failFlag, emptyFlag string }{{"--cask", "--formula"}, {"--formula", "--cask"}} {
		r := brewtest.NewMockRunner()
		r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
			if args[1] == tc.failFlag {
				return nil, brewtest.ErrTest
			}
			return nil, &BrewExitError{Command: "search", ExitCode: 1}
		}
		results, err := NewSearchService(r).Search(context.Background(), "x")
		if !errors.Is(err, brewtest.ErrTest) {
			t.Errorf("%s failed, %s is empty: Search err = %v, want the failed read's error (results %+v)",
				tc.failFlag, tc.emptyFlag, err, results)
		}
	}
}

func TestSearchServiceEmpty(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("==> Formulae\n\n==> Casks"), nil
	}
	svc := NewSearchService(r)

	results, err := svc.Search(context.Background(), "xyznonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestParseSearchOutput(t *testing.T) {
	raw := "==> Formulae\nfzf\nripgrep\n\n==> Casks\nfirefox\nspotify"
	results := parseSearchOutput(raw)
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	if results[0].Name != "fzf" || !results[0].IsFormula {
		t.Error("fzf should be a formula")
	}
	if results[2].Name != "firefox" || !results[2].IsCask {
		t.Error("firefox should be a cask")
	}
}

func TestParseSearchOutputFormulaeOnly(t *testing.T) {
	raw := "fzf\nripgrep"
	results := parseSearchOutput(raw)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Headerless output is UNKNOWN, not formula: defaulting to formula
	// miscategorized entries in odd outputs (r1 finding 21).
	if results[0].IsFormula || results[0].IsCask {
		t.Errorf("headerless entries should be unknown, got IsFormula=%v IsCask=%v",
			results[0].IsFormula, results[0].IsCask)
	}
}

func TestParseSearchOutputCasksOnlyHeader(t *testing.T) {
	// Regression for r1 finding 21: with ONLY a "==> Casks" header (no
	// formulae section), entries used to be labeled formula because
	// seenFormulae stayed false.
	raw := "==> Casks\nfirefox"
	results := parseSearchOutput(raw)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].IsCask {
		t.Errorf("expected cask entry, got IsFormula=%v IsCask=%v", results[0].IsFormula, results[0].IsCask)
	}
	if results[0].IsFormula {
		t.Error("cask-section entry must not be labeled formula")
	}
}

// The real headerless shape of `brew search <q>` without a TTY: the formulae,
// a blank line, then the casks. Each typed read carries its own half.
func TestSearchServiceTypesRowsByReadFlag(t *testing.T) {
	const formulaRead = "claude-code-router\nclaude-code-templates\n"
	const caskRead = "claude-code\nclaude-code@latest\n"
	var calls [][]string
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		calls = append(calls, args)
		if args[1] == "--formula" {
			return []byte(formulaRead), nil
		}
		return []byte(caskRead), nil
	}

	results, err := NewSearchService(r).Search(context.Background(), "claude-code")
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || !slices.Equal(calls[0], []string{"search", "--formula", "claude-code"}) ||
		!slices.Equal(calls[1], []string{"search", "--cask", "claude-code"}) {
		t.Fatalf("calls = %v, want search --formula then search --cask", calls)
	}
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4: %+v", len(results), results)
	}
	wantFormula := map[string]bool{"claude-code-router": true, "claude-code-templates": true, "claude-code": false, "claude-code@latest": false}
	for _, res := range results {
		if res.IsFormula != wantFormula[res.Name] || res.IsCask == wantFormula[res.Name] {
			t.Errorf("%s: IsFormula=%v IsCask=%v, want formula=%v", res.Name, res.IsFormula, res.IsCask, wantFormula[res.Name])
		}
	}
}

// A blank line in a read's output is not a row.
func TestSearchServiceSkipsBlankLines(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if args[1] == "--formula" {
			return []byte("\nfzf\n\n"), nil
		}
		return []byte(""), nil
	}
	results, err := NewSearchService(r).Search(context.Background(), "fzf")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "fzf" || !results[0].IsFormula {
		t.Fatalf("results = %+v, want one formula fzf", results)
	}
}

func TestMarkInstalledMatchesByTypeAndTapSuffix(t *testing.T) {
	results := []SearchResult{
		{Name: "jq", IsFormula: true},
		{Name: "homebrew/core/wget", IsFormula: true},
		{Name: "claude-code", IsCask: true},
		{Name: "homebrew/cask/firefox", IsCask: true},
		{Name: "ripgrep", IsFormula: true},
		{Name: "iterm2", IsCask: true},
	}
	// A cask named jq must not mark the jq formula row, and the reverse.
	formulae := []Formula{{Name: "jq"}, {Name: "wget"}}
	casks := []Cask{{Name: "claude-code"}, {Name: "firefox"}, {Name: "jq"}}

	got := MarkInstalled(results, formulae, casks)
	want := []bool{true, true, true, true, false, false}
	for i, r := range got {
		if r.Installed != want[i] {
			t.Errorf("%s (formula=%v cask=%v): Installed = %v, want %v", r.Name, r.IsFormula, r.IsCask, r.Installed, want[i])
		}
	}
}

func TestMarkInstalledNilListsMarksNothing(t *testing.T) {
	got := MarkInstalled([]SearchResult{{Name: "jq", IsFormula: true}}, nil, nil)
	if got[0].Installed {
		t.Error("a missing installed list marked a row")
	}
}
