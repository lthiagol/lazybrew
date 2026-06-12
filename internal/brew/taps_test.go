package brew

import (
	"context"
	"testing"
	"time"
)

func TestTapsServiceList(t *testing.T) {
	r := NewMockRunner()
	tapCalls := 0
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		tapCalls++
		if len(args) == 1 && args[0] == "tap" {
			return []byte("homebrew/core\nhomebrew/cask\nnicknisi/tap\n"), nil
		}
		if args[0] == "tap-info" {
			return []byte(`[{
				"name": "` + args[len(args)-1] + `",
				"remote": "https://github.com/example/tap.git",
				"formula_count": 10,
				"cask_count": 5,
				"command_count": 1,
				"installed": true,
				"api": false,
				"trusted": true,
				"formula_names": ["foo"],
				"cask_names": ["bar"]
			}]`), nil
		}
		return []byte(`[]`), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsService(r, cache)

	list, err := taps.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 taps, got %d", len(list))
	}

	core := list[0]
	if core.Name != "homebrew/core" {
		t.Errorf("Name = %q, want homebrew/core", core.Name)
	}
	if !core.IsOfficial {
		t.Error("homebrew/core should be official")
	}
	if !core.Trusted {
		t.Error("homebrew/core should be trusted")
	}

	third := list[2]
	if third.Name != "nicknisi/tap" {
		t.Errorf("Name = %q, want nicknisi/tap", third.Name)
	}
	if third.IsOfficial {
		t.Error("nicknisi/tap should not be official")
	}
	if !third.Trusted {
		t.Error("nicknisi/tap should be trusted")
	}
}

func TestTapsServiceTap(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsService(r, cache)

	if err := taps.Tap(context.Background(), "some-org/formulas"); err != nil {
		t.Fatal(err)
	}
}

func TestTapsServiceTapWithURL(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsService(r, cache)

	if err := taps.TapWithURL(context.Background(), "custom/tap", "https://example.com/tap.git"); err != nil {
		t.Fatal(err)
	}
}

func TestTapsServiceUntap(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsService(r, cache)

	if err := taps.Untap(context.Background(), "nicknisi/tap"); err != nil {
		t.Fatal(err)
	}
}

func TestTapsServiceRepair(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsService(r, cache)

	if err := taps.Repair(context.Background(), "nicknisi/tap"); err != nil {
		t.Fatal(err)
	}
}

func TestTapsServiceGet(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`[{
			"name": "homebrew/core",
			"remote": "https://github.com/Homebrew/homebrew-core.git",
			"formula_count": 7000,
			"cask_count": 0,
			"command_count": 0,
			"installed": true,
			"api": true,
			"trusted": true
		}]`), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsService(r, cache)

	tap, err := taps.Get(context.Background(), "homebrew/core")
	if err != nil {
		t.Fatal(err)
	}
	if tap == nil {
		t.Fatal("expected tap")
	}
	if !tap.IsOfficial {
		t.Error("homebrew/core should be official")
	}
	if !tap.Trusted {
		t.Error("homebrew/core should be trusted")
	}
	if tap.FormulaCount != 7000 {
		t.Errorf("FormulaCount = %d, want 7000", tap.FormulaCount)
	}
}
