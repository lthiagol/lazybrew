package brew

import (
	"context"
	"testing"
)

func TestSearchServiceSearch(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("==> Formulae\nneovim\nneovim-qt\n\n==> Casks\nneovide"), nil
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

func TestSearchServiceEmpty(t *testing.T) {
	r := NewMockRunner()
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

func TestSearchServiceSearchDesc(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "--desc" {
			return []byte("ripgrep"), nil
		}
		return []byte("==> Formulae\nripgrep"), nil
	}
	svc := NewSearchService(r)

	results, err := svc.SearchDesc(context.Background(), "search")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("results: %+v", results)
	if len(results) != 1 || results[0].Name != "ripgrep" {
		t.Errorf("got %v, want [ripgrep]", results)
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
	if !results[0].IsFormula {
		t.Error("should default to formula when no section headers")
	}
}
