# Receipt — lb25-search-polish (runner)

Commits, one per AC, in order: AC-01 `b2885f6`, AC-02 `4c14fbf`, AC-03 `897f98f`.
Plan commit (questions, STATE row): `dfa7d36`.

## AC-01 — ✔ is the installed mark
- Test first, `go test ./internal/gui/ -run TestInstalledGlyphIsBrewsCheckmark`:
  `nerd set: Installed = "", want ✔ (U+2714)`; `unicode set: Installed = "✓", want ✔`.
- After: pass. Test also asserts `lipgloss.Width == 1` in both sets.
- `show_icons: false` still draws nothing (`TestIconSelectionMatrix` passes, unchanged).
- Invariant edited: `TestNerdGlyphsArePrivateUse` skips the `Installed` field (✔ is not a Nerd glyph).
- Goldens/expectations that changed (✓ → ✔, or the cell now ✔):
  depviews_test.go: TestFormulaDepsShowInstalledState, TestDepIconsWaitForTheInstalledList,
  TestUsedByListsInstalledDependents, TestSearchResultDependenciesCheckInstalled;
  main_content_cache_test.go: TestUsesTabFollowsTheInstalledList;
  operation_test.go: TestOperationStepMarks (finished-step ticks);
  table_test.go: TestPackageRowsCarryLoadedDataOnly;
  presentation/detail_test.go: TestFormulaStatusRowsCarryIconAndWord, TestCaskStatusNotesAutoUpdate;
  flows/uninstall_test.go: TestUninstallFlow waits for `▸ ✔ testformula`.
- Unchanged: Outdated ↑, tap ✓, service glyphs, errors, `Completed in` row (uses `Icons.OK`, still ✓).
- Not verified: a real terminal's drawing width for ✔. Only `lipgloss.Width` was checked.

## AC-02 — Search rows drawn from cells
- Test first, `search_rows_test.go:47`: `dark w=30: row should be "  ✔ " then the name: "  alpha-cask                  "`
  (old rows were a single string with no type and the word `installed`).
- After: TestSearchRowsAlignOnTheCellSystem, TestSearchRowsWithIconsOffDrawNoMark,
  TestSearchLongNameKeepsType, TestSearchFrameAlignsRows, TestSearchSelectedRowHighlightsWhenFocused pass.
- Frame, 140x40, dark (from the harness):
  `│▸ ✔ alpha-tool-x                           cask│`
  `│    beta-tool                           formula│`
- Long name at 80x24 draws `delta-xxxxxxx… formula`. Search box shows 2 rows at 80x24, 4 at 140x40.
- Both themes checked in the frame test. Items are bare names; `searchCells` draws the row.
- Expectations changed: search_installed_test.go TestSearchRowsCarryTypeAndInstalledState and
  TestSearchInstalledListUnreadableKeepsTypes (now check `Raw` and cells); gui_test.go
  TestSearchEnterShowsInfo, TestSearchInstallKeybind (item fixtures); flows/install_test.go
  waits for `▸   ripgrep` (blank mark cell). README marks line updated.
- Decision: a row with no type shows `unknown` (was blank). Filed Q-lb25-unknown-type.

## AC-03 — `I` removed on Search
- Test first: `keymap_test.go:386: Search answers I; it must answer only i`;
  `gui_misc_test.go:611: I on Search must not open the install confirm (lb25 AC-03)`.
- After: both pass. TestLongCaskNameReachesSearchInstall now presses `i`.
- Removed: `case "I"` in keys.go, `sInstallCap` (keybindings.go and keyTable), its Search footer word.
- README and help had no `I` entry; nothing to remove there.
- Unfocused, `I` now types into the box (Q-lb25-I-types).

## Gate
- Per commit: `make test-fast PKG=./internal/gui/...` ok after each AC (3 of 3).
- `make ci` run as `GOFLAGS=-p=1 make ci`: exit 0, `All coverage floors met`.
  Without `-p=1` the link failed: `mapping output file failed: no space left on device`
  (/tmp is a 2 GB tmpfs at 93%, other sessions' scratch). TMPDIR on the root disk
  instead broke the Logs header tests, which expect the full temp path.
- `GOOS=darwin go build ./...`: ok.
- deadcode `./cmd/...`: no output.
- staticcheck `./...`: 5 findings, all pre-existing: keymap_test.go ParseDir ×2
  (base lines 39, 431); launch_check_test.go:180, 220, 453 SA4006 (file untouched).

## Seen, not done / questions
- Questions in work/questions.md: Q-lb25-unknown-type, Q-lb25-step-tick, Q-lb25-I-types.
- show_icons off: Search rows carry no mark, as Formulae rows do.
- Not checked: real-terminal width of ✔ (macOS); the TUI was not launched against real brew.
- No brew write commands were run.

## Fix r1 — Completed row ✔ (commit 51bd21e, [Q-lb25-step-tick])
- Test first: `TestCompletedRowMatchesTheStepTicks` failed at 80x24 and 140x40: `operation_dismiss_test.go:355: the Completed row should carry ✔`; the row read `✓ Completed in 34s`.
- Fix: `operation.go:152` draws the row with `m.icons.Installed` (smaller than changing `OK`, which also draws the trusted tap and toasts). `Icons.OK` stays ✓ at `sidebar_rows.go:196` (trusted tap), `toasts.go:64` (toast), `dashboard.go:111`, `:198`, `command_log.go:153`. Question: Q-lb25-ok-uses.
- Goldens changed: `operation_test.go` TestOperationFinalStates; `operation_dismiss_test.go` TestFinishedRunsEndWithTheCloseHint (success row).
- Gate: `make test-fast PKG=./internal/gui/...` ok; `GOFLAGS=-p=1 make ci` exit 0, "All coverage floors met"; `GOOS=darwin go build ./...` ok. `-p=1` because /tmp is at 93% (link failed otherwise).
