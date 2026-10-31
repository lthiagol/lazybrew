# Fix brief — slice `detach-master-plan`, reviewer findings F-01..F-04

> **Contract.** The slice contract is `work/briefs/detach-master-plan.md`. This
> brief names the findings it must fix. Where the two differ, this brief wins
> for the fix cycle only.

## Context

The runner lane implemented `detach-master-plan` (commits `3a770a7`, `b95af56`,
`7428739`, `3431b9d`, `180d398`, receipt `2dff233`). The reviewer lane ran the
gate itself, ran the mutations, and returned
**`Verdict: changes-needed`** in `work/inbox/reviewer-detach-master-plan.md`
(commit `a204488`).

Read that receipt before you start. Everything in it that is **not** a finding is
already verified and must not be touched.

The reviewer's evidence checks (V1–V8) and all three mutations (M1–M3) passed.
The gate passes. The detach itself works. **Only the four findings below are
open.** Do not re-do the slice, do not "improve" the docs, do not restructure
anything.

## Findings to fix

One commit per finding, in order, staged by name — never `git add -A`, never
`git add .`.

### F-01 (major) — the two missed questions were never written down

The runner found two defects in its brief but recorded neither in
`work/questions.md`, which still says "_None. No questions are open._" The
contract's Questions section and `AGENTS.md` rule 5 require a brief-wrong or
ambiguous item to be written there. The runner put both only in its receipt.

The two items:

1. **The `.mp-write.lock` contradiction.** AC-04 said the smoke checklist was
   "tracked only inside `master-plan/`", but git also tracked
   `master-plan/.mp-write.lock` — so the checklist was not the only tracked
   file there.
2. **The LB-06 clause ambiguity.** AC-04 required LB-06 to be
   "closed-by-verification with LB-01", while also saying "every row's status
   starts as `active` except the dropped row". The two clauses conflict for
   LB-06 specifically.

**Fix:** append to `work/questions.md` — **append, do not rewrite** — a section
recording both questions. Each entry states: the question, who raised it (the
runner lane), who owns it (**the orchestrator**), and its resolution:

1. resolved — the checklist was copied out to `SMOKE-CHECKLIST.md` and the
   directory was untracked with no on-disk file modified; the extra tracked
   lock file was removed from the index with the rest of the directory.
2. resolved — the more specific clause wins: LB-06's status is
   `closed-by-LB-01` (see F-02).

Because both are resolved, the "no questions are open" line stays true — leave
it in place. The appended section is the record that they were raised at all.

Target file: `work/questions.md` (created in `3431b9d`).

### F-02 (minor) — LB-06's status cell says `active`

A reader of the status column alone cannot see that the M13 review is closed.

**Fix:** in `work/backlog.md`, set LB-06's `status` cell to `closed-by-LB-01`.

Target file: `work/backlog.md` (created in `3431b9d`).

### F-03 (minor) — LB-DROP-01 describes a directory that cannot exist

`master-plan-bkp/` exists nowhere on this machine, so it cannot be "part of the
frozen, untracked, local-only archive".

**Fix:** in `work/backlog.md`, drop the archive parenthetical from the
LB-DROP-01 row and from its reason paragraph. Keep the required reason — the
target directory does not exist on this machine — and keep the row's
`status = dropped`.

Target file: `work/backlog.md` (created in `3431b9d`).

### F-04 (minor) — an unrelated line lost its trailing markdown break

AC-01.7 asked for unrelated content to stay byte-for-byte. The trailing
double-space was stripped from the untouched "Architectural context" line in
the `AGENTS.md` header blockquote. Rendering is unaffected, but the byte change
was not required.

**Fix:** in `AGENTS.md`, restore the two trailing spaces on that line so it
matches `74554ad` exactly.

Target file: `AGENTS.md` (last touched in `3a770a7`).

## Invariants

- **Exactly four commits**, one per finding, in F-01 → F-04 order.
- **No other file changes.** Nothing under `internal/`, `cmd/`, `testdata/`,
  `scripts/`, `.github/`, `Makefile`, `SMOKE-CHECKLIST.md`, `README.md`,
  `DESIGN.md`, `.gitignore`, or `work/STATE.md`.
- **Nothing under `master-plan/` is modified, moved, or deleted.** It stays on
  disk, untracked and gitignored, byte-identical to `3431b9d`.
- Do not re-run the detach, do not re-commit the slice's other commits, do not
  amend or rebase existing commits.
- Do not touch `work/inbox/runner-detach-master-plan.md` or
  `work/inbox/reviewer-detach-master-plan.md`.
- `work/briefs/**` is not editable.
- One commit per finding, staged by name.

## Evidence to produce

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Plus, for the findings themselves:

```bash
grep -n -A20 -i 'raised late' work/questions.md   # F-01: both questions present, owner named
grep -n 'LB-06' work/backlog.md                   # F-02: status closed-by-LB-01
grep -n 'LB-DROP-01' work/backlog.md              # F-03: no archive parenthetical
git diff 74554ad..HEAD -- AGENTS.md | grep -c 'Architectural context'  # F-04: 0 (byte-identical to the original)
```

Record the real output of each in the receipt.

## Report protocol

1. Write `work/inbox/runner-fix-detach-master-plan.md` — concise markdown, at
   most 30 lines: the four commit ids in order, each finding answered with the
   evidence command and its output, the gate result, and anything the next
   session must know.
2. Commit it:
   `git add work/inbox/runner-fix-detach-master-plan.md && git commit -m "inbox: runner reports fix-detach-master-plan"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `runner finished the detach-master-plan fix cycle; receipt at work/inbox/runner-fix-detach-master-plan.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
