# Common rules — code-review slices `cr1` … `cr5`

> Every `cr*` brief says "cr-common.md applies". This file is part of each of
> those contracts. Where a slice brief and this file differ, the slice brief
> wins.

## Source

`code-review.md` at the repo root holds the review these slices come from
(dated 2026-11-20, `wip` @ `c1aa732`). Each AC names the finding ids it
closes (C = correctness, P = performance, R = reuse, D = dead code, Q =
quality). Read each finding in full before you start its AC. The review gives
the evidence; the slice brief gives the contract.

## Before you fix a finding

1. **Find the code by its symbol name**, not by line number. Line numbers in
   the review are approximate (`~:N` at `c1aa732`) and shift as soon as a
   commit lands.
2. **Check that the finding is real.** Read the code. For bugs, write the
   failing test first (below). If the finding is wrong, already fixed, or
   would need a design decision the brief doesn't make, **do not force a
   fix**. Append a question to `work/questions.md` (`Q-<slice>-<id>`), say so
   in the receipt, and go on with the other ACs.
3. **Stay in scope.** If you notice something that isn't in your slice, write
   it in the receipt under "seen, not done". Don't fix it.

## Test rules

- **Bug fixes start with a test that fails** on the code before the fix. Put
  the failing run's output in the receipt: the test name and the line showing
  the failure. The reviewer will revert the fix and expect the test to fail.
- **Refactors change no behaviour.** No golden, snapshot or teatest
  expectation changes. If one has to change, name it in the receipt with the
  reason.
- **Perf ACs ship a Go benchmark** (`Benchmark…`, in the touched package).
  The receipt shows `go test -run '^$' -bench <name> -benchmem` output
  **before** (run it on the commit before the change, e.g. by stashing the
  fix) and **after**.
- No test is deleted, skipped or weakened, except tests for code that a
  dead-code AC removes. Name each removed test in the receipt.
- Inject the runner with `brew.MockRunner` (or `brewtest.MockRunner` once
  `cr5` moves it). Never touch real brew state.

## ⚠ Safety

Never run brew write commands on this machine: `brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|
update` and `brew services start|stop|restart|run|cleanup|kill`. The
`moshi-hook` service is live. Read-only brew commands (`info`, `list`,
`outdated`, `doctor`, `config`, `services list/info`) are allowed for manual
checks. Running `./bin/lazybrew` by hand is allowed only for read-only
navigation.

## Invariants (every slice)

- Follow the AGENTS.md Bubble Tea rules: no `program.Send`, no raw goroutines
  in handlers, every write goes through `task.Manager`.
- Coverage floors stay as they are (`make cover-check`). Don't lower any. If
  removing dead code raises a package's coverage, leave the floor alone too.
- No new dependencies in `go.mod`.
- The worktree has **uncommitted user edits to `AGENTS.md` and `Makefile`**.
  Never stage, commit, revert or reformat them. Stage only by name. Never use
  `git add -A` or `git add .`.
- One commit per AC, in AC order. The subject starts with the slice and AC
  (`cr1 AC-02: …`) and ends with the finding ids in brackets
  (`[C2]`). No amend, rebase, revert or push.
- `GOOS=darwin go build ./...` must still pass (lazybrew ships on macOS).

## Gate

- **Per commit:** the commit builds and passes `make test-fast PKG=<touched
  packages>`.
- **Before the receipt:** `make ci` and `GOOS=darwin go build ./...`.
- **Dead-code and refactor slices also run:**
  `go run golang.org/x/tools/cmd/deadcode@latest ./cmd/...` and
  `go run honnef.co/go/tools/cmd/staticcheck@latest ./...`. The receipt shows
  their output.

## Report protocol

1. Write `work/inbox/runner-<slice>.md`, at most 40 lines:
   - the commit ids, one per AC;
   - each AC's evidence (the command and its real output, not a restatement
     of the intent);
   - the failing-test lines for bug fixes and the benchmark numbers for perf
     ACs;
   - the gate result;
   - any questions you filed, and "seen, not done".
2. Commit the receipt once.
3. Add the slice's row to `work/STATE.md`:
   `| <slice> | <commit ids> | implemented | awaiting reviewer |`.
4. Reply with exactly one line: `done: work/inbox/runner-<slice>.md`.
5. If you are blocked, still write the receipt with the blocker named, and
   reply anyway.

## Fix rounds

The orchestrator reviews the slice. If it has findings, it writes
`work/briefs/fix-<slice>-rN.md`:

- fix every finding, one commit each, with the finding id in the subject;
- run the gate;
- write the receipt `work/inbox/runner-<slice>-fix-rN.md` (at most 30 lines,
  committed once).
