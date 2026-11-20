# Runner receipt — v2-03-look-foundation

Commits, AC order: AC-01 a6f22ab · AC-02 c569ee0 (+ 85f5e3f, escapes follow-up, not amended) · AC-03 0d4593e · AC-04 e56d97c · AC-05 d58fb96 · AC-06 4aaa3b0.

Evidence (test names; `go test ./... -race -v`):
- AC-01: TestPaletteTokensAreOfficialHex (26×2 exact hex), TestLightThemeIsLatte, TestApplyThemeSwitchesEveryRole, TestViewOwnsItsCanvas; the surface-detection hack is gone.
- AC-02: TestIconSetsFillEveryField, TestNerdGlyphsArePrivateUse, TestUnicodeGlyphsAreNotPrivateUse, TestIconSelectionMatrix, config/icons_test.go (default, file, env, unknown, README).
- AC-03: TestPanelTitleAndLabelSitInTheBorder, TestPanelCutsLinesNeverWraps, TestSidebarLabelStates, TestSidebarTitlesCarryIconOnlyWhenShown, TestMainBorderTitleIsBreadcrumb, TestLogBorderLabelNamesRunningTask, TestSidebarBorderClicks; renderBox removed.
- AC-04: TestHeaderIsOneFullWidthRow, TestHeaderShowsKnownCountsAndBadges, TestHeaderUpdateBadge, TestAgoTextUnits, TestHeaderOperationInTheMiddle, TestHeaderDropsBadgesNotTheName.
- AC-05: TestFooterIsOneRowAtEveryWidth (80, 140), TestFooterLeadsWithThePanelsKeys (`? more` at 80, absent at 200), TestFooterShowsTheOpenModalsKeys, TestFooterShowsHelpsKeysWhileHelpIsOpen, TestBottomBarHasLogHint.
- AC-06: before/ (v2-02 after, 22 frames); after/ (22 frames, .ansi at 140x40); result.md, one line per screen.

Gate on 4aaa3b0: make build 0 · make lint 0 · CGO_ENABLED=1 go test ./... -race -count=1 0 (9 packages ok) · make cover-check "All coverage floors met" (brew 72.9≥62, presentation 96.2≥91, gui 75.1≥36, modal 50.5≥41).

Tests: 562 PASS before (d3289ed) → 604 PASS after; 0 FAIL, 0 SKIP. Only name removed: TestLightThemeColors, replaced by TestLightThemeIsLatte (pre-M14 light values replaced, AC-01.1).

Changed expectations (old → new):
- style TestDarkThemeColors: role names; Subtle #6c7086 → #a6adc8 (AC-01.2).
- gui TestLogBorderBusyVsIdle: compares rendered boxes; logBorderStyle is gone.
- panel_loaded TestNeverLoadedPanelSaysNotLoaded: the state reads sidebarLabel.
- review_r1 TestWindowSizeSyncsVisibleRows, review_r3 TestPanelCycleSyncsVisibleRows: visibleRows == heights.
- review_r3 TestCommandLogToggleKeepsCursorVisible: switchPanel fixture; main 14 expanded; offset 17 → 28.
- log_box TestLogIsSiblingBoxBelowMainPanel, TestMouseHitsWhatIsDrawn; mouse TestMouseClickSidebarSelectsRow: title row gone, header row offset.
- gui TestBottomBarHasLogHint: 200 cols; 120 cols ends `? more`.
- modal_float TestModalFloatsOverFrame: footer row exempt (carries the modal's keys).
- help_frame TestHelpKeepsBorderAndHint, flows TestHelpFlow: close marker is `?/esc close`.
- fetch_timeouts: statusDashboard call sites take three results.

Contradictions with the brief (decided; see work/questions.md):
1. AC-02.2 "like other invalid fields": none were validated before; Load now errors on unknown gui.icons / LAZYBREW_ICONS (Q-v2-03-e).
2. AC-03.1 "never wrapping": content renderers still wrap their own text (fitBox); the primitive only cuts.
3. AC-05.2 "help's keys": the help screen's last row is the footer (`?/esc close`).
4. AC-04.1: a refresh with no operation shows in the middle (the old bar showed it).

For the reviewer (Q-v2-03-a..d): the Status count is the dashboard's line count; header counts appear only once the panels load; at 80×24 labels cut to `no…`; a running operation drops from the header at 80 cols. Pre-existing: the toast sits on the Log's bottom border at 80×24 (v2-04).

Harness: real Homebrew, read-only calls; no brew write ran. Tests use the mock runner. Worktree clean at report time.
