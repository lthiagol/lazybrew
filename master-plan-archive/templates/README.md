# Master Plan Templates

Templates and conventions for milestones and status tracking in this repository.

---

## Files

| Template | Use when |
|---|---|
| [milestone.md](milestone.md) | Creating a new milestone under `milestones/` |
| [status.md](status.md) | Bootstrapping or restructuring `master-plan/status.md` |
| [step.md](step.md) | Adding a single step to an existing milestone (copy-paste block) |

Related (not in this folder):

- [../review-template.md](../review-template.md) — project-agnostic code/plan audits
- [../backlog.md](../backlog.md) — deferred scope (B-XX IDs)

---

## Conventions

### Single source of truth

| Artifact | Role |
|---|---|
| **`status.md`** | Portfolio view: what is done, partial, blocked, execution order |
| **`milestones/NN-*.md`** | Execution detail: steps, acceptance criteria, tests |
| **`backlog.md`** | Explicitly deferred work |

**Rule:** Milestone header `Status` must match the row in `status.md`. If they disagree, fix both in the same change.

### Status markers

Use in `status.md` progress block and milestone headers:

| Marker | Meaning | When to use |
|---|---|---|
| `[X]` / ✅ Complete | Done | All DoD checkboxes verified with evidence |
| `[~]` / ⚠️ Partial | In progress or gaps remain | Open DoD items or known gaps documented in **Remaining** |
| `[ ]` / 🔜 Planned | Not started | |
| 🚫 Blocked | Cannot proceed | Add **Blockers** line in status.md |

**Never mark ✅ Complete** if:

- Definition of Done has unchecked items
- Tests listed in Test Plan are missing
- A critical/high review finding in scope is still open

### Milestone sizing

| Label | Duration | Guidance |
|---|---|---|
| **S** | < 1 day | ≤ 3 steps |
| **M** | 1–3 days | ≤ 8 steps |
| **L** | 3–7 days | Split into **phases** if > 10 steps |
| **XL** | > 1 week | Must decompose into multiple milestones |

### Step sizing

| Label | Duration | Guidance |
|---|---|---|
| **S** | < 2 hours | One focused change + tests |
| **M** | 2 hours – 1 day | One feature slice |
| **L** | 1–2 days | Split if possible |

Each step must have:

1. Acceptance criteria (observable)
2. Tests in the **same step** (not deferred to milestone end)
3. Files table (Create / Modify / Delete)

### Step numbering

- Format: `NN.M` — milestone number + sequential step (e.g. `19.6`)
- Optional substeps: `19.6a` only for documentation; prefer new number
- Step `0` allowed for prerequisites moved from other milestones (e.g. `19.0`)

### Phases (large milestones)

For **L** or **XL** milestones, group steps in the Step Index:

```markdown
| Phase | Steps | Theme |
| Phase A — Data truth | 20.1, 20.6 | … |
```

Execute phases **in order** unless status.md parallel tracks say otherwise.

### ADR IDs

- Format: `D{NN}-{n}` — e.g. `D19-1`
- Record in milestone **and** `DESIGN.md` decision log when it affects architecture

### Backlog IDs

- Format: `B-{nn}` in [backlog.md](../backlog.md)
- Reference from **Out of Scope** — never silently drop work

### Parallel tracks

Define in `status.md` only (not every milestone). Milestone header may reference:

```markdown
> **Parallel track:** B (see status.md)
```

### Gate criteria

Milestone header **Gate criteria** = what downstream milestones need before they start.

Post-Milestone Gate section = verification checklist before marking complete.

---

## Workflow

### Adding a new milestone

1. Copy [milestone.md](milestone.md) → `milestones/NN-short-name.md`
2. Fill header, Goal, Step Index, and all steps (use [step.md](step.md) for each)
3. Add row to `status.md` Overall Progress + Milestone Index
4. Link dependencies in **Depends on** / **Enables**
5. If deferring scope → [backlog.md](../backlog.md)

### Updating progress

1. Check off step acceptance criteria in milestone file
2. When all steps done → verify Definition of Done
3. Update milestone header status
4. Update `status.md` marker and **Current phase** / **Execution entry point**
5. Append decision log row if architectural choice was made

### Completing a milestone

- [ ] All steps complete
- [ ] Test Plan tests exist and pass
- [ ] `status.md` synced
- [ ] User-visible docs updated (if applicable)
- [ ] Post-Milestone Gate satisfied

---

## Naming

| Item | Pattern | Example |
|---|---|---|
| Milestone file | `NN-kebab-case.md` | `19-bubble-tea-concurrency-and-task-manager.md` |
| Review doc | `architecture-review-YYYY-MM-DD.md` | |
| Planning challenge | `planning-challenge-YYYY-MM-DD.md` | |

---

## Legacy milestones (M1–M17 in lazybrew)

Older milestones may predate this template. See [milestone-legacy-index.md](../archive/milestone-legacy-index.md).

| Action | When |
|---|---|
| Read legacy steps | Historical context only |
| Execute open work | Follow **Active Work Routing** → M18+ |
| Refine legacy file | Header + routing section; don't rewrite all steps |
| Full template migration | Only pending milestones (e.g. M17) |

---

## Anti-patterns

| Don't | Do instead |
|---|---|
| "Implement feature X" without acceptance criteria | List observable outcomes |
| Tests only in DoD at milestone end | Tests per step |
| ✅ Complete with open gaps | ⚠️ Partial + **Remaining** section |
| New scope in a step without backlog/milestone | Add to backlog or new milestone |
| Duplicate architecture in every milestone | Link to `DESIGN.md` |
| Two sources of truth for status | Sync header ↔ status.md |
