# Review brief — slice `detach-master-plan`

> **Contract.** The slice contract is `work/briefs/detach-master-plan.md`.
> This brief tells you *how to verify* it. The contract wins where the two
> differ.

## What the runner claims

The runner lane implemented slice `detach-master-plan` and wrote its receipt at
`work/inbox/runner-detach-master-plan.md` (committed as `2dff233`).

Its commits, one per acceptance criterion:

| AC | commit | subject |
|---|---|---|
| AC-01 | `3a770a7` | docs: AGENTS.md stops driving work from mp |
| AC-02 | `b95af56` | docs: README drops the plan directory, points at work/ |
| AC-03 | `7428739` | docs: DESIGN.md records the mp/master-plan detach decision |
| AC-04 | `3431b9d` | docs: add tracked lane-flow structure and smoke checklist |
| AC-05 | `180d398` | chore: detach master-plan from git tracking, keep it locally frozen |

The brief itself is `74554ad`.

The runner claims all four gate commands passed. **A receipt may be written
while a gate is still running — do not accept the claim. Run the gate
yourself.**

## Your job

Re-derive the slice from the code and the tree. **Verify by command, not by
reading.** Read to form a hypothesis; run to prove it.

You do not own the implementation. You may mutate files **only** to prove a
check bites, and you must revert each mutation before you report.

## Verify by command

### V1 — AC-01, no live `mp` instruction survives in `AGENTS.md`

```bash
grep -n -iE '\bmp (status|milestone|backlog|decisions|idea|interview|execution|next|show|list|validate|config|review|reviews)\b' AGENTS.md
# must print nothing
grep -n -i 'master-plan' AGENTS.md
# every hit must be part of a sentence saying master-plan is a frozen, untracked,
# local-only archive — not an instruction to read, edit, or run mp on it
```

Also confirm the engineering rules that had nothing to do with planning are
still present and materially unchanged:

```bash
git show 3a770a7 -- AGENTS.md | grep -E '^-' | grep -iE 'program.Send|task.Manager|zero goroutine|teatest|cover-check|Do not commit unless'
# must print nothing (no unrelated rule was deleted)
```

### V2 — AC-02, `README.md`

```bash
grep -n -iE 'master-plan|\bmp\b' README.md
# only the sentence that names master-plan as a frozen local archive, if any
git show b95af56 --stat   # must be README.md only
```

### V3 — AC-03, `DESIGN.md`

```bash
git show 7428739 -- DESIGN.md   # exactly the link fix + one decision-log row
grep -n 'master-plan' DESIGN.md # no link may point into master-plan/
grep -n '2026-10-31' DESIGN.md  # the decision row must exist
```

The link the runner claims was broken must indeed have been broken:
`ls master-plan/milestones/ | grep -i 18` → no match (the directory is
untracked now, but the file list on disk is unchanged).

### V4 — AC-04, the tracked lane-flow structure

```bash
ls work/README.md work/backlog.md work/STATE.md work/questions.md SMOKE-CHECKLIST.md
grep -c '^\- \[x\]' SMOKE-CHECKLIST.md            # must be 28
diff <(grep -v -i 'provenance' SMOKE-CHECKLIST.md) master-plan/smoke-checklist.md
# must be identical (item text and [x] states preserved)
```

`work/backlog.md` must carry exactly these rows, with these sources and
priorities:

| id | source | priority | status |
|---|---|---|---|
| LB-01 | M13 | medium | active |
| LB-02 | B-02 | medium | active |
| LB-03 | B-03 | low | active |
| LB-04 | B-04 | low | active |
| LB-05 | B-05 | medium | active |
| LB-06 | M13 | — | closed-by-LB-01 |
| LB-DROP-01 | B-01 | n/a | dropped, with the "master-plan-bkp does not exist" reason |

`work/STATE.md` must have a `detach-master-plan` row whose state is not
`implementing` (the runner moved it to `implemented` / `awaiting reviewer`) and
whose commits column names the five commit ids above.

### V5 — AC-05, the plan is out of git and still on disk

```bash
git ls-files master-plan | wc -l        # must be 0
grep -n 'master-plan' .gitignore        # must contain /master-plan/
grep -n 'mp-project-write.lock' .gitignore   # must print nothing
git show 180d398 --stat | tail -3      # 33 index removals
ls master-plan | wc -l                # the directory is still here
```

And the on-disk files are **unmodified** — compare against the last commit
that tracked them:

```bash
git ls-tree -r --name-only 3431b9d master-plan | wc -l   # 33
# for each of those paths, compare the blob hash in 3431b9d with the file on disk:
git ls-tree -r 3431b9d master-plan | while read -r mode type hash path; do
  test "$(git hash-object "$path")" = "$hash" || echo "CHANGED: $path"
done
# must print nothing
```

### V6 — scope, no code touched

```bash
git diff --stat 74554ad..180d398
```

Every path must be a doc (AGENTS.md, README.md, DESIGN.md, work/*,
SMOKE-CHECKLIST.md, .gitignore) or a `master-plan/` removal. **No file under
`internal/`, `cmd/`, `testdata/`, `scripts/`, `Makefile`, or
`.github/workflows/` may appear.**

### V7 — the foreign commit

Commit `86d34c1` ("feat(gui): mouse support — click rows, switch tabs, scroll")
landed in this repo **outside** this lane, before the brief was written. Confirm
the lane neither modified nor reverted it:

```bash
git show 86d34c1 --stat
git diff 74554ad..180d398 --stat -- internal/ cmd/
# must print nothing
```

### V8 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output, not the runner's.

## Mutations to run and revert

Each mutation proves that a check actually bites. Revert every one, and prove
the revert with `git status` being clean at the end.

- **M1** — re-add a live instruction line to `AGENTS.md`
  (`git show 86d34c1:AGENTS.md > /dev/null` is irrelevant; simply append
  `Source of truth: run mp status` to `AGENTS.md`), re-run V1's first grep, and
  confirm it now prints a hit. Revert.
- **M2** — remove the `/master-plan/` line from `.gitignore` and run
  `git add -f master-plan/config.json && git ls-files master-plan | wc -l`;
  confirm it is no longer 0. Revert the index change (`git rm --cached
  master-plan/config.json`) and restore `.gitignore`.
- **M3** — delete one `- [x]` item from `SMOKE-CHECKLIST.md` and re-run the
  count in V4; confirm it drops below 28. Revert.

If a mutation does **not** make its check fire, that is a finding: the check
is vacuous, and the corresponding AC is not actually verified.

## Open questions

Read `work/questions.md`. For every question in it, say whether the slice
answered it, missed it, or newly raised one. If the file says no questions are
open, verify that is true by re-reading the receipt and the diffs — the runner
reported one contradiction, so check whether it should have been written down
as a question.

## Scope you do not own

- You do not fix anything. You mutate only to prove a check bites, and you
  revert it.
- You do not rewrite the docs to be "better", clearer, or more complete. Your
  job is to check them against the contract.
- You do not touch the contents of `master-plan/` on disk. It is a frozen
  archive.
- You do not prompt the runner lane and you do not spawn agents.
- You do not commit any change other than your own receipt file.

## Verdict format

Your receipt must end with one line:

```
Verdict: ok
```

or

```
Verdict: changes-needed
```

followed by the findings, each as `F-NN | severity | what is wrong | the command
that proves it | the fix`.

Severity is `blocker`, `major`, or `minor`.

## Report protocol

1. Write `work/inbox/reviewer-detach-master-plan.md` — concise markdown, at most
   60 lines: the verdict, the findings, the evidence (command → result) for
   V1–V8, the outcome of each mutation M1–M3, and the state of the questions.
2. Commit it:
   `git add work/inbox/reviewer-detach-master-plan.md && git commit -m "inbox: reviewer reports detach-master-plan"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished detach-master-plan; verdict in work/inbox/reviewer-detach-master-plan.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report. If it is not, you left a mutation
   behind — fix it before reporting.
