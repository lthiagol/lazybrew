# Milestone NN — Short Title

> **Status:** 🔜 Planned | ⚠️ In Progress | ✅ Complete | 🚫 Blocked  
> **Size estimate:** S (<1 day) | M (1–3 days) | L (3–7 days) | XL (>1 week)  
> **Depends on:** M__ ✅, M__.N (step) — list milestones and specific step gates  
> **Enables:** M__ — what this unblocks  
> **Parallel track:** A | B | C | — (optional; see [status.md](../status.md))  
> **Gate criteria:** One sentence — condition for downstream work to start safely  
> **Remaining:** _(optional; required when Status is ⚠️ Partial)_ — bullet list of open gaps

<!-- Optional links -->
<!-- See [planning-challenge-YYYY-MM-DD.md](../planning-challenge-YYYY-MM-DD.md) -->
<!-- See [backlog.md](../backlog.md) B-XX for deferred items -->

---

## Goal

One paragraph describing the **outcome** for users or developers. State what is true when this milestone is done — not a task list.

---

## Why Now

2–4 sentences: risk mitigated, dependency satisfied, why not deferred.

---

## Challenged Assumptions

<!-- Optional but recommended for L/XL milestones or contested design -->

| Assumption | Challenge | Decision |
|---|---|---|
| | | |

---

## Out of Scope

Explicit boundaries. Every item should point elsewhere.

- **Item** — deferred to [backlog.md](../backlog.md) B-XX / milestone M__ / post-v1

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|---|---|---|---|
| DNN-1 | | | |

Copy stable ADRs to `DESIGN.md` decision log when merged.

---

## Phases

<!-- Required for L/XL milestones. Optional for M. Delete section if not used. -->

Execute phases **in order** unless [status.md](../status.md) parallel tracks document otherwise.

| Phase | Steps | Theme | Phase gate |
|---|---|---|---|
| **A —** | NN.1, NN.2 | | |
| **B —** | NN.3 | | |

---

## Step Index

| Step | Title | Size | Depends | Deliverable |
|---|---|---|---|---|
| NN.1 | | S/M/L | — | |
| NN.2 | | | NN.1 | |

---

## Steps

<!-- Copy one block per step from step.md, or inline below -->

### NN.1 — Step Title

**Size:** S | M | L  
**Phase:** A _(optional)_  
**Track:** A _(optional)_  
**Depends on:** —  
**Blocks:** NN.2

**Context:** Why this step exists (1–3 sentences).

**Preconditions:**
- [ ] Dependency gate met (e.g. M18.5 DESIGN ADR merged)
- [ ] `make test` / project test command passes

**Implementation checklist:**
1. Concrete action
2. Concrete action
3. ...

**Files:**

| File | Action |
|---|---|
| `path/to/file` | Create / Modify / Delete |

**Acceptance criteria:**
- [ ] Observable behavior or artifact
- [ ] No regression in ___

**Tests (same change set as implementation):**
- [ ] `TestName` — what it proves (unit / integration / e2e)

**Out of scope for this step:**
- ...

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| | |

**Rollback:** How to revert if this step fails mid-way.

---

### NN.2 — Next Step Title

<!-- Repeat structure for each step -->

---

## Test Plan (milestone-level)

Consolidated view — must match tests listed in steps.

| Test | Tier | Step | Proves |
|---|---|---|---|
| | unit / integration / e2e / static | NN.x | |

**Verification commands:**

```bash
# Project-specific — example:
make test
go test -race ./...
# make test-integration
```

---

## Definition of Done

- [ ] All steps NN.1–NN._N_ complete; acceptance criteria checked
- [ ] Every Test Plan row has an existing passing test (or documented manual check)
- [ ] Verification commands pass (including race detector if applicable)
- [ ] User-visible or architectural docs updated (README, DESIGN, AGENTS — if applicable)
- [ ] [status.md](../status.md) updated; **this file header Status matches**
- [ ] No open **critical/high** findings in this milestone's scope
- [ ] **Remaining** section empty or removed (if marking ✅ Complete)

---

## Post-Milestone Gate

Before starting **Enables** milestones, confirm:

- [ ] Header **Gate criteria** satisfied
- [ ] [review-template.md](../review-template.md) Dimension 8 (plan hygiene) for this milestone
- [ ] Smoke or manual checklist signed off (if applicable — link path)

---

## Rollback Plan

<!-- Optional; recommended for high-risk milestones (concurrency, migrations) -->

If integration fails mid-milestone:

1. Steps safe to keep independently: NN.0, ...
2. Revert order: ...
3. Minimum hotfix acceptable for ship: ...

---

## Version History

| Date | Change |
|---|---|
| YYYY-MM-DD | Created from [templates/milestone.md](../templates/milestone.md) |
