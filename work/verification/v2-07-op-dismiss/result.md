# v2-07-op-dismiss — frame result

Three sets of frames, all read-only with respect to brew: no write command ran.

## Resting frames (AC-04.1) — not byte-identical to v2-06, one live value differs

`after/` is `frame_dump_test.go.txt` (copied unchanged from
`v2-06-animations/`) run with `LAZYBREW_ANIMATIONS=0`, into
`work/verification/v2-07-op-dismiss/after/`. It reads the real brew (reads only:
`brew services info`, list and info reads); the harness sends no `y` or
`enter` on any confirm, and `S` is cancelled with `esc`.

`diff -rq work/verification/v2-07-op-dismiss/after work/verification/v2-06-animations/after`
is **not empty**. Six files differ, all in the Formulae box's PID row of frames
19 and 20 (`19-services-info`, `20-services-stop-confirm`): the PID is `1332457`
where v2-06 recorded `800`. Nothing else differs: the other 54 frames are
byte-identical to v2-06 (and to v2-05s).

Cause: the live `moshi-hook` service. `brew services info moshi-hook --json`
reports `pid 1332457`, `running: true`, `status: started`, and the service log
shows `gateway: listening` at `2026-11-20T07:01:42`, which is before this slice
began (the brief checkpoint commit `2a7e2a5` is at 07:09 -0300) and before the
harness ran (07:19). The service was restarted before this slice; the bug that
started it was reported after a `restart moshi-hook`. The harness read the PID
that was live then. Normalising the PID in a scratch copy does not give an
empty diff either: the value is seven digits, not three, so the row's padding
moves too. The rule "resting frames unchanged" therefore holds for the code;
the literal `diff -r` in AC-04.1 cannot hold while the live PID changes. Raised
in `work/questions.md` (Q-v2-07-pid).

The stop was never run: the frames 19 and 20 read `1332457`, and a
`brew services info` read after the harness still reports it `started` and
running on that PID.

## Operation view (AC-04.1) — success, error, cancelled

`op/` is `op_dump_test.go.txt` run with `DUMP_DIR` set, over a mock runner: the
run is fed to the Model as `TaskStartedMsg`, `TaskOutputMsg` and
`TaskCompletedMsg`, and the frame is taken at once, before the 2 s close. The
step lines are a transcript written to brew's `==> ` step format, not a captured
brew run. Frames are 140×40 (`.txt` and `.ansi`).

| frame | status row (line 2 of the main area) | footer, first hints |
|---|---|---|
| `op/success` | `✓ Completed in 0s · Esc to close` | `esc close  R refresh …` |
| `op/error` | `✗ Failed: brew services stop moshi-hook: exit status 1 · Esc to close` | `esc close  R refresh …` |
| `op/cancelled` | `Cancelled · Esc to close` | `esc close  R refresh …` |

## Tests (unit, mock runner)

`internal/gui/operation_dismiss_test.go`, one commit per AC:

- AC-01: `TestSuccessfulOperationClosesOnItsTimer`,
  `TestOperationTimerIgnoresANewerOperation`, `TestBatchDoesNotCloseMidBatch`,
  `TestFailedAndCancelledOperationsStartNoTimer`. The timer's message is sent
  by hand; no test waits 2 s.
- AC-02: `TestFinishedRunsEndWithTheCloseHint`, `TestLongErrorKeepsTheCloseHint`,
  `TestFinishedOperationFooterStartsWithEscClose`.
- AC-03: `TestMovingAwayClosesAFinishedOperation` (Tab, Shift+Tab, `1`, sidebar
  click), `TestMovingAwayLeavesARunningOperation`,
  `TestKeyOnTheCurrentPanelKeepsTheFinishedOperation`,
  `TestServicesInfoDrawsTheReloadOnceTheRunCloses`.

No existing test was deleted, skipped or changed.

## Not verified here

- The 2 s close and the Esc hint as seen in a real terminal (the smoke gate).
- A real `brew` batch or service write (none was run, by the brief's safety rule).
