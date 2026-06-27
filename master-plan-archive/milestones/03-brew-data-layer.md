# Milestone 3 — Brew Data Layer

> **Status:** ✅ Complete  
> **Depends on:** Milestone 1 (Foundation — types + runner)  
> **Enables:** Milestone 4 (Read-Only Panels)

---

## Goal

Implement the service layer that calls `brew` CLI commands, parses their JSON/text output into our domain types, and caches results. After this milestone, we have a fully tested data layer that can fetch all the information lazybrew needs — but it's not wired to the UI yet.

---

## Why This Milestone Matters

The data layer is the bridge between brew and the TUI. By building and testing it in isolation (with both mock and real brew output), we ensure the UI layer can trust the data it receives. This is also where we catch JSON parsing edge cases early (missing fields, different brew versions, empty lists, etc.).

---

## Steps

### 3.1 — Formulae Service (Read)

**What:** Parse `brew info --json=v2 --installed` output into `[]Formula`.

**File:** `internal/brew/formulae.go`

> **Design decision:** Read and write operations are split into separate interfaces. Reads are cacheable, concurrent-safe, and used by the panel layer. Writes go through the task manager (M6) and invalidate cache. This separation makes testing easier and prevents accidental cache corruption.

**Interface (Read):**
```go
type FormulaeReader interface {
    // List returns all installed formulae
    List(ctx context.Context) ([]Formula, error)
    
    // Get returns info for a specific formula
    Get(ctx context.Context, name string) (*Formula, error)
    
    // Outdated returns all outdated formulae (does NOT call List internally)
    Outdated(ctx context.Context) ([]Formula, error)
    
    // Leaves returns formulae not depended on by others
    Leaves(ctx context.Context) ([]string, error)
    
    // Deps returns dependency tree for a formula
    Deps(ctx context.Context, name string) (string, error)
    
    // Uses returns reverse dependencies (what depends on X)
    Uses(ctx context.Context, name string) ([]string, error)
}
```

> **Design note — `Outdated()` must NOT call `List()` internally.** The existing code in `formulae.go:211` calls `s.List(ctx)` to build a formula map for enrichment. This creates a hidden dependency: if `List()` fails, `Outdated()` fails even with valid outdated data. Instead, `Outdated()` should parse the outdated JSON independently and return `[]Formula` with only the fields populated from the outdated response. The panel layer can enrich with full formula data by calling `List()` separately if needed.

**Interface (Write):**
```go
type FormulaeWriter interface {
    // Install installs a formula (returns stream for progress)
    Install(ctx context.Context, name string) (<-chan string, <-chan error)
    
    // Uninstall removes a formula (returns stream — cask zap can produce output)
    Uninstall(ctx context.Context, name string) (<-chan string, <-chan error)
    
    // Reinstall reinstalls a formula
    Reinstall(ctx context.Context, name string) (<-chan string, <-chan error)
    
    // Upgrade upgrades a formula (or all if name is empty)
    Upgrade(ctx context.Context, name string) (<-chan string, <-chan error)
    
    // Pin / Unpin
    Pin(ctx context.Context, name string) error
    Unpin(ctx context.Context, name string) error
}
```

> **Design note — Consistent return types:** All write operations that produce output (install, uninstall, reinstall, upgrade) return `(<-chan string, <-chan error)`. Even `Uninstall` can produce output (e.g., cask zap removes files, formula uninstall prints caveats). This is more consistent than the current mix of `error` and stream returns.

**JSON v2 parsing:**
- The `brew info --json=v2` output has a top-level `{ "formulae": [...], "casks": [...] }` structure
- Each formula has nested objects: `versions`, `bottle`, `installed[]`, `dependencies`
- **Homebrew 6.0.0** adds new fields: `binaries` (executables list), `installed_dependents` (reverse deps with install status), `list_versions` (all version history), `shadowed` (PATH warnings)
- Handle edge cases: formula with no installed versions, missing fields, null values
- Map the `installed[0].installed_on_request` boolean to our `InstalledOnReq` field
- Parse `installed[0].time` as the install timestamp

**Acceptance criteria:**
- [ ] `List()` parses the testdata fixture into `[]Formula`
- [ ] All fields correctly mapped (version, deps, caveats, pinned, outdated, etc.)
- [ ] `Get()` calls `brew info --json=v2 <name>` and parses single formula
- [ ] `Outdated()` calls `brew outdated --json=v2 --formula` and enriches with version info
- [ ] `Leaves()` calls `brew leaves` and parses the line-separated output
- [ ] `Deps()` calls `brew deps --tree <name>` and returns raw tree text
- [ ] Install/Uninstall/Upgrade/Pin/Unpin methods build correct command args
- [ ] Unit tests with fixture data cover all fields
- [ ] Fuzz test for JSON parsing edge cases

---

### 3.2 — Casks Service

**What:** Parse cask data into `[]Cask`.

**File:** `internal/brew/casks.go`

**Interface (Read):**
```go
type CasksReader interface {
    List(ctx context.Context) ([]Cask, error)
    Get(ctx context.Context, name string) (*Cask, error)
    Outdated(ctx context.Context) ([]Cask, error)
}
```

**Interface (Write):**
```go
type CasksWriter interface {
    Install(ctx context.Context, name string) (<-chan string, <-chan error)
    Uninstall(ctx context.Context, name string) (<-chan string, <-chan error)
    Zap(ctx context.Context, name string) (<-chan string, <-chan error)
    Upgrade(ctx context.Context, name string) (<-chan string, <-chan error)
    // 6.0.0: cask pinning
    Pin(ctx context.Context, name string) error
    Unpin(ctx context.Context, name string) error
}
```

**Parsing notes:**
- Casks come from the same `--json=v2` output but in the `"casks"` array
- Cask JSON structure differs from formulae (has `artifacts`, `auto_updates`, etc.)
- Handle the `tap` field for attribution

**Acceptance criteria:**
- [ ] `List()` parses cask fixtures correctly
- [ ] All cask fields mapped
- [ ] `Outdated()` works with `--greedy` option awareness
- [ ] `Zap()` builds `brew uninstall --zap` command
- [ ] Unit tests cover all cask-specific fields

---

### 3.3 — Taps Service

**What:** Parse `brew tap` and `brew tap-info --json` output.

**File:** `internal/brew/taps.go`

**Interface:**
```go
type TapsService interface {
    List(ctx context.Context) ([]Tap, error)
    Get(ctx context.Context, name string) (*Tap, error)
    Tap(ctx context.Context, name string) error
    Untap(ctx context.Context, name string) error
}
```

**Parsing notes:**
- `brew tap` returns a simple line-per-tap list
- `brew tap-info --json <name>` returns detailed info per tap; **Homebrew 6.0.0** adds `trusted` (boolean) field and `formula_names`/`cask_names` arrays
- For listing, call `brew tap` first, then batch `brew tap-info --json` for details
- Determine `IsOfficial` by checking if tap name starts with `homebrew/`
- Determine `IsAPI` by checking if tap uses API-sourced mode
- Populate `Trusted` from the `trusted` field in tap-info JSON

**Acceptance criteria:**
- [ ] `List()` returns all taps with basic info
- [ ] `Get()` returns detailed tap info from JSON
- [ ] `IsOfficial` correctly identified
- [ ] `Tap()`/`Untap()` build correct commands
- [ ] Unit tests with fixture data

---

### 3.4 — Trust Service

**What:** Interface for `brew trust` / `brew untrust` commands.

**File:** `internal/brew/trust.go`

**Interface:**
```go
type TrustService interface {
    // ListTrusted returns all explicitly trusted items
    ListTrusted(ctx context.Context) ([]TrustEntry, error)
    
    // TrustTap trusts an entire tap
    TrustTap(ctx context.Context, tapName string) error
    
    // TrustFormula trusts a specific formula in a tap
    TrustFormula(ctx context.Context, fullName string) error
    
    // TrustCask trusts a specific cask in a tap
    TrustCask(ctx context.Context, fullName string) error
    
    // UntarustTap removes trust from a tap
    UntrustTap(ctx context.Context, tapName string) error
    
    // UntrustFormula removes trust from a formula
    UntrustFormula(ctx context.Context, fullName string) error
    
    // UntrustCask removes trust from a cask
    UntrustCask(ctx context.Context, fullName string) error
    
    // GetTapTrustStatus returns the trust status for a specific tap
    GetTapTrustStatus(ctx context.Context, tapName string) (TrustStatus, error)
}
```

**Notes:**
- This wraps `brew trust` and `brew untrust` with the `--formula`, `--cask`, `--command` flags
- **Homebrew 6.0.0:** `brew trust` now has `--json=v1` flag for machine-readable trust state. Use this instead of parsing text output.
- Trust status can also be read from `brew tap-info --json` which now includes a `trusted` field (6.0.0+)
- Official taps (homebrew/*) are always `TrustOfficial`

**Acceptance criteria:**
- [ ] All trust/untrust commands build correct args
- [ ] Official taps return `TrustOfficial` without calling brew
- [ ] Unit tests for command building
- [ ] Integration test that calls `brew trust --help` (validates command exists)

---

### 3.5 — Services Service

**What:** Parse `brew services list --json` output.

**File:** `internal/brew/services.go`

**Interface:**
```go
type ServicesService interface {
    List(ctx context.Context) ([]Service, error)
    Start(ctx context.Context, name string) error
    Stop(ctx context.Context, name string) error
    Restart(ctx context.Context, name string) error
    Run(ctx context.Context, name string) error
}
```

**Platform note:**
- On Linux, `brew services` uses systemd (user units)
- On macOS, it uses launchctl
- lazybrew doesn't care about the underlying mechanism — just calls `brew services`
- On systems without services support, `List()` returns empty with no error

**Acceptance criteria:**
- [ ] JSON parsing of services list
- [ ] All status types handled (started, stopped, error, none)
- [ ] Start/Stop/Restart/Run build correct commands
- [ ] Graceful handling when services not available

---

### 3.6 — Search Service

**What:** Parse `brew search --json=v2` output (JSON, not text).

**File:** `internal/brew/search.go`

> **Design decision:** Use `brew search --json=v2` instead of parsing text output. Text parsing is fragile — brew could change header format, installed markers, etc. JSON is stable and typed.

**Interface:**
```go
type SearchService interface {
    // Search searches for formulae and casks by name (uses --json=v2)
    Search(ctx context.Context, query string) ([]SearchResult, error)
    
    // SearchDesc searches in descriptions too (uses --json=v2 --desc)
    SearchDesc(ctx context.Context, query string) ([]SearchResult, error)
}
```

**Parsing notes:**
- `brew search --json=v2` returns structured JSON with formulae and casks arrays
- Each result includes name, tap, description, and whether it's installed
- No need to parse text headers (`==> Formulae`, `==> Casks`) or detect installed markers (`✓`)
- Handle empty results gracefully (empty arrays, not errors)
- Handle regex search (`/pattern/`) — brew supports this natively

**Acceptance criteria:**
- [ ] Correctly splits formulae and cask results
- [ ] Detects installed markers (`✓`)
- [ ] Handles empty results
- [ ] Handles regex search (`/pattern/`)
- [ ] Unit tests with sample search output

---

### 3.7 — Diagnostics & Misc Commands

**What:** Parse `brew doctor`, `brew config`, `brew --version`, `brew missing`.

**File:** `internal/brew/doctor.go`

> **Design decision:** Split into two interfaces — `DiagnosticsReader` (read-only, cacheable) and `DiagnosticsWriter` (mutations that go through the task manager). This prevents accidental cache corruption and makes the API clearer.

**Interface (Read):**
```go
type DiagnosticsReader interface {
    // Doctor runs brew doctor and returns warnings
    Doctor(ctx context.Context) ([]DoctorWarning, error)
    
    // Missing runs brew missing and returns missing dependencies
    Missing(ctx context.Context) ([]MissingDep, error)
    
    // Config returns brew configuration
    Config(ctx context.Context) (*BrewConfig, error)
    
    // Version returns brew version string
    Version(ctx context.Context) (string, error)
}

// MissingDep represents a missing dependency (from brew missing)
type MissingDep struct {
    Formula string // the formula with the missing dep
    Missing string // the missing dependency name
}
```

**Interface (Write):**
```go
type DiagnosticsWriter interface {
    // Update runs brew update (returns stream for progress)
    Update(ctx context.Context) (<-chan string, <-chan error)
    
    // Cleanup runs cleanup with optional dry-run
    Cleanup(ctx context.Context, dryRun bool) (<-chan string, <-chan error)
    
    // Autoremove removes orphaned dependencies
    Autoremove(ctx context.Context, dryRun bool) (<-chan string, <-chan error)
}
```

**Parsing notes:**
- `brew doctor` outputs warnings as blocks separated by "Warning:" headers
- If all clear, outputs "Your system is ready to brew."
- `brew config` outputs key-value pairs, one per line
- `brew --version` returns version string

**Acceptance criteria:**
- [ ] Doctor warnings parsed into structured objects
- [ ] "All clear" case handled
- [ ] Config key-value pairs parsed
- [ ] Version string extracted
- [ ] Streaming commands (update, cleanup) work via channels

---

### 3.8 — Cache Layer

**What:** In-memory cache with TTL and typed invalidation-on-write.

**File:** `internal/brew/cache.go`

> **Design decision:** Use Go generics for type-safe cache access. The current `interface{}` approach requires type assertions at every call site, which is error-prone. Also, use a typed invalidation registry instead of hardcoded key strings scattered across service methods.

**Design:**
```go
// CacheKey is a typed cache key (prevents string typos)
type CacheKey string

const (
    KeyFormulaeList  CacheKey = "formulae:list"
    KeyCasksList     CacheKey = "casks:list"
    KeyOutdated      CacheKey = "outdated:all"
    KeyTapsList      CacheKey = "taps:list"
    KeyServicesList  CacheKey = "services:list"
    KeyTrustList     CacheKey = "trust:list"
    KeyDoctorResult  CacheKey = "doctor:result"
    KeyConfig        CacheKey = "config"
)

// InvalidateGroup maps a write operation to the keys it invalidates
var InvalidateGroups = map[string][]CacheKey{
    "install":   {KeyFormulaeList, KeyCasksList, KeyOutdated},
    "uninstall": {KeyFormulaeList, KeyCasksList, KeyOutdated},
    "reinstall": {KeyFormulaeList, KeyCasksList, KeyOutdated},
    "upgrade":   {KeyFormulaeList, KeyCasksList, KeyOutdated},
    "update":    {KeyFormulaeList, KeyCasksList, KeyOutdated, KeyTapsList, KeyServicesList, KeyTrustList, KeyDoctorResult, KeyConfig},
    "pin":       {KeyFormulaeList, KeyOutdated},
    "unpin":     {KeyFormulaeList, KeyOutdated},
    "tap":       {KeyTapsList},
    "untap":     {KeyTapsList},
    "trust":     {KeyTrustList, KeyTapsList},
    "untrust":   {KeyTrustList, KeyTapsList},
    "cleanup":   {},
    "autoremove": {KeyFormulaeList, KeyOutdated},
}

type Cache struct {
    mu      sync.RWMutex
    entries map[CacheKey]cacheEntry
    ttl     time.Duration
}

type cacheEntry struct {
    data      any
    timestamp time.Time
}

// Get returns the cached value for key, or (zero, false) if not found/expired.
// Type parameter T must match the type stored by Set.
func (c *Cache) Get[T any](key CacheKey) (T, bool)

func (c *Cache) Set(key CacheKey, data any)
func (c *Cache) Invalidate(keys ...CacheKey)
func (c *Cache) InvalidateFor(operation string) // uses InvalidateGroups
func (c *Cache) InvalidateAll()
```

> **Usage pattern:** Service methods call `s.cache.InvalidateFor("install")` instead of hardcoding key strings. This centralizes invalidation logic and prevents bugs where a key is renamed in one place but not another.

> **Rate limiting:** The cache should also enforce a minimum interval between identical read calls (e.g., 5 seconds). If `List()` was called 2 seconds ago and the cache expired, don't immediately re-fetch — return the stale data with a `Stale` flag. The panel layer can decide whether to refresh.

**Acceptance criteria:**
- [ ] Get/Set/Invalidate work correctly
- [ ] TTL expiry works (entries expire after duration)
- [ ] Thread-safe (RWMutex)
- [ ] InvalidateAll clears everything
- [ ] Unit tests for all cache behaviors including TTL and concurrency

---

### 3.9 — Unified Brew Client

**What:** A single entry point that combines all services.

**File:** `internal/brew/client.go`

```go
type Client struct {
    // Read services (cacheable, concurrent-safe)
    Formulae    FormulaeReader
    Casks       CasksReader
    Taps        TapsService
    Trust       TrustService
    Services    ServicesService
    Search      SearchService
    Diagnostics DiagnosticsReader
    
    // Write services (go through task manager)
    FormulaeWrite    FormulaeWriter
    CasksWrite       CasksWriter
    DiagnosticsWrite DiagnosticsWriter
    
    Cache       *Cache
}

func NewClient(runner Runner) *Client
func NewMockClient() *Client  // for testing
```

> **Design note:** The client exposes read and write services separately. The GUI layer calls read services directly for panel data, and routes write operations through the task manager (M6). This prevents the GUI from accidentally calling a write operation without going through the task queue.

**Acceptance criteria:**
- [ ] `NewClient` wires up all services with a shared runner and cache
- [ ] `NewMockClient` creates a client with mock data for TUI testing
- [ ] All services accessible through the client

---

### 3.10 — Non-Interactive Brew Execution (6.0.0 Requirement)

**What:** Ensure brew runs non-interactively when called from the TUI.

**Context:** Homebrew 6.0.0 made "ask mode" the default, meaning `brew install` and `brew upgrade` will emit a dependency summary and confirmation prompt when stdin is a TTY. Since lazybrew calls brew as a subprocess, we must guarantee it runs without user prompts.

**Implementation in `internal/brew/runner.go`:**
- When creating a brew command (`exec.CommandContext`), set `Stdin` to `nil` or `os.Stdin` to a non-TTY pipe (e.g., `bytes.NewReader(nil)`)
- Alternatively, set `HOMEBREW_NO_ASK=1` in the command's environment
- The runner already abstracts brew execution, so this is a one-line change

**Acceptance criteria:**
- [ ] `brew install`, `brew upgrade`, `brew uninstall` never prompt for stdin confirmation
- [ ] Integration test verifies no interactive prompt when running through the runner
- [ ] `HOMEBREW_NO_ASK` or equivalent non-interactive mechanism is documented

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestFormulaeList` | Unit | `internal/brew/formulae_test.go` | JSON parsing of formulae list |
| `TestFormulaeFieldMapping` | Unit | `internal/brew/formulae_test.go` | Every field mapped correctly from fixture |
| `TestFormulaeEmptyList` | Unit | `internal/brew/formulae_test.go` | Empty install list handled |
| `TestCasksList` | Unit | `internal/brew/casks_test.go` | JSON parsing of casks list |
| `TestTapsList` | Unit | `internal/brew/taps_test.go` | Text + JSON parsing |
| `TestTrustCommands` | Unit | `internal/brew/trust_test.go` | Command arg building |
| `TestServicesList` | Unit | `internal/brew/services_test.go` | JSON parsing |
| `TestSearchParsing` | Unit | `internal/brew/search_test.go` | Text output parsing with sections |
| `TestSearchInstalled` | Unit | `internal/brew/search_test.go` | Installed marker detection |
| `TestDoctorParsing` | Unit | `internal/brew/doctor_test.go` | Warning block parsing |
| `TestDoctorAllClear` | Unit | `internal/brew/doctor_test.go` | Clean system output |
| `TestConfigParsing` | Unit | `internal/brew/doctor_test.go` | Key-value config parsing |
| `TestCacheTTL` | Unit | `internal/brew/cache_test.go` | Entries expire after TTL |
| `TestCacheInvalidation` | Unit | `internal/brew/cache_test.go` | Specific key invalidation |
| `TestCacheConcurrency` | Unit | `internal/brew/cache_test.go` | Thread safety under concurrent access |
| `FuzzFormulaeJSON` | Fuzz | `internal/brew/formulae_fuzz_test.go` | JSON parsing doesn't panic on malformed input |
| `FuzzSearchOutput` | Fuzz | `internal/brew/search_fuzz_test.go` | Search parsing doesn't panic |
| `TestRealBrewList` | Integration | `internal/brew/formulae_integration_test.go` | Real brew call returns parseable output |
| `TestRealBrewSearch` | Integration | `internal/brew/search_integration_test.go` | Real search returns results |

---

## Definition of Done

- [ ] All read services implement their interfaces (FormulaeReader, CasksReader, etc.)
- [ ] All write services implement their interfaces (FormulaeWriter, CasksWriter, etc.)
- [ ] Read and write services are separate interfaces
- [ ] All JSON/text parsing works against synthetic fixture data
- [ ] Cache layer uses generics (`Get[T]`) and typed `CacheKey` constants
- [ ] Cache invalidation uses `InvalidateFor(operation)` with centralized groups
- [ ] Search uses `--json=v2` (not text parsing)
- [ ] `Outdated()` does NOT call `List()` internally
- [ ] `brew missing` is implemented (DiagnosticsReader.Missing)
- [ ] Unified `Client` struct wires everything together
- [ ] `MockClient` available for TUI testing
- [ ] All unit tests pass
- [ ] Fuzz tests run without panics (at least 10s each)
- [ ] Integration tests pass when brew is available
- [ ] No data layer depends on any GUI code
