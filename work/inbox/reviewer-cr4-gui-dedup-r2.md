verdict: changes-required

# reviewer: cr4-gui-dedup r2 (base db7d464, HEAD c53e3e4; fixes fe72f96, 187a8cb)
## Gate and harness
- V1: `make ci` exit 0 (gofmt, vet, race, all 10 floors OK, build); `GOOS=darwin go build ./...` exit 0.
- V3 frames: my r1 harness re-run at HEAD and, fresh, at db7d464 (the Services Logs tab tails the live
  moshi-hook log, so the r1 base frames were stale). 832 frames (dark+light × 80x24+140x40); 79 differ
  raw, all on header line 0 only ("updated Ns ago"); 0 differ after normalising it.
- V3 brew args: 164 recorded write calls per version, identical (`diff` empty).
- modal.Hint.Keys: no production code reads it (grep `.Keys` outside tests: none); the frames above
  include every confirm/input/menu modal with its key line and footer: unchanged.
- V7: deadcode/staticcheck not run (fetch denied).

## Earlier findings
| id | status | evidence |
|---|---|---|
| F-r1-01 | fixed | task_build.go at fe72f96^ in a scratch worktree: TestLinesTaskCancelMidStream, ...CancelledBeforeStart, ...CancelThroughManager/{doctor,missing,vulns} all FAIL (`errCh delivered nothing within 2s`, `no TaskCompletedMsg within 2s`); pass at HEAD. My probe through the real m.runDoctor/runMissing/runVulns on a Model (MockRunner, 200 lines), CancelCurrent after the first chunk, 200 runs each with -race: every run completes (doctor/missing `context canceled`; vulns already done, nil); pre-fix: hang on doctor run 0. |
| F-r1-02 | partly | see F-r2-01. My stDoctor mutation now fails (`Status: the dispatch answers "d"…`). |

Exit paths checked: errTask sends run's error once, synchronously. linesTask: one send on every path
(fetch error, or sendLines → ctx.Err(), nil only if all lines went out uncancelled). streamTask
passes brew's errCh through: DefaultRunner (defer close + once), LoggingRunner/sudo (receive once,
then close), streamWrite (one send, close), MockRunner default (closed chans); a nil errCh is
skipped by readChunk. readChunk's loop ends only with err in or errCh nil. No path leaves it waiting.

## Mutations on the keymap tests (each restored; worktree clean)
| # | mutation | result |
|---|---|---|
| M1 | stDoctor off the Status footer | FAIL keymap_test.go:355 |
| M2 | fFetch off the Formulae help page | FAIL :125 and :355 |
| M3 | sFetch off the Search help page | FAIL :125 and :355 |
| M4 | `n` off ConfirmModal.Hints | FAIL :490 |
| M5 | `enter` off InputModal.Hints | FAIL :490 |
| M6 | keyNavigate (`j/k navigate`) off the Outdated footer | **whole `./internal/gui/...` passes** |
| M7 | keyPanelTabs (`[ ] tabs`) off the Formulae footer | **whole `./internal/gui/...` passes** |
| M8 | keyNavigate off tapsHints | caught only by tap_actions_test, not the keymap test |
Not vacuous: dispatch is parsed from handleKey's source, footers and help come from the real
footerHints/buildHelp, and modal keys from each Update's source. keyTable is still hand-listed in the
test, but a missing entry only makes the check stricter (a dispatched key it lacks fails).

## Findings
F-r2-01 (minor) — internal/gui/keymap_test.go:362-373 (onFooterAndHelp) — a key is accepted if ANY
table entry with that key satisfies "(no footer word or on the footer) and (no help line or on the
help)". j/k/down/up have keyNavigate (footer only) and keyScroll (help only); [ ]/left/right have
keyPanelTabs (footer) and keyTabs/stTabs (help). keyScroll/keyTabs sit on the Global help page, so
they always satisfy the rule and the footer entry is never required: M6 and M7 drop a footer key and
no test in the package fails. Fix must achieve: footer and help are checked separately — for each key
a panel answers, if some table entry with that key has a footer word, one such entry is on that
panel's footer; likewise for help lines. M6 and M7 must fail; M1-M5 must still fail; frames unchanged.

## New problems
None beyond F-r2-01. The fix commits touch only task_build.go, tests, and modal Hint data.
