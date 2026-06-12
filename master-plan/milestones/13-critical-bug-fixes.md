# Milestone 13 — Critical Bug Fixes

> **Status:** 🔲 Not Started  
> **Depends on:** Milestone 12 (Test Infrastructure)  
> **Enables:** Milestone 14 (Wire Dead Code)

---

## Goal

Fix bugs that will crash the program or produce incorrect results. These are not theoretical — they will manifest under real usage.

---

## Steps

### 13.1 — Fix Cache.Get RLock/Delete Race

**Problem:** `Cache.Get` holds `RLock` (read lock) but calls `delete(c.entries, key)` when an entry is expired. Multiple goroutines can hold `RLock` simultaneously. Go maps are not safe for concurrent read+write. This will cause `fatal error: concurrent map iteration and map write`.

**File:** `internal/brew/cache.go`

**Fix:** Check expiry under `RLock`, but acquire `Lock` for deletion:
```go
func (c *Cache) Get(key CacheKey) (any, bool) {
    c.mu.RLock()
    entry, ok := c.entries[key]
    if !ok {
        c.mu.RUnlock()
        return nil, false
    }
    if time.Since(entry.timestamp) > c.ttl {
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
```

**Test:** Add concurrent test that reads while TTL expires entries.

---

### 13.2 — Fix Shared KeyOutdated Cache Key

**Problem:** `formulaeReader.Outdated()` and `casksReader.Outdated()` both use `KeyOutdated`. Formulae caches `[]Formula`, casks caches `[]Cask`. The type assertion `cached.([]Formula)` fails when casks were cached first, causing a cache miss every time.

**Files:** `internal/brew/cache.go`, `internal/brew/formulae.go`, `internal/brew/casks.go`

**Fix:** Add separate cache keys:
```go
const (
    KeyOutdatedFormulae CacheKey = "outdated:formulae"
    KeyOutdatedCasks    CacheKey = "outdated:casks"
)
```

Update `InvalidateGroups` to map relevant operations to both keys.

**Test:** Verify formulae and casks outdated results are cached independently.

---

### 13.3 — Fix ConfirmModal Default Selection

**Problem:** `ConfirmModal.selected` defaults to 0, which maps to "Yes" (line 48-51: `if m.selected == 0 { m.result = true }`). Pressing Enter without navigating confirms the action. For destructive operations (uninstall, zap), this is dangerous.

**File:** `internal/gui/modal/confirm.go`

**Fix:** Default to index 1 ("No"):
```go
func NewConfirmModal(title, message string) *ConfirmModal {
    return &ConfirmModal{
        title:    title,
        message:  message,
        selected: 1,  // default to No
    }
}
```

Also fix the test: `TestConfirmModalDefaultNo` asserts `selected == 0` but should assert `selected == 1`.

---

### 13.4 — Fix padRight UTF-8 Handling

**Problem:** `padRight` uses byte length (`len(s)`) and byte slicing (`s[:length]`). Truncation at a byte boundary can split a multi-byte UTF-8 character, producing invalid output.

**File:** `internal/gui/presentation/formatters.go`

**Fix:** Use `utf8.RuneCountInString` and `utf8.DecodeRuneInString` for rune-based padding:
```go
func padRight(s string, length int) string {
    runes := []rune(s)
    count := len(runes)
    if count >= length {
        return string(runes[:length])
    }
    return s + strings.Repeat(" ", length-count)
}
```

**Test:** Add test with multi-byte characters (e.g., emoji, CJK).

---

## Tests

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestCacheGetConcurrentExpiry` | Unit | `internal/brew/cache_test.go` | No race under concurrent read + TTL expiry |
| `TestOutdatedCacheIndependent` | Unit | `internal/brew/cache_test.go` | Formulae and casks outdated cached separately |
| `TestConfirmModalDefaultIsNo` | Unit | `internal/gui/modal/modal_test.go` | Default selection is No (index 1) |
| `TestPadRightUTF8` | Unit | `internal/gui/presentation/formatters_test.go` | Multi-byte characters handled correctly |

---

## Definition of Done

- [ ] Cache.Get no longer deletes under RLock
- [ ] Formulae and casks outdated use separate cache keys
- [ ] ConfirmModal defaults to "No"
- [ ] padRight handles UTF-8 correctly
- [ ] All new tests pass
- [ ] `go test -race ./...` passes
- [ ] No regressions in existing tests
