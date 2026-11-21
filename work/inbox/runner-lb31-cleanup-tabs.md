# Receipt — runner lb31-cleanup-tabs (LB-29)

## Commits (AC order)
- AC-01 `baf820b` brew layer: read-only cleanup and autoremove previews
- AC-02 `8534342` Cleanup and Unused Status tabs
- AC-03 `6b8284c` `c` / `A` open the tabs, then confirm the real run

## Evidence
**AC-01.** Real output on this machine: `brew cleanup -n` and `brew autoremove -n` both
printed nothing and exited 0, so there was no listing to anonymise. Fixtures follow brew's
source (`Library/Homebrew/cleanup.rb`: `Would remove:`, `Would prune`, `Would remove (broken
link)`, `==> Would autoremove N unneeded formulae:`, the closing total).
Tests first: `go test ./internal/brew/ -run 'CleanupPreview|AutoremovePreview|ParseSize'` failed
with `diag.CleanupPreview undefined` and `undefined: CleanupItem` (build failed). After the
change: 9 tests PASS. Exit 1 is treated as a result when it has output, as doctor and missing
do; exit 1 with no output is an error. Exit 1 was not observed here.

**AC-02.** Tabs `Cleanup`, `Unused` read when opened, kept until R. Typed cache, not tabContent
text (deviation from the brief): a middle cut needs the width of the draw. y draws at copyWidth,
so it is whole (`TestCleanupTabCopyIsUncut`). Tests: `TestCleanupViewTable`, `TestCutMiddle`,
`TestPreviewTabFrames` (2 tabs x data/empty/error x 80x24 and 140x40 x both themes, 24 subtests),
`TestCleanupTabReadsOnOpen`, `TestStatusTabRowWhole140x40`. `TestNeedsTabFetchTable` now lists
the two tabs (they are brew reads).
**80x24 not met:** names need 60 cells, the bar has about 50 (`… Vulns  Clea` is cut).
Single-space separators need 52. Nothing abbreviated or clipped on purpose. Q-lb31-tabs-80.

**AC-03.** `c`/`A` off the tab open it and read brew; on the tab they confirm. Confirm text:
`Run brew cleanup, freeing about 230.3MB?` and `Uninstall 2 unused formulae?`.
Removed: `brewCleanup`, `runAutoremove`, the `cleanup-preview` / `autoremove-preview` cases in
`finishedMutation`, the dry-run parameter of `DiagnosticsWriter.Cleanup` / `Autoremove`.
Tests removed: `TestBrewCleanupFlow`, `TestRunAutoremove` (replaced by
`TestPreviewKeysOpenTabThenConfirm`). Tests changed: the cleanup and autoremove cases of
`confirm_kind_test.go` (they now open through the key path); `writes_invalidate_test.go` (the
`false` argument dropped). No golden file changed. teatest flows `TestCleanupTabFlow` and
`TestAutoremoveTabFlow`: open, confirm, `y`, one `brew cleanup` / `brew autoremove` streamed,
tab re-reads and shows `Nothing to clean up` / `Nothing unused`. Both PASS.

## Gate
- `GOFLAGS=-p=1 make ci`: exit 0. fmt-check, vet, `go test ./... -race -count=1` (all ok),
  cover-check (`All coverage floors met`, floors unchanged), build.
- `GOOS=darwin go build ./...`: ok.
- `deadcode ./cmd/...`: no output.
- `staticcheck ./...`: 5 findings, none in lb31 files (keymap_test.go:39, 498 ParseDir;
  launch_check_test.go:180, 220, 453 SA4006); those files are untouched.
- Per commit: `make test-fast PKG=./internal/brew/...` and `./internal/gui/...` ok.

## Questions filed
- Q-lb31-tabs-80: seven tabs at 80x24 (open; two options, first recommended).
- Q-lb31-cleanup-autoremove: `brew cleanup` also uninstalls unused formulae. Shown as a note;
  the plain run is kept. Ruling needed.

## Seen, not done
- Unused rows show names only: brew prints no versions in autoremove output.
- AGENTS.md "coverage tracking file": none in the repo; no row added. `cleanup -n` and
  `autoremove -n` are reads of commands already wrapped, not new command support.
- The preview reads use `fetchTabContentTimeout` (15 s). A very slow `cleanup -n` on a large
  cache shows the timeout error in the tab.
- Path cut: at a `/` where one is near; with no `/` near, the cut falls where the width ends.

## Fix r1 — Q-lb31-tabs-80 (code commit 9170eab)
- Window: `tabWindowIn` grows from the active tab, alternating right and left, while the shown whole tabs plus one `‹`/`›` per cut side fit. Rows that fit are drawn unwindowed, as before.
- 80x24 Status (content 48 cells; the row is 60): Dashboard shows ` Dashboard  Config  Doctor  Missing  Vulns ›`; Vulns shows `‹ Doctor  Missing  Vulns  Cleanup  Unused `. Underline spans the active name in both.
- Red before the fix: `TestTabWindowEveryTabAt80x24` ("names row is 60 cells, want at most 48"), `TestTabWindowNarrow` ("60 cells, want at most 40"), `TestStatusTabWindowRows`, `TestStatusTabWindowClicks` ("click on the left marker selected tab 0, want 3").
- 60 columns: Status (60) and Formulae (47) fit whole with no marker: `TestTabWindowAt60Columns` (green before and after).
- Clicks: `handleTabBarClick` maps through the window; a click on a marker calls `selectTab` for the neighbouring tab.
- Tests changed: `TestStatusTabRowShowsEveryName` is now `TestStatusTabRowShowsTheActiveTab` (it required Dashboard at 80x24, which the window hides with Vulns active). `TestSlidingTabBarAtRestIsTheRestingBar`: signature only. No golden changed.
- Gate: `GOFLAGS=-p=1 make ci` exit 0 (fmt, vet, race tests, floors met, build). `GOOS=darwin go build ./...` ok. deadcode: no output. staticcheck: the same five pre-existing findings, none in lb31 files.

## Fix r2 — F-r1-01, F-r1-02 (commits 9a5b520, 66f79c8; rename 2ea1e4b)
- F-r1-01, partly met. `cutMiddle` moves the cut to a `/` only within `cutSnap` (4) cells, else cuts at the width; a cut already after a `/` keeps that segment. At 80x24 (Path 22 cells): `<prefix>/Cellar/llvm/17.0.6` → `<prefix>/…llvm/17.0.6` (was `…/17.0.6`); python keeps `…/3.12.4_1`. Red on HEAD: `TestCutMiddle`, `TestCutMiddleKeepsWholeSegmentAtCut`, `TestCleanupViewRowsAt80x24`.
- Not met: the two bottle files still both render `~/Library/….bottle.tar.gz` at 80x24 (they differ before the last 11 cells). The cut rule cannot separate them; no test asserts it. Question Q-lb31-cleanup-path-cut.
- F-r1-02: `TestCleanupPreviewReadsThousandsSeparators` (`1,700 files` → 1700). Mutation `atoiComma` without `ReplaceAll` → `Files:0, want 1700`, FAIL; source restored.
- Rename: `TestTabWindowAt60Columns` → `TestTabWindowAt60CellContent`.
- Gate: `GOFLAGS=-p=1 make ci` exit 0 (fmt, vet, race, floors met, build); `GOOS=darwin go build ./...` ok; `make test-fast` on brew and gui ok.
- Commits 9a5b520, 66f79c8, 2ea1e4b lack the Co-Authored-By trailer (missed); not amended, AGENTS.md forbids amending unless asked.
