package brew

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// TestBrewReadsConcurrentCallsRunOnce: every cached reader runs its brew
// command once when many callers ask for it at the same time on an empty
// cache. The runner blocks until release, so every caller is in flight
// together.
func TestBrewReadsConcurrentCallsRunOnce(t *testing.T) {
	cases := []struct {
		name    string
		command string
		output  string
		read    func(c *Client) error
	}{
		{"formulae list", "info --json=v2 --installed", `{"formulae":[]}`, func(c *Client) error {
			_, err := c.Formulae.List(context.Background())
			return err
		}},
		{"casks list", "info --json=v2 --installed --cask", `{"casks":[]}`, func(c *Client) error {
			_, err := c.Casks.List(context.Background())
			return err
		}},
		{"formulae outdated", "outdated --json=v2 --formula", `{"formulae":[]}`, func(c *Client) error {
			_, err := c.Formulae.Outdated(context.Background())
			return err
		}},
		{"casks outdated", "outdated --json=v2 --cask", `{"casks":[]}`, func(c *Client) error {
			_, err := c.Casks.Outdated(context.Background())
			return err
		}},
		{"taps", "tap", "", func(c *Client) error {
			_, err := c.Taps.List(context.Background())
			return err
		}},
		{"services list", "services list --json", `[]`, func(c *Client) error {
			_, err := c.Services.List(context.Background())
			return err
		}},
		{"services info", "services info --all --json", `[]`, func(c *Client) error {
			_, err := c.Services.Info(context.Background())
			return err
		}},
		{"trust", "trust --json=v1", `{"taps":[],"formulae":[],"casks":[],"commands":[]}`, func(c *Client) error {
			_, err := c.Trust.ListTrusted(context.Background())
			return err
		}},
		{"doctor", "doctor", "Your system is ready to brew.", func(c *Client) error {
			_, err := c.Diagnostics.Doctor(context.Background())
			return err
		}},
		{"config", "config", "HOMEBREW_VERSION: 4.0.0", func(c *Client) error {
			_, err := c.Diagnostics.Config(context.Background())
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var mu sync.Mutex
				calls := 0
				release := make(chan struct{})
				runner := &brewtest.MockRunner{ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
					if strings.Join(args, " ") != tc.command {
						t.Errorf("unexpected command %q", strings.Join(args, " "))
						return nil, nil
					}
					mu.Lock()
					calls++
					mu.Unlock()
					<-release
					return []byte(tc.output), nil
				}}
				client := NewClient(runner)

				var wg sync.WaitGroup
				for i := 0; i < 20; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := tc.read(client); err != nil {
							t.Error(err)
						}
					}()
				}
				synctest.Wait()
				close(release)
				wg.Wait()

				if calls != 1 {
					t.Errorf("%s ran %d times for 20 concurrent callers, want 1", tc.command, calls)
				}
			})
		})
	}
}

// TestCachedDropsFetchThatRacedAWrite: a fetch that started before an
// invalidation must not store its result after the write.
func TestCachedDropsFetchThatRacedAWrite(t *testing.T) {
	c := NewCache(time.Minute)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = cached(context.Background(), c, KeyFormulaeList, 0, func(context.Context) ([]Formula, bool, error) {
			close(started)
			<-release
			return []Formula{{Name: "before-write"}}, true, nil
		})
	}()

	<-started
	c.Invalidate(KeyFormulaeList)
	close(release)
	<-done

	if _, ok := c.Get(KeyFormulaeList); ok {
		t.Fatal("a fetch that started before the write stored its result after it")
	}
}

// TestCachedCallAfterWriteDoesNotJoinOldFetch: a caller that starts after a
// write does not wait on, or take the result of, a fetch that began before it.
func TestCachedCallAfterWriteDoesNotJoinOldFetch(t *testing.T) {
	c := NewCache(time.Minute)
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = cached(context.Background(), c, KeyFormulaeList, 0, func(context.Context) ([]Formula, bool, error) {
			close(started)
			<-release
			return []Formula{{Name: "before-write"}}, true, nil
		})
	}()

	<-started
	c.Invalidate(KeyFormulaeList)
	got, err := cached(context.Background(), c, KeyFormulaeList, 0, func(context.Context) ([]Formula, bool, error) {
		return []Formula{{Name: "after-write"}}, true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "after-write" {
		t.Fatalf("caller after the write got %+v, want after-write", got)
	}

	close(release)
	<-done
	val, ok := c.Get(KeyFormulaeList)
	if !ok {
		t.Fatal("expected the post-write result to stay cached")
	}
	if formulae := val.([]Formula); formulae[0].Name != "after-write" {
		t.Fatalf("cache holds %q, want after-write", formulae[0].Name)
	}
}

// TestCachedCancelledCallerDoesNotFailJoiner: the fetch runs on the flight's
// context, not the context of the caller that started it. That caller leaves
// when its ctx is cancelled; the fetch must not see the cancel, so a joiner
// with a live ctx still gets the value.
func TestCachedCancelledCallerDoesNotFailJoiner(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewCache(time.Minute)
		var runs atomic.Int32
		release := make(chan struct{})
		fetch := func(ctx context.Context) ([]Formula, bool, error) {
			runs.Add(1)
			<-release
			if err := ctx.Err(); err != nil {
				return nil, false, err
			}
			return []Formula{{Name: "fetched"}}, true, nil
		}

		leaderCtx, cancelLeader := context.WithCancel(context.Background())
		leaderErr := make(chan error, 1)
		go func() {
			_, err := cached(leaderCtx, c, KeyFormulaeList, 0, fetch)
			leaderErr <- err
		}()
		synctest.Wait()

		type result struct {
			formulae []Formula
			err      error
		}
		joiner := make(chan result, 1)
		go func() {
			formulae, err := cached(context.Background(), c, KeyFormulaeList, 0, func(context.Context) ([]Formula, bool, error) {
				t.Error("joiner must not run fetch")
				return nil, false, nil
			})
			joiner <- result{formulae, err}
		}()
		synctest.Wait()

		cancelLeader()
		if err := <-leaderErr; !errors.Is(err, context.Canceled) {
			t.Errorf("leader err = %v, want context.Canceled", err)
		}
		close(release)
		r := <-joiner
		if r.err != nil || len(r.formulae) != 1 || r.formulae[0].Name != "fetched" {
			t.Errorf("joiner = (%+v, %v), want the fetched list", r.formulae, r.err)
		}
		if got := runs.Load(); got != 1 {
			t.Errorf("fetch ran %d times, want 1", got)
		}
	})
}

// blockedFetch starts cached(key) in the background. Its fetch counts itself in
// runs and blocks until release is closed. It returns once the fetch has
// started, and a channel that closes when the read returns.
func blockedFetch(c *Cache, key CacheKey, runs *atomic.Int32, release <-chan struct{}) <-chan struct{} {
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = cached(context.Background(), c, key, 0, func(context.Context) ([]Formula, bool, error) {
			runs.Add(1)
			close(started)
			<-release
			return []Formula{{Name: "fetched"}}, true, nil
		})
	}()
	<-started
	return done
}

// TestCachedUnrelatedInvalidationKeepsFetch: a write to another group while a
// fetch runs neither drops its result nor makes a later caller run the command
// again.
func TestCachedUnrelatedInvalidationKeepsFetch(t *testing.T) {
	c := NewCache(time.Minute)
	var runs atomic.Int32
	release := make(chan struct{})
	first := blockedFetch(c, KeyOutdatedFormulae, &runs, release)

	c.InvalidateFor("taps")
	second := make(chan struct{})
	go func() {
		defer close(second)
		_, _ = cached(context.Background(), c, KeyOutdatedFormulae, 0, func(context.Context) ([]Formula, bool, error) {
			runs.Add(1)
			return []Formula{{Name: "again"}}, true, nil
		})
	}()
	time.Sleep(20 * time.Millisecond)
	close(release)
	<-first
	<-second

	if got := runs.Load(); got != 1 {
		t.Fatalf("outdated ran %d times, want 1: the unrelated write made a second caller fetch", got)
	}
	if _, ok := c.Get(KeyOutdatedFormulae); !ok {
		t.Fatal("the fetch's result was dropped by an invalidation of another group")
	}
}

// TestCachedGroupInvalidationDropsItsFetch: a write to the group that holds the
// fetched key drops the result of a fetch that is still running, and a caller
// after the write starts a new fetch rather than joining the old one.
func TestCachedGroupInvalidationDropsItsFetch(t *testing.T) {
	c := NewCache(time.Minute)
	var runs atomic.Int32
	release := make(chan struct{})
	first := blockedFetch(c, KeyOutdatedFormulae, &runs, release)

	c.InvalidateFor("outdated")
	second := make(chan struct{})
	go func() {
		defer close(second)
		_, _ = cached(context.Background(), c, KeyOutdatedFormulae, 0, func(context.Context) ([]Formula, bool, error) {
			runs.Add(1)
			return []Formula{{Name: "after-write"}}, true, nil
		})
	}()
	time.Sleep(20 * time.Millisecond)
	close(release)
	<-first
	<-second

	if got := runs.Load(); got != 2 {
		t.Fatalf("outdated ran %d times, want 2: a caller after the write must not join the old fetch", got)
	}
	val, ok := c.Get(KeyOutdatedFormulae)
	if !ok {
		t.Fatal("expected the post-write result to be cached")
	}
	if formulae := val.([]Formula); formulae[0].Name != "after-write" {
		t.Fatalf("cache holds %q, want after-write: the pre-write fetch was stored", formulae[0].Name)
	}
}

// TestCachedWrongTypeIsRefetched: an entry of another type is a miss, and the
// fetch result replaces it.
func TestCachedWrongTypeIsRefetched(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set(KeyFormulaeList, 42)
	fetched := false
	got, err := cached(context.Background(), c, KeyFormulaeList, 0, func(context.Context) ([]Formula, bool, error) {
		fetched = true
		return []Formula{{Name: "ripgrep"}}, true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !fetched || len(got) != 1 {
		t.Fatalf("fetched=%v got=%+v, want a fetch", fetched, got)
	}
}

// TestCachedNotStorableResultIsNotCached: a result the fetch marks not
// storable is returned, and the next call fetches again.
func TestCachedNotStorableResultIsNotCached(t *testing.T) {
	c := NewCache(time.Minute)
	calls := 0
	fetch := func(context.Context) (TapList, bool, error) {
		calls++
		return TapList{TrustUnknown: true}, false, nil
	}
	for i := 0; i < 2; i++ {
		got, err := cached(context.Background(), c, KeyTapsList, 0, fetch)
		if err != nil || !got.TrustUnknown {
			t.Fatalf("call %d: got %+v, %v", i, got, err)
		}
	}
	if calls != 2 {
		t.Errorf("fetch ran %d times, want 2 (nothing stored)", calls)
	}
}
