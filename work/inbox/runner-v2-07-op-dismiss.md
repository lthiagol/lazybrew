# runner receipt — v2-07-op-dismiss

## Commits, one per AC, in order

- AC-01 `0dfd9ff` successful operation closes itself after 2 s (tea.Tick, names its op).
- AC-02 `ede3959` close hint on error/cancelled/success rows; footer first hint `esc close`.
- AC-03 `21b3b21` moving to another panel clears a finished operation; running stays.
- AC-04 `21fb1a1` harness + op frames (`work/verification/v2-07-op-dismiss/`), STATE row, question.

## Evidence per AC

- AC-01: `TestSuccessfulOperationClosesOnItsTimer` (success -> `Completed in` ->
  opAutoCloseMsg -> `Services › Info` on View); `TestOperationTimerIgnoresANewerOperation`
  (stale and running timers close nothing); `TestBatchDoesNotCloseMidBatch`
  (`autoCloseCmd()` nil mid-batch, timer fired by hand still shows the op);
  `TestFailedAndCancelledOperationsStartNoTimer`. No test sleeps 2 s.
- AC-02: `TestFinishedRunsEndWithTheCloseHint` (exact rows, running keeps
  `Esc to cancel`); `TestLongErrorKeepsTheCloseHint` (60 cols, error cut,
  hint kept); `TestFinishedOperationFooterStartsWithEscClose`.
- AC-03: `TestMovingAwayClosesAFinishedOperation` (Tab, S-Tab, `1`, sidebar
  click; View no longer shows `Completed in`); `TestMovingAwayLeavesARunningOperation`;
  `TestServicesInfoDrawsTheReloadOnceTheRunCloses` (Info Status row reads `none`
  after the reload, not `Loading...`).
- AC-04: `op/success|error|cancelled` (.txt + .ansi, 140×40) show the rows
  above and footer `esc close`. Resting frames: 54 of 60 byte-identical to
  v2-06; the other 6 differ only in the live moshi-hook PID row (see below).

## Gate (final tree, 21fb1a1)

- `make build`: ok
- `make lint` (go vet ./...): ok
- `CGO_ENABLED=1 go test ./... -race -count=1`: all packages ok
- `make cover-check`: all floors met (brew 76.5/62, presentation 98.2/91,
  gui 84.1/36, modal 90.1/41)
- `TestEveryFrameIsTerminalSize` passes (in the gui package run).

## For the next session

- **Contradiction, AC-04.1:** `after/` is not `diff -r` empty against v2-06.
  The harness reads the live moshi-hook PID (`1332457`; v2-06 had `800`). The
  service started at 07:01:42, before this slice. Kept as captured, not edited.
  Raised as Q-v2-07-pid in `work/questions.md` for a ruling.
- Decision, not in the brief: a key or click on the panel already shown clears
  nothing (only a change of panel does). Search results that switch to Search
  from another panel do clear a finished run.
- The status row cuts a long error with `…` so the close hint stays on the row.
- Batch guard is in two places: the timer is started only when `batchCount == 0`,
  and the timer's handler re-checks it.
- Not verified: the 2 s close and the hint in a real terminal; a real brew
  write. No brew write ran; the frame harness only reads.
- Temporary `internal/gui/dump/` used to run the harness was removed; not committed.
