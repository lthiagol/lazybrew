# runner receipt — cr3-render-perf

Commits, AC order: AC-01 d8650ab · AC-02 ec0df54 · AC-03 99b737b · AC-04 c043ba2 · AC-05 8da6ed8 · AC-06 1ebf46c · AC-07 5010219

## Evidence per AC
- **AC-01 [P4, C11]** `parsePackageInfo` runs once, in the `SearchInfoLoadedMsg` handler (`parseSearchInfo`). Brew's error is the `Error:` line of `BrewExitError.Stderr` (`searchInfoErr`). Red before the fix: `search_info_test.go:189: … got: Parse error: invalid character 'E' looking for beginning of value`.
  Bench `SearchPanelView` 140x40: 2431008 → 2466669 ns/op, 2802 → 2775 allocs (payload is small; the per-frame decode is gone).
- **AC-02 [P5]** Step state kept on `Operation` by `addStepLine` in `AppendLine`; body and progress read it; `Lines` capped at `operationMaxLines` = 5000; body built from the end, only the rows shown. `TestOperationStepsMatchParseSteps` checks after every line; `TestOperationLinesAreCapped` gives 1999/2000 after the cap.
  Bench `OperationRenderPerLine` (10 000 lines, render after each): 131.5 s, 29.2 GB, 726 M allocs → 2.61 s, 205 MB, 5.48 M allocs.
- **AC-03 [P5, C10]** `TaskOutputMsg.Lines`, batches ≤ 64 without blocking; after the error, `readChunk` reads `outCh` to its close (or ctx done). Red on pre-fix code (single-line message, 500 runs, `-race`): `manager_test.go:476: iteration 3: 19 of 20 lines arrived before the completion`. Green after: `-count=20 -race` passes.
  Bench `TaskOutput`: 10 000 → 248.9 msgs/run; 1.30 → 1.27 ms/op (producer-bound here; the saving is per Update+View).
- **AC-04 [P6, P7]** Files/Used By wrapped rows cached (`wrapCache`, key contentKey|width, valid for tab source, installed generation and theme); installed list built by `indexInstalled` on landing. `TestScrollingFilesTabRendersOnce`: 50 PgDn + G + g → 0 extra renders.
  Bench `FilesTabPageDown` (5000 files): 122.9 ms, 48.2 MB, 222 533 allocs → 2.95 ms, 0.99 MB, 2 967 allocs.
- **AC-05 [P8]** `CommandLog.Len`/`Last`; `logHeight` uses `Len`, `View` uses `Last(h)`; one `screenLayout` per render (`renderMainPanelIn`); sidebar rows take `bodyRows()`. `TestCommandLogLenAndLast`: empty, partial, wrapped, 0, negative.
  Bench `ViewCommandLog`: 2281445 → 2284050 ns/op. No change: the log holds 20 entries.
- **AC-06 [P9, P10]** `TextStyle`, `MutedText`, `SuccessText`, `WarningText`, `ErrorText` set in `ApplyTheme`; 24 of 28 per-render sites replaced (left: 2× non-bold `AccentColor`, 2× box edge once per panel). Sidebar separator drawn once per row, panel edge once per panel row. `TestSharedTextStylesDrawWhatTheyReplace` (both themes).
  Bench `FormulaePanelView` (300 formulae): 2646606 → 2496624 ns/op, 3365 → 2571 allocs.
- **AC-07 [P11, P12, P13]** `needsTabFetch` is a switch (`AllocsPerRun` = 0 test). `helpRows` memo by width and theme. `naturalWidth` replaces the `Detail(left, 0)` render in `Info` (test matches `lipgloss.Width(Detail(f, 0))`, incl. multi-line and tab). Toast motion keyed by `Toast.ID()`. `sortFormulae` decorate-sort + `SortStableFunc`; `TestSortFormulaeKeepsTheOrder` against the old comparator.
  Bench `HelpScreenView`: 994543 → 794176 ns/op, 1816 → 485 allocs. `SortFormulae` (1000): 1072524 → 922239 ns/op, 22331 → 1001 allocs (bytes 179 → 393 KB).

## Gate
- `make ci` exit 0: fmt-check, vet, `go test ./... -race -count=1` (12 packages ok), cover-check (gui 87.5% ≥ 83%, all floors met), build.
- `GOOS=darwin go build ./...` ok.
- Frames: 152 frames (80x24, 140x40, dark and light, all frame states plus added long-operation, scrolled-Files and Used By states), dumped at dc4df13 and at HEAD. Only the 4 `search_info_error` frames differ: C11 text, `Parse error: invalid character 'E'…` → `Error: No available formula with the name "zz".`

## Questions and seen, not done
- No questions filed; no contradiction between brief and repo found.
- AC-02 tests needed the new field, so their pre-fix red is a compile error, not a behaviour failure.
- Not done: `screenLayout` still runs once per key event (`mouse.go`, `pageRows`, `contentLast`); the Logs tab is not cached; `motionSidebar` still runs in `Update`.
- Not done: step titles are not capped (one per `==>` line); only each step's output is.
- Note: the C10 drain relies on producers closing `outCh` (`runner.go` and `closedCh` do); ctx cancel is the escape.
- Note: `scrolledContent` now joins only the window's rows. h = 0 cannot happen at the 24-row minimum.
