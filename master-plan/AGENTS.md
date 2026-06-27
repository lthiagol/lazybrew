# Master Plan — Agent Instructions

> **Rust readiness:** [docs/AGENT-READINESS.md](../docs/AGENT-READINESS.md) — command matrix.  
> **Repo session start:** [../AGENTS.md](../AGENTS.md) — read meta plan and propose next move.

This project uses **spec-driven development**. The `master-plan/` directory is the
single source of truth for what to build, in what order, and how to verify it.

**You must follow these rules when planning or implementing work in this project.**

---

## 1. Non-Negotiable Rules

1. **Never read or edit files under `master-plan/` directly.** Use the `mp` CLI for
   all reads and writes.
2. **Spec before code.** Do not modify application source code until the relevant
   milestone has `spec_status: ready` (approved).
3. **Two-phase milestones.** Phase 1 is the spec (what/why). Phase 2 is the
   implementation plan (how), created only after spec approval.
4. **Reads use JSON.** `mp <command> --format json`
5. **User-facing output uses human format** or a natural-language summary of JSON.
6. **After every write, validate.** `mp validate`
7. **Plan-only mode.** When asked to plan without implementing, stop after `mp`
   writes. Do not touch application code.
8. **Execution mode.** Check `mp execution status`. In `planning` mode, do not implement
   unless the user directs a specific milestone/step. In `autonomous` mode, run the
   `next-step` loop until blocked — then escalate. Never change spec without pausing.

---

## 1a. Plan zone vs code zone

| Zone | What | Rules |
|------|------|-------|
| **Plan** | `master-plan/` | All reads/writes via `mp`. Never hand-edit plan files. |
| **Code** | Application source (`src/`, `tests/`, configs) | Search, read, implement. Harness tools OK (ripgrep, LSP). |

Use code zone to learn **current behavior** during interviews. Record findings in spec
fields (`context.references`, scenarios, ACs) via `mp milestone create --json @-` — not
by editing TOML directly.

See toolkit `docs/BROWNFIELD.md` for greenfield vs brownfield routing.

---

## 2. Tooling

| Tool | Location |
|------|----------|
| CLI | `mp` (Master Plan CLI — `~/.agents/master-plan/bin/mp`) |
| Plan directory | `./master-plan/` |
| Spec reference | Project docs or `mp plan show --format json` |

If `mp` is not found, tell the user to install the master-plan toolkit. Do not fall
back to editing plan files by hand.

### Intake routing (which lane?)

```text
Too vague / later?     → idea (P1.6) or conversation note
Small fix / polish?    → track (bugfix / tweak)     ✅ works today
Defer scope formally?  → backlog (P3)
Feature / behavior?    → milestone (full spec)
Prod emergency?        → track bugfix — see docs/EMERGENCY.md
```

---

## 3. Workflows

### 3.0 Project brief (first session after init)

Use when `brief.status = in_progress` or `planning_phase = brief`.

```text
1. mp brief todo --format json
2. Ask 1–2 questions per pending topic; user brain-dumps freely
3. mp brief edit T01 --body "..."  (repeat for each topic)
4. mp brief add ...                (optional custom topics)
5. mp brief list --format json     # established context
6. mp brief done
7. → charter interview (§3.1) or milestone planning
```

Optional: `mp interview checklist --checklist-type brief --format json` for suggested question rounds.

Do not start milestone specs until `mp brief done` unless the user explicitly skips the brief.

---

### 3.1a Brownfield change (behavior change on existing code)

Use when changing existing behavior — not a greenfield subsystem.

**Small** → track (`§3.7`). **Large** → milestone with explicit before/after.

**Today (pre-P4):**

```text
1. mp doctor --format json
2. Code zone: locate current behavior (files, tests)
3. mp interview checklist --checklist-type milestone --format json
4. Spec must state: what exists today, what changes, what stays the same
5. context.references → source paths as evidence
6. mp milestone create --json @-  (change_kind: greenfield is OK until P4)
7. Approve → decompose → implement
```

**P4 (delta milestones):** `change_kind: delta`, `delta.domain`, ADDED/MODIFIED/REMOVED;
`mp specs show <domain>` before create; `mp brownfield scan` optional assist.

---

### 3.1 Plan a new feature or bug (interview mode)

Use when the user asks to plan, groom, or spec work without implementing yet.

```text
1. mp interview checklist --type milestone --format json
2. Ask the user 2–4 questions per round (skip topics already answered)
3. Propose defaults from codebase analysis; user confirms or corrects
4. mp milestone create --json @-   (spec fields only — no WPs/steps yet)
5. mp milestone set-spec-status <id> review
6. Summarize the spec in natural language
7. On user approval → mp milestone approve <id>
8. mp validate
9. Stop. Do not write application code.
```

**Interview topics — Phase 1 (spec):**
1. Intent — outcome, problem
2. Scenarios — given/when/then, priorities (P1–P3), edge cases
3. Requirements — FR-XX functional requirements, NC-XX needs clarification
4. Success — measurable SSC-XX criteria, assumptions
5. Interface — APIs, CLI, config, entities (if applicable)
6. Scope — in-scope, out-of-scope (minimum 2 exclusions)
7. Acceptance — AC-XX verification methods
8. Design — technical choices
9. Sequencing — dependencies, effort, risk
10. Gaps — open questions, risks (resolve before approval)

**Interview topics — Phase 2 (implementation plan, after approval):**
1. Technical context — language, deps, testing, constraints
2. Decomposition — work packages, steps, files, tests, rollback

---

### 3.2 Decompose into implementation plan (phase 2)

Use only after the user approves the spec (`spec_status: ready`).

Triggered by: *"break this into steps"*, *"plan implementation for M03"*, *"decompose M03"*.

```text
1. mp groom milestone <id> --format json
2. mp milestone decompose <id>
3. mp plan gaps <id> --format json
4. mp wp add <id> ... (work package grouping)
5. mp step add <id> --wp WP1 ... (steps S1, S2, … with files, tests, done-when)
6. mp validate
7. Present the implementation plan to the user for confirmation
```

Do not start coding until the user confirms the implementation plan (or explicitly
says to proceed).

---

### 3.2a Split a step (step too large)

```text
1. mp list steps --milestone <id> --format json
2. mp step split <id> <step> --json @-   # e.g. S3 → S3, S3.1, S3.2
3. mp validate
```

---

### 3.2b Challenge a plan (stress-test)

Use when the user wants to review, challenge, or find gaps in a spec or implementation plan.

```text
1. mp show milestone <id> --format json
2. mp challenge start <id> --scope plan    # or spec | full
3. mp challenge audit <id> --format json
4. mp challenge list <id> --format json
5. Discuss findings with user
6. mp challenge resolve <id> F-01 --action update-step --payload ...
7. mp validate
8. mp challenge done <id>
```

---

### 3.3 Execute work

Use when implementing approved, decomposed milestones.  
**State transitions:** toolkit `docs/AGENT-PLAYBOOK.md` (start → in-progress → done → complete).

```text
1. mp execution status --format json
2. mp next-step --format json
3. mp milestone set-status <id> in-progress     # first step on this milestone only
4. mp step set-status <id> <step> in-progress   # BEFORE code changes
5. Implement application code (outside master-plan/)
6. mp step done <id> <step> --evidence "..."
7. mp validate
8. Repeat 2–7 until all steps done
9. mp milestone criterion pass <id> <ac-id> --evidence "..."   # each AC
10. mp milestone complete <id> --evidence "..."
11. mp validate
```

**Blocked?** `mp milestone block <id> --reason "..."` → `mp execution pause` → escalate to user.  
**Resume:** `mp milestone unblock <id>`.

### 3.3a Execution handoff (autonomous mode)

When the user says “go execute”, “work through the plan”, or similar:

```text
1. mp execution check --format json
2. Present execution_ready milestones and blockers
3. User confirms → mp execution handoff
4. Loop: next-step → step in-progress → code → step done → validate
5. On ambiguity, validate fail, or new scope → mp execution pause + escalate
```

See toolkit `docs/EXECUTION-MODES.md` and `docs/WALKTHROUGH.md`.

---

### 3.4 Query and report

Use when the user asks for status, summaries, or what's next.

| User intent | Command |
|-------------|---------|
| Overall status | `mp status --format json` → summarize |
| All milestones (compact) | `mp list milestones --format human` |
| Done / pending / in progress / partial | `mp list milestones --filter <preset>` |
| Needs grooming | `mp list milestones --filter grooming` |
| Pending milestones | `mp list milestones --spec-status ready,interview --format json` |
| What's next? | `mp path --format json` or `mp next-step --format json` |
| Full work queue | `mp path --format human` |
| Do M4 before M3 | `mp path pin 04 --before 03` |
| What should we do with M03? | `mp groom milestone 03 --format json` |
| Steps for a milestone | `mp list steps --milestone <id> --format json` |
| Park idea for later | `mp idea create ...` |
| Small bugfix | `mp track add bugfix ...` |
| Show one milestone | `mp show milestone <id> --format human` (pass through) |
| Archived items | `mp list archived --format json` |
| Full validation | `mp validate --format json` |

> For grooming, challenge, and step filters use the commands in
> [docs/AGENT-READINESS.md](../docs/AGENT-READINESS.md).

---

Use when the user defers a topic **without** asking for a plan or fix:

- “Let’s handle this later”
- “Park this idea”
- “Remind me about the installer approach”

```text
mp idea create --title "App installer design" --body "..." --tags installer
mp validate
```

Do **not** create a milestone, track item, or backlog entry unless the user wants formal deferred scope or actionable work.

**Later:**
```text
mp idea list --format json
mp idea promote ID-01 --to-milestone    # or --to-backlog | --to-track bugfix
```

---

### 3.6 Defer scope (backlog)

Use during **grooming** when scope is formally deferred from a milestone or charter:

```text
mp backlog add --desc "..." --priority medium --source planning
mp validate
```

---

### 3.7 Track item (lightweight bugfix/tweak)

Use for **small, independent fixes** that do not need a full milestone spec.

**Choose track when:**
- Effort is hours, not days
- No new feature behavior — correctness, polish, or small improvement
- Work can be done and verified in one pass
- Item does not need scenarios, FR-XX, or design decisions

**Choose milestone when:**
- New feature or significant behavior change
- Multiple work packages or cross-cutting design
- Needs interview, spec approval, and acceptance criteria

```text
1. mp interview checklist --checklist-type track-item --kind bugfix --format json
2. Ask 1–3 quick questions if fields missing
3. mp track add bugfix --title "..." --problem "..." --verification "..."
4. mp track start bugfix BF-01              # → in-progress
5. Implement fix
6. mp track done bugfix BF-01 --evidence "..."
7. mp validate
```

**Blocked?** Tell the user; do not call `track done`. **Cancel?** `mp track cancel bugfix BF-01`.

If a track item grows large: `mp track promote bugfix BF-03 --to-milestone`

---

### 3.8 Bootstrap (first time only)

```text
mp init
mp doctor
mp brief todo --format json    # first agent task with user
# ... fill brief (see §3.0) ...
mp brief done
```

If charter is empty after the brief, run a charter interview
(`mp interview checklist --checklist-type charter`) before planning milestones.

Use `mp brief list --format json` as context — do not re-ask what the brief already covers.

---

## 4. Spec Lifecycle Gates

| Gate | Rule |
|------|------|
| No code before ready | `spec_status` must be `ready` before `in-progress` |
| No open questions at ready | Resolve all `Q-XX` before approval |
| Min 2 out-of-scope items | Required before `review` |
| Min 1 acceptance criterion | Required before `review` |
| No impl plan before ready | WPs/steps only after spec approval |
| Verified before done | All `AC-XX` must pass before `complete` |

If `mp validate` fails, fix via `mp` commands. Do not patch files manually.

---

## 5. Output Conventions

| Audience | Format | Example |
|----------|--------|---------|
| Agent reads | `json` | `mp show milestone 03 --format json` |
| User sees rendered plan | `human` | `mp show milestone 03 --format human` |
| Debug source | `raw` | `mp show milestone 03 --format raw` |

- **Talking to the user:** summarize JSON in clear prose, or pass through `--format human`.
- **Never dump raw TOML** unless the user explicitly asks.

---

## 6. Triggers

Activate this workflow when the user:

- Mentions master plan, milestone, roadmap, spec, backlog, or ideas
- Mentions “later”, “park this”, “remind me about”, or defers a topic without planning it
- Uses `/mp` or `/master-planner`
- Asks "what's next?", "plan X", "break down X", or "what's the status?"

---

## 7. References

- **Agent playbook (state updates):** toolkit `docs/AGENT-PLAYBOOK.md`
- Full spec model: toolkit `docs/SPEC.md`
- Command reference: toolkit `docs/MP-COMMANDS.md`
- Walkthrough example: toolkit `docs/WALKTHROUGH.md`
- Global skill: `~/.agents/skills/master-planner/SKILL.md`
