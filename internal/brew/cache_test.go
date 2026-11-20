package brew

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheGetSet(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set(KeyFormulaeList, []Formula{{Name: "ripgrep"}})
	val, ok := c.Get(KeyFormulaeList)
	if !ok {
		t.Fatal("expected cache hit")
	}
	formulae, ok := val.([]Formula)
	if !ok {
		t.Fatal("expected []Formula")
	}
	if len(formulae) != 1 || formulae[0].Name != "ripgrep" {
		t.Errorf("got %+v, want [ripgrep]", formulae)
	}
}

func TestCacheMiss(t *testing.T) {
	c := NewCache(time.Minute)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

// newCacheWithClock returns a cache on a clock that stands still until advance
// is called, so a TTL test moves time forward instead of sleeping.
func newCacheWithClock(ttl time.Duration) (c *Cache, advance func(time.Duration)) {
	now := time.Now()
	c = NewCache(ttl)
	c.now = func() time.Time { return now }
	return c, func(d time.Duration) { now = now.Add(d) }
}

func TestCacheTTL(t *testing.T) {
	c, advance := newCacheWithClock(50 * time.Millisecond)
	c.Set(KeyFormulaeList, "data")
	advance(100 * time.Millisecond)
	_, ok := c.Get(KeyFormulaeList)
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestCacheInvalidate(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Invalidate("key1")
	_, ok1 := c.Get("key1")
	if ok1 {
		t.Error("expected key1 to be invalidated")
	}
	_, ok2 := c.Get("key2")
	if !ok2 {
		t.Error("expected key2 to still be present")
	}
}

func TestCacheInvalidateFor(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set(KeyFormulaeList, "f")
	c.Set(KeyOutdatedFormulae, "o")
	c.InvalidateFor("upgrade")
	_, fOk := c.Get(KeyOutdatedFormulae)
	_, oOk := c.Get(KeyOutdatedCasks)
	if fOk || oOk {
		t.Error("expected both keys invalidated by upgrade group")
	}
}

func TestCacheInvalidateAll(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)
	c.InvalidateAll()
	_, ok := c.Get("a")
	if ok {
		t.Error("expected all keys invalidated")
	}
}

func TestCacheConcurrency(t *testing.T) {
	c := NewCache(time.Minute)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			c.Set(KeyFormulaeList, i)
			c.Get(KeyFormulaeList)
			c.Invalidate(KeyFormulaeList)
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			c.Get(KeyFormulaeList)
			c.InvalidateAll()
			c.Set(KeyOutdatedFormulae, i)
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}

func TestCacheZeroTTL(t *testing.T) {
	c := NewCache(0)
	c.Set(KeyFormulaeList, "data")
	time.Sleep(time.Millisecond)
	_, ok := c.Get(KeyFormulaeList)
	if ok {
		t.Error("expected cache miss with zero TTL")
	}
}

func TestCacheConcurrentExpiry(t *testing.T) {
	c := NewCache(time.Millisecond)
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				c.Set(KeyFormulaeList, j)
				c.Get(KeyFormulaeList)
				time.Sleep(time.Millisecond)
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestTypedCacheWrongTypeReturnsMiss(t *testing.T) {
	c := NewCache(time.Minute)
	tc := NewTypedCache[string](c, KeyFormulaeList)
	c.Set(KeyFormulaeList, 42)
	val, ok := tc.Get()
	if ok {
		t.Fatal("expected miss on wrong type")
	}
	if val != "" {
		t.Errorf("expected zero value, got %q", val)
	}
	_, cached := c.Get(KeyFormulaeList)
	if cached {
		t.Error("expected key to be invalidated on type mismatch")
	}
}

func TestCacheSeparateOutdatedKeys(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set(KeyOutdatedFormulae, []Formula{{Name: "ripgrep"}})
	c.Set(KeyOutdatedCasks, []Cask{{Name: "firefox"}})

	f, fOk := c.Get(KeyOutdatedFormulae)
	if !fOk {
		t.Fatal("expected KeyOutdatedFormulae hit")
	}
	formulae, _ := f.([]Formula)
	if len(formulae) != 1 || formulae[0].Name != "ripgrep" {
		t.Error("wrong formulae cached")
	}

	f2, f2Ok := c.Get(KeyOutdatedCasks)
	if !f2Ok {
		t.Fatal("expected KeyOutdatedCasks hit")
	}
	casks, _ := f2.([]Cask)
	if len(casks) != 1 || casks[0].Name != "firefox" {
		t.Error("wrong casks cached")
	}
}

func TestCacheSetWithTTLUsesPerEntryTTL(t *testing.T) {
	c, advance := newCacheWithClock(50 * time.Millisecond)
	c.SetWithTTL(KeyOutdatedFormulae, "outdated", 5*time.Second)

	// Per-entry TTL overrides cache default; entry should still be fresh
	// after the cache default would have expired.
	advance(75 * time.Millisecond)
	v, ok := c.Get(KeyOutdatedFormulae)
	if !ok {
		t.Fatal("expected hit; per-entry TTL should override cache default")
	}
	if v.(string) != "outdated" {
		t.Errorf("got %v, want outdated", v)
	}
}

func TestCacheSetWithTTLZeroInheritsDefault(t *testing.T) {
	c, advance := newCacheWithClock(50 * time.Millisecond)
	c.SetWithTTL(KeyOutdatedFormulae, "outdated", 0)

	advance(75 * time.Millisecond)
	if _, ok := c.Get(KeyOutdatedFormulae); ok {
		t.Error("expected miss after cache default TTL elapsed (per-entry ttl=0 inherits)")
	}
}

func TestCacheSetWithoutTTLStillUsesDefault(t *testing.T) {
	c, advance := newCacheWithClock(50 * time.Millisecond)
	c.Set(KeyOutdatedFormulae, "outdated")

	advance(75 * time.Millisecond)
	if _, ok := c.Get(KeyOutdatedFormulae); ok {
		t.Error("expected miss after cache default TTL elapsed (plain Set, no per-entry ttl)")
	}
}

// TestCacheSetWithTTLAllKeys locks the M12 promise: every per-class cache
// key (formulae/casks/taps/services/doctor/outdated) can carry its own
// TTL override without coupling the entries together.
func TestCacheSetWithTTLAllKeys(t *testing.T) {
	c, advance := newCacheWithClock(50 * time.Millisecond)
	c.SetWithTTL(KeyFormulaeList, "f", 1*time.Hour)
	c.SetWithTTL(KeyCasksList, "c", 1*time.Hour)
	c.SetWithTTL(KeyTapsList, "t", 1*time.Hour)
	c.SetWithTTL(KeyServicesList, "s", 1*time.Hour)
	c.SetWithTTL(KeyDoctorResult, "d", 1*time.Hour)
	c.SetWithTTL(KeyOutdatedFormulae, "of", 1*time.Hour)
	c.SetWithTTL(KeyOutdatedCasks, "oc", 1*time.Hour)

	// Wait past cache default; per-entry TTL should keep all entries fresh.
	advance(75 * time.Millisecond)
	for key, want := range map[CacheKey]string{
		KeyFormulaeList:     "f",
		KeyCasksList:        "c",
		KeyTapsList:         "t",
		KeyServicesList:     "s",
		KeyDoctorResult:     "d",
		KeyOutdatedFormulae: "of",
		KeyOutdatedCasks:    "oc",
	} {
		v, ok := c.Get(key)
		if !ok {
			t.Errorf("%s: expected hit (per-entry TTL 1h, cache default 50ms expired)", key)
			continue
		}
		if s, _ := v.(string); s != want {
			t.Errorf("%s: got %q, want %q", key, s, want)
		}
	}
}

// TestCacheExpiredGetKeepsConcurrentSet is a stress test. Get used to find an
// expired entry under the read lock, drop the lock, and delete the key under
// the write lock without checking again, so a fresh Set that landed in between
// was deleted with the old entry.
func TestCacheExpiredGetKeepsConcurrentSet(t *testing.T) {
	var nowNanos atomic.Int64
	nowNanos.Store(time.Now().UnixNano())
	c := NewCache(time.Minute)
	c.now = func() time.Time { return time.Unix(0, nowNanos.Load()) }

	for i := 0; i < 20000; i++ {
		c.Set(KeyFormulaeList, "old")
		nowNanos.Add(int64(2 * time.Minute)) // the "old" entry is now expired

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			c.Get(KeyFormulaeList)
		}()
		go func() {
			defer wg.Done()
			<-start
			c.Set(KeyFormulaeList, "new")
		}()
		close(start)
		wg.Wait()

		if v, ok := c.Get(KeyFormulaeList); !ok || v != "new" {
			t.Fatalf("iteration %d: fresh Set was lost (got %v, ok=%v)", i, v, ok)
		}
	}
}
