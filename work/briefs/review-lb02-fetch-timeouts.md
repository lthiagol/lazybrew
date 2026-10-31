# Review brief — slice `lb02-fetch-timeouts`

> **Contract.** The slice contract is `work/briefs/lb02-fetch-timeouts.md`. This
> brief tells you *how to verify* it. The contract wins where the two differ.

## What the runner claims

The runner lane bounded the four brew fetch paths. Its commits:

| AC | commit | subject |
|---|---|---|
| AC-01 | `0027596` | feat(gui): bound panel and status brew reads with timeouts |
| AC-02 | `f7bc287` | feat(gui): bound the search brew reads with timeouts |
| AC-03 | `00f73f2` | test(gui): cover the brew fetch timeouts on all four paths |
| AC-04 | `16c24f4` | docs(backlog): close LB-02 and record LB-07 |

Its receipt is `work/inbox/runner-lb02-fetch-timeouts.md` (commit `a64219f`).

The runner claims all four gate commands passed, and that its own mutation
checks showed: detecting the timeout from the error's type instead of
`ctx.Err()` fails 5 tests; dropping the total budget fails
`TestFetchStatusDataTotalBudgetBoundsTheAggregate`; dropping the per-read cap
makes the panel/search tests hang and fail. **A receipt is a claim, not
evidence.** Run the gate yourself and reproduce the mutations.

## What this slice must guarantee

From the contract:

| id | guarantee |
|---|---|
| G-1 | every individual brew read is capped — 30s for panel/status reads, 15s for search/search-info — via package-level `var`s a test can override |
| G-2 | the status dashboard resolves within a **total** budget (45s) even if every source hangs, and the total budget is a separate bound from the per-read cap |
| G-3 | a timeout surfaces as `*brew.TimeoutError` (the type at `internal/brew/errors.go:42`, previously unused), **detected from the context's own error**, never from the type of the returned error |
| G-4 | the tab-content path (`fetchTabContentTimeout`, `fetchTabContentCmd`) is untouched |

The runner's implementation: `brewReadTimeout` (30s), `brewSearchTimeout`
(15s), `brewStatusTotalTimeout` (45s), a `boundedRead`/`boundedReadIn` helper,
and `timeoutError` at `internal/gui/commands.go:60`.

## Verify by command

### V1 — G-1: the caps exist, with the right values, and are overridable

```bash
grep -n "brewReadTimeout\|brewSearchTimeout\|brewStatusTotalTimeout" internal/gui/commands.go
```

30s / 15s / 45s, all package-level `var`s. Then confirm each of the four paths
actually uses its cap — no path may still call the brew client with an
unbounded context:

```bash
grep -n "context.Background()" internal/gui/commands.go
```

Every remaining hit must be accounted for: the bounded-read helper, the
tab-content path, the two `LB-07` reads, and the `serviceCleanup` write path.
**A bare `context.Background()` passed straight to a brew client in one of the
four paths is a blocker.**

### V2 — G-2: the total budget is real and separate

```bash
grep -n -A25 "func statusDashboard" internal/gui/commands.go
```

There must be a parent context with a total deadline that the per-read
deadlines are derived from, so that one slow source cannot consume the whole
dashboard's time. Confirm the two bounds are distinct numbers and that the
per-read deadline is *trimmed* by the parent's remaining budget.

### V3 — G-3: the timeout is detected from the context, not the error type

```bash
grep -n -A12 "func timeoutError" internal/gui/commands.go
```

The detection must be `ctx.Err() == context.DeadlineExceeded`. The reason is
in the code's own comment and it matters: `exec.CommandContext` kills the brew
process when the deadline fires, so the runner returns `"signal: killed"`,
**not** `context.DeadlineExceeded`. A detection based on the returned error's
type silently never fires. The runner's mock reproduces exactly this.

Also confirm the error type is reused, not reinvented:

```bash
grep -rn "TimeoutError" internal/gui/ internal/brew/errors.go
```

### V4 — AC-03: the tests exist, and they bite

Six tests, in `internal/gui/fetch_timeouts_test.go`:

- `TestFetchPanelDataTimesOut`
- `TestFetchStatusDataTimesOut`
- `TestFetchStatusDataRespectsTotalBudget`
- `TestFetchStatusDataTotalBudgetBoundsTheAggregate`
- `TestExecuteSearchTimesOut`
- `TestFetchSelectedSearchInfoTimesOut`

Each must assert that the call **returns** (every wait is bounded — a test that
can hang is worse than no test), that the error is a `*brew.TimeoutError` via
`errors.As`, and that the message names the command and the duration. Judge
the assertions, not just their presence.

### V5 — G-4 and scope: nothing else moved

```bash
# the tab-content path is byte-identical to its pre-slice state
diff <(git show 1ec84ec:internal/gui/commands.go | sed -n '/func fetchTabContentCmd/,/^}/p') \
     <(git show HEAD:internal/gui/commands.go | sed -n '/func fetchTabContentCmd/,/^}/p')
# must print nothing

git diff --name-only 1ec84ec..16c24f4
# must be only internal/gui/commands.go, internal/gui/fetch_timeouts_test.go,
# work/backlog.md, work/STATE.md

git diff --name-only 1ec84ec..16c24f4 -- internal/brew/ go.mod go.sum .github/ Makefile scripts/
# must print nothing

# no existing test was weakened
git diff 1ec84ec..16c24f4 -- internal/gui/gui_test.go
# must print nothing
```

### V6 — the dashboard's success path is unchanged

This slice restructured `fetchStatusData`. Confirm the user-visible behaviour
is the same when nothing times out:

- the dashboard items are formatted from the same six sources in the same way
  (`presentation.FormatStatusDashboard` with the same arguments);
- the doctor status line is still appended after the items
  (`presentation.FormatDoctorStatus(warnings, docErr)`);
- the aggregated `⚠` lines appear in the same **order** as before — formulae,
  casks, taps, services, doctor;
- `Diagnostics.Config`'s error is still **ignored** (the dashboard has always
  tolerated a missing config) — it must not start surfacing an error;
- the outdated count is still read from the **cache only** (the M9 lazy
  behaviour), never from a fresh `brew outdated`.

### V7 — the concurrency decision

The runner made the six dashboard reads run **concurrently** under the total
budget, instead of sequentially. The contract specified behaviour, not shape,
so this is allowed — but it is the most consequential choice in the slice and
it is yours to judge.

Judge it on:

- **Soundness.** Each goroutine writes only its own local variable and the
  function joins before reading them. Confirm there is no shared mutable state
  and that the race detector agrees: `go test ./internal/gui/ -race -count=5`.
- **The brew cache.** Six concurrent brew reads may write to the same
  `brew.Client` cache (`internal/brew/cache.go` — a typed cache with
  reader/writer splits). Read it and judge whether concurrent writes are safe.
  Report it as a finding if you are not satisfied.
- **Resource shape.** Six simultaneous `brew` processes at dashboard time is a
  real load on the user's machine, and `brew doctor` is heavy. Is the
  concurrency worth it against simply bounding the reads sequentially? Say so
  plainly — this is a judgement the orchestrator needs, not a nit.
- **Determinism.** The `⚠` line order must not vary between runs.

### V8 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output, including the coverage numbers.

### V9 — AC-04: the bookkeeping

```bash
git show 16c24f4 -- work/backlog.md work/STATE.md
```

`LB-02` must be `closed` with a note; a new row `LB-07` must record the two
still-unbounded reads (the uninstall dependency pre-check and `toggleLeaves`),
low priority, `active`, sourced "found by lb02-fetch-timeouts", with the
**current** line numbers. `LB-01`, `LB-03`, `LB-04`, `LB-05`, `LB-06` and
`LB-DROP-01` must be untouched.

## Mutations to run and revert

Prove the new tests bite. **Revert every mutation**; `git status` must be
clean when you report.

- **M1 (G-3's detection rule)** — in `timeoutError`, change
  `ctx.Err() == context.DeadlineExceeded` to
  `errors.Is(err, context.DeadlineExceeded)`. Run
  `go test ./internal/gui/ -run 'TimesOut|RespectsTotalBudget|TotalBudgetBoundsTheAggregate' -count=1`.
  The timeout tests must **fail** (the mock returns "signal: killed"). Revert.
- **M2 (G-2's total budget)** — in `statusDashboard`, give the parent context
  no deadline (e.g. `context.Background()` instead of the total-budget
  context). Run
  `go test ./internal/gui/ -run 'TestFetchStatusDataTotalBudgetBoundsTheAggregate' -count=1`.
  It must **fail**. Revert.
- **M3 (G-1's per-read cap)** — in `boundedReadIn`, remove the deadline
  (`context.WithTimeout(parent, cap)` → `parent`). Run
  `go test ./internal/gui/ -run 'TestFetchPanelDataTimesOut|TestExecuteSearchTimesOut' -count=1`.
  They must **fail** (the bounded wait must catch the hang). Revert.
- **M4 (G-3's error type)** — in `timeoutError`, return the raw `err` instead
  of `&brew.TimeoutError{...}`. Run the same timeout tests. The
  `errors.As` assertions must **fail**. Revert.
- **M5 (G-4)** — break the untouched tab-content path (remove its
  `context.WithTimeout`). Run
  `go test ./internal/gui/ -run TestFetchTabContentCmdTimesOut -count=1`.
  It must **fail** — this proves the pre-existing coverage is still live and
  that the slice did not quietly weaken it. Revert.

A mutation that does not make its test fail is a finding: the test is vacuous.

## Open questions

Read `work/questions.md`. Report whether anything in it was answered, missed,
or newly raised by this slice.

## Scope you do not own

- You do not fix anything.
- You mutate only to prove a check bites, and you revert it.
- You do not rewrite the implementation to be simpler, faster, or nicer.
- You do not touch `internal/brew/**`, the tab-content path, or the contents
  of `master-plan/`.
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

1. Write `work/inbox/reviewer-lb02-fetch-timeouts.md` — concise markdown, at
   most 45 lines: the verdict, the findings, the evidence (command → result)
   for V1–V9, the outcome of each mutation M1–M5, your judgement on the
   concurrency decision (V7), and the state of the questions.
2. Commit it:
   `git add work/inbox/reviewer-lb02-fetch-timeouts.md && git commit -m "inbox: reviewer reports lb02-fetch-timeouts"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished lb02-fetch-timeouts; verdict in work/inbox/reviewer-lb02-fetch-timeouts.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
