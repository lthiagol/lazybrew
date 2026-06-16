# M27 — Tiered Refresh & Polling Strategy

> **Status:** 🔜 Planned  
> **Size estimate:** M (1–3 days)  
> **Depends on:** M24.2 ✅ (request coalescing), M24.3 ✅ (lazy Outdated load), M24.4 ✅ (configurable TTL), M25 ✅ (doctor error handling), M26 ✅ (Taps batch)  
> **Enables:** future controller split (B-01), background sync features  
> **Parallel track:** —  
> **Gate criteria:** lazybrew no longer hammers Homebrew on a fixed 60 s loop; different data classes refresh at appropriate rates and panels only fetch when viewed.

> **Note:** M24 must be functionally complete before M27 starts, because M27 generalizes the cache TTL and lazy-loading mechanisms introduced in M24.

<!-- See [backlog.md](../backlog.md) B-02 — routed to M24 -->
<!-- See [archive/architecture-review-2026-06-13.md](../archive/architecture-review-2026-06-13.md) startup / performance findings -->

---

## Goal

Replace the current "refresh everything every 60 s" model with a tiered strategy: volatile data (formulae, casks, outdated, services) refreshes reasonably often, stable data (taps, config, version) refreshes rarely, and expensive diagnostics (doctor) refresh only on demand. Panels that are not visible should not fetch data at all.

---

## Readiness to Start

Before executing this milestone, confirm:

- [ ] M24 is functionally complete (combined Outdated, coalescing, lazy Outdated load, configurable TTL).
- [ ] M25 is complete (`brew doctor` no longer logs false failures).
- [ ] M26 is complete (Taps batch loading).
- [ ] Config fallback chain (M24's `cache_ttl_seconds` / old `auto_refresh_seconds` → M27's `refresh.volatile_seconds`) is accepted.
- [ ] Status dashboard design for stale/placeholder data is agreed (e.g., "—" for counts still loading).

---

## Why Now

- The debug log shows a ~60 s firehose of 15–20 brew subprocesses, including expensive `tap-info`, `config`, and `doctor` calls.
- M24–M26 fix the per-panel call volume, but without a refresh strategy the app still refreshes everything on the same schedule.
- This milestone is the capstone of the post-v0.2.0 performance pass.

---

## Challenged Assumptions

| Assumption | Challenge | Decision |
|---|---|---|
| All panels should refresh on the same `AutoRefreshSeconds` tick. | Taps, config, and doctor change far less often than packages/services. | Assign data classes with independent TTLs/refresh triggers. |
| `Init()` must load every panel so the UI is immediately complete. | Users usually look at one panel at a time; eager loading wastes work. | Lazy-load each panel on first focus; only the active panel + Status dashboard fetch at startup. |
| Auto-refresh should run even while a modal or long task is active. | Background refresh competes with the user's foreground action. | Pause auto-refresh while any task/modal is active. |
| Cache TTL is a single global value. | Different data has different freshness requirements. | Support per-key TTL presets via config. |

---

## Out of Scope

- **Controller refactor** — deferred to [backlog.md](../backlog.md) B-01.
- **Background prefetch / predictive refresh** — future enhancement; this milestone focuses on visible-panel-only + TTL.
- **Network/offline behavior** — not a Homebrew concern.

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|---|---|---|---|
| D27-1 | Data classes: volatile, stable, manual. | Per-panel TTL; single TTL. | Three classes are enough to capture real behavior and keep config simple. |
| D27-2 | Lazy panel loading for all non-Status panels. | Keep eager Init fetch. | Avoids work the user never sees. |
| D27-3 | Pause auto-refresh while tasks/modals are active. | Always refresh; cancel refresh on interaction. | Respects foreground work and avoids competing brew locks. |
| D27-4 | Config exposes TTL presets, not per-key knobs. | Expose every cache key separately. | Usable defaults with optional override. |

Copy stable ADRs to `DESIGN.md` decision log when merged.

---

## Phases

| Phase | Steps | Theme | Phase gate |
|---|---|---|---|
| **A — Data classes & TTL** | 27.1, 27.2 | Assign TTLs per class, wire config | Unit tests prove TTLs applied |
| **B — Lazy loading** | 27.3 | Fetch only visible panels | E2E confirms no brew calls for hidden panels |
| **C — Refresh behavior** | 27.4, 27.5 | Pause on interaction, manual refresh | Manual smoke checklist passes |

---

## Step Index

| Step | Title | Size | Depends | Deliverable |
|---|---|---|---|---|
| 27.1 | Define data classes and per-class TTLs | S | M24.4 | Cache layer supports per-key TTL presets |
| 27.2 | Wire refresh policy config | S | 27.1 | Config file controls auto-refresh class by class |
| 27.3 | Lazy-load all panels on first focus | S | M24.3 | `Init()` only fetches Status + active panel |
| 27.4 | Pause auto-refresh during tasks/modals | S | 27.3 | No background brew during foreground mutation |
| 27.5 | Manual refresh keys (`r`/`R`) bypass TTL | S | 27.2, 27.3 | Immediate refresh on user request |
| 27.6 | Regression tests & smoke | S | 27.4, 27.5 | All tests pass |

---

## Steps

### 27.1 — Define Data Classes and Per-Class TTLs

**Size:** S  
**Phase:** A  
**Depends on:** M24.4  
**Blocks:** 27.2

**Context:** `internal/brew/cache.go` uses a single `ttl`. We need different lifetimes for different cache keys. Example defaults:
- **Volatile** (30 s): formulae list, casks list, outdated, services.
- **Stable** (5 min): taps list, config, version.
- **Manual** (∞ until explicit refresh or mutation): doctor, leaves, missing, vulns.

**Implementation checklist:**
1. Define a `DataClass` enum: `DataClassVolatile`, `DataClassStable`, `DataClassManual`.
2. Add a `CacheKey → DataClass` map in `internal/brew/cache.go` (or adjacent file).
3. Change `Cache` constructor to accept a `map[DataClass]time.Duration` (or `func(CacheKey) time.Duration`) instead of a single TTL.
4. Provide a default mapping:
   - Volatile: 60 s
   - Stable: 5 min
   - Manual: 0 (no expiry; rely on invalidation)
5. Keep `NewCache(defaultTTL)` as a convenience wrapper that maps all keys to `Volatile` with the given TTL, preserving backward compatibility for tests and callers.

**Files:**

| File | Action |
|---|---|
| `internal/brew/cache.go` | Add data-class TTL mapping |
| `internal/brew/client.go` | Wire class TTLs; pass TTL map from config/app |

**Acceptance criteria:**
- [ ] Volatile keys expire quickly; stable keys expire slowly; manual keys never expire until invalidated.
- [ ] Default TTL still works for unclassified keys.

**Tests (same change set):**
- [ ] `TestCacheTTLByDataClass` — different classes expire at different rates.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Manual class makes UI look stale. | Provide clear "last updated" indicator and manual refresh. |

**Rollback:** Revert to single global TTL.

---

### 27.2 — Wire Refresh Policy Config

**Size:** S  
**Phase:** A  
**Depends on:** 27.1  
**Blocks:** 27.5

**Context:** Users should be able to tune or disable auto-refresh. Replace/extend `AutoRefreshSeconds` and `brew.cache_ttl_seconds` (from M24.4) with class-level controls.

**Proposed config shape:**

```yaml
# New section (M27)
refresh:
  enabled: true
  volatile_seconds: 60      # formulae, casks, outdated, services
  stable_seconds: 300       # taps, config, version

# Existing keys remain for backward compatibility:
gui:
  auto_refresh_seconds: 60  # deprecated; maps to refresh.volatile_seconds
brew:
  cache_ttl_seconds: 60     # from M24; maps to refresh.volatile_seconds default
```

**Resolution order when loading config:**
1. `refresh.volatile_seconds` if present.
2. `gui.auto_refresh_seconds` if present (deprecated fallback).
3. `brew.cache_ttl_seconds` if present (M24 fallback).
4. Hard-coded default (60 s).

**Implementation checklist:**
1. Add `RefreshConfig` struct with `Enabled`, `VolatileSeconds`, `StableSeconds`.
2. Implement the fallback chain above in `config.Load` or a helper.
3. Pass TTLs to `brew.NewClient` as a map: `{Volatile: volatile, Stable: stable, Manual: 0}`.
4. Use `refresh.enabled` to decide whether `autoRefreshCmd()` returns a tick.
5. Log a deprecation warning if `auto_refresh_seconds` is used.

**Files:**

| File | Action |
|---|---|
| `internal/config/config.go` | Add `RefreshConfig`; implement fallback chain |
| `internal/brew/client.go` | Accept TTL map |
| `internal/app/app.go` | Pass config TTLs to client |
| `internal/gui/gui.go` | Use `refresh.enabled` for tick scheduling |

**Acceptance criteria:**
- [ ] Config values control refresh cadence.
- [ ] `refresh.enabled: false` stops auto-refresh entirely.
- [ ] Existing configs with `auto_refresh_seconds` or `cache_ttl_seconds` still work.

**Tests (same change set):**
- [ ] `TestRefreshConfigDisablesAutoRefresh` — no tick command when disabled.
- [ ] `TestRefreshConfigUsesClassIntervals` — cache TTLs match config.
- [ ] `TestRefreshConfigFallbackChain` — old keys map correctly to new keys.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Config migration confusion. | Document fallback chain in config comments, AGENTS.md, and milestone ADR. |

**Rollback:** Restore `AutoRefreshSeconds` as sole knob; ignore new `refresh` section.

---

### 27.3 — Lazy-Load All Panels on First Focus

**Size:** S  
**Phase:** B  
**Depends on:** M24.3  
**Blocks:** 27.4

**Context:** `Init()` and `RefreshMsg` currently fetch every panel. Only the active panel (and Status dashboard) should fetch eagerly; others load when focused.

**Implementation checklist:**
1. Remove `PanelFormulae`, `PanelCasks`, `PanelOutdated`, `PanelTaps`, `PanelServices` from `Init()` and `RefreshMsg` batch.
2. Keep `fetchStatusData` (Status is always the default panel).
3. On panel switch, if the target panel has no items and is not loading, enqueue `fetchPanelData(target)`.
4. Show loading state for panels that are fetching.

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Update `Init()`, `RefreshMsg`, panel switch logic |
| `internal/gui/commands.go` | Keep `fetchPanelData` usable |

**Acceptance criteria:**
- [ ] Startup brew calls are limited to Status + active panel data.
- [ ] Switching to a panel triggers its fetch.
- [ ] Refresh (`r`/`R`) still refreshes the active panel/Status.

**Tests (same change set):**
- [ ] `TestLazyLoadAllPanels` — e2e: switching panels triggers fetch.
- [ ] `TestStartupOnlyFetchesStatus` — mock records no hidden-panel fetches.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| User sees empty panel on first switch. | Show spinner and panel-specific empty/loading message. |

**Rollback:** Restore eager `Init()` fetch.

---

### 27.4 — Pause Auto-Refresh During Tasks/Modals

**Size:** S  
**Phase:** C  
**Depends on:** 27.3  
**Blocks:** 27.6

**Context:** A background refresh firing during `brew upgrade` creates confusion and potential races. Auto-refresh should pause while the TaskManager is busy or a modal is open.

**Implementation checklist:**
1. In the `RefreshMsg` handler, return early if `m.tasks.IsBusy()` or `m.activeModal != nil`.
2. Still schedule the next tick so refresh resumes when idle.
3. Add `tasks.IsBusy()` if it does not exist.

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Guard `RefreshMsg` |
| `internal/gui/task/manager.go` | Add `IsBusy()` if missing |

**Acceptance criteria:**
- [ ] Auto-refresh does not fetch while a task is running.
- [ ] Auto-refresh does not fetch while a modal is open.

**Tests (same change set):**
- [ ] `TestRefreshPausedDuringTask` — task running → RefreshMsg is no-op.
- [ ] `TestRefreshPausedDuringModal` — modal open → RefreshMsg is no-op.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Long-running task stalls refresh indefinitely. | Resume on task completion; manual refresh always works. |

**Rollback:** Remove guards.

---

### 27.5 — Manual Refresh Keys Bypass TTL

**Size:** S  
**Phase:** C  
**Depends on:** 27.2, 27.3  
**Blocks:** 27.6

**Context:** Users need a way to force refresh regardless of TTL. `r`/`R` already trigger `RefreshMsg`; ensure it invalidates the active panel's cache.

**Implementation checklist:**
1. On `RefreshMsg`, invalidate cache keys for the active panel + Status before fetching.
2. Distinguish user-triggered refresh from auto-tick refresh (optional: add `Force bool` to `RefreshMsg`).
3. Keep auto-refresh from invalidating cache if data is still fresh (cache TTL handles this).

**Files:**

| File | Action |
|---|---|
| `internal/gui/gui.go` | Update `RefreshMsg` handling |
| `internal/brew/cache.go` | Ensure `Invalidate` works per key |

**Acceptance criteria:**
- [ ] Pressing `r` fetches fresh data for the active panel even if cache TTL has not expired.
- [ ] `R` runs `brew update` when `UpdateOnStart` is enabled (existing behavior preserved).

**Tests (same change set):**
- [ ] `TestManualRefreshBypassesCache` — cached data ignored on `RefreshMsg`.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Accidental flood of manual refreshes. | Rate-limit not required; user-initiated. |

**Rollback:** Restore current refresh behavior.

---

### 27.6 — Regression Tests & Smoke

**Size:** S  
**Phase:** C  
**Depends on:** 27.4, 27.5  
**Blocks:** —

**Context:** Validate that the new strategy does not break existing flows.

**Implementation checklist:**
1. Run full test suite: `make test`, `go vet ./...`, `make lint`.
2. Verify E2E flows still pass.
3. Update `smoke-checklist.md` refresh behavior section.

**Files:**

| File | Action |
|---|---|
| `internal/gui/flows/*` | Add/adjust tests |
| `master-plan/smoke-checklist.md` | Update refresh expectations |

**Acceptance criteria:**
- [ ] All tests pass.
- [ ] Debug log shows far fewer brew calls per minute.
- [ ] No `program.Send` introduced.

**Tests (same change set):**
- [ ] Existing E2E flows.
- [ ] New tests from 27.1–27.5.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Lazy loading breaks snapshot tests. | Update snapshots; assert on loading state. |

**Rollback:** Revert M27 commits.

---

## Test Plan (milestone-level)

| Test | Tier | Step | Proves |
|---|---|---|---|
| `TestCacheTTLByDataClass` | unit | 27.1 | Per-class expiry |
| `TestRefreshConfigDisablesAutoRefresh` | unit | 27.2 | Config control |
| `TestRefreshConfigUsesClassIntervals` | unit | 27.2 | TTLs match config |
| `TestLazyLoadAllPanels` | e2e | 27.3 | Hidden panels don't fetch |
| `TestStartupOnlyFetchesStatus` | unit | 27.3 | Minimal startup |
| `TestRefreshPausedDuringTask` | unit | 27.4 | No background fetch during mutation |
| `TestRefreshPausedDuringModal` | unit | 27.4 | No background fetch during modal |
| `TestManualRefreshBypassesCache` | unit | 27.5 | User can force refresh |
| Existing flows | e2e | 27.6 | No regression |

**Verification commands:**

```bash
make test
go test -race ./...
make vet
```

---

## Definition of Done

- [ ] All steps 27.1–27.6 complete; acceptance criteria checked
- [ ] Every Test Plan row has an existing passing test
- [ ] Verification commands pass
- [ ] `DESIGN.md` / `AGENTS.md` updated if ADRs are stable
- [ ] [status.md](../status.md) updated; **this file header Status matches**
- [ ] No open **critical/high** findings in this milestone's scope
- [ ] **Remaining** section empty or removed (if marking ✅ Complete)

---

## Post-Milestone Gate

Before starting **Enables** milestones, confirm:

- [ ] Header **Gate criteria** satisfied
- [ ] [review-template.md](../review-template.md) Dimension 8 (plan hygiene) for this milestone
- [ ] `smoke-checklist.md` updated for lazy load + refresh pause

---

## Rollback Plan

If integration fails mid-milestone:

1. Steps safe to keep independently: 27.1 (TTL infra) is useful on its own.
2. Revert order: 27.5 → 27.4 → 27.3 → 27.2 → 27.1.
3. Minimum hotfix: disable auto-refresh by default.

---

## Version History

| Date | Change |
|---|---|
| 2026-06-15 | Created from [templates/milestone.md](../templates/milestone.md) |
