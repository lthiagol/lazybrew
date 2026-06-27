# mp Adoption Log — lazybrew

> **Purpose:** Running progress + field-test feedback for the **mp** (Master Plan) CLI, captured during lazybrew's adoption migration. The user will import this file into the **mp repository** to analyze real-world usage and drive improvements.
>
> **Filled in by:** the executing agent, **as work happens** (not retroactively).
> **Companion to:** `mp-adoption-plan.md` (the plan). This log is **retained**; the plan is deleted at the end.
> **Rule:** detail is the deliverable. When in doubt, write more — commands, output, surprises, confusion.

---

## Context (fill once at start)

| Field | Value |
|---|---|
| mp version | _(run `mp --version`)_ |
| mp path | `/home/thiago/.agents/master-plan/bin/mp` |
| OS / shell | _ |
| Repo | `lazybrew` @ branch `chore/mp-adoption` |
| Plan profile | `full` (per D1) |
| Plan-dir strategy | `.mp/` parallel, cutover later (per D3) |
| Session start | _(date/time)_ |
| Executing agent | _ |

### Pre-flight baseline
- `master-plan/` file manifest captured at: _(path, e.g. `/tmp/mp-pre-manifest.txt`)_
- `git status` clean before Phase 1: ☐ yes / ☐ no

---

## Running `[IMPROVE]` candidates

Append every concrete mp improvement idea here as you find it (also reference it from the phase where it occurred). One per row.

| # | Tag | Area | Observation | Suggested improvement | Where seen |
|---|---|---|---|---|---|
| I-01 | `[IMPROVE]` | _cli/docs_ | _e.g. "docs say `mp idea add` but real cmd is `mp idea create`"_ | _ | Phase _ |
| I-02 | | | | | |

---

## Phase 0 — Prep
- **Started / ended:** _
- **Commands run:** _(exact)_
- **What happened:** _
- **Actual vs expected:** _
- **Friction / surprises:** _
- **Checkpoint passed:** ☐ · `git status` clean: ☐
- **`[IMPROVE]` items raised:** _

## Phase 1 — Bootstrap (collision guard)
- **Started / ended:** _
- **1.0 Discovery — how is plan-dir location set?** _(env var? flag? config file? did `mp init --help` document it?)_ **Be precise — this is the #1 mp feedback item.**
  - Mechanism that worked: _
  - What the docs claimed: _
  - Mismatch (if any): _
- **Commands run:** _(exact `mp init …` invocation)_
- **1.3 Collision guard result:**
  - `.mp/` created with mp artifacts: ☐ yes / ☐ no
  - `master-plan/` byte-identical to baseline (`git ls-files` matches manifest, `git status` clean): ☐ yes / ☐ no
  - If mp wrote into `master-plan/` or refused init: _describe exactly what it did, verbatim error, and how you recovered_
- **mp doctor output (summary):** _
- **mp config show — confirmed plan dir:** _
- **Gitignore decision for `.mp/`:** _ignored / tracked — rationale: _
- **Friction / surprises:** _
- **Checkpoint passed:** ☐
- **`[IMPROVE]` items raised:** _ _(location/collision handling is expected here)_

## Phase 2 — Brief & charter
- **Started / ended:** _
- **Commands run:** _
- **Brief topics filled — source mapping used:** _(from `status.md` header)_
- **Charter goals/non-goals source:** _(from Challenged Decisions)_
- **Validation errors hit (B1/B3…):** _trigger + resolution_
- **Friction / surprises:** _
- **Checkpoint passed:** ☐ · `mp validate` green: ☐
- **`[IMPROVE]` items raised:** _

## Phase 3 — Import milestones (M24–M28)
- **Started / ended:** _
- **Per-milestone log (repeat block):**

  **M24 — Smoke Test Fixes**
  - `mp milestone create` JSON shape used: _
  - Approve → decompose result: _
  - Step import (mark done except S24.13): _did mp allow steps to be marked done at import? how?_
  - Errors hit (G1 spec-before-code on a code-complete milestone? G4 out-of-scope?): _trigger + resolution_

  **M25 — Outdated** _(same fields)_
  **M26 — Diagnostics** _(...)_
  **M27 — Taps** _(...)_
  **M28 — Tiered Refresh** _(...)_
- **Schema mapping friction** (§5 mapping — what fit cleanly, what didn't): _
- **Validation errors hit:** _
- **Friction / surprises:** _e.g. where did why-now / ADRs / rollback prose have to go?_
- **Checkpoint passed:** ☐ · 5 milestones `ready`: ☐
- **`[IMPROVE]` items raised:** _

## Phase 4 — Import backlog
- **Started / ended:** _
- **Commands run:** _
- **Tracks created (B-01, B-07):** _
- **Ideas created (B-09, B-11, B-12, B-13):** _
- **Gotcha check — `mp idea create` vs `add`:** _did it bite?_
- **Every backlog item represented:** ☐
- **Checkpoint passed:** ☐
- **`[IMPROVE]` items raised:** _

## Phase 5 — Cutover
- **User confirmed D2 (history) + D3 (location) before starting:** ☐ yes / date _
- **Commands run:** _(move/reename/relocate)_
- **Rename `.mp/` → `master-plan/` behavior:** _did mp re-detect correctly after rename? did config need updating?_
- **Operational docs relocated to:** _
- **`AGENTS.md` Planning Rules rewrite:** _summary of new content_
- **Adoption ADR added to `DESIGN.md`:** ☐
- **Friction / surprises:** _ _(high-value: how mp handles a location/rename cutover)_
- **Checkpoint passed:** ☐ · `mp status` reads from new location: ☐
- **`[IMPROVE]` items raised:** _

## Phase 6 — Merge
- **Started / ended:** _
- **CI result:** _(`go test -race`, `go vet`, `make lint`)_
- **`mp doctor` + `mp status` on `main`:** _
- **Checkpoint passed:** ☐
- **`[IMPROVE]` items raised:** _

---

## Final outcome summary

- **Result:** ☐ success / ☐ partial / ☐ failed
- **Total elapsed:** _
- **What mp did well:** _
- **Top 3 friction points:** _
  1. _
  2. _
  3. _
- **Artifacts mp now owns** _(paths)_: _
- **Open issues handed back to mp maintainers:** _(link the `[IMPROVE]` rows above)_
- **Would we adopt mp again on a similar project?** _yes/no — why_

---

### Logging checklist (for the executing agent)
- [ ] Filled in **as work happened**, not at the end.
- [ ] Every command captured (or faithfully summarized).
- [ ] Every validation error logged with code + trigger + resolution.
- [ ] Every CLI≠docs discovery has a `[IMPROVE]` row.
- [ ] Final outcome summary completed.
- [ ] File **retained** (do not delete — ships as mp feedback).
