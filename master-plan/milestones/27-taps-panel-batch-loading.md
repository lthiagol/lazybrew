# M27 — Taps Panel Batch Loading

> **Status:** 🔜 Planned  
> **Size estimate:** S (<1 day)  
> **Depends on:** M22 ✅ (release tag)  
> **Enables:** M28 (lazy panel loading, longer tap TTL)  
> **Parallel track:** —  
> **Gate criteria:** Taps panel loads in one `tap-info --json` call instead of N sequential per-tap calls, with FormulaNames/CaskNames fetched only when the trust menu needs them.

> **Note:** M25.2 request coalescing is *recommended* but not blocking. If M27 lands before M25, add a follow-up task to adopt coalescing for `KeyTapsList`.

<!-- See [archive/architecture-review-2026-06-13.md](../archive/architecture-review-2026-06-13.md) startup findings -->

---

## Goal

The Taps panel is one of the slowest parts of startup: `internal/brew/taps.go` calls `tap-info --json <tap>` sequentially for every tap. With ~10 taps this is ~10 s before the UI is usable. This milestone batches that work into a single `tap-info --json` call and defers the expensive FormulaNames/CaskNames detail until the user actually opens a trust menu.

---

## Readiness to Start

Before executing this milestone, confirm:

- [ ] M22 release tag is done.
- [ ] `brew tap-info --json` output shape has been verified to contain the fields in 27.1 (or the lazy-detail approach in 27.2 is adjusted).
- [ ] Trust menu flows are covered by existing tests or E2E flows.

---

## Why Now

- Debug log shows 8–10 sequential `tap-info --json <tap>` calls every refresh cycle, each ~1 s.
- This is the single biggest subprocess cost after `brew doctor` (which is addressed separately in **M26**).
- M25 introduces request coalescing, which makes this refactor safer and easier to test.

---

## Challenged Assumptions

| Assumption | Challenge | Decision |
|---|---|---|
| Each tap needs its own `tap-info --json` call to get full metadata. | `tap-info --json` (no args) returns all installed taps in one JSON array. | Use one batch call for the list; lazy-load per-tap detail only for trust menus. |
| `FormulaNames`/`CaskNames` must be in the Tap list model. | They are only used by the trust-specific menu code. | Store them optionally; fetch on demand when the menu opens. |

---

## Out of Scope

- **General controller refactor** — deferred to [backlog.md](../backlog.md) B-01.
- **Refresh policy / TTL** — covered by **M28** tiered refresh strategy.
- **Other panels** — Outdated in **M25**, formulae/casks dedupe in **M28**.

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|---|---|---|---|
| D27-1 | `Taps.List()` calls `tap-info --json` once for all taps. | Sequential per-tap calls; parallel per-tap goroutines. | One subprocess is simplest and fastest; Homebrew already supports it. |
| D27-2 | `FormulaNames`/`CaskNames` are lazy-loaded via `Taps.Get(name)` only for trust menus. | Always include them in list; remove the feature. | Keeps trust menu functional without paying the cost on every startup/refresh. |

Copy stable ADRs to `DESIGN.md` decision log when merged.

---

## Phases

| Phase | Steps | Theme | Phase gate |
|---|---|---|---|
| **A — Batch list** | 27.1 | One `tap-info --json` for all taps | Unit test shows one subprocess |
| **B — Lazy detail** | 27.2 | Detail only for trust menu | teatest confirms no per-tap info on startup |

---

## Step Index

| Step | Title | Size | Depends | Deliverable |
|---|---|---|---|---|
| 27.1 | Batch `tap-info` in `Taps.List()` | S | — | Single subprocess tap list |
| 27.2 | Lazy FormulaNames/CaskNames for trust menu | S | 27.1 | Trust menu still works, no detail fetch at list time |
| 27.3 | Regression tests | S | 27.2 | Tests pass |

---

## Steps

### 27.1 — Batch `tap-info` in `Taps.List()`

**Size:** S  
**Phase:** A  
**Depends on:** —  
**Blocks:** 27.2

**Context:** `internal/brew/taps.go:55-96` fetches tap names with `brew tap`, then loops calling `fetchTapInfo` for each. Replace the loop with one `brew tap-info --json` call.

**Pre-implementation verification:** Confirm that `brew tap-info --json` (no tap names) returns the fields we need: `name`, `remote`, `formula_count`, `cask_count`, `command_count`, `private`, `installed`, `manifest`, `api`, `auto_publish`, `trusted`. If `formula_names`/`cask_names` are also present, we can optionally keep them; if absent, 27.2 must fetch them per tap.

**Implementation checklist:**
1. Run `brew tap-info --json` locally and inspect the shape; record result in this milestone.
2. Add a new `fetchAllTapInfo(ctx)` method that calls `runner.ExecuteJSON(ctx, &data, "tap-info", "--json")`.
3. Map the returned array directly to `[]Tap`.
4. Cache under `KeyTapsList`.
5. Keep `fetchTapInfo(name)` for single-tap use in `Get()` and lazy detail loading.

**Files:**

| File | Action |
|---|---|
| `internal/brew/taps.go` | Refactor `List()` to batch call |
| `internal/brew/taps_test.go` | Update mocks/expectations |

**Acceptance criteria:**
- [ ] `Taps.List()` issues exactly one `tap-info --json` subprocess.
- [ ] All taps still appear in the panel with Remote/FormulaCount/CaskCount/etc.

**Tests (same change set):**
- [ ] `TestTapsListUsesSingleTapInfoCall` — mock records one `tap-info --json`.
- [ ] `TestTapsListMapsAllTaps` — multiple taps parsed correctly.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Bulk `tap-info --json` omits some fields present in single-tap call. | Compare output shape; fallback to per-tap if needed. |

**Rollback:** Revert to per-tap loop.

---

### 27.2 — Lazy FormulaNames/CaskNames for Trust Menu

**Size:** S  
**Phase:** B  
**Depends on:** 27.1  
**Blocks:** 27.3

**Context:** `FormulaNames` and `CaskNames` are only used in `commands.go:95-103` to populate the trust-specific formula/cask menus. They do not need to be fetched for the list view.

**Implementation checklist:**
1. Remove population of `FormulaNames`/`CaskNames` from `Taps.List()` (batch response may not include them anyway).
2. In `startTrustMenu()` / `executeTrustMenuAction()`, if `tap.FormulaNames`/`CaskNames` are empty, call `client.Taps.Get(tapName)` to fetch detail.
3. Cache single-tap info under a per-tap key or rely on `KeyTapsList` refresh when mutations happen.

**Files:**

| File | Action |
|---|---|
| `internal/brew/taps.go` | `List()` no longer fetches names; `Get()` does |
| `internal/gui/commands.go` | Trust menu triggers lazy detail fetch |

**Acceptance criteria:**
- [ ] Startup/refresh `tap-info` calls do not include FormulaNames/CaskNames fetch.
- [ ] Trust formula/cask menus still populate correctly.

**Tests (same change set):**
- [ ] `TestTrustMenuLazyLoadsTapNames` — menu opens → `Taps.Get` called.
- [ ] `TestTapsListDoesNotFetchNames` — list response has empty names.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| User sees empty trust menu briefly. | Show loading spinner or "Loading formulas..." in menu. |

**Rollback:** Populate names eagerly again.

---

### 27.3 — Regression Tests

**Size:** S  
**Phase:** B  
**Depends on:** 27.2  
**Blocks:** —

**Context:** Ensure existing tap-related flows (untap, repair, trust) still work.

**Implementation checklist:**
1. Run `make test`, `go vet ./...`, `make lint`.
2. Verify existing tap tests pass.

**Files:**

| File | Action |
|---|---|
| `internal/brew/taps_test.go` | Update/add tests |
| `internal/gui/flows/*` | Add/adjust E2E if needed |

**Acceptance criteria:**
- [ ] All tap tests pass.
- [ ] Race detector clean.

**Tests (same change set):**
- [ ] Existing tap test suite.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Menu tests rely on eager names. | Update test fixtures or mocks. |

**Rollback:** Revert M27 commits.

---

## Test Plan (milestone-level)

| Test | Tier | Step | Proves |
|---|---|---|---|
| `TestTapsListUsesSingleTapInfoCall` | unit | 27.1 | One subprocess for all taps |
| `TestTapsListMapsAllTaps` | unit | 27.1 | Correct mapping |
| `TestTrustMenuLazyLoadsTapNames` | unit/e2e | 27.2 | Trust menu triggers detail fetch |
| `TestTapsListDoesNotFetchNames` | unit | 27.2 | No eager detail fetch |
| Existing tap tests | unit/e2e | 27.3 | No regression |

**Verification commands:**

```bash
make test
go test -race ./...
```

---

## Definition of Done

- [ ] All steps 27.1–27.3 complete; acceptance criteria checked
- [ ] Every Test Plan row has an existing passing test
- [ ] Verification commands pass
- [ ] `DESIGN.md` updated if ADR is stable
- [ ] [status.md](../status.md) updated; **this file header Status matches**
- [ ] No open **critical/high** findings in this milestone's scope
- [ ] **Remaining** section empty or removed (if marking ✅ Complete)

---

## Post-Milestone Gate

Before starting **Enables** milestones, confirm:

- [ ] Header **Gate criteria** satisfied
- [ ] [review-template.md](../review-template.md) Dimension 8 (plan hygiene) for this milestone
- [ ] Debug log shows only one `tap-info --json` per refresh cycle

---

## Rollback Plan

If integration fails mid-milestone:

1. Steps safe to keep independently: none — 27.2 depends on 27.1.
2. Revert order: 27.2 → 27.1.
3. Minimum hotfix: revert to per-tap calls.

---

## Version History

| Date | Change |
|---|---|
| 2026-06-15 | Created from [templates/milestone.md](../templates/milestone.md) |
