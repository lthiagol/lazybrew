package brew

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

func TestTrustReaderParsesRealShape(t *testing.T) {
	raw, err := os.ReadFile("testdata/trust_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return raw, nil
	}
	svc := NewTrustReader(r, NewCache(time.Minute))

	entries, err := svc.ListTrusted(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := []TrustEntry{
		{Name: "octo/tools", Type: TrustTypeTap, Tap: "octo/tools"},
		{Name: "acme/fonts", Type: TrustTypeTap, Tap: "acme/fonts"},
		{Name: "octo/tools/alpha", Type: TrustTypeFormula, Tap: "octo/tools"},
		{Name: "acme/fonts/beta", Type: TrustTypeFormula, Tap: "acme/fonts"},
		{Name: "acme/fonts/gamma-app", Type: TrustTypeCask, Tap: "acme/fonts"},
		{Name: "octo/tools/zeta", Type: TrustTypeCommand, Tap: "octo/tools"},
	}
	if !reflect.DeepEqual(entries, want) {
		t.Errorf("entries = %+v\nwant %+v", entries, want)
	}
}

func TestTrustReaderTapDerivation(t *testing.T) {
	cases := []struct {
		name string
		json string
		want TrustEntry
	}{
		{"tap is its own tap", `{"taps":["owner/repo"]}`, TrustEntry{Name: "owner/repo", Type: TrustTypeTap, Tap: "owner/repo"}},
		{"three parts: owner/repo/name", `{"formulae":["owner/repo/name"]}`, TrustEntry{Name: "owner/repo/name", Type: TrustTypeFormula, Tap: "owner/repo"}},
		{"one part has no tap", `{"formulae":["wget"]}`, TrustEntry{Name: "wget", Type: TrustTypeFormula, Tap: ""}},
		{"two parts on a formula has no tap", `{"casks":["owner/repo"]}`, TrustEntry{Name: "owner/repo", Type: TrustTypeCask, Tap: ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := brewtest.NewMockRunner()
			r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
				return []byte(tc.json), nil
			}
			entries, err := NewTrustReader(r, NewCache(time.Minute)).ListTrusted(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || entries[0] != tc.want {
				t.Errorf("entries = %+v, want [%+v]", entries, tc.want)
			}
		})
	}
}

func TestTrustReaderEmptyLists(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{"taps":[],"formulae":[],"casks":[],"commands":[]}`), nil
	}
	entries, err := NewTrustReader(r, NewCache(time.Minute)).ListTrusted(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no entries, got %+v", entries)
	}
}

func TestTrustReaderMalformedJSONIsError(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{"taps":`), nil
	}
	if _, err := NewTrustReader(r, NewCache(time.Minute)).ListTrusted(context.Background()); err == nil {
		t.Fatal("expected a parse error")
	}
}

func TestTrustWriterUntrustArgs(t *testing.T) {
	cases := []struct {
		name string
		call func(TrustWriter) error
		want string
	}{
		{"untrust formula", func(w TrustWriter) error { return w.UntrustFormula(context.Background(), "x/y/z") }, "untrust --formula x/y/z"},
		{"untrust cask", func(w TrustWriter) error { return w.UntrustCask(context.Background(), "x/y/z") }, "untrust --cask x/y/z"},
		{"untrust command", func(w TrustWriter) error { return w.UntrustCommand(context.Background(), "x/y/z") }, "untrust --command x/y/z"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			r := brewtest.NewMockRunner()
			r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
				got = strings.Join(args, " ")
				return []byte(""), nil
			}
			if err := tc.call(NewTrustWriter(r, NewCache(time.Minute))); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("args = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTrustServiceTrustTap(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustWriter(r, cache)

	if err := svc.TrustTap(context.Background(), "nicknisi/tap"); err != nil {
		t.Fatal(err)
	}
}

func TestTrustServiceUntrustTap(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustWriter(r, cache)

	if err := svc.UntrustTap(context.Background(), "nicknisi/tap"); err != nil {
		t.Fatal(err)
	}
}

func TestTrustServiceTrustFormula(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustWriter(r, cache)

	if err := svc.TrustFormula(context.Background(), "nicknisi/tap/some-formula"); err != nil {
		t.Fatal(err)
	}
}

func TestTrustServiceTrustCask(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustWriter(r, cache)

	if err := svc.TrustCask(context.Background(), "nicknisi/tap/some-cask"); err != nil {
		t.Fatal(err)
	}
}

// A Taps refresh drops the trust list with the taps, so R and the auto-refresh
// read the trust state again (lb30 AC-03). The taps group is the one the panel
// refresh invalidates.
func TestTapsRefreshReadsTheTrustListAgain(t *testing.T) {
	var runs int
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		runs++
		return []byte(`{"taps":[],"formulae":[],"casks":[],"commands":[]}`), nil
	}
	cache := NewCache(time.Minute)
	svc := NewTrustReader(r, cache)

	if _, err := svc.ListTrusted(context.Background()); err != nil {
		t.Fatal(err)
	}
	cache.InvalidateFor("taps")
	if _, err := svc.ListTrusted(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runs != 2 {
		t.Fatalf("trust list ran %d times, want 2: the taps refresh kept the cached trust list", runs)
	}
}
