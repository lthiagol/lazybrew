# Re-review brief — fix cycle for slice `detach-master-plan`

> **Contract.** The slice contract is `work/briefs/detach-master-plan.md`; the
> fix cycle contract is `work/briefs/fix-detach-master-plan.md`. This brief
> tells you how to verify the fix. The contracts win where they differ.

## Context

The first review of `detach-master-plan` returned **`Verdict: changes-needed`**
with four findings (F-01..F-04), recorded in
`work/inbox/reviewer-detach-master-plan.md` (commit `a204488`).

A fix cycle then ran and produced four commits, one per finding:

| Finding | commit | subject |
|---|---|---|
| F-01 | `17b1b8d` | docs(questions): record the two missed detach-master-plan questions |
| F-02 | `dce1468` | docs(backlog): mark LB-06 closed-by-LB-01 |
| F-03 | `26f44cf` | docs(backlog): drop archive parenthetical from LB-DROP-01 |
| F-04 | `88dbed9` | docs(agents): restore trailing markdown break on the header line |

Its receipt is `work/inbox/runner-fix-detach-master-plan.md` (commit `0bcff0b`).

Your job: **re-derive the fix from the tree and prove it by command.** The
fix receipt is a claim, not evidence — the gate in particular must be run by
you. Every check in the first review that passed (V1–V8, M1–M3) must still
pass after the fix; a fix that breaks an earlier verified property is a
finding.

You do not own the implementation. You may mutate files only to prove a check
bites, and you must revert each mutation before you report.

## Verify by command

### R1 — F-01: the two questions are recorded, appended not rewritten

```bash
grep -n -i -E 'raised late|\.mp-write\.lock|LB-06' work/questions.md
# both questions present, each naming the raiser (runner lane) and the owner (the orchestrator)
git diff 3431b9d..HEAD -- work/questions.md
# additions only — the original sections must not have been rewritten
```

### R2 — F-02: LB-06's status cell

```bash
grep -n 'LB-06' work/backlog.md
# the status cell must read closed-by-LB-01
```

### R3 — F-03: LB-DROP-01

```bash
grep -n 'LB-DROP-01' work/backlog.md
grep -n -A3 'LB-DROP-01 reason' work/backlog.md
# no archive parenthetical; status dropped; the reason keeps "does not exist on this machine"
ls -d master-plan-bkp ../master-plan-bkp 2>&1   # still must not exist
```

### R4 — F-04: the AGENTS.md header line is byte-identical again

```bash
diff <(git show 74554ad:AGENTS.md | sed -n '5p') <(git show HEAD:AGENTS.md | sed -n '5p')
# must print nothing
```

Note: the fix receipt's own explanation of this check was wrong — it said the
fix brief's literal command prints `1`. It prints `0`: the pattern `^[+-]`
cannot match a diff *context* line, which begins with a space. The receipt's
recorded output (`0`) was right; its explanation was not. Judge F-04 on
byte-identity, as above. (Corrected after the re-review; the brief's
requirements were unchanged.)

### R5 — the fix touched nothing else

```bash
git diff --stat 8d6fa5d..HEAD
# only AGENTS.md, work/backlog.md, work/questions.md
git diff --name-only 8d6fa5d..HEAD -- internal/ cmd/ testdata/ scripts/ Makefile .github/ work/STATE.md work/briefs work/inbox SMOKE-CHECKLIST.md README.md DESIGN.md .gitignore
# must print nothing
git log --oneline 8d6fa5d..HEAD   # exactly four fix commits + the receipt commit
```

### R6 — the archive is still intact

```bash
git ls-files master-plan | wc -l    # must be 0
ls master-plan | wc -l              # the directory is still on disk
git ls-tree -r 3431b9d master-plan | while read -r mode type hash path; do
  test "$(git hash-object "$path")" = "$hash" || echo "CHANGED: $path"
done
# must print nothing
```

### R7 — no regression of the first review's verified checks

Re-run the substance of V1, V2, V3, V4 and V5 from
`work/briefs/review-detach-master-plan.md`:

- `AGENTS.md` still has no live `mp` instruction;
- `README.md` still has no `mp` planning prose;
- `DESIGN.md` still has no link into `master-plan/` and still carries the
  2026-10-31 decision row;
- the five lane-flow files and `SMOKE-CHECKLIST.md` still exist, and
  `grep -c '^\- \[x\]' SMOKE-CHECKLIST.md` is still **28**;
- `git diff --stat 74554ad..180d398` still shows no `internal/`, `cmd/`,
  `testdata/`, `scripts/`, `Makefile` or `.github/` path.

Also check that the appended questions section did not introduce a **live**
`mp` instruction anywhere:

```bash
git grep -n -i -E 'master-plan|mp status|mp milestone|mp backlog' -- . ':(exclude)master-plan'
# every hit must be a historical mention, a frozen-archive statement, a brief, or a receipt — none may instruct an agent to run mp or read the plan
```

### R8 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output.

## Mutations to run and revert

Prove the checks bite. Revert each one; `git status` must be clean when you
report.

- **N1** — revert LB-06's status to `active` in `work/backlog.md`; R2's grep
  must fail to find `closed-by-LB-01`. Revert.
- **N2** — delete the appended questions section from `work/questions.md`; R1's
  grep must stop finding it. Revert.
- **N3** — strip the trailing two spaces from `AGENTS.md` line 5; R4's `diff`
  must report a difference. Revert.

A mutation that does not make its check fire is itself a finding: the check is
vacuous.

## Open questions

Read `work/questions.md`. Report whether each of the two recorded questions is
**answered**, **missed**, or **newly raised**, and whether any new question
was raised by this fix cycle.

## Scope you do not own

- You do not fix anything. You mutate only to prove a check bites, and revert.
- You do not touch `work/STATE.md` — the slice's row is the orchestrator's, and
  the orchestrator closes the slice after your verdict.
- You do not touch the contents of `master-plan/` on disk.
- You do not rewrite the docs to be "better". You check them against the
  contracts.
- You do not prompt the runner lane and you do not spawn agents.
- You commit only your own receipt file.

## Verdict format

Your receipt must end with one line:

```
Verdict: ok
```

or

```
Verdict: changes-needed
```

followed by findings, each as `F-NN | severity | what is wrong | the command
that proves it | the fix`. Severity is `blocker`, `major`, or `minor`.

## Report protocol

1. Write `work/inbox/reviewer-fix-detach-master-plan.md` — concise markdown, at
   most 45 lines: the verdict, any findings, the evidence (command → result)
   for R1–R8, the outcome of each mutation N1–N3, and the state of the
   questions.
2. Commit it:
   `git add work/inbox/reviewer-fix-detach-master-plan.md && git commit -m "inbox: reviewer reports fix-detach-master-plan"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished the detach-master-plan fix cycle; verdict in work/inbox/reviewer-fix-detach-master-plan.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
