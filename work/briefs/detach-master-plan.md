# Slice: detach master-plan / mp from this repo

> **Contract.** This brief wins over any plan, note, or prior convention where they
> differ. If you find something in the repo that contradicts the brief, follow the
> brief and report the contradiction — do not stop to ask.

## Goal

`lazybrew` no longer uses the `mp` CLI or the `master-plan/` directory as its
planning source of truth. Planning moves to the **herdr lane flow** that lives in
`work/` (briefs → runner lane → reviewer lane → receipt → state row).

The `master-plan/` directory must **stop being part of this repo** while the files
**stay on this machine**. Nothing under `master-plan/` is deleted, edited, or
moved — it simply stops being tracked by git and stops being the plan.

## Context (already established — do not re-derive)

- `master-plan/` = 33 tracked files, ~332K of `mp` JSON. 16 milestones, all
  `complete`, zero open findings, zero ideas, zero active tracks.
- The plan directory is **frozen**: `mp status` shows no actionable steps
  (`mp next` → "No actionable steps found"). There is no in-flight work.
- Residue that must be adopted before the directory is untracked:
  - `B-01` — edit a header in `master-plan-bkp/` → **DEAD**: `master-plan-bkp/`
    exists nowhere on this machine. Drop it, with that reason.
  - `B-02` — apply context timeouts to `fetchPanelData`, `fetchStatusData`,
    `fetchSelectedSearchInfo`, `executeSearch` (residual from M13 F-02/F-07).
    Priority medium. → adopt.
  - `B-03` — status dashboard outdated count stays 0 until first refresh
    (M09 F-02 residual). Priority low. → adopt.
  - `B-04` — chunk `brew tap-info --json` batches when tap count risks ARG_MAX
    (M11 F-01 residual). Priority low. → adopt.
  - `B-05` — optional full interactive human smoke before stable release tag.
    Human gate. → adopt as the tracked smoke checklist, not as a milestone.
  - M13 — execution done, independent review never passed (all 7 findings fixed).
    → adopt as a verification slice (verify the fix by running the Deps tests).
  - `AGENTS.md` release checklist — M22.1b (CI green on push/PR) and M22.4
    (release checklist signed off) stay open in `AGENTS.md`. Not plan work.
- `master-plan/smoke-checklist.md` (28 items, all `[x]`) is a **release gate
  artifact** and must survive in the repo. It is currently tracked only inside
  `master-plan/`.

## Acceptance criteria

One commit per criterion, in order. Stage by name — never `git add -A`, never
`git add .`.

### AC-01 — `AGENTS.md` stops driving work from `mp`

Rewrite the planning half of `AGENTS.md`:

1. Delete the "Planning Rules" section (source of truth `mp status` /
   `raul status`, `mp milestone create`, `mp backlog add`, `mp decisions add`,
   "never edit `master-plan/*.json`").
2. Delete the `master-plan/` row from the Repository Map, and delete the
   `master-plan/AGENTS.md` entry from References.
3. Delete the "Planning context" line from the header blockquote.
4. Add a short **"Planning"** section that states, in the agent's own terms:
   - this project does **not** use the `mp` CLI or `master-plan/`;
   - `master-plan/` is a **frozen local archive**, untracked by git, kept on this
     machine for history only; it is **not** a spec, not a plan, not an input to
     any decision; do not read it, do not update it, do not run `mp` here;
   - `work/briefs/<slice>.md` is the contract for a slice; `work/STATE.md` is the
     resume record; `work/backlog.md` is the parked-work list;
     `work/questions.md` holds open questions.
5. Add the **lane protocol** rules that replace `mp`:
   - one commit per acceptance criterion, staged by name;
   - the gate must pass before a lane reports;
   - a lane reports by writing `work/inbox/<lane>-<slice>.md`, committing it,
     then nudging the orchestrator — the file is the record, the nudge is the bell;
   - the slice's row in `work/STATE.md` is the resume record and is updated in the
     same session as the slice;
   - a question that is not written in `work/questions.md` is a lost question.
6. Fix the release-readiness section: it must point at the tracked smoke
   checklist path chosen in AC-04, **not** `master-plan/smoke-checklist.md`.
7. Keep every engineering rule that is unrelated to planning (Bubble Tea rules,
   testing rules, git rules, verification commands) byte-for-byte where
   possible.

### AC-02 — `README.md` drops the plan directory

Remove the `master-plan/` line from the README repository map and any prose
mentioning `mp` planning. Add one line for `work/` describing it as the
brief/inbox/state area for the lane flow. Do not otherwise restructure the
README.

### AC-03 — `DESIGN.md` stops linking into the plan and records the decision

1. The "See [config ADR in M18.9](master-plan/milestones/18-documentation-and-project-hygiene.md)"
   link is **broken** — no milestone 18 exists in `master-plan/milestones/`.
   Replace it with a plain-text reference to the config ADR topic (no link into
   `master-plan/`).
2. Append one row to the **Decision Log** table:
   `| 2026-10-31 | Detached the mp/master-plan planning workflow; planning now runs through the herdr lane flow in work/ (briefs, inbox, state row). master-plan/ is kept locally as a frozen archive and is no longer tracked by git. | Slice detach-master-plan |`
3. Change nothing else in `DESIGN.md`.

### AC-04 — the tracked lane-flow structure replaces the plan

Create these files, all tracked:

1. `work/README.md` — the lane protocol in one page: the brief, the receipt,
   the gate, the commits, the state row, the questions file, who owns what.
2. `work/backlog.md` — the adopted residue, as a table with columns
   `id | work | source | priority | status`, where:
   - `LB-01` = verify the M13 Deps-tab fix by running the Deps test suite
     (source: M13, awaiting independent review), priority medium;
   - `LB-02` = context timeouts on the four brew fetch paths (source: B-02),
     priority medium;
   - `LB-03` = status dashboard outdated count (source: B-03), priority low;
   - `LB-04` = chunk `brew tap-info --json` batches (source: B-04), priority low;
   - `LB-05` = full interactive human smoke before the stable tag (source: B-05),
     priority medium, marked as a human gate;
   - `LB-06` = the M13 review itself — state it as closed-by-verification with
     LB-01, so no work is silently dropped;
   - `LB-DROP-01` = the dead `master-plan-bkp` header fix (source: B-01),
     priority n/a, status `dropped`, with the reason recorded: the target
     directory does not exist on this machine.
   Every row's `status` starts as `active` except the dropped row.
3. `work/STATE.md` — the resume record: a table with columns
   `slice | commits | state | verdict`, one row for `detach-master-plan`
   (state `implementing`, commits `—` for now), plus a one-paragraph note on how
   a cold session resumes work.
4. `work/questions.md` — the open-questions file, initially stating that no
   questions are open, with the rule that the reviewer reports each question as
   answered, missed, or newly raised.
5. `SMOKE-CHECKLIST.md` at the repo root — the smoke checklist copied out of
   `master-plan/smoke-checklist.md`, with its 28 items preserved and its paths
   updated so it no longer points inside `master-plan/`. Keep the item text and
   the `[x]` states as they are.

### AC-05 — the plan directory leaves git, stays on disk

1. Add `/master-plan/` to `.gitignore`.
2. Remove the now-dead `# mp CLI lock file (transient)` / `.mp-project-write.lock`
   entry from `.gitignore`.
3. `git rm -r --cached master-plan` — the working-tree files **stay exactly
   where they are**; only the tracking goes away. Do **not** delete the
   directory, do not move it, do not edit anything inside it.
4. The commit records the 33 file removals from the index.

## Invariants

- **No Go source changes.** `internal/**` and `cmd/**` are untouched. No test
  changes. `Makefile` and `.github/workflows/**` untouched.
- **Nothing under `master-plan/` is modified, moved, or deleted.** After the
  slice, `master-plan/` is still on disk with the same 33 files.
- **Git history is not rewritten.** The plan's history stays in the repo's past
  commits.
- The brief file itself (`work/briefs/detach-master-plan.md`) is not edited.
- One commit per acceptance criterion, in AC order, staged by name.
- No new dependencies, no toolchain changes.

## Scope — do NOT touch

- `internal/**`, `cmd/**`, `testdata/**`, `go.mod`, `go.sum`, `Makefile`
- `.github/workflows/**`
- `master-plan/**` (contents)
- `.pi-subagents/**` (gitignored scratch)
- `scripts/**`
- Any file outside the AC list.

## Evidence to produce

Run these and record the exact result of each in the receipt:

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

Plus these detach checks:

```bash
git ls-files master-plan | wc -l          # must print 0
ls master-plan | wc -l                    # must still print 15+ (files remain on disk)
git grep -n -i -E 'master-plan|mp status|mp milestone|mp backlog' -- . ':(exclude)master-plan' ; echo "exit=$?"
# every remaining hit must be inside a sentence that explicitly says master-plan is a
# frozen, untracked, local-only archive — no live instruction may survive
```

If any command fails, fix it or report it as a blocker in the receipt. Do not
paper over a failure.

## Report protocol

1. Write `work/inbox/runner-detach-master-plan.md` — concise markdown, at most
   40 lines, stating:
   - the commit ids, one per AC, in order;
   - each AC answered with the evidence that proves it (the command and its
     output, not a restatement of the intent);
   - the gate result (all four commands, pass or fail);
   - anything the next session must know, including any command that failed and
     any contradiction you found between this brief and the repo.
2. Commit it with: `git add work/inbox/runner-detach-master-plan.md && git commit -m "inbox: runner reports detach-master-plan"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished detach-master-plan; receipt at work/inbox/runner-detach-master-plan.md`
   - send it with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. If you are blocked and cannot finish a criterion, write the receipt anyway
   with the blocker named, and say so in the nudge. A blocked lane that reports
   is better than a lane that goes quiet.

## Questions

If something in this brief is wrong or ambiguous, write it in
`work/questions.md` (append, do not rewrite) and keep going on the criteria that
are unaffected. The orchestrator owns the open questions; do not answer them
yourself.
