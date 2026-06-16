# M24 — Outdated Panel Performance

> **Status:** 🔜 Planned  
> **Size estimate:** M (1–3 days)  
> **Depends on:** M19 ✅ (TaskManager stable), M20 ✅ (Outdated panel wired), M22 ✅ (release tag)  
> **Enables:** M26 (Taps batch loading uses M24.2 coalescing), M27 (tiered refresh builds on M24.3 lazy load + M24.4 TTL)  
> **Parallel track:** —  
> **Gate criteria:** `brew outdated` perceived load time is measurably reduced without regressing correctness or mutation side effects.

<!-- See [backlog.md](../backlog.md) B-02 for deferred item -->
<!-- See [archive/architecture-review-2026-06-13.md](../archive/architecture-review-2026-06-13.md) H10/M10 startup findings -->
<!-- Related: M25 fixes `brew doctor` exit=1; M26 batch tap-info; M27 tiered refresh -->

---

## Goal

The Outdated panel should appear quickly when the user opens it. Today it is one of the slowest parts of startup because it triggers redundant `brew outdated` subprocesses. This milestone removes those redundancies: combine the formula/cask outdated fetch into a single Homebrew call, deduplicate concurrent requests, and only fetch Outdated data when the panel is first viewed.

---

## Readiness to Start

Before executing this milestone, confirm:

- [ ] M22 release tag is done.
- [ ] Combined-reader design in 24.1 is accepted (new `Client.Outdated()` method).
- [ ] Existing Outdated tests (`TestBatchUpgradeCallsBrewWithSelectedNames`, `TestOutdatedFetchSurfacesError`) pass on `main`.
- [ ] Decision on M24.4 → M27.2 TTL migration is understood (M24's `cache_ttl_seconds` becomes volatile TTL in M27).

---

## Why Now

- The architecture review (2026-06-13) flagged startup as a performance concern, and the Outdated panel is the clearest user-visible symptom.
- All functional/release milestones (M17–M23) are complete, so this is the right time for a focused performance pass without destabilizing the v0.2.0 tag.
- The fix is small and localized (cache + `fetchPanelData` + `Outdated` readers), but it materially improves daily use.

---

## Challenged Assumptions

| Assumption | Challenge | Decision |
|---|---|---|
| Separate `brew outdated --formula` and `brew outdated --cask` calls are clearer. | They double subprocess overhead and both calls can block concurrently. | Use a single `brew outdated --json=v2` call that returns both formulae and casks. |
| Eagerly fetching Outdated on `Init()` is required for the Status count. | Status can read the same cached result once it is available; eager fetch wastes work if the user never opens Outdated. | Lazy-load Outdated on first focus; Status shows spinner/"—" until data arrives. |
| The 30 s hardcoded cache TTL is sufficient. | Users with many packages may want to tune freshness vs speed. | Expose `cache_ttl_seconds` in config (default 30 s) as part of this milestone. |

---

## Out of Scope

- **General controller refactor** — deferred to [backlog.md](../backlog.md) B-01.
- **Lazy loading for other panels** — covered by **M27** tiered refresh strategy.
- **Background prefetch / auto-refresh policy redesign** — covered by **M27**; this milestone only wires a single cache TTL knob.
- **`brew doctor` exit-code bug** — covered by **M25**.
- **Taps `tap-info` batch loading** — covered by **M26**.

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|---|---|---|---|
| D24-1 | Single `brew outdated --json=v2` call returns both formulae and casks. | Two separate calls; one call plus manual filtering. | Homebrew natively returns both in one JSON document; fewer process spawns and less JSON parsing. |
| D24-2 | Request coalescing for in-flight `Outdated()` calls. | Leave cache-only dedupe; add a global brew lock. | A per-key in-flight promise/lock is precise and keeps other panels concurrent. |
| D24-3 | Lazy Outdated fetch on first panel focus. | Keep eager `Init()` fetch; prefetch after Status. | Avoids work when the user never visits Outdated; Status count updates as soon as data arrives. |
| D24-4 | Configurable cache TTL (`cache_ttl_seconds`, default 30). | Hardcoded TTL; stale-while-revalidate. | Simplest win for users who prefer speed over freshness. |

Copy stable ADRs to `DESIGN.md` decision log when merged.

---

## Phases

Execute phases **in order** unless [status.md](../status.md) parallel tracks document otherwise.

| Phase | Steps | Theme | Phase gate |
|---|---|---|---|
| **A — Measure & combine** | 24.1, 24.2 | Reduce `brew outdated` call count | Unit tests prove one subprocess returns both lists |
| **B — Lazy load** | 24.3 | Fetch only when needed | teatest/e2e confirms Outdated loads on focus, not at startup |
| **C — Tune & verify** | 24.4, 24.5 | Config + regression | All tests pass, no mutation regressions |

---

## Step Index

| Step | Title | Size | Depends | Deliverable |
|---|---|---|---|---|
| 24.1 | Add combined Outdated reader | S | — | `client.Outdated()` returning formulae + casks |
| 24.2 | Request coalescing for Outdated | S | 24.1 | In-flight lock/promise in `brew` cache layer |
| 24.3 | Lazy-load Outdated panel | S | 24.2 | `Init()` no longer fetches Outdated; focus triggers fetch |
| 24.4 | Configurable cache TTL | S | — | `cache_ttl_seconds` in config, wired to `NewClient` |
| 24.5 | Regression tests & smoke | S | 24.3, 24.4 | Unit + E2E coverage, `make test` green |

---

## Steps

### 24.1 — Add Combined Outdated Reader

**Size:** S  
**Phase:** A  
**Depends on:** —  
**Blocks:** 24.2

**Context:** `Formulae.Outdated()` and `Casks.Outdated()` each call `brew outdated` with a type flag. A single `brew outdated --json=v2` returns both `formulae` and `casks` keys, so we can halve the subprocess cost.

**Design decision:** Add a new top-level reader method on `Client`:

```go
// Client.Outdated returns both outdated formulae and casks from one brew call.
Outdated(ctx context.Context) ([]Formula, []Cask, error)
```

Implementation lives in a small unexported helper in `internal/brew/outdated.go` (new file). `Formulae.Outdated()` and `Casks.Outdated()` become cache-aware delegates: they first check `KeyOutdatedFormulae` / `KeyOutdatedCasks`; on miss they call `Client.Outdated()` and return only their slice.

**Implementation checklist:**
1. Create `internal/brew/outdated.go` with `combinedOutdatedReader` and `parseOutdatedJSON`.
2. Add `Outdated(ctx) ([]Formula, []Cask, error)` to `Client`.
3. Refactor `Formulae.Outdated()` and `Casks.Outdated()` to delegate to `Client.Outdated()` after a cache check.
4. Cache results under both `KeyOutdatedFormulae` and `KeyOutdatedCasks`.
5. Update `fetchPanelData(PanelOutdated)` to call `client.Outdated()` directly.
6. Update `fetchStatusData()` to call `client.Outdated()` directly instead of the separate readers.

**Files:**

| File | Action |
|---|---|
| `internal/brew/outdated.go` | Create combined reader |
| `internal/brew/formulae.go` | Make `Outdated()` a cache + delegate |
| `internal/brew/casks.go` | Make `Outdated()` a cache + delegate |
| `internal/brew/client.go` | Expose `Outdated()` on `Client` |
| `internal/gui/commands.go` | Use `client.Outdated()` in Outdated/Status fetch |

**Acceptance criteria:**
- [ ] `brew outdated --json=v2` is invoked once when the Outdated panel loads.
- [ ] Panel still renders formula and cask outdated items separately.

**Tests (same change set):**
- [ ] `TestCombinedOutdatedReturnsBothTypes` — combined reader parses both formulae and casks from one JSON document.
- [ ] `TestOutdatedPanelFetchUsesSingleBrewCall` — mock runner records exactly one `outdated --json=v2` call.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Older Homebrew versions behave differently without flags. | Target is 6.0.0+ per `status.md`; document if older versions need flags. |

**Rollback:** Revert to separate `Formulae.Outdated()` / `Casks.Outdated()` calls.

---

### 24.2 — Request Coalescing for Outdated

**Size:** S  
**Phase:** A  
**Depends on:** 24.1  
**Blocks:** 24.3

**Context:** `Init()` currently starts `fetchPanelData(PanelOutdated)` and `fetchStatusData()` in parallel. Both call `Outdated()`, so before the cache is populated two `brew outdated` subprocesses run. A request coalescing layer ensures only one in-flight call per cache key.

**Implementation checklist:**
1. Add a `type cachePromise struct { once sync.Once; val any; err error; done chan struct{} }`.
2. Add `promises map[CacheKey]*cachePromise` to `Cache`, protected by the existing mutex or a separate mutex.
3. In `Cache.Get`, on miss check `promises[key]`. If found, wait on `done` and return the promise's `val`/`err`.
4. In `Cache.Get`, on miss with no promise: create a promise, store it, release the lock, execute the fetch, set promise result, close `done`, remove the promise from the map, and call `Set`.
5. On panic or error, store the error in the promise so waiters see it, then remove the promise.
6. Apply this generically to all cache keys; Outdated is the first beneficiary.

**Files:**

| File | Action |
|---|---|
| `internal/brew/cache.go` | Add in-flight promise map |
| `internal/brew/cache_test.go` | Add concurrent-dedupe test |

**Acceptance criteria:**
- [ ] Two concurrent `Outdated()` calls result in exactly one `brew outdated` subprocess.
- [ ] Both callers receive the same result.

**Tests (same change set):**
- [ ] `TestCacheCoalescesConcurrentOutdatedCalls` — race-free, single subprocess.
- [ ] `make test -race` passes.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Promise leaks on panic/error. | Use `sync.Once` or deferred cleanup; test error path. |

**Rollback:** Disable coalescing by skipping the promise map.

---

### 24.3 — Lazy-Load Outdated Panel

**Size:** S  
**Phase:** B  
**Depends on:** 24.2  
**Blocks:** 24.5

**Context:** The Outdated panel is fetched on every startup even though the user may never open it. With request coalescing in place, lazy loading is safe: the first focus will trigger the combined fetch, and Status will pick up the result when ready.

**Implementation checklist:**
1. Remove `fetchPanelData(m.client, PanelOutdated)` from `Init()`.
2. On panel switch to `PanelOutdated`, if `panel.items` is empty and not loading, enqueue `fetchPanelData(PanelOutdated)`.
3. Ensure Status dashboard shows a placeholder count (e.g., "—") until outdated data arrives, then updates via `DataLoadedMsg`.
4. Keep `r`/`R` refresh key functional for Outdated.

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Remove Outdated from `Init()`; add lazy fetch on focus |
| `internal/gui/commands.go` | Ensure refresh path still works |

**Acceptance criteria:**
- [ ] `brew outdated` is not invoked during `Init()` when the app starts on another panel.
- [ ] Switching to Outdated triggers the fetch and shows a spinner/loading state.
- [ ] Status dashboard count updates once Outdated data loads.

**Tests (same change set):**
- [ ] `TestOutdatedLazyLoadedOnFocus` — teatest/e2e: startup View does not contain outdated data; after focus it does.
- [ ] `TestStatusUpdatesAfterOutdatedLoads` — unit test around `DataLoadedMsg` handling.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| User sees empty panel briefly on first focus. | Show spinner/"Loading outdated packages..." state. |

**Rollback:** Add Outdated back to `Init()`.

---

### 24.4 — Configurable Cache TTL

**Size:** S  
**Phase:** C  
**Depends on:** —  
**Blocks:** 24.5

**Context:** `Client.NewCache(30 * time.Second)` is hardcoded. Users with large Homebrew installs may prefer a longer TTL to avoid repeated slow `brew info --installed` and `brew outdated` calls.

**Relationship to M27:** M24.4 introduces a single global knob. M27 will generalize this into per-class TTLs (volatile/stable/manual). `brew.cache_ttl_seconds` therefore becomes the **default/volatile TTL** in M27 and must remain backward-compatible.

**Implementation checklist:**
1. Add `CacheTTLSeconds int` to `config.BrewConfig` (default 30).
2. Wire it through `app.New` into `brew.NewClient`.
3. Validate: minimum 0 (no cache), maximum reasonable cap (e.g., 3600) to avoid stale data surprises.
4. Document in the config file comment that this value will be used as the volatile-data TTL once M27 lands.

**Files:**

| File | Action |
|---|---|
| `internal/config/config.go` | Add `cache_ttl_seconds` |
| `internal/brew/client.go` | Accept TTL parameter |
| `internal/app/app.go` | Pass configured TTL |

**Acceptance criteria:**
- [ ] Config file `brew.cache_ttl_seconds` controls cache expiry.
- [ ] Default behavior remains 30 s when unset.

**Tests (same change set):**
- [ ] `TestCacheTTLFromConfig` — cache respects configured TTL.
- [ ] `TestDefaultCacheTTL` — default is 30 s.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Long TTL makes mutations appear not to update the panel. | Invalidate groups already handle this; document behavior. |

**Rollback:** Hardcode 30 s again.

---

### 24.5 — Regression Tests & Smoke

**Size:** S  
**Phase:** C  
**Depends on:** 24.3, 24.4  
**Blocks:** —

**Context:** Performance changes must not break existing Outdated interactions: batch select, upgrade, pin/unpin cache invalidation, and error surfacing.

**Implementation checklist:**
1. Run `make test`, `go vet ./...`, `make lint`.
2. Verify `TestBatchUpgradeCallsBrewWithSelectedNames` still passes.
3. Verify `TestOutdatedFetchSurfacesError` still passes.
4. Add/update teatest if needed for lazy-load behavior.

**Files:**

| File | Action |
|---|---|
| `internal/gui/flows/*` | Add/adjust E2E tests |

**Acceptance criteria:**
- [ ] All existing Outdated tests pass.
- [ ] No `program.Send` introduced.
- [ ] Race detector clean.

**Tests (same change set):**
- [ ] Existing Outdated test suite.
- [ ] `TestOutdatedLazyLoadedOnFocus`.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Lazy load breaks smoke checklist expectations. | Update `smoke-checklist.md` Outdated section to mention lazy fetch. |

**Rollback:** Revert all M24 commits.

---

## Test Plan (milestone-level)

Consolidated view — must match tests listed in steps.

| Test | Tier | Step | Proves |
|---|---|---|---|
| `TestCombinedOutdatedReturnsBothTypes` | unit | 24.1 | Single JSON parse yields formulae + casks |
| `TestOutdatedPanelFetchUsesSingleBrewCall` | unit | 24.1 | Only one `brew outdated` subprocess |
| `TestCacheCoalescesConcurrentOutdatedCalls` | unit / race | 24.2 | Concurrent callers share one brew call |
| `TestOutdatedLazyLoadedOnFocus` | e2e | 24.3 | Outdated not fetched at startup |
| `TestStatusUpdatesAfterOutdatedLoads` | unit | 24.3 | Status count reacts to shared data |
| `TestCacheTTLFromConfig` | unit | 24.4 | Configurable TTL |
| `TestDefaultCacheTTL` | unit | 24.4 | Default unchanged |
| Existing Outdated tests | unit/e2e | 24.5 | No regression |

**Verification commands:**

```bash
make test
go test -race ./...
make vet
```

---

## Definition of Done

- [ ] All steps 24.1–24.5 complete; acceptance criteria checked
- [ ] Every Test Plan row has an existing passing test (or documented manual check)
- [ ] Verification commands pass (including race detector)
- [ ] `AGENTS.md` / `DESIGN.md` updated if ADRs are stable
- [ ] [status.md](../status.md) updated; **this file header Status matches**
- [ ] No open **critical/high** findings in this milestone's scope
- [ ] **Remaining** section empty or removed (if marking ✅ Complete)

---

## Post-Milestone Gate

Before starting **Enables** milestones, confirm:

- [ ] Header **Gate criteria** satisfied
- [ ] [review-template.md](../review-template.md) Dimension 8 (plan hygiene) for this milestone
- [ ] `smoke-checklist.md` Outdated section updated for lazy load

---

## Rollback Plan

If integration fails mid-milestone:

1. Steps safe to keep independently: 24.4 (config TTL) is orthogonal.
2. Revert order: 24.3 → 24.2 → 24.1.
3. Minimum hotfix acceptable for ship: revert to eager two-call fetch if combined call has compatibility issues.

---

## Version History

| Date | Change |
|---|---|
| 2026-06-15 | Created from [templates/milestone.md](../templates/milestone.md) |
