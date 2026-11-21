verdict: changes-required

# Review: cr5-dead-code, round r1 (base f22848f, HEAD 3223bc3)

**Gate (V1).** fmt-check, vet, `go test -race ./...` (CGO_ENABLED=1), cover-check (all 10 floors met, gui 89.9, style 89.5, task 96.2), build: all pass; `GOOS=darwin go build ./...` ok. I ran it in pieces: the whole `make ci` with GOTMPDIR on the big disk fails 3 `TestLogs*` tests (a long temp path is elided in the header). This comes from my environment, not the slice: they pass with the default TMPDIR. No change to `scripts/`, `Makefile` or `.github`.
**Tools (V3).** `deadcode ./cmd/lazybrew` → only `brew.SetDebug`; `deadcode -test ./...` → nothing; `staticcheck ./...` → only SA1019 at keymap_test.go:38 and :434 (accepted). Matches the receipt.

| AC | Verdict | Evidence |
|---|---|---|
| 01 | ok | `go tool nm bin/lazybrew \| grep -c MockRunner` = 0; brewtest imports nothing from brew, so there is no cycle; `assertAnError` is gone and `ErrTest` is used everywhere |
| 02 | ok | items hold bare names (Status/Search keep their drawn text, documented on `panelData.items`); Format{Formula,Cask,Tap,Service,Outdated*} and padRight are gone; FormatFormulaInfo/FormatCaskInfo stay |
| 03 | ok | every deleted symbol grepped at f22848f: SearchDesc, GetTapTrustStatus, TapWithURL, Taps.Get, InvalidateAll had test-only callers. TrustStatus enum: only GetTapTrustStatus and ParseTrustStatus used it, so it was dead (TapList.TrustUnknown is a different field and is kept). fetchTapInfo: its only caller was Taps.Get. Cask.ConflictsWith was in fact filled by the JSON decode, but nothing read it outside types_test, so it is dead UI. No `p` on Casks: keys.go gates pin on PanelFormulae; my harness shows `p` on Casks makes no brew call at base or at HEAD |
| 04 | ok | the Taps arms of executeMenuAction were unreachable: trust-menu has its own case, and Brewfile (`B`) opens only on Status. Both `searchInput.Focused()` checks were dead: keys.go:33 handles esc/enter first when the input is focused. TestDepsTabShowsWithoutFetch proves the Deps tab draws without a fetch. task.Status deleted; tests assert on TaskCompletedMsg.Err |
| 05 | ok | the 8 styles have no production reader; the tests moved to SurfaceColor/PanelTitleActive/SubtleText, all read by production; the SA9003 empty branch is removed |
| 06 | ok | `--version` → `v1.0.0-rc1.0.2026…-3223bc31cd56`; `-X main.version=9.9.9` → `9.9.9`; `-buildvcs=false` → `dev`; Makefile has no -ldflags |
| 07 | ok* | d0a6a86 is 7 pure renames at 100% similarity. SA4006 checked at 4 sites: preload_test `inFlight==0` (a real assertion), sidebar_rows re-reads p after leaves-off, hot_path keeps the AllocsPerRun result and asserts it, toastName compares against the first call. The SA4017/SA4000 pair is no longer vacuous. *See F-r1-03 |

**Deleted or changed tests (V4).** Each belongs to code this slice deletes, and the behaviour it guarded is still covered:
- TestFetchTabContentCmdTimesOut now covers the timeout through Used By (`uses`).
- TestCancelledTrustMenuDoesNotPoisonBrewfileMenu still asserts that a cancelled trust menu is followed by a `brewfile-dump` task.
- TestNeedsTabFetchShifted flips Deps to "no fetch", as intended.
- TestMutation/ProgressMessageTypes were field-echo tests for deleted fields.
- TestFetchDepsTabReturnsContent covered the deleted fetch arm.
- task/types_test.go tested the deleted Status.String.
- TestManager_CancelledCompletion still asserts `errors.Is(Err, context.Canceled)`.

**Frames and brew args (V8).** I wrote my own harness (an external `gui_test` with a fake runner and fixtures, the Services log frozen to a fixed file). It covers 80x24 and 140x40, dark and light, all 7 panels with 6 tab steps each, and 29 action keys per panel with each modal step (key, `y`, enter, esc, esc): about 1100 frames per combination, including help, filter, confirms, menus, inputs and the operation overlay. Plain-text and raw-ANSI frames are **identical** between f22848f and d0a6a86 in all four combinations, and stable across two HEAD runs. Recorded brew args, sorted per scenario, are **identical**: install/uninstall/zap/reinstall/upgrade/fetch for formula and cask, unpin, services stop/restart/run/cleanup, tap --repair, cleanup, autoremove, bundle dump. No golden or testdata is in the diff.

**Judgements.**
- (4) newestToast moved to toast_stack_test.go: fine. It is a test helper that 15 tests read, and it is out of the binary. parseSteps kept as the oracle: fine. It folds addStepLine, and the test compares the Operation's incremental AppendLine state against that fold, which still checks the wiring.
- (5) brew.SetDebug kept: acceptable. app/debug_log_test.go needs a cross-package logger reset, and only `_test` code calls it. Note: its `true` branch logs to stderr, the C13 hazard. A later slice could narrow it to a reset-only helper.
- Note (not a finding): TestSingleflightPropagatesError now waits on a 20 ms sleep instead of the `waiters` poll that v2-11 AC-02 added to avoid wall-clock waits. If the second goroutine is not queued within 20 ms, the test fails. The brief ordered `waiters` removed, and without a hook no race-free check exists, so I accept the sleep. Suggest a backlog line.

## Findings

F-r1-01 (minor): `internal/gui/render.go:113`. Comments name deleted code. `contentWidthIn` is documented as "mainContentWidth for the layout l", but mainContentWidth was deleted in c978a89, together with the only explanation of what the width means. The same happened to `renderMainPanelIn` (:77): its doc lost the "main column, exactly l.mainW × l.bodyH" text. Fix: fold the deleted doc text into these comments so that no comment names a removed function.

F-r1-02 (minor): `internal/gui/task/manager_test.go:409` says the test "locks the StatusCancelled wiring … recorded as cancelled", but StatusCancelled is gone. `internal/gui/gui_test.go:1257/1275/1391/1415` still talk about "the fast path" and keep the name TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply; the fast path was renderCachedTabContent, deleted in c978a89. Fix: reword them to describe what the tests assert now (the completion carries context.Canceled; Deps needs no fetch).

F-r1-03 (minor): the receipt does not give the full old → new rename mapping that AC-07.2 asks for. It names only 4 of the 7 files, and only under "seen, not done". Missing: review_r2→batch_upgrade_test.go, runner_stream_r1→runner_stream_errors_test.go, runner_stream_r2→runner_stream_cancel_test.go. Fix: list all 7 in the receipt or in the fix-round receipt.
