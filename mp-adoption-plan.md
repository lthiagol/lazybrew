# mp Adoption Plan — lazybrew

> **Status:** 🟡 Planning — execution deferred to a later agentic session
> **Branch:** `chore/mp-adoption` (created from `main` @ `4e87066`)
> **Authored:** 2026-06-26
> **Scope:** Migrate lazybrew's bespoke `master-plan/` directory onto the **mp** (Master Plan) CLI, v1.4.0+.
>
> ⚠️ **This file is intentionally OUTSIDE mp.** It is the bootstrap plan that brings mp into the repo. It must be **deleted** once adoption is complete (see §10 DoD). Do not import it as an mp artifact.
>
> 📒 **Companion file:** `mp-adoption-plan-log.md` lives next to this file and is the executing agent's running progress log. Unlike this plan, the log is **retained** — it is real-world feedback for the mp project itself (see §8 "Progress log").

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
- Re-planning the product roadmap — the *content* (M25–M28 + backlog) is re-shaped, not re-thought. M24 and M1–M23 stay historical (see D2).
- Migrating operational docs (release/smoke/coverage checklists) into mp.
- Changing application source code.

---

## 2. Current state inventory

| Artifact | Disposition | mp target |
|---|---|---|
| `status.md` header (project/stack/platforms) | Distill | **brief** |
| `status.md` "Challenged Decisions" + release scope | Distill | **charter** goals/non-goals |
| `milestones/24-smoke-test-fixes.md` (code-complete; smoke pending) | **Archive, do not import** | release gate stays in `release-checklist.md`; noted in brief/charter |
| `milestones/25-outdated-panel-performance.md` | Import | **milestone** (planned) |
| `milestones/26-diagnostics-error-handling.md` | Import | **milestone** (planned) |
| `milestones/27-taps-panel-batch-loading.md` | Import | **milestone** (planned) |
| `milestones/28-tiered-refresh-and-polling-strategy.md` | Import | **milestone** (planned) |
| `milestones/01–23-*.md` (all complete) | **Archive, do not import** | keep `archive/milestone-legacy-index.md` as history |
| `backlog.md` B-01 … B-13 | Import | **`mp backlog add`** (native B-xx IDs — match yours) |
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
| **D2** | Historical milestones M1–M24 | **Archive; import only M25–M28** ⚠️ confirm | mp plans forward. M24 is code-complete (only human smoke remains) — importing done work fights mp's spec-before-code gates (G1/G6/G7) for little value. The v0.2.0 gate stays in `release-checklist.md` and is noted in the charter. |
| **D3** | Plan-dir location during migration | **Parallel `.mp/`, cutover later** ⚠️ confirm | Rebuilding `master-plan/` in place is irreversible and mixes old+new during transition. `.mp/` lets you validate mp output first, then rename at cutover. |
| **D4** | Operational checklists | **Keep outside mp** | They are runbooks, not plan artifacts. mp has no concept for them. |
| **D5** | Backlog routing | **All B-xx → `mp backlog add`** | mp has a native backlog with **B-xx IDs** (matching yours). Tracks are for in-flight `bugfix`/`tweak` work; ideas are quick reminders — wrong fit for deferred backlog items. |
| **D6** | Plan committed to git? | **Yes — `workflow.plan.in_repo = true`** ⚠️ confirm | mp can gitignore the plan dir (`workflow.plan.in_repo`; when `false`, `mp init` appends it to `.gitignore`). lazybrew's bespoke `master-plan/` is committed and shared across agents/CI, so commit the mp plan too. Verify mp's default for the `full` profile and set `in_repo` explicitly in Phase 1. |

---

## 4. Target state (post-adoption)

```
lazybrew/
├── master-plan/            # mp-managed (renamed from .mp/ at cutover) OR .mp/ per config
│   ├── brief.md            # from status.md header
│   ├── charter.*           # goals/non-goals
│   ├── milestones/
│   │   ├── 25/ …          # mp milestone artifacts (M25–M28)
│   │   └── 28/
│   ├── backlog.toml        # B-01 … B-13 (native mp backlog, B-xx IDs)
│   └── config (mp-managed)
├── docs/operational/       # (or repo root) release-checklist, smoke-checklist, coverage-audit, review-template
├── master-plan-archive/    # OLD bespoke docs (kept until confident, then deleted)
├── AGENTS.md               # Planning Rules rewritten → mp CLI workflow
└── DESIGN.md               # + adoption ADR in decision log
```

`mp validate` is green; `mp status` reflects M25–M28 (planned). The M24/v0.2.0 release gate stays in `release-checklist.md` (operational).

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
  ],
  "depends_on": ["25", "26"]          // optional: milestone IDs this depends on (G8)
}
```

**Strictness & AC coverage (G10):** the `full` profile defaults `workflow.gates.strictness = full`, so `mp validate` **errors** if any AC has no covering step. During the bulk import (Phase 3), temporarily `mp config set workflow.gates.strictness relaxed`, then restore `full` afterward. When decomposing, link steps to ACs via `mp step … --covers-ac AC-01,AC-02`.

### Field mapping (existing milestone → mp)
| Existing section | mp field |
|---|---|
| `## Goal` | `intent.outcome` + `problem.description` |
| `## Out of Scope` | `scope.out_of_scope` |
| Step deliverables / `## Goal` deliverables | `scope.in_scope` |
| `## Architecture Decisions (ADRs)` | `context.references` + `DESIGN.md` decision log |
| `Gate criteria` (header) | `acceptance_criteria[]` (one per observable criterion) |
| `## Step Index` / `## Steps` | generated by `mp milestone decompose` → `S1`, `S1.1`, … |
| `Depends on:` (header) / inter-milestone deps | `depends_on` (e.g. `["25","26","27"]`) — M28 depends on M25/M26/M27 |
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

**Per-phase checkpoint (do not skip):** at the end of every phase the executing agent must (1) re-verify that phase's AC checkboxes, (2) confirm the repo is still on the expected path — `git status` shows **only intended changes** and no collateral edits, (3) run `mp validate`, and (4) append a detailed entry to `mp-adoption-plan-log.md` (see §8 "Progress log"). **Do not advance to the next phase until the checkpoint passes.** If a step's actual behavior differs from what this plan predicts, STOP, log the divergence, and re-confirm with the user before continuing.

### Phase 0 — Prep
- [ ] Branch `chore/mp-adoption` exists and is current (already done).
- [ ] Record this plan + D1–D5 as a decision in `DESIGN.md` (adoption ADR).
- **AC:** decision logged; no plan files touched yet.
- **Checkpoint ☐:** `git status` shows only `DESIGN.md` (+ this plan/log) changed; append Phase 0 entry to log.

### Phase 1 — Bootstrap (with collision guard)
> The repo already has a `master-plan/` folder. mp defaults to that **same name**. You must force mp to use `.mp/` and **prove** the bespoke folder is never touched. This is the highest-risk step — verify before advancing.

- [ ] **1.0 Confirm the mechanism (known — verify it):** mp resolves the plan dir in priority order — **(1) `--plan-dir` CLI flag** > (2) `config.workflow.plan.location` > (3) `<root>/master-plan/` (default, **collides with the bespoke folder**). The config key lives in `master-plan/config.toml` (`[workflow]` → `plan.location`), settable via `mp config set workflow.plan.location .mp` — but that file lives *inside* the plan dir, so to bootstrap a non-default location you **must use the flag**, not a pre-edited file. Confirm `mp init --help` lists `--plan-dir` and record whether reality matches the docs (prime feedback).
- [ ] **1.1 Snapshot (baseline):** capture the existing plan dir so any change is detectable:
      ```bash
      git status --short                         # must be clean
      git ls-files master-plan | sort > /tmp/mp-pre-manifest.txt
      ```
- [ ] **1.2 Init into `.mp/`:** `mp init --profile full --from-repo --plan-dir .mp` (the flag is what avoids the `master-plan/` collision).
- [ ] **Expect (not collisions):** `--from-repo` scans the repo and **drafts a charter** (and may propose backlog candidates) — that's Phase 2/4 input, not a problem. `mp doctor` may *note* the existing `master-plan/` folder; with `--plan-dir .mp` that's benign — confirm it's a note, not an error.
- [ ] **1.3 Collision guard — VERIFY (hard gate):**
      - `.mp/` exists and contains mp artifacts.
      - `master-plan/` is **byte-identical** to pre-init: `git status` still clean **and** `git ls-files master-plan | sort` matches `/tmp/mp-pre-manifest.txt`.
      - **If mp wrote anything into `master-plan/`, or refused init because it "detected" an existing plan dir: ABORT.** Do not retry blindly. Log the exact behavior, move any stray mp files out, re-run with `--plan-dir .mp`, and re-confirm with the user before retrying.
- [ ] `mp doctor --format json` → healthy.
- [ ] `mp config show --format json` → confirms plan dir is `.mp/`; set `workflow.plan.in_repo` per **D6** via `mp config set workflow.plan.in_repo true` (verify the `full`-profile default first), and confirm `.gitignore` was **not** given the plan path.
- **AC:** `mp doctor` healthy; `.mp/` exists; `master-plan/` **provably unchanged** (guard passed).
- **Checkpoint ☐:** re-run the 1.3 assertions; `git status` clean; `mp validate`; append Phase 1 entry to log including the discovered location-forcing mechanism, mp's actual init behavior, and any doc mismatch.

### Phase 2 — Brief & charter (review the `--from-repo` draft)
> `mp init --from-repo` already drafted a charter (and may have proposed backlog candidates) from the repo. Don't build from zero — review and refine the draft against the source docs.

- [ ] `mp brief todo --format json` → fill topics from current `status.md` header (`mp brief edit`) → `mp brief done`.
- [ ] **Review the `--from-repo` charter draft** against `status.md` "Challenged Decisions": goals = ship v0.2.0 → perf pass; non-goals from challenged decisions. Refine via `mp interview checklist --checklist-type charter`.
- [ ] **Record the current release gate** in charter/brief: *"v0.2.0 is blocked by M24.13 manual smoke (human), tracked in `release-checklist.md`. mp forward plan starts at M25."*
- [ ] **Reconcile backlog:** if `--from-repo` proposed backlog candidates, dedupe against the explicit B-01…B-13 import in Phase 4.
- **AC:** `mp validate` green; brief `status: done`; charter reflects goals/non-goals + the M24 gate note.
- **Gotcha:** charter requires `brief done` first (error `B1`/`B3` otherwise).
- **Checkpoint ☐:** verify brief done + validate green; append Phase 2 entry (note what `--from-repo` drafted vs what you changed).

### Phase 3 — Import forward milestones (M25–M28)
> M24 is **not** imported (D2) — it's historical; the release gate lives in `release-checklist.md`.

- [ ] **Relax gates for bulk import:** `mp config set workflow.gates.strictness relaxed` (avoids G3/G4/G10 friction during import; restore at end of phase).
- [ ] For each of M25, M26, M27, M28: `mp milestone create --json @-` (§5 mapping) → `set-spec-status review` → `approve` → `decompose`.
- [ ] **Set dependencies:** M28 `depends_on = ["25","26","27"]` (G8) — via create JSON or `mp milestone` update.
- [ ] **Link steps → ACs:** when decomposing, `--covers-ac AC-xx` so every AC is covered (required once `strictness` is back to `full`).
- [ ] **Preserve per-step detail** (implementation checklists, file tables, test names) in step notes / `context.references` — mp's step fields are tighter than the current prose.
- [ ] **Restore:** `mp config set workflow.gates.strictness full`; then `mp validate`.
- **AC:** 4 milestones (M25–M28) with `spec_status: ready`; steps decomposed; ACs covered; M28 deps set.
- **Gotcha:** ≥2 out-of-scope items per milestone (error `G4`) — existing Out-of-Scope sections satisfy this.
- **Gotcha:** leave M25–M28 steps `pending` (planned, not implemented) — don't push them to `done`, or G6/G7 (evidence/verified) will block.
- **Checkpoint ☐:** verify 4 milestones ready + M28 deps; `git status` shows only `.mp/` changes; append Phase 3 entry — log **each** create/decompose, the strictness relax/restore, and every validation error (code + trigger + resolution).

### Phase 4 — Import backlog (`mp backlog`, B-xx IDs)
- [ ] For each backlog item (B-01, B-07, B-09, B-11, B-12, B-13): `mp backlog add --desc "…" --priority <medium|low> --source "backlog.md"`.
- [ ] mp assigns B-xx IDs itself — they may not match the old numbers. Record an **old→new ID map** in the log so nothing is lost.
- [ ] Dedupe against any candidates `--from-repo` proposed in Phase 2.
- [ ] `mp validate`
- **AC:** every backlog item appears in `mp list backlog`; content matches `master-plan/backlog.md`.
- **Checkpoint ☐:** cross-check every old B-xx maps to an mp backlog item (old→new table in log); append Phase 4 entry.

### Phase 5 — Cutover ⚠️ confirm D2/D3 with user first
- [ ] **Pause and confirm D2 (history) + D3 (location) with the user before any irreversible action.**
- [ ] Move old `master-plan/*.md` → `master-plan-archive/` (preserve history; delete later when confident).
- [ ] Rename `.mp/` → `master-plan/` **or** keep `.mp/` and set `config.workflow.plan.location`.
- [ ] Move operational docs (`release-checklist.md`, `smoke-checklist.md`, `coverage-audit.md`, `review-template.md`) out of the plan dir to `docs/operational/` (or repo root).
- [ ] Rewrite `AGENTS.md` "Planning Rules": replace hand-edit rules with the mp session-start sequence (see §8) and "all plan I/O via `mp`; never hand-edit plan files".
  - **Scope:** edit ONLY the "Planning Rules" section. Do **not** touch Git Rules, Bubble Tea Rules, Testing Rules, or Verification Commands.
  - **Dual AGENTS.md:** mp manages its own `<plan-dir>/AGENTS.md` — leave that alone; only the **root** `AGENTS.md` is edited here.
- [ ] Add adoption ADR to `DESIGN.md` decision log.
- [ ] `mp validate`
- **AC:** `master-plan/` is mp-managed and the single source of truth; `AGENTS.md` reflects mp workflow; old docs archived.
- **Rollback:** restore `master-plan/` from `master-plan-archive/`; revert `AGENTS.md`.
- **Checkpoint ☐:** confirm mp reads from the new location (`mp config show` + `mp status`); append Phase 5 entry — log the rename/cutover behavior in detail (this is high-value mp feedback).

### Phase 6 — Merge
- [ ] Single PR `chore/mp-adoption` → `main`.
- [ ] CI green (`go test -race ./...`, `go vet ./...`, `make lint`).
- [ ] `mp doctor` + `mp status` green on `main` after merge.
- **AC:** merged; plan reads correctly via `mp status`.
- **Checkpoint ☐:** append the **final outcome summary** to the log (success/partial/failed, what mp did well, top friction points) and confirm the `[IMPROVE]` list is complete.

---

## 7. Risks & mitigations

| Risk | Mitigation |
|---|---|
| Narrative loss (mp schema is tighter than current prose) | Preserve why-now/ADR/rollback in `context.references` and `DESIGN.md`; don't discard. |
| "Spec before code" fights imported milestones | Relax `strictness` during bulk import (Phase 3); leave M25–M28 steps `pending`. M24 stays out of mp entirely (D2). |
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
5. **Verify, then advance.** After each phase: re-check its ACs, confirm `git status` shows only intended changes, then append to `mp-adoption-plan-log.md`. If reality ≠ this plan, **stop** and re-confirm with the user.
6. **Never let mp touch the bespoke `master-plan/` during Phases 1–4.** The Phase 1 collision guard (1.3) is a hard gate — if it fails, abort and log.

### mp gotchas (CLI ≠ docs)
| Doc says | Use instead |
|---|---|
| `mp idea add` | `mp idea create` |
| `mp interview --type X` | `--checklist-type X` |
| `mp plan show --all` | `mp plan show` (no flag) |
| `mp step done --evidence "x"` | **verify + log**: skill says no flag; `AGENT-PLAYBOOK` shows `step done <m> <s> --evidence "…"` |
| `mp milestone criterion pass --evidence "x"` | `--evidence` is **positional**, not a flag |
| backlog import | use **`mp backlog add`** (B-xx IDs), not tracks/ideas |

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
- `mp status` shows M25–M28 (planned); the M24/v0.2.0 gate remains in `release-checklist.md`.
- `mp validate` green.
- This file (`mp-adoption-plan.md`) **deleted** (the log is retained).

### Progress log — `mp-adoption-plan-log.md` (required)

Maintain a companion file `mp-adoption-plan-log.md` (repo root, next to this file) throughout execution. This is **field-test feedback for the mp project itself** — the user will import it into the mp repo to drive improvements. Be exhaustive; detail is the deliverable, not a burden.

A template with per-phase sections is already created at `mp-adoption-plan-log.md` — fill it in as you go (do not write it all at the end).

For **every phase** (and any notable event in between), record:
- **Exact commands run** + verbatim output (or a faithful summary if huge).
- **Actual vs expected** — did mp behave as this plan / the docs predicted? Quote any divergence.
- **CLI ≠ docs discoveries** — wrong command names, missing flags, misleading errors, confusing output format.
- **Validation errors** — every code hit (`B1`/`G1`/`G4`/…) with the trigger and how it was resolved.
- **Friction & confusion points** — anything that made you pause, guess, or read mp source.
- **Time/effort** — rough per-phase effort and where it went.
- **Improvement candidates** — concrete suggestions for mp maintainers, each tagged `[IMPROVE]`.
- **Artifact diffs** — which files mp created/changed (paths), so the cutover is auditable.

End the log with a **final outcome summary**: success / partial / failed, what mp did well, and the top friction points. Do **not** delete this log at the end — it ships as feedback, unlike `mp-adoption-plan.md`.

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
- [ ] M25–M28 imported as mp milestones with specs, steps, ACs, deps (M24 historical — not imported)
- [ ] Backlog B-01…B-13 → `mp backlog` (B-xx IDs)
- [ ] `master-plan/` is mp-managed and the single source of truth
- [ ] Operational checklists moved out of plan dir
- [ ] `AGENTS.md` Planning Rules rewritten for mp workflow
- [ ] Adoption ADR in `DESIGN.md`
- [ ] `mp validate` green; CI green; merged to `main`
- [ ] `mp-adoption-plan-log.md` complete and **retained** as mp feedback (not deleted)
- [ ] **This plan file deleted** (the log stays)
