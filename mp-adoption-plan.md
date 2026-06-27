# mp Adoption Plan — lazybrew

> **Status:** 🟡 Planning — execution deferred to a later agentic session
> **Branch:** `chore/mp-adoption` (created from `main` @ `4e87066`)
> **Authored:** 2026-06-26
> **Scope:** Migrate lazybrew's bespoke `master-plan/` directory onto the **mp** (Master Plan) CLI, v1.4.0+.
>
> ⚠️ **This file is intentionally OUTSIDE mp.** It is the bootstrap plan that brings mp into the repo. It must be **deleted** once adoption is complete (see §10 DoD). Do not import it as an mp artifact.

---

## 1. Why migrate

### Current state
lazybrew uses a **bespoke, hand-authored** `master-plan/` layout (status.md, milestones/*.md, backlog.md, templates/). It follows conventions documented in `AGENTS.md`, but:

- All plan I/O is **freehand markdown editing** — no schema, no validation.
- No enforced "spec before code" gate.
- Merge conflicts are content-based and semantic (a number collision between two branches just cost a full renumbering pass — see `status.md` decision log 2026-06-26).
- Milestone numbering, step IDs, and cross-references drift and must be hand-maintained.

mp is already installed locally (`/home/thiago/.agents/master-plan/bin/mp`, v1.4.0). This project does not yet use it.

### Goals
- Single source of truth for plans, driven by `mp` CLI (validated TOML/JSON).
- Enforced spec-before-code and acceptance-criteria coverage.
- Stable, generated step IDs (no more hand-renumbering).
- Reproducible plan reads/writes for agents (`mp … --format json`).

### Non-goals
- Re-planning the product roadmap — the *content* (M24–M28, backlog) is preserved, only re-shaped.
- Migrating operational docs (release/smoke/coverage checklists) into mp.
- Changing application source code.

---

## 2. Current state inventory

| Artifact | Disposition | mp target |
|---|---|---|
| `status.md` header (project/stack/platforms) | Distill | **brief** |
| `status.md` "Challenged Decisions" + release scope | Distill | **charter** goals/non-goals |
| `milestones/24-smoke-test-fixes.md` (code-complete) | Import | **milestone** — `spec_status: ready`, steps done except S24.13 |
| `milestones/25-outdated-panel-performance.md` | Import | **milestone** (greenfield-style, planned) |
| `milestones/26-diagnostics-error-handling.md` | Import | **milestone** (planned) |
| `milestones/27-taps-panel-batch-loading.md` | Import | **milestone** (planned) |
| `milestones/28-tiered-refresh-and-polling-strategy.md` | Import | **milestone** (planned) |
| `milestones/01–23-*.md` (all complete) | **Archive, do not import** | keep `archive/milestone-legacy-index.md` as history |
| `backlog.md` B-01, B-07 (Medium, near-term) | Import | **tracks** |
| `backlog.md` B-09, B-11, B-12, B-13 (Low/conditional) | Import | **ideas** |
| Milestone prose (why-now, ADRs, rollback, phase gates) | Preserve | **`context.references`** + milestone notes / `DESIGN.md` |
| `templates/` (custom) | **Retire** | mp ships its own |
| `release-checklist.md`, `smoke-checklist.md`, `coverage-audit.md`, `review-template.md` | **Keep as repo docs** (outside mp) | operational, not plan artifacts |
| `archive/*` | Keep | historical |

---

## 3. Adoption decisions

Documented as **recommended defaults**. D2 and D3 have real trade-offs — **confirm with the user before the irreversible cutover (Phase 5)**.

| ID | Decision | Adopted | Rationale / trade-off |
|---|---|---|---|
| **D1** | Profile | **`full`** | Project has multi-step milestones with specs/ADRs/ACs. `hybrid`/`session` are too lightweight. |
| **D2** | Historical milestones M1–M23 | **Archive; import only M24–M28** ⚠️ confirm | mp plans forward. Importing 23 done milestones is heavy and low-value. Alternative: import all as `complete` (full history in mp, but noisy). |
| **D3** | Plan-dir location during migration | **Parallel `.mp/`, cutover later** ⚠️ confirm | Rebuilding `master-plan/` in place is irreversible and mixes old+new during transition. `.mp/` lets you validate mp output first, then rename at cutover. |
| **D4** | Operational checklists | **Keep outside mp** | They are runbooks, not plan artifacts. mp has no concept for them. |
| **D5** | Backlog routing | **B-01, B-07 → tracks; rest → ideas** | Tracks = near-term work; ideas = parking lot. Matches backlog priorities. |

---

## 4. Target state (post-adoption)

```
lazybrew/
├── master-plan/            # mp-managed (renamed from .mp/ at cutover) OR .mp/ per config
│   ├── brief.md            # from status.md header
│   ├── charter.*           # goals/non-goals
│   ├── milestones/
│   │   ├── 24/ …          # mp milestone artifacts (M24–M28)
│   │   └── 28/
│   ├── ideas.toml          # B-09, B-11, B-12, B-13
│   ├── tracks/             # B-01, B-07
│   └── config (mp-managed)
├── docs/operational/       # (or repo root) release-checklist, smoke-checklist, coverage-audit, review-template
├── master-plan-archive/    # OLD bespoke docs (kept until confident, then deleted)
├── AGENTS.md               # Planning Rules rewritten → mp CLI workflow
└── DESIGN.md               # + adoption ADR in decision log
```

`mp validate` is green; `mp status` reflects M24 (near-done) + M25–M28 (planned).

---

## 5. Content mapping & schema

### mp milestone schema (minimal valid)
```json
{
  "title": "...",
  "intent": { "outcome": "What users can do after this ships." },
  "problem": { "description": "The gap this fills." },
  "scope": { "in_scope": ["..."], "out_of_scope": ["...", "..."] },
  "acceptance_criteria": [
    { "description": "Observable behavior", "verification": "test command or manual check" }
  ]
}
```

### Field mapping (existing milestone → mp)
| Existing section | mp field |
|---|---|
| `## Goal` | `intent.outcome` + `problem.description` |
| `## Out of Scope` | `scope.out_of_scope` |
| Step deliverables / `## Goal` deliverables | `scope.in_scope` |
| `## Architecture Decisions (ADRs)` | `context.references` + `DESIGN.md` decision log |
| `Gate criteria` (header) | `acceptance_criteria[]` (one per observable criterion) |
| `## Step Index` / `## Steps` | generated by `mp milestone decompose` → `S1`, `S1.1`, … |
| `## Why Now`, `## Challenged Assumptions`, `## Rollback Plan` | `context.references` / milestone notes |

### Worked example — M26 (Diagnostics)
```json
{
  "title": "Diagnostics Error Handling",
  "intent": { "outcome": "brew doctor warnings display correctly and the debug log stops filling with false failures." },
  "problem": { "description": "brew doctor exits 1 on warnings; lazybrew treats that as a hard failure and logs WARN every refresh." },
  "scope": {
    "in_scope": [
      "Accept exit=1 from brew doctor and parse stdout for warnings",
      "Audit other readers for non-zero-exit misclassification"
    ],
    "out_of_scope": [
      "Reducing how often doctor runs (covered by M28)",
      "Changing doctor parsing format"
    ]
  },
  "acceptance_criteria": [
    { "description": "exit=1 yields parsed warnings, no error", "verification": "TestDoctorExitOneReturnsWarnings" },
    { "description": "exit=2 yields error", "verification": "TestDoctorExitTwoReturnsError" }
  ]
}
```
M25, M27, M28 follow the same pattern — their existing `## Goal`, `## Out of Scope`, and gate criteria map directly.

---

## 6. Execution plan (phased)

Execute in order. Each phase has acceptance criteria. Run `mp validate` after every write phase.

### Phase 0 — Prep
- [ ] Branch `chore/mp-adoption` exists and is current (already done).
- [ ] Record this plan + D1–D5 as a decision in `DESIGN.md` (adoption ADR).
- **AC:** decision logged; no plan files touched yet.

### Phase 1 — Bootstrap
- [ ] `mp init --profile full --from-repo`
- [ ] `mp doctor --format json` → healthy
- [ ] `mp config show --format json` → confirm plan dir (`.mp/` per D3) and that it is gitignored
- **AC:** `mp doctor` reports healthy; `.mp/` exists and is ignored by git (or intentionally tracked — decide).
- **Files:** `.mp/` created.

### Phase 2 — Brief & charter
- [ ] `mp brief todo --format json`
- [ ] Fill brief topics from current `status.md` header (`mp brief edit`)
- [ ] `mp brief done`
- [ ] `mp interview checklist --checklist-type charter` → goals (ship v0.2.0 → perf pass) / non-goals (from Challenged Decisions)
- **AC:** `mp validate` green; brief `status: done`.
- **Gotcha:** charter requires `brief done` first (error `B1`/`B3` otherwise).

### Phase 3 — Import forward milestones (M24–M28)
- [ ] M24: `mp milestone create --json @-` → `set-spec-status review` → `approve` → `decompose`; mark steps `done` except the manual smoke step (S24.13 stays pending).
- [ ] M25, M26, M27, M28: create from §5 mapping → `approve` → `decompose`.
- [ ] Set statuses: M24 `in-progress` (or appropriate); M25–M28 planned.
- [ ] `mp validate`
- **AC:** 5 milestones exist with `spec_status: ready`; steps decomposed; ACs present.
- **Gotcha:** mp blocks `in-progress` until `spec_status: ready` (error `G1`). M24 is already code-complete — import carefully so this doesn't flag.
- **Gotcha:** each milestone needs ≥2 out-of-scope items (error `G4`) — existing Out-of-Scope sections satisfy this.

### Phase 4 — Import backlog
- [ ] B-01, B-07 → `mp track add` (near-term, Medium)
- [ ] B-09, B-11, B-12, B-13 → `mp idea create` (parking lot)
- [ ] `mp validate`
- **AC:** backlog items present as tracks/ideas; `master-plan/backlog.md` content fully represented.

### Phase 5 — Cutover ⚠️ confirm D2/D3 with user first
- [ ] Move old `master-plan/*.md` → `master-plan-archive/` (preserve history; delete later when confident).
- [ ] Rename `.mp/` → `master-plan/` **or** keep `.mp/` and set `config.workflow.plan.location`.
- [ ] Move operational docs (`release-checklist.md`, `smoke-checklist.md`, `coverage-audit.md`, `review-template.md`) out of the plan dir to `docs/operational/` (or repo root).
- [ ] Rewrite `AGENTS.md` "Planning Rules": replace hand-edit rules with the mp session-start sequence (see §8) and "all plan I/O via `mp`; never hand-edit plan files".
- [ ] Add adoption ADR to `DESIGN.md` decision log.
- [ ] `mp validate`
- **AC:** `master-plan/` is mp-managed and the single source of truth; `AGENTS.md` reflects mp workflow; old docs archived.
- **Rollback:** restore `master-plan/` from `master-plan-archive/`; revert `AGENTS.md`.

### Phase 6 — Merge
- [ ] Single PR `chore/mp-adoption` → `main`.
- [ ] CI green (`go test -race ./...`, `go vet ./...`, `make lint`).
- [ ] `mp doctor` + `mp status` green on `main` after merge.
- **AC:** merged; plan reads correctly via `mp status`.

---

## 7. Risks & mitigations

| Risk | Mitigation |
|---|---|
| Narrative loss (mp schema is tighter than current prose) | Preserve why-now/ADR/rollback in `context.references` and `DESIGN.md`; don't discard. |
| "Spec before code" flags M24 (already code-complete) | Import M24 with `spec_status: ready` and steps already `done`; only S24.13 pending. |
| Step ID renumbering breaks external references | mp generates stable IDs; re-map any stray refs. (This is what mp exists to prevent.) |
| Big-bang doc commit / merge conflicts | Do on dedicated branch; one PR; not during active `wip` work. |
| mp version drift | Pin to v1.4.0+; record version in adoption ADR. |
| `.mp/` accidentally committed with secrets/state | Confirm gitignore in Phase 1; review before PR. |

---

## 8. Session handoff (for the executing agent)

**You are picking up a planned, not-yet-executed migration. Read this whole file first.**

### Starting commands
```bash
mp doctor --format json          # toolkit + project readiness
mp config show --format json     # confirm profile + plan dir
mp execution status              # should be mode=planning
```

### Hard rules
1. **Plan zone vs code zone:** under the plan dir, use **only `mp` commands** — never hand-edit plan files. (See `AGENTS.md` after Phase 5.)
2. **Spec before code:** no app source changes until a milestone has `spec_status: ready`.
3. **Confirm with the user** before Phase 5 (irreversible cutover) — especially D2 and D3.
4. Run `mp validate` after every write phase.

### mp gotchas (CLI ≠ docs)
| Doc says | Use instead |
|---|---|
| `mp idea add` | `mp idea create` |
| `mp interview --type X` | `--checklist-type X` |
| `mp plan show --all` | `mp plan show` (no flag) |
| `mp step done --evidence "x"` | no `--evidence` flag on `step done` |
| `mp milestone criterion pass --evidence "x"` | `--evidence` is **positional**, not a flag |

### Key validation errors
| Code | Meaning |
|---|---|
| `B1`/`B3` | brief not done before charter |
| `G1` | `in-progress` without `spec_status: ready` |
| `G4` | fewer than 2 out-of-scope items |
| `G6`/`G7` | ACs not passed / `done` before `verified` |
| `G8` | dependency milestone not complete |

### What "done" looks like
- §10 DoD all checked.
- `mp status` shows M24 (near-complete) + M25–M28 (planned).
- `mp validate` green.
- This file (`mp-adoption-plan.md`) **deleted**.

---

## 9. Rollback (abandon adoption)

If the migration goes wrong before merge:
```bash
git checkout main
git branch -D chore/mp-adoption     # discard the branch
rm -rf .mp/                          # if bootstrapped locally
```
The bespoke `master-plan/` on `main` is untouched until Phase 5 cutover, so pre-cutover rollback is lossless.

---

## 10. Definition of Done

- [ ] mp bootstrapped (`full` profile), `mp doctor` healthy
- [ ] Brief + charter imported from `status.md`
- [ ] M24–M28 imported as mp milestones with specs, steps, ACs
- [ ] Backlog B-01/B-07 → tracks; B-09/B-11/B-12/B-13 → ideas
- [ ] `master-plan/` is mp-managed and the single source of truth
- [ ] Operational checklists moved out of plan dir
- [ ] `AGENTS.md` Planning Rules rewritten for mp workflow
- [ ] Adoption ADR in `DESIGN.md`
- [ ] `mp validate` green; CI green; merged to `main`
- [ ] **This file deleted**
