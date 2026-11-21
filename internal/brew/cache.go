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

// invalidateGroups names the cache keys each group drops: the write groups,
// which every writer calls once its command has finished, and the panel
// groups the GUI drops before a refresh (see InvalidateFor).
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

	"formulae":    {KeyFormulaeList},
	"casks":       {KeyCasksList},
	"outdated":    {KeyOutdatedFormulae, KeyOutdatedCasks},
	"taps":        {KeyTapsList},
	"status":      {KeyFormulaeList, KeyCasksList, KeyTapsList, KeyServicesList},
	"diagnostics": {KeyDoctorResult, KeyConfig},
}

type cacheEntry struct {
	data      any
	timestamp time.Time
	ttl       time.Duration // 0 = use cache default TTL
}

type Cache struct {
	mu         sync.RWMutex
	entries    map[CacheKey]cacheEntry
	seq        uint64              // the last invalidation's stamp; see generationLocked
	keyGen     map[CacheKey]uint64 // the stamp of the last invalidation that dropped each key
	defaultTTL time.Duration
	now        func() time.Time // time.Now; a test swaps in a clock it moves by hand
	sf         *singleflight    // coalesces concurrent fetches of one key
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries:    make(map[CacheKey]cacheEntry),
		keyGen:     make(map[CacheKey]uint64),
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
	c.seq++
	for _, key := range keys {
		delete(c.entries, key)
		c.keyGen[key] = c.seq
	}
}

func (c *Cache) InvalidateFor(group string) {
	keys, ok := invalidateGroups[group]
	if !ok {
		return
	}
	c.Invalidate(keys...)
}

// generation is the stamp of the last invalidation that dropped key. A read
// notes it before it fetches, and its result is stored only if it has not moved
// since. Every stamp comes from seq, so a change always yields a new value.
func (c *Cache) generation(key CacheKey) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.generationLocked(key)
}

// generationLocked is generation for a caller that already holds c.mu.
func (c *Cache) generationLocked(key CacheKey) uint64 {
	return c.keyGen[key]
}

// setIfCurrent stores data only when key's generation is still gen. A read that
// started before a write finished must not put its pre-write result back after
// the write invalidated the key.
func (c *Cache) setIfCurrent(key CacheKey, data any, ttl time.Duration, gen uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.generationLocked(key) != gen {
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

	// The flight key carries the key's generation, so a caller that starts after
	// a write to this key finished never joins a fetch that began before it.
	gen := c.generation(key)
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
