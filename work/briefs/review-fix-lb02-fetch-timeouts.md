# Re-review brief — fix cycle for slice `lb02-fetch-timeouts`

> **Contract.** The slice contract is `work/briefs/lb02-fetch-timeouts.md`; the
> fix contract is `work/briefs/fix-lb02-fetch-timeouts.md`. This brief tells
> you how to verify the fix. The contracts win where they differ.

## Context

The review of `lb02-fetch-timeouts` returned **`Verdict: ok`** with two minor
findings, in `work/inbox/reviewer-lb02-fetch-timeouts.md` (commit `633dc50`):

- **F-01** — `boundedReadIn` could report a **negative** duration when the
  parent's budget was already spent (`brew list timed out after -1s`).
- **F-02** — the fan-out puts both heavy `brew info` loads and `brew doctor` in
  the same instant. The reviewer judged it acceptable; it is a shape note.

**F-02 was accepted and closed.** The fix cycle must not change the fan-out
shape, the timeout values, the detection rule, `timeoutError`, or the
tab-content path.

The fix cycle produced:

| AC | commit | subject |
|---|---|---|
| AC-01 | `8647165` | fix(gui): clamp negative budget and skip reads with a spent parent |
| AC-02 | `0cfb213` | test(gui): prove the timeout budget is never negative |

Its receipt is `work/inbox/runner-lb02-timeout-budget-fix.md` (commit
`e851ad2`).

Your job: **re-derive the fix from the tree and prove it by command.** The
runner claims the new test fails against the pre-fix code — check that claim
yourself.

## Verify by command

### S1 — the budget is never negative, and a spent parent skips the read

```bash
grep -n -A22 "func boundedReadIn" internal/gui/commands.go
```

There must be:

1. a clamp so the reported budget is never negative;
2. an early return when the parent's remaining budget is spent, so `read(ctx)`
   is **never invoked** — the brew client must not be called with a dead
   context;
3. the error must still be a `*brew.TimeoutError` naming the command.

The chosen behaviour for a spent parent is a reported duration of **0**
(`brew list timed out after 0s`). That is acceptable; a negative duration is
not.

### S2 — the test exists and bites

```bash
cat internal/gui/fetch_timeouts_budget_test.go
go test ./internal/gui/ -run 'SpentParent|Negative|Budget' -count=1 -v
```

The test must assert: the call returns promptly (bounded wait), the error is a
`*brew.TimeoutError` via `errors.As`, the duration is `>= 0`, the message
contains no negative duration, and the read function was **not** called.

**Prove it bites:** revert the fix (restore the unclamped `time.Until` and the
unconditional `read(ctx)` call) and confirm the test **fails**; then restore the
fix with `git checkout --`. A test that passes against the broken code is
vacuous.

### S3 — nothing else moved

```bash
git diff --name-only 633dc50..HEAD
# must be only internal/gui/commands.go and the new test file

git diff 633dc50..HEAD -- internal/gui/commands.go
# must show only the clamp and the early return inside boundedReadIn
```

Check specifically that these are **unchanged**:

```bash
grep -n "brewReadTimeout = \|brewSearchTimeout = \|brewStatusTotalTimeout = " internal/gui/commands.go
# still 30s / 15s / 45s
grep -n -A8 "func timeoutError" internal/gui/commands.go
# still detects ctx.Err() == context.DeadlineExceeded
diff <(git show 633dc50:internal/gui/commands.go | sed -n '/func fetchTabContentCmd/,/^}/p') \
     <(git show HEAD:internal/gui/commands.go | sed -n '/func fetchTabContentCmd/,/^}/p')
# must print nothing
```

The fan-out shape is unchanged: the six dashboard reads still run concurrently
under the total budget.

### S4 — scope

```bash
git diff --name-only 633dc50..HEAD -- internal/brew/ go.mod go.sum .github/ Makefile scripts/ work/
# must print nothing
```

No `work/` file is touched by the fix cycle — the state row is the
orchestrator's to close.

### S5 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output.

## Mutations to run and revert

- **Q1** — in `boundedReadIn`, restore the unclamped budget
  (`budget = time.Until(parentDeadline)` without a floor) **and** remove the
  early return so the read is always invoked. Run the new test. It must
  **fail**. Revert.
- **Q2** — keep the clamp but remove only the early return (so the read runs
  with a dead context). Run the new test. The "read was never called"
  assertion must **fail**. Revert.

`git status` must be clean when you report.

## Open questions

Read `work/questions.md`. Report whether anything in it was answered, missed,
or newly raised by this cycle.

## Scope you do not own

- You do not fix anything.
- You mutate only to prove a check bites, and you revert it.
- You do not change the fan-out shape, the timeout values, or the detection
  rule.
- You do not touch `work/STATE.md` or `work/backlog.md` — the orchestrator
  closes the slice.
- You do not touch the contents of `master-plan/`.
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

1. Write `work/inbox/reviewer-lb02-timeout-budget-fix.md` — concise markdown,
   at most 30 lines: the verdict, any findings, the evidence (command →
   result) for S1–S5, the outcome of each mutation Q1–Q2, and the state of the
   questions.
2. Commit it:
   `git add work/inbox/reviewer-lb02-timeout-budget-fix.md && git commit -m "inbox: reviewer reports lb02-timeout-budget-fix"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished the lb02 timeout-budget fix; verdict in work/inbox/reviewer-lb02-timeout-budget-fix.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
