package brew

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTapsServiceList(t *testing.T) {
	r := NewMockRunner()
	var tapListCalls, tapInfoCalls int
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "tap" {
			tapListCalls++
			return []byte("homebrew/core\nhomebrew/cask\nnicknisi/tap\n"), nil
		}
		if args[0] == "tap-info" {
			tapInfoCalls++
			// M11: batch tap-info — args[2:] are the tap names requested.
			names := args[2:]
			entries := make([]string, 0, len(names))
			for _, name := range names {
				entries = append(entries, `{
					"name": "`+name+`",
					"remote": "https://github.com/example/`+name+`.git",
					"formula_count": 10,
					"cask_count": 5,
					"command_count": 1,
					"installed": true,
					"api": false,
					"trusted": true,
					"formula_names": ["foo-from-`+name+`"],
					"cask_names": ["bar-from-`+name+`"]
				}`)
			}
			return []byte("[" + strings.Join(entries, ",") + "]"), nil
		}
		return []byte(`[]`), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsReader(r, cache)

	list, err := taps.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 taps, got %d", len(list))
	}

	// M11 AC-01: exactly one `brew tap` + one batched `brew tap-info --json`.
	if tapListCalls != 1 {
		t.Errorf("expected exactly 1 'brew tap' call, got %d", tapListCalls)
	}
	if tapInfoCalls != 1 {
		t.Errorf("expected exactly 1 batched 'brew tap-info' call, got %d", tapInfoCalls)
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
	if len(core.FormulaNames) != 1 || core.FormulaNames[0] != "foo-from-homebrew/core" {
		t.Errorf("FormulaNames = %v, want [foo-from-homebrew/core]", core.FormulaNames)
	}
	if len(core.CaskNames) != 1 || core.CaskNames[0] != "bar-from-homebrew/core" {
		t.Errorf("CaskNames = %v, want [bar-from-homebrew/core]", core.CaskNames)
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

// TestTapsBatchLoadIsSingleInvocation is the explicit M11 AC-01 reproducer:
// a list of N taps must result in exactly one `brew tap` plus one batched
// `brew tap-info --json name1 name2 …`. Per-tap loops would scale linearly;
// the batch path stays constant.
func TestTapsBatchLoadIsSingleInvocation(t *testing.T) {
	const taps = 12

	r := NewMockRunner()
	var tapListCalls, tapInfoCalls int
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "tap" {
			tapListCalls++
			names := make([]string, taps)
			for i := range names {
				names[i] = "thirdparty/tap"
			}
			return []byte(strings.Join(names, "\n") + "\n"), nil
		}
		if args[0] == "tap-info" {
			tapInfoCalls++
			entries := make([]string, 0, len(args)-2)
			for _, name := range args[2:] {
				entries = append(entries, `{"name":"`+name+`","installed":true,"trusted":true}`)
			}
			return []byte("[" + strings.Join(entries, ",") + "]"), nil
		}
		return []byte(`[]`), nil
	}
	cache := NewCache(time.Minute)

	list, err := NewTapsReader(r, cache).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != taps {
		t.Fatalf("expected %d taps, got %d", taps, len(list))
	}
	if tapListCalls != 1 {
		t.Errorf("expected 1 'brew tap' call, got %d", tapListCalls)
	}
	if tapInfoCalls != 1 {
		t.Errorf("expected exactly 1 batched 'brew tap-info' call for %d taps, got %d (linear loop regression)", taps, tapInfoCalls)
	}
}

// TestTapsBatchLoadFallsBackToNameOnlyOnError locks M11's robustness
// policy: if the batched tap-info call fails, the tap list still returns
// (names only) rather than an empty list. The GUI keeps rendering the
// tap names; detail tabs may show no data until the next refresh.
func TestTapsBatchLoadFallsBackToNameOnlyOnError(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "tap" {
			return []byte("homebrew/core\nhomebrew/cask\n"), nil
		}
		if args[0] == "tap-info" {
			return nil, errors.New("tap-info batch failed")
		}
		return nil, nil
	}
	cache := NewCache(time.Minute)

	list, err := NewTapsReader(r, cache).List(context.Background())
	if err != nil {
		t.Fatalf("List should swallow batch error and return names; got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 tap names even on batch failure, got %d", len(list))
	}
	for _, tap := range list {
		if tap.Name == "" {
			t.Error("expected non-empty tap name in fallback list")
		}
	}
}

func TestTapsServiceTap(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsWriter(r, cache)

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
	taps := NewTapsWriter(r, cache)

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
	taps := NewTapsWriter(r, cache)

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
	taps := NewTapsWriter(r, cache)

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
	taps := NewTapsReader(r, cache)

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
