package brew

import (
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
	KeyTrustList        CacheKey = "trust:list"
	KeyDoctorResult     CacheKey = "doctor:result"
	KeyConfig           CacheKey = "config"
)

var InvalidateGroups = map[string][]CacheKey{
	"install":    {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"uninstall":  {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"reinstall":  {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"upgrade":    {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks},
	"update":     {KeyFormulaeList, KeyCasksList, KeyOutdatedFormulae, KeyOutdatedCasks, KeyTapsList, KeyServicesList, KeyTrustList, KeyDoctorResult, KeyConfig},
	"pin":        {KeyFormulaeList, KeyOutdatedFormulae},
	"unpin":      {KeyFormulaeList, KeyOutdatedFormulae},
	"tap":        {KeyTapsList},
	"untap":      {KeyTapsList},
	"trust":      {KeyTrustList, KeyTapsList},
	"untrust":    {KeyTrustList, KeyTapsList},
	"cleanup":    {},
	"autoremove": {KeyFormulaeList, KeyOutdatedFormulae},
	"doctor":     {KeyDoctorResult},
}

type cacheEntry struct {
	data      any
	timestamp time.Time
	ttl       time.Duration // 0 = use cache default TTL
}

type Cache struct {
	mu         sync.RWMutex
	entries    map[CacheKey]cacheEntry
	defaultTTL time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries:    make(map[CacheKey]cacheEntry),
		defaultTTL: ttl,
	}
}

func (c *Cache) Get(key CacheKey) (any, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	if !ok {
		c.mu.RUnlock()
		return nil, false
	}
	ttl := entry.ttl
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	if time.Since(entry.timestamp) > ttl {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}
	data := entry.data
	c.mu.RUnlock()
	return data, true
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
		timestamp: time.Now(),
		ttl:       ttl,
	}
}

func (c *Cache) Invalidate(keys ...CacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		delete(c.entries, key)
	}
}

func (c *Cache) InvalidateFor(operation string) {
	keys, ok := InvalidateGroups[operation]
	if !ok {
		return
	}
	c.Invalidate(keys...)
}

func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[CacheKey]cacheEntry)
}

type TypedCache[T any] struct {
	cache *Cache
	key   CacheKey
}

func NewTypedCache[T any](cache *Cache, key CacheKey) *TypedCache[T] {
	return &TypedCache[T]{cache: cache, key: key}
}

func (tc *TypedCache[T]) Get() (T, bool) {
	val, ok := tc.cache.Get(tc.key)
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := val.(T)
	if !ok {
		tc.cache.Invalidate(tc.key)
		var zero T
		return zero, false
	}
	return typed, true
}

func (tc *TypedCache[T]) Set(val T) {
	tc.cache.Set(tc.key, val)
}
