# Receipt — runner, v2-02-layout-bugs

Commits, in AC order: AC-01 2611e00 · AC-02 a9ccbbe · AC-03 a56fd8e · AC-04 2498d26 · AC-05 23a031c · AC-06 5525dfd · AC-07 9046165 · AC-08 8d9af6b

## Evidence per AC
- AC-01 `TestEveryFrameIsTerminalSize`: 18 states × 140x40, 80x24, 200x60, 70x20 pass. On the old geometry the same test fails 50 subtests.
- AC-02 `TestModalFloatsOverFrame`: cells outside the modal box unchanged, box centred (confirm, input, menu). `TestMouseIgnoredWhileModalOpen` passes.
- AC-03 `TestModalsShowContentOnFirstFrame` (uninstall, tap-add, Brewfile, trust menus; no `Loading...`), `TestModalSurvivesResizeWhileOpen`.
- AC-04 `TestToastFloatsAboveBottomBar`: no cell outside the toast changes; toast flush right on row h-3.
- AC-05 `TestNeverLoadedPanelSaysNotLoaded`, `TestFailedFirstFetchIsLoadedNotNotLoaded`.
- AC-06 `TestLogIsSiblingBoxBelowMainPanel`, `TestClickOnLogBoxDoesNothing`, `TestMouseHitsWhatIsDrawn`. The sibling move itself landed in AC-01: the Log cannot stay inside the border and still fill the height. AC-06 adds the tests.
- AC-07 `TestBashShowsRealRuntimeDependencies` (fixture: trimmed real `brew info --json=v2 bash fish`): bash → ncurses, readline. `TestFlexStringsReadsDependencyObjects`.
- AC-08 harness `work/verification/v2-02-layout-bugs/frame_dump_test.go.txt`, run against real Homebrew: 22 frames in `after/`, all exact size (checked by script). `result.md` classifies every changed region of the nine screens.

## Gate (tree at 8d9af6b)
- `make build`: ok
- `make lint` (`go vet ./...`): clean
- `CGO_ENABLED=1 go test ./... -race -count=1`: all nine packages ok
- `make cover-check`: All coverage floors met (brew 72.9, presentation 96.2, gui 74.3, modal 50.5)
- Tests: 410 before, 425 after (`grep -rho "^func Test\w*" --include=*_test.go . | wc -l`). None deleted or skipped.

## Expectations changed (each encoded a bug)
- `TestLogCollapsedFreesMainHeight` (gui_test.go): rewritten on `screenLayout`; the Log is a sibling box.
- `TestCommandLogToggleKeepsCursorVisible` (review_r3_test.go): collapsed window 16 → 17 rows at 24 (the dead rows). Expanded 13 unchanged.
- `TestFormulaeReaderList` (formulae_test.go): uninstalled fish Dependencies `[]` → `[pcre2]` (AC-07.2).

## Found beyond B1–B7 (fixed in the AC-08 commit)
- The main list was capped by the sidebar box's rows: 80x24 Status showed 1 row, the baseline 9. The list now uses the panel height (`TestMainListFillsMainPanel`).
- The help text overflowed its box at 80x24 and lost the border and hint. It is now cut inside the box (`TestHelpKeepsBorderAndHint`).

## Decisions and contradictions
- AC-03.2: modals have a fixed width and need no size; the compositor re-centres every frame. WindowSizeMsg is still not forwarded to modals.
- AC-05: the R-refresh rule keeps `items == nil` (switching would change the error path). Search sets `loaded` on SearchDoneMsg, so it says `not loaded` before any search.
- The brief says the 140x40 main view is 38 rows; the measured baseline frame is 39 rows, 138 used cells.
- Another lane committed `docs(questions)` to wip during this slice (5107598, d155b51). Not touched.

## Open (work/questions.md)
Q-v2-hints (bottom-bar hints wider than 80), Q-v2-help (help taller than 80x24), Q-v2-small (80x24 sidebar hides inactive panels' rows), Q-v2-cask-deps (cask `depends_on` shape not read, pre-existing).

State: tree clean; `internal/gui/dump` absent; nothing pushed; STATE row added (`awaiting reviewer`).
