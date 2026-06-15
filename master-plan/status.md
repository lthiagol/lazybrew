# Lazybrew — Master Plan Status

> **Project:** lazybrew — A TUI for managing Homebrew  
> **Stack:** Go + Bubble Tea + Lip Gloss + Bubbles  
> **Platforms:** macOS + Linux  
> **Created:** 2026-06-11  
> **Last Updated:** 2026-06-14 (milestone 23 complete)  
> **Target Homebrew:** 6.0.0+

---

## Planning Documents

| Document | Purpose |
|---|---|
| [architecture-review-2026-06-13.md](architecture-review-2026-06-13.md) | First-pass findings |
| [planning-challenge-2026-06-13.md](planning-challenge-2026-06-13.md) | Challenged decisions + revised sequencing |
| [review-template.md](review-template.md) | **Project-agnostic** template for future reviews |
| [templates/](templates/) | Milestone, status, and step templates + conventions |
| [milestone-legacy-index.md](milestone-legacy-index.md) | M1–M17 format audit + routing to M18–M22 |
| [backlog.md](backlog.md) | Deferred items (out of M18–M22 scope) |
| [smoke-checklist.md](smoke-checklist.md) | Manual verification (M20.11) |
| [coverage-audit.md](coverage-audit.md) | Brew command coverage |

---

## Overall Progress

```
[X]  Milestone 1   — Foundation
[~]  Milestone 2   — TUI Shell (small terminal → M20.7)
[X]  Milestone 3   — Brew Data Layer
[~]  Milestone 4   — Read-Only Panels (Info tab → M20.2)
[X]  Milestone 5   — Modals & Search
[~]  Milestone 6   — Package Mutations (TaskManager → M19)
[~]  Milestone 7–11 — Features (verify DoD during M20 smoke)
[~]  Milestone 12  — Test Infrastructure (E2E → M21)
[X]  Milestone 13  — Critical Bug Fixes
[~]  Milestone 14–16 — Wire/cleanup/tests (partial)
[X]  Milestone 23  — TUI Layout Rework & Debug Logging
[ ]  Milestone 17  — Lazygit UI (after M19–M22)
[ ]  Milestone 18  — Documentation & Hygiene
[ ]  Milestone 19  — Concurrency & TaskManager
[ ]  Milestone 20  — Functional & UX (phases A–F)
[ ]  Milestone 21  — Test Strategy v2 (tiers T0–T3)
[ ]  Milestone 22  — CI & Release
```

**Legend:** `[X]` complete · `[~]` partial · `[ ]` not started

**Current phase:** Milestone 23 complete  
**Execution entry point:** M18.1 + M19.0 in parallel, then M18.5 → M19.1

---

## Parallel Execution Tracks

Tracks can run concurrently when dependencies allow.

| Track | Milestones | Owner focus | Starts | Blocks |
|---|---|---|---|---|
| **A — Docs** | M18 | README, DESIGN, AGENTS, audit | Day 1 | M19.6 needs 18.8 |
| **B — Concurrency** | M19 | TaskManager | After M18.5 | M20.3, M17 |
| **C — UX** | M20 | Tab cache, Info, batch | M20.A during M19.8 | M17 |
| **D — Quality** | M21 T0–T3 | tests | T0 immediately | M22 full |
| **E — Ops** | M22.1 early | CI | After M19.5 | Release |

### Critical path

```
M18.5 → M19.1 → M19.5 → M19.6 → M20.A → M20.B → M21.T2 → M22.4 → M17
```

---

## Milestone Index (M18–M22 detail)

| # | Milestone | Steps | Size | Gate |
|---|---|---|---|---|
| 18 | [Documentation](milestones/18-documentation-and-project-hygiene.md) | 18.1–18.10 | M | DESIGN + AGENTS exist |
| 19 | [TaskManager](milestones/19-bubble-tea-concurrency-and-task-manager.md) | 19.0–19.10 | L | No program.Send |
| 20 | [Functional UX](milestones/20-functional-completeness-and-ux.md) | 20.1–20.11 | L | smoke-checklist pass |
| 21 | [Tests v2](milestones/21-test-strategy-v2.md) | 21.0–21.5 | M–L | 8 E2E + 5 integration |
| 22 | [CI & Release](milestones/22-ci-and-release-hardening.md) | 22.1–22.6 | M | CI green + goreleaser |
| 23 | [TUI Layout Rework](milestones/23-tui-layout-and-debug-logging.md) | 23.1–23.8 | M | Layout fills space, two-line bar, debug log |
| 17 | [Lazygit UI](milestones/17-lazygit-tui-and-auto-update.md) | 17.1–17.11 (phases A–D) | L | After M19–M22; refined template |

Legacy milestones M1–M17: see [milestone-legacy-index.md](milestone-legacy-index.md)

---

## Challenged Decisions (summary)

Full rationale: [planning-challenge-2026-06-13.md](planning-challenge-2026-06-13.md)

| Decision | Outcome |
|---|---|
| Docs before all code? | **No** — M19 starts after M18.5 only |
| M20 monolith? | **Split** into phases A–F |
| M21 after all M20? | **Tiered** — T0/T1 parallel with M19 |
| M22 after all M21? | **Split** — 22.1 after M19.5 |
| M17 defer? | **Yes** — visual polish last |
| TypedCache fix? | **Moved to M19.0** |
| Config dead fields? | **Document M18.9, wire M20.8** |

---

## Metrics Baseline (2026-06-13)

| Metric | Value |
|---|---|
| Go lines | ~7,821 |
| Test functions | 162 |
| `brew/` coverage | 65.2% |
| `gui/` coverage | 31.5% |
| Integration tests | 0 |
| teatest E2E | 0 |
| CI workflows | 0 |

---

## Testing Strategy

| Tier | When | What |
|---|---|---|
| T0 | During M19 | TypedCache, cache race, MockRunner recorder |
| T1 | M19.5+ | teatest helper, integration file |
| T2 | M20.A+ | E2E flows, regression tests |
| T3 | Pre-release | Makefile coverage floors |

Future reviews: use [review-template.md](review-template.md).

---

## Decision Log (recent)

| Date | Decision | Context |
|---|---|---|
| 2026-06-13 | Planning pass 2 | Granular steps, challenged M18–M22 sequencing |
| 2026-06-13 | review-template.md | Project-agnostic audit template for agents |
| 2026-06-13 | M19.0 TypedCache | Moved from M21 — before concurrency |
| 2026-06-13 | M22.1 early CI | Don't wait for full test pyramid |
| 2026-06-13 | backlog.md | B-01–B-10 deferred explicitly |
| 2026-06-13 | M17 last | Visual work after correctness + CI |
| 2026-06-14 | M23 complete | TUI layout rework (fill, command log, two-line bar, spinner, debug log) |

---

## Adoption Readiness

Planning for M18–M22 is **execution-ready** when:

- [x] Each step has size, deps, acceptance criteria, tests
- [x] Challenged decisions documented
- [x] Parallel tracks defined
- [x] Out-of-scope in backlog.md
- [x] Review template for future audits
- [x] Milestone/status templates in [templates/](templates/)
- [ ] Human sign-off on config ADR (ShowIcons deferred — see M18.9)

**Ready for execution:** ☑ Yes (pending ShowIcons sign-off only)
