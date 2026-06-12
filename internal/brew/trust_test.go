package brew

import (
	"context"
	"testing"
	"time"
)

func TestTrustServiceListTrusted(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		data := result.(*trustJSON)
		*data = trustJSON{
			Entries: []trustEntryJSON{
				{Name: "nicknisi/tap", Type: "tap", Tap: "nicknisi/tap"},
				{Name: "nicknisi/tap/some-formula", Type: "formula", Tap: "nicknisi/tap"},
			},
		}
		return nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustService(r, cache)

	entries, err := svc.ListTrusted(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Type != TrustTypeTap {
		t.Errorf("expected tap type, got %s", entries[0].Type)
	}
	if entries[1].Type != TrustTypeFormula {
		t.Errorf("expected formula type, got %s", entries[1].Type)
	}
}

func TestTrustServiceGetTapTrustStatusOfficial(t *testing.T) {
	r := NewMockRunner()
	cache := NewCache(time.Minute)
	svc := NewTrustService(r, cache)

	status, err := svc.GetTapTrustStatus(context.Background(), "homebrew/core")
	if err != nil {
		t.Fatal(err)
	}
	if status != TrustOfficial {
		t.Errorf("expected TrustOfficial, got %s", status)
	}
}

func TestTrustServiceGetTapTrustStatusUntrusted(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		data := result.(*trustJSON)
		*data = trustJSON{}
		return nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustService(r, cache)

	status, err := svc.GetTapTrustStatus(context.Background(), "nicknisi/tap")
	if err != nil {
		t.Fatal(err)
	}
	if status != TrustUntrusted {
		t.Errorf("expected TrustUntrusted, got %s", status)
	}
}

func TestTrustServiceTrustTap(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustService(r, cache)

	if err := svc.TrustTap(context.Background(), "nicknisi/tap"); err != nil {
		t.Fatal(err)
	}
}

func TestTrustServiceUntrustTap(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustService(r, cache)

	if err := svc.UntrustTap(context.Background(), "nicknisi/tap"); err != nil {
		t.Fatal(err)
	}
}

func TestTrustServiceTrustFormula(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustService(r, cache)

	if err := svc.TrustFormula(context.Background(), "nicknisi/tap/some-formula"); err != nil {
		t.Fatal(err)
	}
}

func TestTrustServiceTrustCask(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustService(r, cache)

	if err := svc.TrustCask(context.Background(), "nicknisi/tap/some-cask"); err != nil {
		t.Fatal(err)
	}
}
