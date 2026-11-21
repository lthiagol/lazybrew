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
