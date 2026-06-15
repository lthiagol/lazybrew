# Lazybrew — Master Plan Status

> **Project:** lazybrew — A TUI for managing Homebrew  
> **Stack:** Go + Bubble Tea + Lip Gloss + Bubbles  
> **Platforms:** macOS + Linux  
> **Created:** 2026-06-11  
> **Last Updated:** 2026-06-14 (plan synced to code reality; release readiness tracked)  
> **Target Homebrew:** 6.0.0+

---

## Planning Documents

| Document | Purpose |
|---|---|
| [architecture-review-2026-06-13.md](archive/architecture-review-2026-06-13.md) | First-pass findings |
| [planning-challenge-2026-06-13.md](archive/planning-challenge-2026-06-13.md) | Challenged decisions + revised sequencing |
| [review-template.md](review-template.md) | **Project-agnostic** template for future reviews |
| [templates/](templates/) | Milestone, status, and step templates + conventions |
| [milestone-legacy-index.md](archive/milestone-legacy-index.md) | M1–M17 format audit + routing to M18–M22 |
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
[X]  Milestone 17  — Lazygit UI (parse summary toast → 17.3)
[X]  Milestone 18  — Documentation & Hygiene
[X]  Milestone 19  — Concurrency & TaskManager
[X]  Milestone 20  — Functional & UX (phases A–F)
[X]  Milestone 21  — Test Strategy v2 (8 teatest flows done)
[X]  Milestone 22  — CI & Release (22.1a/22.1b/22.2/22.3a/22.5/22.6 done; 22.4 remaining)
```

**Legend:** `[X]` complete · `[~]` partial · `[ ]` not started

**Current phase:** Release readiness  
**Execution entry point:** M22.1b (verify CI green), then M17.3 + M21.2a/b in parallel, then M22.3 → M22.4

---

## Parallel Execution Tracks

Tracks can run concurrently when dependencies allow.

| Track | Milestones | Owner focus | Starts | Blocks |
|---|---|---|---|---|
| **A — Docs** | M18 | Complete | — | — |
| **B — Polish** | M17.3 | Update summary toast | Now | none |
| **C — Quality** | M21.2a/b | Install + uninstall teatest flows | After M22.1 green | Release tag |
| **D — Ops** | M22 | Verify CI green, goreleaser, release checklist | Now | Release tag |

### Critical path

```
M22.1b → M21.2a/b → M22.3a → M22.3b → M22.4
```

M17.3 can land in parallel with the critical path.

---

## Milestone Index (M17–M23 detail)

| # | Milestone | Steps | Size | Gate | Status |
|---|---|---|---|---|---|
| 17 | [Lazygit UI](milestones/17-lazygit-tui-and-auto-update.md) | 17.1–17.3 (phases A–D) | S remaining | Update summary toast | ~95% done |
| 18 | [Documentation](milestones/18-documentation-and-project-hygiene.md) | 18.1–18.10 | Done | `AGENTS.md` | Done |
| 19 | [TaskManager](milestones/19-bubble-tea-concurrency-and-task-manager.md) | 19.0–19.10 | L | No `program.Send` | Done |
| 20 | [Functional UX](milestones/20-functional-completeness-and-ux.md) | 20.1–20.11 | L | smoke-checklist pass | Done |
| 21 | [Tests v2](milestones/21-test-strategy-v2.md) | 21.0–21.5 | S remaining | 8 E2E flows | ~80% done |
| 22 | [CI & Release](milestones/22-ci-and-release-hardening.md) | 22.1a–22.4 | M | CI green + goreleaser | Partial |
| 23 | [TUI Layout Rework](milestones/23-tui-layout-and-debug-logging.md) | 23.1–23.8 | M | Layout fills space, two-line bar, debug log | Done |

Legacy milestones M1–M16: see [milestone-legacy-index.md](archive/milestone-legacy-index.md). Open items from M1–M16 are either done in M17–M23 or tracked in the backlog.

---

## Challenged Decisions (summary)

Full rationale: [planning-challenge-2026-06-13.md](archive/planning-challenge-2026-06-13.md)

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

## Metrics Baseline (2026-06-14)

| Metric | Value |
|---|---|
| Go lines | ~8,200 |
| Test functions | ~180 |
| `brew/` coverage | 62.2% |
| `gui/` coverage | 36.6% |
| `gui/presentation/` coverage | 91.6% |
| `gui/modal/` coverage | 41.4% |
| Integration tests | 5 |
| teatest E2E | 6 |
| CI workflows | 2 (ci.yml + integration.yml) |

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
| 2026-06-14 | Plan/code reality sync | M17/M19–M21/M23 found implemented; status + milestones updated to truth |
| 2026-06-14 | Release readiness focus | Remaining work: AGENTS.md, CI, 2 teatest flows, update summary toast |
| 2026-06-14 | Retire first-version branch | Merged to main; default branch changed to main; branch deleted |
| 2026-06-14 | M17.3, M21.2 complete | parseUpdateSummary + toast, install teatest, uninstall teatest done; 8 flows total |

---

## Release Readiness

Before tagging v0.2.0:

- [x] M19 TaskManager done; zero `program.Send`
- [x] M20 functional correctness done; smoke-checklist pass
- [x] M21 T0–T1 done; regression tests linked to architecture review
- [x] M18.8 `AGENTS.md` exists and linked
- [x] M17.3 update summary toast implemented
- [x] M21.2 8 teatest flows done
- [x] M22.1b CI green on push/PR (confirmed 5/5 green)
- [x] M22.2 integration workflow file exists
- [x] first-version branch retired
- [x] M22.3a goreleaser snapshot succeeds (4 tarballs + checksums, no deprecations)
- [ ] M22.4 release checklist signed off
- [x] M22.6 branch protection recommendations documented
- [x] Coverage floors raised to current actuals

**Ready for release:** ☐ No — remaining items above block v0.2.0 tag

---

## Release Readiness Execution Plan

Remaining work is small and mostly parallelizable. Suggested order:

### Phase 1 — Independent implementation (can run in parallel)

| Step | Owner | Deliverable | Notes |
|---|---|---|---|
| **M17.3** | Any | `parseUpdateSummary` + toast | See milestone file for exact handler change and tests |
| **M21.2a** | Any | `install_test.go` teatest | Requires custom MockRunner to record install args |
| **M21.2b** | Any | `uninstall_test.go` teatest | Seed Formulae panel with one item, confirm modal, assert uninstall args |

### Phase 2 — External verification (wait after push)

| Step | Owner | Deliverable | Notes |
|---|---|---|---|
| **M22.1b** | GitHub Actions | CI badge green | Already triggered; check Actions tab. Fix any flakiness before release |
| **M22.3a** | Human/agent | `goreleaser release --snapshot --clean` green | Install goreleaser CLI first |
| **M22.3b** | Human/agent | `.goreleaser.yml` fixed if needed | Only if 22.3a fails |

### Phase 3 — Final gate

| Step | Owner | Deliverable | Notes |
|---|---|---|---|
| **M22.4** | Human | Signed `release-checklist.md` | All pre-release checkboxes verified |
| **Tag** | Human | `v0.2.0` tag + goreleaser publish | Needs `GITHUB_TOKEN` |

### Risks to watch

- **CI flakiness:** `-race` on GitHub Actions can be slower than locally; if it times out, increase test timeout or split jobs.
- **Goreleaser version:** The config uses `version: 2`. Ensure installed goreleaser is v2.x.
- **SSH/agent:** Future pushes from this environment may need `SSH_AUTH_SOCK` pointed at the active socket.
