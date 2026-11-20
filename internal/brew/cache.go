package brew

import (
	"context"
	"strconv"
	"sync"
	"time"
)

type CacheKey string

const (
	KeyFormulaeList     CacheKey = "formulae:list"
	KeyCasksList        CacheKey = "casks:list"
	KeyOutdatedFormulae CacheKey = "outdated:formulae"
	KeyOutdatedCasks    CacheKey = "outdated:casks"
	KeyTapsList         CacheKey = "taps:list"
	KeyServicesList     CacheKey = "services:list"
	KeyServicesInfo     CacheKey = "services:info"
	KeyTrustList        CacheKey = "trust:list"
	KeyDoctorResult     CacheKey = "doctor:result"
	KeyConfig           CacheKey = "config"
)

// invalidateGroups names the cache keys each write group drops. Every writer
// calls InvalidateFor with its group, once its command has finished.
var invalidateGroups = map[string][]CacheKey{
	"install":    {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"uninstall":  {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"reinstall":  {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"upgrade":    {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"update":     {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks, KeyTapsList, KeyServicesList, KeyServicesInfo, KeyTrustList, KeyDoctorResult, KeyConfig},
	"pin":        {KeyFormulaeList, KeyOutdatedFormulae},
	"unpin":      {KeyFormulaeList, KeyOutdatedFormulae},
	"tap":        {KeyTapsList},
	"untap":      {KeyTapsList},
	"trust":      {KeyTrustList, KeyTapsList},
	"untrust":    {KeyTrustList, KeyTapsList},
	"services":   {KeyServicesList, KeyServicesInfo},
	"cleanup":    {KeyFormulaeList, KeyCasksList},
	"autoremove": {KeyFormulaeList, KeyOutdatedFormulae},
}

type cacheEntry struct {
	data      any
	timestamp time.Time
	ttl       time.Duration // 0 = use cache default TTL
}

type Cache struct {
	mu         sync.RWMutex
	entries    map[CacheKey]cacheEntry
	gen        uint64 // bumped by every invalidation; see cached
	defaultTTL time.Duration
	now        func() time.Time // time.Now; a test swaps in a clock it moves by hand
	sf         *singleflight    // coalesces concurrent fetches of one key
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries:    make(map[CacheKey]cacheEntry),
		defaultTTL: ttl,
		now:        time.Now,
		sf:         newSingleflight(),
	}
}

// Get returns a fresh entry. An expired entry is a miss and stays in the map:
// Set overwrites it, and deleting it here would drop a fresh Set that landed
// between the read and a second lock.
func (c *Cache) Get(key CacheKey) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	ttl := entry.ttl
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	if c.now().Sub(entry.timestamp) > ttl {
		return nil, false
	}
	return entry.data, true
}

func (c *Cache) Set(key CacheKey, data any) {
	c.SetWithTTL(key, data, 0)
}

// SetWithTTL stores data with a per-entry TTL. Pass ttl=0 to inherit the
// cache default. Per-entry TTL is used by M9 to keep Outdated entries
// fresh for 30m while other entries use the cache default (30s).
func (c *Cache) SetWithTTL(key CacheKey, data any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		data:      data,
		timestamp: c.now(),
		ttl:       ttl,
	}
}

func (c *Cache) Invalidate(keys ...CacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gen++
	for _, key := range keys {
		delete(c.entries, key)
	}
}

func (c *Cache) InvalidateFor(group string) {
	keys, ok := invalidateGroups[group]
	if !ok {
		return
	}
	c.Invalidate(keys...)
}

func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gen++
	c.entries = make(map[CacheKey]cacheEntry)
}

// generation is the invalidation count. A read notes it before it fetches, and
// its result is stored only if no invalidation came in between.
func (c *Cache) generation() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gen
}

// setIfCurrent stores data only when gen is still the current generation. A
// read that started before a write finished must not put its pre-write result
// back after the write invalidated the key.
func (c *Cache) setIfCurrent(key CacheKey, data any, ttl time.Duration, gen uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gen != gen {
		return
	}
	c.entries[key] = cacheEntry{
		data:      data,
		timestamp: c.now(),
		ttl:       ttl,
	}
}

// streamWrite starts a streaming brew write. The group's keys are dropped once
// the command has finished, whether it succeeded or failed: a read that ran
// while the command was going may have cached data from before it. The
// stream's error is forwarded unchanged, after the invalidation.
func streamWrite(ctx context.Context, runner Runner, cache *Cache, group string, args ...string) (<-chan string, <-chan error) {
	lines, errs := runner.ExecuteStream(ctx, args...)
	out := make(chan error, 1)
	go func() {
		defer close(out)
		err := <-errs
		cache.InvalidateFor(group)
		out <- err
	}()
	return lines, out
}

// cached is the one read path every brew reader uses. It returns the cached
// value of type T for key when there is one. Otherwise it runs fetch, once for
// all concurrent callers, and stores the result for ttl (0 = the cache
// default). fetch reports whether its result may be stored; a result it marks
// as not storable (a partial tap list) is returned but not cached.
func cached[T any](c *Cache, key CacheKey, ttl time.Duration, fetch func() (T, bool, error)) (T, error) {
	if val, ok := cachedValue[T](c, key); ok {
		return val, nil
	}

	// The flight key carries the generation, so a caller that starts after a
	// write finished never joins a fetch that began before it.
	gen := c.generation()
	flightKey := string(key) + "@" + strconv.FormatUint(gen, 10)
	val, err := c.sf.Do(flightKey, func() (any, error) {
		// A flight that finished just before this caller started has already
		// stored its result.
		if val, ok := cachedValue[T](c, key); ok {
			return val, nil
		}
		val, store, err := fetch()
		if err != nil {
			return nil, err
		}
		if store {
			c.setIfCurrent(key, val, ttl, gen)
		}
		return val, nil
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return val.(T), nil
}

func cachedValue[T any](c *Cache, key CacheKey) (T, bool) {
	var zero T
	val, ok := c.Get(key)
	if !ok {
		return zero, false
	}
	typed, ok := val.(T)
	return typed, ok
}
