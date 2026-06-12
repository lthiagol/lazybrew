package brew

import (
	"context"
	"testing"
)

func TestSearchServiceSearch(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		data := result.(*searchJSON)
		*data = searchJSON{
			Formulae: []searchItemJSON{
				{Name: "neovim", FullName: "neovim", Description: "Vim-fork focused on extensibility", Installed: []interface{}{map[string]interface{}{"version": "0.10.4"}}},
				{Name: "neovim-qt", FullName: "neovim-qt", Description: "Neovim client library and GUI"},
			},
			Casks: []searchItemJSON{
				{Name: "neovide", FullName: "neovide", Description: "Neovim client in Rust"},
			},
		}
		return nil
	}
	svc := NewSearchService(r)

	results, err := svc.Search(context.Background(), "neovim")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	installed := results[0]
	if installed.Name != "neovim" {
		t.Errorf("Name = %q, want neovim", installed.Name)
	}
	if !installed.IsFormula {
		t.Error("neovim should be a formula")
	}
	if !installed.Installed {
		t.Error("neovim should be marked installed")
	}
	if installed.Version != "0.10.4" {
		t.Errorf("Version = %q, want 0.10.4", installed.Version)
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
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		data := result.(*searchJSON)
		*data = searchJSON{}
		return nil
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
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		data := result.(*searchJSON)
		*data = searchJSON{
			Formulae: []searchItemJSON{
				{Name: "ripgrep", Description: "Search tool like grep but faster"},
			},
		}
		return nil
	}
	svc := NewSearchService(r)

	results, err := svc.SearchDesc(context.Background(), "search")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "ripgrep" {
		t.Errorf("got %v, want [ripgrep]", results)
	}
}
