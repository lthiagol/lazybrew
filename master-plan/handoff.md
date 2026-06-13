# Lazybrew — Session Handoff

> **Created:** 2026-06-13  
> **Branch:** `first-version` (24 ahead of origin)  
> **State:** All milestones M1–M22 + M17 adopted. Backlog remains.

---

## What Was Built

### Milestone Summary

| # | Milestone | Status | Key Deliverables |
|---|---|---|---|
| M18 | Documentation | ✅ | README, LICENSE, DESIGN.md, AGENTS.md, status.md |
| M19 | TaskManager Concurrency | ✅ | Zero `program.Send`, zero `isBusy`, sequential write queue, streaming cmd pattern |
| M20 | Functional UX | ✅ | Tab cache with itemName, Info tab formatters, empty states, viewport scrolling, batch upgrade, pin toggle fix, config wiring, auto-refresh tick, small terminal warning |
| M21 | Test Strategy v2 | ✅ | teatest E2E flows (6), integration test suite (5 tests), regression tests (5), coverage gates in Makefile, concurrent cache expiry test |
| M22 | CI & Release | ✅ | GitHub Actions CI, goreleaser config, release checklist, dependabot |
| M17 | Lazygit TUI | ✅ | Auto-update via TaskManager, bottom bar update ticker, per-panel sidebar boxes with accordion heights, breadcrumb, search info preview |

### Test Counts
- `go test -race ./...` — all pass
- `go vet ./...` — clean
- `go build ./...` — clean
- `make cover-check` — floors met: brew 65%, gui/presentation 90%, gui 30%, modal 40%

---

## What's Left — Backlog (6 items)

| ID | Item | Priority | Effort | Why now |
|---|---|---|---|---|
| **B-01** | Split `internal/gui/controllers/` per panel | Medium | L | Largest remaining architecture debt |
| **B-07** | Runner SIGKILL after 5s on cancel | Medium | S | Real-world UX: stuck brew processes |
| B-02 | Lazy panel loading | Low | M | Performance: 6 parallel brew calls on every Init/Refresh |
| B-09 | Homebrew formula for lazybrew | Low | S | User install |
| B-11 | TypedCache serialization for config hot-reload | Low | S | Future-proofing |
| B-12 | Config migration path | Low | S | If ShowIcons semantics change |

---

## Recommended Adoption Sequence

```
B-07 → B-01 → B-02 → B-09/B-11/B-12 (any order)
```

### 1. B-07 — Runner SIGKILL (S, Medium)
Smallest effort, highest user impact. When a user cancels a running brew operation (e.g., Esc in progress modal), the subprocess may linger. After a 5s grace period, send SIGKILL to the process group.

**Files:** `internal/brew/runner.go` — `DefaultRunner.ExecuteStream`
**Pattern:** Launch with process group; on context cancel, wait 5s then `syscall.Kill(-pgid, syscall.SIGKILL)`

**Why first:** Self-contained, exercises the cancel path we built in M19, immediate UX improvement.

### 2. B-01 — Split controllers (L, Medium)
Largest remaining debt. Currently `gui.go` + `commands.go` ~1700 lines. Split into per-panel controller files.

**Suggested split:**
```
internal/gui/controllers/
├── status.go      (runDoctor, runVulns, runMissing, brewCleanup, etc.)
├── formulae.go    (doMutation for Formulae, togglePin, toggleLeaves)
├── casks.go       (doMutation for Casks, togglePin)
├── outdated.go    (batchUpgrade, select-all)
├── taps.go        (startTapAdd, startTrustMenu, executeTrustAction, etc.)
├── services.go    (serviceAction, serviceCleanup)
└── search.go      (executeSearch, fetchSelectedSearchInfo)
```

**Pattern:** Each file receives `*Model` or `Model` methods (same package, just split files). No interface extraction needed yet.

**Why second:** Improves maintainability before any larger feature work.

### 3. B-02 — Lazy Panel Loading (M, Low)
Currently `Init()` fires 6 parallel brew subprocesses. Change to fetch only the active panel first, then lazy-load remaining panels on first activation.

**Files:** `gui.go` — `Init()`, `Update()` on panel switch, `commands.go` — `fetchPanelData`

**Pattern:** Add `panelData.needsFetch bool`. `Init()` fetches only PanelStatus. On `switchPanel`, if target panel hasn't been fetched, fetch it. Cache results as before.

**Why third:** Performance improvement that builds on existing infrastructure.

### 4. B-09 / B-11 / B-12 (any order)
- **B-09:** Create Homebrew tap formula pointing to GitHub releases. Depends on M22 goreleaser release pipeline.
- **B-11:** Add JSON marshal/unmarshal to TypedCache for config hot-reload. Small, isolated.
- **B-12:** If `ShowIcons` config field semantics change, add migration in `config.Load()`.

---

## Session Start Command

To resume adoption, run:
```
cd /home/thiago/code/lazybrew
cat master-plan/handoff.md  # read this file
```

Then proceed with the recommended sequence above, or ask for guidance on a specific item.
