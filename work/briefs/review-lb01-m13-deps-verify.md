# Review brief — slice `lb01-m13-deps-verify`

> **Contract.** The slice contract is `work/briefs/lb01-m13-deps-verify.md`. This
> brief tells you *how to verify* it. The contract wins where the two differ.

## What the runner claims

The runner lane verified the M13 Deps-tab fix and closed `LB-01`. Its commits:

| AC | commit | subject |
|---|---|---|
| AC-01 | `789d506` | docs(verification): record the M13 Deps-tab fix verification (LB-01) |
| AC-02 | `a7d5498` | docs(backlog): close LB-01 with its verification evidence |
| AC-03 | `196575b` | docs(state): add the lb01-m13-deps-verify row to the resume record |

Its receipt is `work/inbox/runner-lb01-m13-deps-verify.md`. Note the receipt
was committed twice (`5841cc9`, then `7cc8084` "tighten the runner receipt to
the brief's 30-line limit"). The contract asked for **one** receipt commit —
judge whether that is a finding or harmless, and say which.

The runner claims every gate command passed and that no code, test, Makefile or
workflow was touched. **A receipt is a claim, not evidence.** Run the gate
yourself.

## The point of this review

This slice is a *verification* slice. Its whole value rests on one question:

> Do the M13 regression tests actually fail when the M13 fix is broken?

A verification record that lists passing tests proves only that the tests pass
today. **You must break the fix on purpose and watch the named test fail.**
If a test still passes with its invariant broken, the verification is worthless
and that is a finding — a major one.

## The five invariants and the test that must bite

| id | invariant | anchor | test that must fail when broken |
|---|---|---|---|
| I-1 | tab-content fetch is bounded by a context timeout | `internal/gui/commands.go:22`, `:833` | `TestFetchTabContentCmdTimesOut` (`internal/gui/gui_test.go:1276`) |
| I-2 | after `DataLoadedMsg` the tab content load is re-kicked | `internal/gui/gui.go:455-456` | `TestDepsSelectionChangeUpdatesContent` (`internal/gui/gui_test.go:1370`) |
| I-3 | the Deps tab fast-paths from the cached formula's deps | `internal/gui/commands.go:779`, `:794`, `:810` | `TestDepsFastPathRendersCachedDependencies` (`:1314`), `TestDepsTabViewLeavesLoadingAfterFastPath` (`:1418`), `TestDepsTabLoads` (`:1446`) |
| I-4 | `parseFormula` maps build deps from `build_dependencies` | `internal/brew/formulae.go:318` (field `:64`) | `TestParseFormulaBuildDepsFromBuildDependencies`, `TestFormulaeReaderList` (`internal/brew/formulae_test.go`) |
| I-5 | tab/panel key handlers return early so the fetch cmd is not dropped | `internal/gui/gui.go:627-634` | `TestTabKey` (`:454`), `TestTabSwitching` (`:185`), `TestDataLoadedOnNonFetchTabDoesNotReissueFetch` (`:1254`) |

## Verify by command

### V1 — AC-01: the verification record is complete and truthful

```bash
cat work/verification/lb01-m13-deps.md
git show 789d506 --stat
```

For each of I-1..I-5: the record must name the invariant, a command that proves
it, that command's **real output**, and the line. Re-run each command yourself
and confirm the recorded output matches reality — a record that quotes output
which does not reproduce is a finding.

`work/README.md` must have gained exactly one row for `work/verification/` and
nothing else changed:

```bash
git show 789d506 -- work/README.md
```

### V2 — AC-02: the backlog close-out

```bash
git show a7d5498 -- work/backlog.md
grep -n -E 'LB-0[1-6]|LB-DROP-01' work/backlog.md
```

`LB-01` must be `closed`, with a note naming the evidence file and the commit.
`LB-06` must still be `closed-by-LB-01`. `LB-02`, `LB-03`, `LB-04`, `LB-05` and
`LB-DROP-01` must be **byte-identical** to their state before the slice:

```bash
git diff 46e9baa..HEAD -- work/backlog.md
```

Only the `LB-01` row and the added note may appear in that diff.

### V3 — AC-03: the state row

```bash
git show 196575b -- work/STATE.md
```

One row added, with the four columns; the `detach-master-plan` row untouched.

### V4 — scope: nothing but the four `work/` files

```bash
git diff --name-only 46e9baa..196575b
# must list only work/verification/lb01-m13-deps.md, work/README.md,
# work/backlog.md, work/STATE.md
git diff --name-only 46e9baa..196575b -- internal/ cmd/ testdata/ scripts/ Makefile .github/ go.mod go.sum
# must print nothing
```

### V5 — the gate, run by you

```bash
make build
make lint
go test ./... -race -count=1
make cover-check
```

All four must pass. Report the real output.

## Mutations — the heart of this review

Break each invariant and watch its test fail. **Revert every mutation.** The
worktree must be clean when you report.

- **M1 (I-1)** — in `internal/gui/commands.go`, remove the timeout: replace the
  `context.WithTimeout(context.Background(), fetchTabContentTimeout)` at `:833`
  with `context.Background()`. Run
  `go test ./internal/gui/ -run TestFetchTabContentCmdTimesOut -count=1`.
  It must **fail** (the test waits for the deadline). Revert.
- **M2 (I-2)** — in `internal/gui/gui.go`, delete the re-kick at `:455-456`
  (`return m, m.loadTabContent()`). Run
  `go test ./internal/gui/ -run TestDepsSelectionChangeUpdatesContent -count=1`.
  It must **fail**. Revert.
- **M3 (I-3)** — in `internal/gui/commands.go`, make the fast path never apply
  (e.g. change the `panelID == PanelFormulae && tabIdx == 2` condition at
  `:795` to `false`). Run
  `go test ./internal/gui/ -run 'TestDepsFastPathRendersCachedDependencies|TestDepsTabViewLeavesLoadingAfterFastPath|TestDepsTabLoads' -count=1`.
  At least one must **fail**. Revert.
- **M4 (I-4)** — in `internal/brew/formulae.go:318`, change
  `buildDeps := append([]string(nil), f.BuildDependencies...)` back to
  `buildDeps := []string(f.Dependencies)`. Run
  `go test ./internal/brew/ -run 'TestParseFormulaBuildDepsFromBuildDependencies|TestFormulaeReaderList' -count=1`.
  It must **fail**. Revert.
- **M5 (I-5)** — in `internal/gui/gui.go`, remove one of the early returns in
  `:627-634` so the fetch command is dropped. Run
  `go test ./internal/gui/ -run 'TestTabKey|TestTabSwitching|TestDataLoadedOnNonFetchTabDoesNotReissueFetch' -count=1`.
  At least one must **fail**. Revert.

**If a mutation does not make its test fail, that is a blocker finding.** It
means the M13 fix is not actually covered, and `LB-01` cannot be closed on
evidence — say so plainly.

Note: mutating a **test file** is out of scope — you are proving the tests
bite, not editing them. If a test genuinely cannot fail, that is the finding.

## The stale paragraph the runner flagged

The runner reported that the trailing paragraph of `work/STATE.md` still says
"Next slice: `LB-01`", which its own closure contradicts. That paragraph was
written by the orchestrator when the previous slice closed, and the state row
for a slice is the orchestrator's to close. **Judge it, do not fix it** —
report whether it is a finding, and at what severity. The orchestrator will
correct it.

## Open questions

Read `work/questions.md`. Report whether anything in it was answered, missed,
or newly raised by this slice.

## Scope you do not own

- You do not fix anything, including the stale `work/STATE.md` paragraph.
- You mutate only to prove a check bites, and you revert it.
- You do not touch `master-plan/` (a frozen local archive).
- You do not rewrite the verification record to be nicer.
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

1. Write `work/inbox/reviewer-lb01-m13-deps-verify.md` — concise markdown, at
   most 45 lines: the verdict, the findings, the evidence (command → result)
   for V1–V5, the outcome of each mutation M1–M5 (which test failed, with its
   failure output), and the state of the questions.
2. Commit it:
   `git add work/inbox/reviewer-lb01-m13-deps-verify.md && git commit -m "inbox: reviewer reports lb01-m13-deps-verify"`
3. Nudge the orchestrator:
   - pane id: `w27:p1`
   - text (one line): `reviewer finished lb01-m13-deps-verify; verdict in work/inbox/reviewer-lb01-m13-deps-verify.md`
   - send with `herdr pane send-text "w27:p1" "<text>"` then
     `herdr pane send-keys "w27:p1" enter`.
4. `git status` must be clean when you report.
