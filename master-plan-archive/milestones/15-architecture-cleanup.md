# Milestone 15 — Architecture Cleanup

> **Status:** ⚠️ Mostly done  
> **Remaining:** TypedCache panic path → M19.0; verify Reader/Writer splits complete  
> **Depends on:** Milestone 14 (Wire Dead Code)  
> **Enables:** Milestone 16 (Test Coverage)

---

## Active Work Routing

> **Format:** Legacy.

| Open item | Execute in |
|---|---|
| TypedCache safe get (panic on wrong type) | [M19.0](19-bubble-tea-concurrency-and-task-manager.md) |
| Remaining DoD items | [M18.4](18-documentation-and-project-hygiene.md) audit |

---

## Goal

Fix design inconsistencies and improve type safety. These issues don't cause crashes but make the code harder to maintain, test, and extend.

---

## Steps

### 15.1 — Consistent Read/Write Interface Splits

**Problem:** `FormulaeReader`/`FormulaeWriter` and `CasksReader`/`CasksWriter` split read/write operations. But `TapsService`, `ServicesService`, `TrustService` combine both in single interfaces. This inconsistency makes the API confusing.

**Files:** `internal/brew/taps.go`, `internal/brew/services.go`, `internal/brew/trust.go`

**Fix:** Split each into Reader/Writer:
```go
type TapsReader interface {
    List(ctx context.Context) ([]Tap, error)
    Get(ctx context.Context, name string) (*Tap, error)
}

type TapsWriter interface {
    Tap(ctx context.Context, name string) error
    TapWithURL(ctx context.Context, name, url string) error
    Untap(ctx context.Context, name string) error
    Repair(ctx context.Context, name string) error
}
```

Apply same pattern to `ServicesService` and `TrustService`.

Update `Client` struct to use the split interfaces.

---

### 15.2 — Remove interface{} from Cache

**Problem:** `Cache.Get` returns `any` and `Cache.Set` accepts `any`. Every caller must type-assert, and the `KeyOutdated` collision (fixed in M13) demonstrated the danger.

**File:** `internal/brew/cache.go`

**Fix:** Use a typed cache with generics. Since Go 1.26 doesn't support method-level generics, use a wrapper pattern:
```go
type TypedCache[T any] struct {
    cache *Cache
    key   CacheKey
}

func (tc *TypedCache[T]) Get() (T, bool) {
    val, ok := tc.cache.Get(tc.key)
    if !ok {
        var zero T
        return zero, false
    }
    return val.(T), ok
}

func (tc *TypedCache[T]) Set(val T) {
    tc.cache.Set(tc.key, val)
}
```

Each service creates its own `TypedCache[[]Formula]`, `TypedCache[[]Cask]`, etc.

---

### 15.3 — Remove interface{} from panelData

**Problem:** `panelData.rawData` is `interface{}`. Every accessor (`selectedFormula`, `selectedCask`, etc.) must type-assert. This is error-prone.

**File:** `internal/gui/panel.go`

**Fix:** Replace `rawData interface{}` with typed fields:
```go
type panelData struct {
    // ... existing fields ...
    formulae []brew.Formula
    casks    []brew.Cask
    taps     []brew.Tap
    services []brew.Service
}
```

Update `DataLoadedMsg` to include typed fields instead of `RawData interface{}`.

---

### 15.4 — Remove interface{} from Messages

**Problem:** `DataLoadedMsg.RawData` and `ModalDoneMsg.Result` use `interface{}`.

**Files:** `internal/gui/messages.go`

**Fix:** 
- For `DataLoadedMsg`: Use typed fields (see 15.3).
- For `ModalDoneMsg`: Remove it (it's dead code, see M14 step 14.8).

---

### 15.5 — Consistent Receiver Types

**Problem:** Methods on `Model` use mixed value and pointer receivers. `doMutation` uses value receiver but modifies `m.activeModal`. `startSearch` uses pointer receiver. This is inconsistent.

**Files:** `internal/gui/gui.go`, `internal/gui/commands.go`

**Fix:** Use pointer receivers for all methods that modify the model. Use value receivers only for read-only methods (like `View()`, `renderContent()`).

Update all methods in `commands.go` to use consistent receivers.

---

### 15.6 — Remove Duplicate itoa

**Problem:** Custom `itoa` function exists in `panel.go` and `menu.go`. `strconv.Itoa` exists in the standard library.

**Files:** `internal/gui/panel.go`, `internal/gui/modal/menu.go`

**Fix:** Replace all uses of custom `itoa` with `strconv.Itoa`. Remove the custom implementations.

---

### 15.7 — Remove jsonUnmarshal Inconsistency

**Problem:** `casks.go` uses `jsonUnmarshal` (a package-level variable function) while `formulae.go` and `services.go` use `json.Unmarshal` directly. The pattern is applied inconsistently.

**File:** `internal/brew/casks.go`

**Fix:** Remove `jsonUnmarshal` variable and use `json.Unmarshal` directly everywhere.

---

### 15.8 — Fix Global Logger/Theme Race Potential

**Problem:** `Logger` in `logger.go` and theme variables in `theme.go` are package-level variables replaced without synchronization. If any goroutine reads them while they're being replaced, there's a data race.

**Files:** `internal/brew/logger.go`, `internal/gui/style/theme.go`

**Fix:** Use `sync.Once` for initialization and `atomic.Value` for updates:
```go
var logger atomic.Value

func init() {
    logger.Store(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelWarn,
    })))
}

func Logger() *slog.Logger {
    return logger.Load().(*slog.Logger)
}

func SetDebug(enabled bool) {
    if enabled {
        logger.Store(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })))
    } else {
        logger.Store(slog.New(slog.NewTextHandler(io.Discard, nil)))
    }
}
```

Apply same pattern to theme variables.

---

### 15.9 — Unexport Program Field

**Problem:** `Model.Program` is exported, allowing external code to call `Send()` and bypass the Update loop.

**File:** `internal/gui/gui.go`

**Fix:** Rename to `program` (unexported). Provide a `SetProgram(p *tea.Program)` method for initialization.

---

### 15.10 — Move NewMockClient to Test File

**Problem:** `NewMockClient` is only useful for tests but lives in production code (`client.go`).

**File:** `internal/brew/client.go`

**Fix:** Move `NewMockClient` to `client_test.go` or a separate `testutil` package.

---

## Tests

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestTypedCacheGetSet` | Unit | `internal/brew/cache_test.go` | Typed cache works without type assertions |
| `TestPanelDataTypedFields` | Unit | `internal/gui/panel_test.go` | Typed fields work correctly |
| `TestLoggerConcurrency` | Unit | `internal/brew/logger_test.go` | Logger replacement is race-free |
| `TestThemeConcurrency` | Unit | `internal/gui/style/theme_test.go` | Theme replacement is race-free |

---

## Definition of Done

- [ ] All services use consistent Reader/Writer interface splits
- [ ] Cache uses typed wrappers (no interface{} in public API)
- [ ] panelData uses typed fields (no interface{})
- [ ] Messages use typed fields (no interface{})
- [ ] All methods use consistent receiver types
- [ ] Custom itoa removed, strconv.Itoa used everywhere
- [ ] jsonUnmarshal removed, json.Unmarshal used everywhere
- [ ] Logger and Theme use atomic.Value for race safety
- [ ] Program field is unexported
- [ ] NewMockClient moved to test file
- [ ] All new tests pass
- [ ] `go test -race ./...` passes
- [ ] No regressions in existing tests
