package brew

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

func TestTapsServiceList(t *testing.T) {
	r := brewtest.NewMockRunner()
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

	tapList, err := taps.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	list := tapList.Taps

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
	const n = 12

	r := brewtest.NewMockRunner()
	var tapListCalls, tapInfoCalls int
	var batchNames []string
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "tap" {
			tapListCalls++
			names := make([]string, n)
			for i := range names {
				names[i] = fmt.Sprintf("thirdparty/tap-%d", i)
			}
			return []byte(strings.Join(names, "\n") + "\n"), nil
		}
		if args[0] == "tap-info" {
			tapInfoCalls++
			batchNames = append([]string(nil), args[2:]...)
			entries := make([]string, 0, len(args)-2)
			for _, name := range args[2:] {
				entries = append(entries, `{"name":"`+name+`","installed":true,"trusted":true,"formula_names":["f-`+name+`"]}`)
			}
			return []byte("[" + strings.Join(entries, ",") + "]"), nil
		}
		return []byte(`[]`), nil
	}
	cache := NewCache(time.Minute)

	tapList, err := NewTapsReader(r, cache).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	list := tapList.Taps
	if len(list) != n {
		t.Fatalf("expected %d taps, got %d", n, len(list))
	}
	if tapListCalls != 1 {
		t.Errorf("expected 1 'brew tap' call, got %d", tapListCalls)
	}
	if tapInfoCalls != 1 {
		t.Errorf("expected exactly 1 batched 'brew tap-info' call for %d taps, got %d (linear loop regression)", n, tapInfoCalls)
	}
	if len(batchNames) != n {
		t.Fatalf("batch args must include all %d names, got %d: %v", n, len(batchNames), batchNames)
	}
	for i := 0; i < n; i++ {
		want := fmt.Sprintf("thirdparty/tap-%d", i)
		found := false
		for _, got := range batchNames {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("batch missing name %q; got %v", want, batchNames)
		}
	}
	if len(list[0].FormulaNames) != 1 {
		t.Errorf("expected FormulaNames from batch on list[0], got %v", list[0].FormulaNames)
	}
}

// TestTapsBatchLoadFallsBackToNameOnlyOnError locks M11's robustness
// policy: if the batched tap-info call fails, the tap list still returns
// (names only) rather than an empty list. The GUI keeps rendering the
// tap names; detail tabs may show no data until the next refresh.
func TestTapsBatchLoadFallsBackToNameOnlyOnError(t *testing.T) {
	r := brewtest.NewMockRunner()
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

	tapList, err := NewTapsReader(r, cache).List(context.Background())
	if err != nil {
		t.Fatalf("List should swallow batch error and return names; got %v", err)
	}
	list := tapList.Taps
	if len(list) != 2 {
		t.Fatalf("expected 2 tap names even on batch failure, got %d", len(list))
	}
	for _, tap := range list {
		if tap.Name == "" {
			t.Error("expected non-empty tap name in fallback list")
		}
	}
}

func TestTapsBatchFailureDoesNotPoisonCache(t *testing.T) {
	r := brewtest.NewMockRunner()
	var tapInfoCalls int
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "tap" {
			return []byte("homebrew/core\n"), nil
		}
		if args[0] == "tap-info" {
			tapInfoCalls++
			if tapInfoCalls == 1 {
				return nil, errors.New("transient batch failure")
			}
			return []byte(`[{"name":"homebrew/core","formula_names":["wget"],"trusted":true}]`), nil
		}
		return []byte(`[]`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewTapsReader(r, cache)

	firstList, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	first := firstList.Taps
	if !firstList.TrustUnknown {
		t.Error("a list read after a failed tap-info must say its trust state is unknown")
	}
	if len(first) != 1 || len(first[0].FormulaNames) != 0 {
		t.Fatalf("first List should be name-only after batch fail, got %+v", first)
	}

	secondList, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second := secondList.Taps
	if secondList.TrustUnknown {
		t.Error("a list read after tap-info recovers must not say its trust state is unknown")
	}
	if tapInfoCalls != 2 {
		t.Errorf("expected second List to retry batch (tap-info calls=2), got %d", tapInfoCalls)
	}
	if len(second) != 1 || len(second[0].FormulaNames) != 1 || second[0].FormulaNames[0] != "wget" {
		t.Fatalf("second List should be enriched after retry, got %+v", second)
	}
}

// TestTapsTrustUnknownOnTapInfoTimeout is the AC-00 case of a slow tap-info:
// the timed-out read comes back as an error from the runner, and the list is
// flagged TrustUnknown with the names it has, not read as untrusted.
func TestTapsTrustUnknownOnTapInfoTimeout(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "tap" {
			return []byte("homebrew/core\nuser/tap\n"), nil
		}
		if args[0] == "tap-info" {
			return nil, context.DeadlineExceeded
		}
		return nil, nil
	}
	cache := NewCache(time.Minute)

	got, err := NewTapsReader(r, cache).List(context.Background())
	if err != nil {
		t.Fatalf("a timed-out tap-info must not fail the list; got %v", err)
	}
	if !got.TrustUnknown {
		t.Fatal("a timed-out tap-info must flag the list TrustUnknown")
	}
	if len(got.Taps) != 2 || got.Taps[1].Name != "user/tap" {
		t.Fatalf("the name-only list must keep every tap, got %+v", got.Taps)
	}
	for _, tap := range got.Taps {
		if tap.Trusted || len(tap.FormulaNames) != 0 {
			t.Errorf("tap %s was not read, so it carries no trust or formula data: %+v", tap.Name, tap)
		}
	}
}

// TestTapsTrustKnownWhenTapInfoReads checks the flag is clear on a full read,
// so the GUI only refuses trust actions when the state really is missing.
func TestTapsTrustKnownWhenTapInfoReads(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "tap" {
			return []byte("user/tap\n"), nil
		}
		if args[0] == "tap-info" {
			return []byte(`[{"name":"user/tap","trusted":false}]`), nil
		}
		return nil, nil
	}
	got, err := NewTapsReader(r, NewCache(time.Minute)).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.TrustUnknown {
		t.Fatal("a full tap-info read must not flag TrustUnknown")
	}
	if len(got.Taps) != 1 || got.Taps[0].Trusted {
		t.Fatalf("the untrusted tap must read as untrusted, got %+v", got.Taps)
	}
}

func TestTapsServiceTap(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	taps := NewTapsWriter(r, cache)

	if err := taps.Tap(context.Background(), "some-org/formulas"); err != nil {
		t.Fatal(err)
	}
}

func TestTapsServiceUntap(t *testing.T) {
	r := brewtest.NewMockRunner()
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
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	cache.Set(KeyTapsList, []Tap{{Name: "nicknisi/tap"}})
	taps := NewTapsWriter(r, cache)

	if err := taps.Repair(context.Background(), "nicknisi/tap"); err != nil {
		t.Fatal(err)
	}

	// Repair must invalidate the taps cache like every other taps write
	// (r1 finding 20).
	if _, ok := cache.Get(KeyTapsList); ok {
		t.Error("expected KeyTapsList to be invalidated by Repair")
	}
}

// tapInfoStub answers `brew tap` with tapList and each `brew tap-info --json`
// with one entry per name, recording the names of every tap-info call. fail,
// when set, is asked with the 1-based call number whether that call fails.
type tapInfoStub struct {
	tapList string
	calls   [][]string
	fail    func(call int) bool
}

func (st *tapInfoStub) execute(ctx context.Context, args ...string) ([]byte, error) {
	if len(args) == 1 && args[0] == "tap" {
		return []byte(st.tapList), nil
	}
	if args[0] != "tap-info" {
		return []byte(`[]`), nil
	}
	st.calls = append(st.calls, append([]string(nil), args[2:]...))
	if st.fail != nil && st.fail(len(st.calls)) {
		return nil, errors.New("tap-info failed")
	}
	entries := make([]string, 0, len(args)-2)
	for _, name := range args[2:] {
		entries = append(entries, `{"name":"`+name+`","trusted":true,"formula_names":["f-`+name+`"]}`)
	}
	return []byte("[" + strings.Join(entries, ",") + "]"), nil
}

// setTapInfoChunkBounds lowers the chunk bounds for one test; they are restored
// when the test ends.
func setTapInfoChunkBounds(t *testing.T, names, bytes int) {
	t.Helper()
	origMax, origBytes := tapInfoChunkMax, tapInfoChunkBytes
	tapInfoChunkMax, tapInfoChunkBytes = names, bytes
	t.Cleanup(func() { tapInfoChunkMax, tapInfoChunkBytes = origMax, origBytes })
}

// LB-04 (a): a few taps are one tap-info call, with the same args as before.
func TestTapInfoBatchFewTapsIsOneCallWithSameArgs(t *testing.T) {
	st := &tapInfoStub{}
	var calls int
	var argv []string
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if args[0] == "tap-info" {
			calls++
			argv = append([]string(nil), args...)
		}
		return st.execute(ctx, args...)
	}

	got, err := (&tapsReader{runner: r}).fetchTapInfoBatch(context.Background(), []string{"a/x", "b/y", "c/z"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 tap-info call, got %d", calls)
	}
	if want := "tap-info --json a/x b/y c/z"; strings.Join(argv, " ") != want {
		t.Errorf("args = %q, want %q", strings.Join(argv, " "), want)
	}
	if len(got) != 3 || got[2].Name != "c/z" {
		t.Errorf("entries = %+v, want three in input order", got)
	}
}

// LB-04 (b): more names than the count bound make several calls, each within
// both bounds, every name sent once, and the entries merged in input order.
func TestTapInfoBatchSplitsByNameCount(t *testing.T) {
	setTapInfoChunkBounds(t, 3, 64<<10)
	var names []string
	for i := 0; i < 7; i++ {
		names = append(names, fmt.Sprintf("user/tap-%d", i))
	}
	st := &tapInfoStub{}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = st.execute

	got, err := (&tapsReader{runner: r}).fetchTapInfoBatch(context.Background(), names)
	if err != nil {
		t.Fatal(err)
	}
	wantSizes := []int{3, 3, 1}
	if len(st.calls) != len(wantSizes) {
		t.Fatalf("expected %d tap-info calls, got %d: %v", len(wantSizes), len(st.calls), st.calls)
	}
	seen := map[string]int{}
	for i, call := range st.calls {
		if len(call) != wantSizes[i] || len(call) > tapInfoChunkMax {
			t.Errorf("call %d has %d names, want %d (bound %d)", i, len(call), wantSizes[i], tapInfoChunkMax)
		}
		for _, name := range call {
			seen[name]++
		}
	}
	for _, name := range names {
		if seen[name] != 1 {
			t.Errorf("name %q sent %d times, want once", name, seen[name])
		}
	}
	if len(got) != len(names) {
		t.Fatalf("got %d entries, want %d", len(got), len(names))
	}
	for i, name := range names {
		if got[i].Name != name {
			t.Errorf("entry %d = %q, want %q (input order)", i, got[i].Name, name)
		}
	}
}

// LB-04 (c): long names split by the byte budget, not the count; a name over
// the budget on its own still goes in a call of its own.
func TestTapInfoBatchSplitsByByteBudget(t *testing.T) {
	setTapInfoChunkBounds(t, 100, 70)
	var names []string
	for i := 0; i < 5; i++ {
		names = append(names, strings.Repeat(string(rune('a'+i)), 30))
	}
	st := &tapInfoStub{}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = st.execute

	if _, err := (&tapsReader{runner: r}).fetchTapInfoBatch(context.Background(), names); err != nil {
		t.Fatal(err)
	}
	// Each 30-byte name costs 31 with its NUL, so 70 fits two of them.
	wantSizes := []int{2, 2, 1}
	if len(st.calls) != len(wantSizes) {
		t.Fatalf("expected %d tap-info calls, got %d: %v", len(wantSizes), len(st.calls), st.calls)
	}
	for i, call := range st.calls {
		if len(call) != wantSizes[i] {
			t.Errorf("call %d has %d names, want %d", i, len(call), wantSizes[i])
		}
		size := 0
		for _, name := range call {
			size += len(name) + 1
		}
		if size > tapInfoChunkBytes {
			t.Errorf("call %d carries %d bytes, over the %d budget", i, size, tapInfoChunkBytes)
		}
	}

	setTapInfoChunkBounds(t, 100, 40)
	st = &tapInfoStub{}
	r.ExecuteFn = st.execute
	oversize := []string{strings.Repeat("s", 30), strings.Repeat("L", 50), "tiny/tap"}
	if _, err := (&tapsReader{runner: r}).fetchTapInfoBatch(context.Background(), oversize); err != nil {
		t.Fatal(err)
	}
	if len(st.calls) != 3 {
		t.Fatalf("an oversize name must sit in a call of its own: got %d calls %v", len(st.calls), st.calls)
	}
	if st.calls[1][0] != oversize[1] {
		t.Errorf("the 50-byte name should be alone in call 1, got %v", st.calls[1])
	}
}

// LB-04: a cancelled ctx stops the batch before the next chunk.
func TestTapInfoBatchStopsBeforeNextChunkOnCancel(t *testing.T) {
	setTapInfoChunkBounds(t, 1, 64<<10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st := &tapInfoStub{fail: func(int) bool {
		cancel()
		return false
	}}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = st.execute

	_, err := (&tapsReader{runner: r}).fetchTapInfoBatch(ctx, []string{"a/x", "b/y", "c/z"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if len(st.calls) != 1 {
		t.Errorf("expected the batch to stop after 1 call, got %d: %v", len(st.calls), st.calls)
	}
}

// LB-04 (d): when the second chunk fails, the whole list falls back to names
// only, flagged TrustUnknown, and is not cached, so the next read retries.
func TestTapsListSecondChunkFailureFallsBackAndDoesNotCache(t *testing.T) {
	setTapInfoChunkBounds(t, 2, 64<<10)
	st := &tapInfoStub{
		tapList: "a/x\nb/y\nc/z\n",
		fail:    func(call int) bool { return call == 2 },
	}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = st.execute
	reader := NewTapsReader(r, NewCache(time.Minute))

	first, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !first.TrustUnknown {
		t.Error("a failed second chunk must flag the list TrustUnknown")
	}
	if len(first.Taps) != 3 {
		t.Fatalf("the name-only list must keep every tap, got %d", len(first.Taps))
	}
	for _, tap := range first.Taps {
		if tap.Trusted || len(tap.FormulaNames) != 0 {
			t.Errorf("tap %s was not read, so it carries no trust or formula data: %+v", tap.Name, tap)
		}
	}

	second, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if second.TrustUnknown {
		t.Error("a list read after both chunks succeed must not be TrustUnknown")
	}
	if len(st.calls) != 4 {
		t.Errorf("expected 4 tap-info calls (2 per read, nothing cached), got %d: %v", len(st.calls), st.calls)
	}
	if len(second.Taps) != 3 || second.Taps[2].FormulaNames[0] != "f-c/z" {
		t.Errorf("second read should be enriched in input order, got %+v", second.Taps)
	}
}
