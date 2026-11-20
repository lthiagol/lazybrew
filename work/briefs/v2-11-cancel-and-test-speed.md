# Slice: v2-11-cancel-and-test-speed — cancel stops brew's whole process tree; fast tests

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Found by the orchestrator, ruling `Q-v2-test-speed` in `work/questions.md`.
The gate's race run takes ~39 s; `TestExecuteStreamCancelReportsCanceled`
alone takes **30 s**. Cause, and a real bug: `internal/brew/runner.go` (both
`exec.CommandContext` calls, ~:74 and ~:158) kills only brew's own process on
cancel. Children brew started (curl, git, ruby helpers) keep the stdout pipe
open, so a cancelled stream does not close until they exit: in the app, Esc on
a running install/upgrade leaves the Operation view on `Running` until the
download or git step finishes by itself. The test's fake brew runs `sleep 30`
and shows exactly that.

## ⚠ Safety

Never run brew write commands on this machine (`brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|update`,
`brew services start|stop|restart|run|cleanup|kill`). `moshi-hook` is a live
service. Tests use the mock runner only.

The frame harness may press `t`/`x` **only** to open a confirm or menu and
must then press `esc`; never `y` or `enter` on a tap confirm or menu, and
never submit the add-tap input.

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes `go test` (no `-race`) for the packages it touches; the full gate (with `-race`) runs once,
before the receipt (ruling `Q-v2-builder-speed`).

### AC-01 — cancel kills the process group

1. Both runner paths start brew in its own process group
   (`SysProcAttr{Setpgid: true}`) and on context cancel kill the whole group
   (`cmd.Cancel` sending SIGKILL — or SIGTERM then SIGKILL — to `-pgid`), with
   a short `cmd.WaitDelay` (e.g. 2 s) so a stray holder of the pipe can never
   hang the stream. Build-tagged for Unix (Linux and macOS); keep it compiling
   on both (`GOOS=darwin go build ./...`).
2. `TestExecuteStreamCancelReportsCanceled` keeps its `sleep 30` fake brew and
   now finishes in well under 1 s; add a test that a grandchild
   (`sh -c 'sleep 30 & wait'` or similar) is gone after cancel (its pid no
   longer exists) and the stream closed. Same for `Execute`.
3. Nothing else in the runner changes (output, errors, logging).

### AC-02 — gui tests stop waiting on the wall clock

1. Find every gui/brew test that sleeps or waits real time ≥ 100 ms
   (`go test -json`, list them in `result.md` with before/after times):
   at least `TestToastTickMsgEventuallyClearsToast` (3.3 s),
   `TestNonRRefreshEmittersStayCacheOnly` (2 s), `TestSingleflightPropagatesError`
   (2 s) and the 0.5 s batch tests. Drive them with messages/fake time or
   shorter injected durations instead — **without weakening what they
   assert** (each changed test named with what it still proves).
2. Production durations unchanged (inject through a field/variable that tests
   override, as `brewReadTimeout` already is).

### AC-02b — CI runs the race detector

1. `.github/workflows/wip-ci.yml` and `stable-ci.yml`: the `go test` step runs
   `go test ./... -race -count=1` with `CGO_ENABLED: "1"` (ruling
   `Q-v2-builder-speed` amendment). Nothing else in the workflows changes.
2. After the push of this slice, the `wip-ci` run must be green — the
   orchestrator checks it; say in the receipt that it is to be checked.

### AC-03 — evidence and bookkeeping

1. `result.md` in `work/verification/v2-11-cancel-and-test-speed/`: the gate
   race run time before and after on this machine, the slowest 10 tests
   before/after.
2. `work/STATE.md` row
   `| v2-11-cancel-and-test-speed | — | implemented | awaiting reviewer |`.

## Invariants

- No GUI behaviour or frame change. Workflow change limited to AC-02b. No production timing change except the
  cancel path.
- Every frame exactly the terminal size. No test deleted, skipped or
  weakened; changed expectations named. Coverage floors unchanged. No new
  dependencies.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-11-cancel-and-test-speed.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-v2-11-cancel-and-test-speed.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-11-cancel-and-test-speed-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-v2-11-cancel-and-test-speed-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
