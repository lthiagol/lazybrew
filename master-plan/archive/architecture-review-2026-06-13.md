# Lazybrew — Architecture & Code Review

> **Reviewer:** Code audit (planning session)  
> **Date:** 2026-06-13  
> **Branch reviewed:** `first-version`  
> **Scope:** Full codebase + `master-plan/` (no code changes applied)

---

## Executive Summary

Lazybrew has a **solid foundation** in the `internal/brew/` layer: typed domain models, a clean `Runner` abstraction, read/write service splits, cache invalidation groups, and good unit test coverage for parsers and services (~65%). The Bubble Tea shell works, most P0 Homebrew commands are wired, and the milestone planning is unusually thorough.

The main risks are **not** in brew parsing — they are in the **GUI layer's concurrency model**, **plan-vs-code drift**, and **test strategy gaps**. Several milestones are marked complete in their own files while `status.md` and the code tell a different story. The README still claims the TUI is not built.

**Recommended sequencing:** fix correctness and concurrency (M19–M20) → establish real test infrastructure (M21) → CI/release (M22) → then pursue M17 visual polish on a stable base.

---

## What Is Done Well

### Brew data layer (`internal/brew/`)

| Strength | Detail |
|---|---|
| Runner abstraction | `Execute`, `ExecuteJSON`, `ExecuteStream` with `HOMEBREW_NO_ASK` and `HOMEBREW_NO_AUTO_UPDATE` |
| Service boundaries | Formulae/Casks/Taps/Services/Trust/Diagnostics split into Reader/Writer |
| Typed errors | `BrewNotFoundError`, `BrewExitError`, `JSONParseError` |
| Cache design | TTL, typed wrappers, invalidation groups per operation |
| Test fixtures | Synthetic JSON in `testdata/` — avoids brittle real-brew snapshots |
| Fuzz tests | JSON parsing edge cases |

### GUI foundations

| Strength | Detail |
|---|---|
| Elm-style messages | `DataLoadedMsg`, `TabContentMsg`, `ProgressCompleteMsg`, etc. |
| Modal subsystem | Reusable confirm/input/menu/progress/toast in `internal/gui/modal/` |
| Presentation layer | Formatters separated from rendering logic |
| Config + theming | YAML config, dark/light themes, help overlay, CLI flags |
| Reader/Writer split reflected in Client | `brew.NewClient(runner)` is the right injection point |

### Planning

The milestone documents capture real decisions (non-interactive brew, trust granularity, accordion layout for M17). The coverage audit and decision log are valuable. The problem is **tracking accuracy**, not planning quality.

---

## Architecture Assessment

### Current layering

```
cmd/lazybrew/main.go
    └── internal/app/app.go          (bootstrap: config, theme, runner, gui.New)
            ├── internal/config/
            ├── internal/brew/       (domain + subprocess)
            └── internal/gui/        (Bubble Tea model — monolithic)
                    ├── modal/
                    ├── presentation/
                    └── style/
```

This is a reasonable 3-layer split for a TUI app. The brew layer is testable in isolation. The GUI layer has grown into a **god module** (`gui.go` + `commands.go` ≈ 1,500 lines) without the planned `controllers/` decomposition from M4/M10.

### Intended vs actual GUI structure

| Planned (M1, M4, M10) | Actual |
|---|---|
| `internal/gui/controllers/*_panel.go` | Does not exist — all logic in `gui.go`, `commands.go`, `render.go`, `panel.go` |
| `internal/gui/presentation/formulae.go` | Only `formatters.go` (list line formatting, not Info tab content) |
| `internal/gui/task/manager.go` | `task.go` contains only `batchState` — **TaskManager never built** |
| Per-panel controllers | Single `Model` struct owns everything |

### Concurrency model — critical design issue

Bubble Tea requires all model mutations through `Update()`. The codebase uses **three inconsistent patterns**:

1. **Correct:** `tea.Cmd` returning messages (`fetchPanelData`, `serviceAction` for short ops)
2. **Problematic:** `tea.Cmd` that calls `m.program.Send()` from inside the command (`runDoctor`, `runMissing`, `runVulns`, `brewCleanup` preview)
3. **Dangerous:** Raw `go func()` in `doMutation` that streams via `program.Send()` while mutating shared state (`isBusy`, `activeModal`)

**Why this matters:**

- `program.Send()` bypasses the normal `Update()` cmd return path — harder to reason about ordering
- `doMutation` sets `isBusy = true` but returns early without clearing it when selection is invalid
- No queue: concurrent user actions can race despite `isBusy` (trust actions don't set it)
- M6's TaskManager (single write queue, cancel, retry) was the designed fix — still missing

**Recommendation:** Introduce a proper `TaskManager` (or at minimum unify all long-running ops as `tea.Cmd` + message streaming, zero raw goroutines in handlers). See **Milestone 19**.

### Data flow on startup — performance concern

`Init()` fires **6 parallel brew subprocesses** immediately (formulae, casks, outdated, taps, services, status). Each may invoke JSON or list commands. On a typical machine this is 6–10 brew invocations before the user sees data.

| Issue | Impact |
|---|---|
| No request deduplication | Status panel re-fetches formulae+casks lists already fetched for sidebar |
| Fixed 30s cache TTL, not configurable | Stale data or excess refresh depending on usage |
| `AutoRefreshSeconds` in config | **Defined but never used** |
| `Brew.Path` in config | **Defined but never used** — runner ignores it |

**Recommendation:** Lazy-load panels (fetch active panel first), dedupe status dashboard from cached lists, wire config fields or remove them.

### Tab content lifecycle — UX correctness bug

`loadTabContent()` runs only on `[` / `]` tab switches. It does **not** run when:

- User presses `j`/`k` to change selection
- User switches panels via Tab/number keys
- Data refresh completes

`tabContent` is keyed by `panel:tab` only — **not by selected item**. Result: Deps/Used By/Files tabs can show data for the wrong formula after navigation.

**Recommendation:** Include item name in cache key; invalidate/refetch on selection change. See **Milestone 20**.

### Main panel Info tab — incomplete vs design

M4 specifies "info in the main panel" for Formulae/Casks. Current behavior:

- Formulae tab 0 (`Info`) calls `renderList()` — **same list as sidebar**, not formatted package info
- No `FormatFormulaInfo()` or equivalent exists

The presentation formatters only produce single-line sidebar entries.

---

## Findings by Severity

### Critical

| # | Category | Finding | Location |
|---|---|---|---|
| C1 | Correctness | `doMutation` sets `isBusy=true` then returns without clearing on invalid selection / empty name / nil program | `commands.go:240–267` |
| C2 | Architecture | Raw goroutine + `program.Send()` for mutations violates Bubble Tea model; races with concurrent updates | `commands.go:269–312` |
| C3 | Correctness | Tab content cache ignores selected item — stale Deps/Used By/Files | `commands.go:471–495`, `render.go:72–77` |
| C4 | Maintainability | Plan marks M6 TaskManager complete; code has no task manager | `task.go`, M6 doc |

### High

| # | Category | Finding | Location |
|---|---|---|---|
| H1 | Correctness | Batch select (`Space`) toggles selection but **no batch upgrade** implementation | `task.go`, `gui.go:322–325` |
| H2 | Correctness | Pin toggle tries `Unpin` first, then `Pin` on any error — wrong semantics | `commands.go:393–406` |
| H3 | Correctness | `fetchPanelData` for Outdated silently ignores errors (`formulae, _ := ...`) | `commands.go:786–787` |
| H4 | Documentation | README says TUI not built; app runs full TUI | `README.md:29` |
| H5 | Documentation | `status.md` references design doc; **DESIGN.md does not exist** | `master-plan/status.md:80` |
| H6 | Documentation | No AGENTS.md for agent/human contributor conventions | repo root |
| H7 | Release | `.goreleaser.yml` references `LICENSE` file that does not exist | `.goreleaser.yml:32` |
| H8 | Testing | Zero integration tests despite `make test-integration` target | Makefile, no `//go:build integration` files |
| H9 | Testing | No `teatest` E2E tests — GUI tests call `Update()` without `View()` assertions | `gui_test.go` |

### Medium

| # | Category | Finding | Location |
|---|---|---|---|
| M1 | UX | Formulae/Casks Info tab shows list duplicate, not info view | `render.go:68–71` |
| M2 | UX | Generic empty states ("No items") vs panel-specific messages planned in M4 | `panel.go:191` |
| M3 | UX | Small terminal layout not implemented (M2 remaining) | M2 milestone |
| M4 | Config | `AutoRefreshSeconds`, `ShowIcons`, `Brew.Path` unused | `config.go`, grep |
| M5 | Architecture | Root `viewport` updated but never rendered in `View()` | `gui.go:24,395` |
| M6 | Architecture | `handleModalResult(result interface{})` — still untyped at boundary | `commands.go:19` |
| M7 | Architecture | `TypedCache.Get` panics on wrong type stored (no ok check) | `cache.go:121` |
| M8 | Plan drift | Milestone headers say ✅ Complete; `status.md` index says ⚠️ Partial for same milestones | M6–M16 |
| M9 | Plan drift | M16 claims 186 tests, 80% coverage; actual: **162 tests**, gui **31.5%**, modal **41.4%** | `go test -cover` |
| M10 | Performance | 6 parallel brew calls on every Init/Refresh | `gui.go:59–67` |
| M11 | CI | No GitHub Actions workflow | no `.github/` |

### Low

| # | Category | Finding |
|---|---|---|
| L1 | Structure | Planned `internal/utils/` never created |
| L2 | Dependencies | M1 mentions `testify`; not in `go.mod` |
| L3 | Search panel | M17.11 plans search info preview; Search tab shows same list as sidebar today |
| L4 | Outdated panel | No typed formulae/casks on Outdated — limits Info/Versions tabs |
| L5 | Modal overlay | Modal appended below full view rather than centered overlay (minor UX) |

---

## Test Strategy Assessment

### What exists (162 test functions)

| Package | Coverage | Quality |
|---|---|---|
| `internal/brew/` | 65.2% | Good — parsers, cache, services with mocks |
| `internal/gui/presentation/` | 95.6% | Good — formatters + snapshots |
| `internal/gui/style/` | 100% | Good |
| `internal/app/` | 76.9% | Adequate — bootstrap paths |
| `internal/config/` | 66.7% | Missing validation tests (invalid theme clamping planned in M16 but not implemented) |
| `internal/gui/` | 31.5% | Shallow — state flag checks, no View output |
| `internal/gui/modal/` | 41.4% | Partial — some View tests |

### What's missing (planned since M2, still absent)

| Type | Planned in | Status |
|---|---|---|
| `teatest` E2E flows | M2, M4–M7, M12 | **0 tests** |
| Integration (`//go:build integration`) | M1, M3, M12 | **0 files** |
| Concurrent cache expiry test | M13 | Not verified in `cache_test.go` |
| Race-sensitive GUI tests | M15 | Only logger/theme tested |

### Tests that matter vs tests that don't

**High value additions:**

- Tab content refetch when selection changes (would catch C3)
- `doMutation` early-return does not leave `isBusy` stuck (would catch C1)
- Batch upgrade with mocked runner verifying correct brew args
- Integration: `brew --version`, `brew list --json`, `brew search --json` against real brew
- teatest: install flow, modal capture, refresh key → mock data update in View

**Lower value (avoid for now):**

- Testing every keybinding in isolation without View assertions
- Chasing 80% line coverage in `gui.go` without E2E infrastructure

---

## Performance & Reliability

| Area | Assessment |
|---|---|
| Brew subprocess | Appropriate for CLI wrapper; streaming for long ops is correct |
| Cache | Good invalidation mapping; TTL upgrade-to-Lock on expiry fixed (M13) |
| Startup | Heavy — consider lazy panel loading |
| Memory | `tabContent` map grows unbounded — add LRU or clear on refresh |
| Cancel | Progress modals have cancel func; `doMutation` goroutine respects stream cancel via context |
| Error propagation | Mixed — some paths use toasts, some silent `_` ignores |

---

## Master Plan Health

### Strengths

- Clear milestone numbering and dependencies
- Decision log is excellent
- Coverage audit maps brew commands to features
- M13–M16 correctly identified real audit findings

### Problems

1. **Dual truth:** Milestone file headers (`✅ Complete`) vs `status.md` index (`⚠️ Partial`) vs code reality
2. **Missing canonical design doc:** Architecture reference points to non-existent DESIGN.md
3. **M16 marked complete** but coverage targets unmet and E2E/integration absent
4. **M17 (visual polish)** planned before functional/test debt is resolved
5. **No AGENTS.md** despite user expectation and status.md implying documented agent workflow

### Recommended milestone reorder

```
Done (with caveats): M1–M5, M9–M10, M13
Partial (keep open):  M6–M8, M11–M12, M14–M16
New (this review):    M18–M22
Deferred polish:      M17 (after M19–M22)
```

---

## Proposed New Milestones

See individual files:

| # | Milestone | Purpose |
|---|---|---|
| 18 | [Documentation & Project Hygiene](milestones/18-documentation-and-project-hygiene.md) | README, DESIGN.md, AGENTS.md, LICENSE, status sync |
| 19 | [Bubble Tea Concurrency & Task Manager](milestones/19-bubble-tea-concurrency-and-task-manager.md) | Fix C1/C2, implement M6 TaskManager properly |
| 20 | [Functional Completeness & UX Correctness](milestones/20-functional-completeness-and-ux.md) | Info tabs, tab cache, batch upgrade, pin, empty states |
| 21 | [Test Strategy v2](milestones/21-test-strategy-v2.md) | teatest, integration, meaningful coverage gates |
| 22 | [CI & Release Hardening](milestones/22-ci-and-release-hardening.md) | GitHub Actions, goreleaser dry-run, release checklist |
| 17 | Lazygit TUI & Auto-Update *(existing)* | **Move after M19–M22** |

---

## Definition of "Production Ready"

Before calling lazybrew daily-driver reliable:

- [ ] Single write-operation queue with cancel (TaskManager)
- [ ] No raw goroutines mutating model state outside Update
- [ ] Tab content correct after j/k navigation
- [ ] Info tabs show package details, not sidebar duplicate
- [ ] Integration test suite runnable via `make test-integration`
- [ ] At least 5 teatest flows covering install, uninstall, search, refresh
- [ ] CI on push (test + vet + build)
- [ ] README accurate; DESIGN.md and AGENTS.md exist
- [ ] LICENSE file present for goreleaser

---

## Appendix: Metrics (2026-06-13)

| Metric | Value |
|---|---|
| Go source lines | ~7,821 |
| Test functions | 162 |
| Packages with tests | 7/8 (cmd untested — acceptable) |
| Overall coverage (weighted est.) | ~55% |
| Integration test files | 0 |
| teatest E2E tests | 0 |
| CI workflows | 0 |
