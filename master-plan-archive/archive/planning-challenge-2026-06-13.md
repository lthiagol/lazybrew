# Planning Challenge — M18–M22 Decisions

> **Date:** 2026-06-13 (second planning pass)  
> **Purpose:** Stress-test the first review's milestone split, sequencing, and scope before execution.  
> **Rule:** If a challenge changes a decision, the milestone files and `status.md` are the source of truth.

---

## Executive Summary

The first review correctly identified **concurrency**, **tab cache bugs**, and **test/CI gaps** as top risks. After challenging those decisions:

| Original decision | Verdict | Change |
|---|---|---|
| M18 before M19 | **Partially wrong** | M19 can start after M18.5 (DESIGN skeleton only); full docs not blocking |
| Single TaskManager in M19 | **Correct** | Keep, but split into 10 sub-steps with ADR first |
| M20 as one milestone | **Too large** | Split into execution phases A→F with strict order inside |
| M21 after all of M20 | **Too late for infra** | Integration tests + teatest helper can start during M19 |
| M22 after M21 | **Too late for feedback** | Minimal CI (22.1) can land after M19.5; full release gate stays after M21 |
| Defer M17 entirely | **Correct** | M17 still last; M20.7 only adds minimum-size warning, not accordion |
| TypedCache fix in M21 | **Wrong placement** | Move to M19.0 — defensive fix before concurrency work |
| Config fields "document in M18" | **Incomplete** | M18 locks decisions; M20 implements; no ambiguity |

---

## Challenge 1 — Must documentation block engineering?

**Original:** M18 → M19 (docs before TaskManager)

**Challenge:** TaskManager is the highest technical risk. Waiting for full DESIGN.md + AGENTS.md + milestone reconciliation delays the fix by ~1–2 days.

**Resolution:**

| Track | Milestone steps | Can start when |
|---|---|---|
| **Docs track** | M18.1–M18.4, M18.8–M18.10 | Immediately |
| **Design minimum** | M18.5 (DESIGN skeleton + concurrency ADR) | Day 1 morning |
| **Engineering track** | M19.1+ | After M18.5 only |
| **Full docs** | M18.6–M18.7, M18.9 | Parallel with M19 |

**Gate:** M19 must not merge without AGENTS.md concurrency rules (M18.8) — can be a short section added before M19.6 merges.

---

## Challenge 2 — Is TaskManager the right abstraction?

**Original:** Full M6 TaskManager with queue

**Alternatives considered:**

| Option | Pros | Cons | Verdict |
|---|---|---|---|
| A. Fix `isBusy` only | Fast | Doesn't fix `program.Send`; no queue | ❌ Reject |
| B. Collect output in single `tea.Cmd` (no streaming) | Simple | Bad UX for long installs | ❌ Reject |
| C. TaskManager + message streaming | Matches M6; testable | More code | ✅ **Keep** |
| D. Extract `controllers/` first | Cleaner files | Large refactor; delays fix | ❌ Defer to post-M22 backlog |

**Refinements to TaskManager design (D19-1 through D19-5):**

| ID | Decision |
|---|---|
| D19-1 | Package: `internal/gui/task/` — not top-level; keeps GUI cohesion |
| D19-2 | **Zero `program.Send`** — TaskManager returns `tea.Cmd` chains; streaming via `TaskOutputMsg` handled in `Update()` |
| D19-3 | Reads (`fetchPanelData`, tab fetch) **outside** TaskManager — only writes + long-running diagnostic streams |
| D19-4 | Progress modal stays on `Model`; manager emits messages only |
| D19-5 | Queue max depth: 10 tasks; overflow → toast "Queue full" (prevent unbounded memory) |

---

## Challenge 3 — Is M20 overloaded?

**Original:** 10 steps in one milestone (~1 week)

**Challenge:** Mixes correctness bugs (tab cache), feature completion (batch upgrade), config wiring, and layout — different risk profiles and testers.

**Resolution — internal phases (execute in order):**

| Phase | Steps | Theme | Gate |
|---|---|---|---|
| **A — Data truth** | 20.1, 20.6, 20.9 | Tab keys, typed outdated data, errors | Tab tests pass |
| **B — Display truth** | 20.2, 20.5, 20.10 | Info tab, empty states, scroll viewport | Snapshot tests pass |
| **C — Action truth** | 20.3, 20.4 | Batch upgrade, pin semantics | TaskManager integration tests |
| **D — Config truth** | 20.8 | Wire or remove dead config fields | Config tests + DESIGN updated |
| **E — Layout minimum** | 20.7 | Small terminal warning only | Manual 80×24 check |
| **F — Verification** | 20.11 | End-to-end manual smoke script | Checklist signed off |

**Moved out of M20:**

| Item | New home | Reason |
|---|---|---|
| Controller package split | Backlog B-01 | Refactor, not correctness |
| M17 accordion/boxes | M17 | Visual polish |
| Search info preview | M17.11 (existing) | Depends on stable tab/content model from 20.1 |

---

## Challenge 4 — Should tests wait for perfect UX?

**Original:** M21 entirely after M20

**Challenge:** Without CI and integration tests, M19–M20 regressions won't be caught during execution.

**Resolution — split M21 into tiers:**

| Tier | Steps | Start when | Purpose |
|---|---|---|---|
| **T0 — Safety net** | 21.0, 21.6, 21.7 | Before / during M19 | TypedCache, cache race, regression stubs |
| **T1 — Infra** | 21.1, 21.3 | M19.5 in progress | teatest helper, integration file |
| **T2 — Behavior** | 21.2, 21.4 | M20 phase A done | E2E flows, bug regression tests |
| **T3 — Gates** | 21.5 | M21 T2 done | Makefile coverage floors |

**Critical rule:** Every M19 and M20 step lists its tests **in the same step** — not at milestone end.

---

## Challenge 5 — Should CI wait for full test pyramid?

**Original:** M22 after M21

**Challenge:** Developers (and agents) executing M19–M20 without CI will repeat the "no integration tests" mistake.

**Resolution:**

| Step | When | What |
|---|---|---|
| M22.1 minimal CI | After M19.5 + M21.0 | `lint` + `test` + `build` on Ubuntu |
| M22.2 integration workflow | After M21.3 | Manual macOS brew job |
| M22.3–22.5 | After M21 T3 | Goreleaser, checklist, dependabot |

---

## Challenge 6 — Is M17 deferral still correct?

**Yes.** M17 changes layout math (`computeContentHeights`, per-panel boxes). Doing that before tab cache and TaskManager fixes means:

- Visual work on wrong Info tab content
- Re-test layout after every correctness fix
- Higher churn in `render.go`

**M20.7 allowance:** Warning banner only at `<80×24` — no accordion. Accordion stays M17.5.

---

## Challenge 7 — Config dead fields: document or implement?

**Original:** "Choose one per field" in M18 — ambiguous for executors.

**Locked decisions (implement in M20.8):**

| Field | M18 action (document) | M20 action (implement) |
|---|---|---|
| `AutoRefreshSeconds` | Mark **planned** in DESIGN + config comment | Tick `RefreshMsg` when > 0; 0 = disabled |
| `Brew.Path` | Mark **planned** | Pass to runner constructor; empty = auto-detect |
| `ShowIcons` | Mark **deferred** (P2) | Remove from user-facing docs until M17 icons; keep field, no-op with comment |

**Rationale:** Removing fields breaks existing user configs; better to wire or no-op with honesty.

---

## Challenge 8 — Missing backlog items from review

These are real gaps but **not in M18–M22** — tracked to avoid scope creep:

| ID | Item | When |
|---|---|---|
| B-01 | `internal/gui/controllers/` split | Post v0.2.0 |
| B-02 | Lazy panel loading (fetch active first) | Performance pass after M22 |
| B-03 | `tabContent` LRU / max entries | If memory issue observed |
| B-04 | Search info preview | M17.11 |
| B-05 | testify adoption | Optional; stdlib sufficient for now |
| B-06 | Official tap untap UX polish | M7 remainder verification |

---

## Revised Execution Timeline (suggested)

```
Week 1
├── M18.1–M18.5  (LICENSE, README, audit, DESIGN skeleton)
├── M19.0–M19.5  (TypedCache, task types, manager core, Model wire)
├── M21.0, M21.6, M21.7  (safety tests — parallel)
└── M22.1  (minimal CI — after M19.5)

Week 2
├── M19.6–M19.10  (migrate all handlers, remove isBusy)
├── M18.6–M18.10  (complete docs — parallel)
├── M21.1, M21.3  (teatest helper, integration)
└── M20 phase A  (tab cache, outdated typed, errors)

Week 3
├── M20 phase B–C  (info tab, empty states, viewport, batch, pin)
├── M21.2, M21.4  (E2E flows, regression)
└── M20 phase D–F  (config, small term, smoke)

Week 4
├── M21.5  (coverage gates)
├── M22.2–M22.5  (integration CI, goreleaser, release checklist)
└── M17  (visual polish — now safe)
```

Estimates assume single developer/agent; parallel tracks reduce calendar time.

---

## Adoption Readiness Checklist

Before marking "planning complete" for M18–M22:

- [x] Each milestone has step index with sizes and dependencies
- [x] Each step has acceptance criteria + tests in same step
- [x] Contested decisions recorded (this file + per-milestone ADRs)
- [x] Out-of-scope explicit per milestone
- [x] Parallel tracks documented in status.md
- [x] Review template for future audits
- [x] Backlog IDs for deferred items
- [ ] **Human sign-off** on locked config decisions (ShowIcons deferred)

---

## References

- [architecture-review-2026-06-13.md](architecture-review-2026-06-13.md) — first pass findings
- [review-template.md](review-template.md) — future review process
- [templates/README.md](templates/README.md) — milestone/status conventions
