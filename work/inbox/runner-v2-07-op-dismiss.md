# runner receipt — v2-07-op-dismiss

## Commits, one per AC, in order

- AC-01 `0dfd9ff` success closes itself after 2 s (tea.Tick, names its op).
- AC-02 `ede3959` close hint on error/cancelled/success rows; footer `esc close` first.
- AC-03 `21b3b21` moving to another panel clears a finished operation.
- AC-04 `21fb1a1` harness + op frames in `work/verification/v2-07-op-dismiss/`, STATE row, Q-v2-07-pid.

## Evidence per AC

- AC-01: `TestSuccessfulOperationClosesOnItsTimer`, `TestOperationTimerIgnoresANewerOperation`,
  `TestBatchDoesNotCloseMidBatch`, `TestFailedAndCancelledOperationsStartNoTimer`.
  The timer's message is sent by hand; no test sleeps 2 s.
- AC-02: `TestFinishedRunsEndWithTheCloseHint` (exact rows), `TestLongErrorKeepsTheCloseHint`
  (60 cols), `TestFinishedOperationFooterStartsWithEscClose`.
- AC-03: `TestMovingAwayClosesAFinishedOperation` (Tab, S-Tab, `1`, sidebar click),
  `TestMovingAwayLeavesARunningOperation`, `TestServicesInfoDrawsTheReloadOnceTheRunCloses`.
- AC-04: `op/success|error|cancelled` (140×40) show the rows above and footer `esc close`.
  Resting frames: 54 of 60 byte-identical to v2-06; the other 6 differ only in the
  live moshi-hook PID row (below).

## Gate (final tree, 21fb1a1)

- `make build`, `make lint`: ok
- `CGO_ENABLED=1 go test ./... -race -count=1`: all packages ok (incl. TestEveryFrameIsTerminalSize)
- `make cover-check`: all floors met

## For the next session

- **AC-04.1 contradiction:** `after/` is not `diff -r` empty against v2-06. The
  harness reads the live moshi-hook PID (`1332457`; v2-06 had `800`); the service
  started at 07:01:42, before this slice. Kept as captured. Raised as
  Q-v2-07-pid in `work/questions.md`.
- Decision not in the brief: a key or click on the panel already shown clears
  nothing; only a change of panel does.
- Not verified: the 2 s close and hint in a real terminal; any real brew write
  (none ran; the harness only reads).
