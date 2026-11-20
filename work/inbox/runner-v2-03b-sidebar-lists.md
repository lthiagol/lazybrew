# runner receipt — v2-03b-sidebar-lists

Commits, in AC order (lazygit-style: lists in the sidebar, details in the main area):
- AC-01 `0fa39ff` sidebar rows from typed data; sidebar 35%, cap 60
- AC-02 `4b16b65` Formulae List tab removed; Info is the default
- AC-03 `c30158e` other main tabs show details; typed status summary
- AC-04 `33f6348` wheel follows the pointer; main content scrolls
- AC-05 `020011f` frames, result.md, state row

Evidence per AC (test names are in `internal/gui`):
- AC-01.1 width: default 35 (config, README, config_test 30→35), cap 60, floor 20. `TestSidebarWidthRule`.
- AC-01.2 rows: `sidebar_rows.go` builds formulae, casks, outdated, taps, services from typed data; `1.2 → 1.3` in the warning role; names cut with `…`. `TestFormulaRowShowsVersionArrowAndPin`, `TestSidebarRowCutsNameWithEllipsis`, `TestOutdatedRowsCountFormulaeThenCasks`, `TestTapAndServiceRowsShowTheirState`.
- AC-01.3 highlight: surface background across the row, focused box only. `TestSelectedRowHighlightOnlyWhenFocused`.
- AC-02: tabs are named constants (`formulaInfoTab`…); `table.go` deleted; `sortFormulae` moved to `formulae.go`. `TestFormulaeDefaultTabIsInfo`, `TestFormulaInfoShowsTap`, `TestLeavesFilterKeepsFormulaeAligned`.
- AC-03.1: `renderList` and `listWindow` deleted; the four main tabs describe the selected item. `TestMainAreaShowsDetailsNotTheList`, `TestStatusDashboardTabShowsTheFullDashboard`.
- AC-03.2: batch from the sidebar; Versions shows `N selected — u to upgrade, U all`. `TestOutdatedVersionsShowsBatchSummary`.
- AC-03.3: the sidebar window is `p.offset`, set only in the Update loop. `TestSidebarListFillsItsBox`, `TestSidebarListScrollsWithCursor`, `TestFormulaeSidebarKeepsSelectedInWindow`.
- AC-03.4: `statusSummary` / `knownCount` from the dashboard load; header shows counts on first launch; unknown omitted. `TestHeaderShowsCountsOnFirstLaunch`, `TestUnknownCountIsLeftOutNotZero`, `TestStatusRowsAreCompactSummary`.
- AC-04.1: wheel routed by position; click unchanged. `TestWheelOverAnotherBoxFocusesIt`, `TestWheelOverMainScrollsItsContent`, `TestMouseWheelMovesSelection`.
- AC-04.2 keys: no key added, removed or remapped. AC-04.3: help and `panelHints` name no List tab, so unchanged.

Gate on `020011f` (HEAD, all four green):
- `make build`: exit 0. `make lint` (go vet): exit 0.
- `CGO_ENABLED=1 go test ./... -race -count=1`: every package ok.
- `make cover-check`: brew 72.9 (≥62), presentation 96.2 (≥91), gui 77.5 (≥36), modal 50.5 (≥41).
- Tests: 625 PASS (start 615). `TestEveryFrameIsTerminalSize` passes at every commit.

Changed expectations and replacements (floor kept, nothing weakened):
- Removed with named replacements: `TestFormulaVersionAndStatus`, `TestFormulaStatusStyledUsesBadges`, `TestShortenTap`, `TestRenderFormulaeTableFullPane`, `TestFormulaeTableColWidthsSum`, `TestRenderFormulaeTableEmpty`, `TestRenderContentListTabIsTable`, `TestFormulaeTableKeepsSelectedInWindow`, `TestMouseClickFormulaeTableRow`, `TestMainListFillsMainPanel`, `TestMainListScrollsWithCursor`.
- Config default 30→35 (`config_test`). The 80-column label tests pin `sidebar_width: 30` (24 cells), so their expectations are unchanged.
- Formulae tab numbers: Deps 2→1, Used By 3→2 (via constants). `TestMainBorderTitleIsBreadcrumb` now presses `]` once and expects Deps.
- `TestMouseHitsWhatIsDrawn` clicks the sidebar row, not the table row. `TestBatchSelectionShowsIndicator` drops its `renderList` block.
- Tests that read the dashboard text use `DataLoadedMsg.Dashboard`, not `Items`.

Frames: `work/verification/v2-03b-sidebar-lists/` (`before/` = v2-03 frames; `after/` = real Homebrew at `33f6348`). `result.md` reads every screen. The harness copy differs in screens 03 and 11 (one `]` fewer). Read commands only.

Open for the orchestrator (in `work/questions.md`): Q-v2-03b-a (pre-existing: batch markers draw on every sidebar list; one-line fix), -b (unknown Status counts omitted), -c (Tap Info and Trust overlap), -d (viewport removed, per-item scroll offset).

Worktree clean at report time. Branch `wip`, not pushed.
