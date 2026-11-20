# runner — v2-05-content-polish

## Commits (AC order)
- AC-01 1192a40 · AC-02 ca14a11 · AC-03 c1b4b07 · AC-04 dfb79fb · AC-05 158e76e · AC-06 1a7a0bf

## Evidence per AC
- **AC-01** one label/value helper, `presentation.Detail` (detail.go, detail_test.go). Used by Formula/Cask Info (icon + role status), Outdated Info, Tap Info, Trust, Outdated Versions, Services Status, search info, Status Config. Tests: `TestDetailAlignsLabelsAndOmitsEmptyValues`, cut/wrap/role tests, `TestServiceStatusOmitsEmptyValues`, `TestConfigTabCutsRowsToTheWidth`.
- **AC-02** deps with installed icons (depviews.go; `depviews_test.go`); Files count + muted dirs. Cask `depends_on` object parsed in brew (`Cask.UnmarshalJSON`, `parseCaskDepends`). Fixture `internal/brew/testdata/casks_installed.json` is real `brew info --json=v2 --cask --installed` output: Google Cloud CLI → `DependsOnFormulae = [python@3.14]`. Tests: `TestCaskDependsOnRealBrewShape`, `TestParseCaskDependsShapes`.
- **AC-03** five cards (dashboard.go, dashboard_test.go); gauge; Health states; Doctor blocks (`doctorView`).
- **AC-04** scrollable help (help.go, help_scroll_test.go): keys, wheel, `1–37 of 40` hint, two columns at ≥80 text width, whole-page flow. Every line reachable at 80×24 (`TestHelpScrollReachesEveryLineAt80x24`).
- **AC-05** operation.go: steps, elapsed (`formatElapsed`), final states, batch bar (`charm.land/bubbles/v2/progress`). Tests in operation_test.go, including the transcript-shaped `brew upgrade` fixture (no upgrade run).
- **AC-06** frames 01–18 in `work/verification/v2-05-content-polish/{before,after}`, `result.md`, STATE row.

## Gate (last run, after AC-06 code state)
- `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` all packages ok
- `make cover-check`: brew 76.1 ≥ 62 · presentation 97.9 ≥ 91 · gui 81.5 ≥ 36 · modal 88.3 ≥ 41

## Changed expectations and signatures (named, no test deleted)
- `TestTapTrustTabShowsTheTrustStory`: fixture gained a Remote so `Remote:` still shows (empty values are omitted now).
- `TestStatusDashboardTabShowsTheFullDashboard`: checks the card titles and summary values, not the old text lines.
- `TestOperationCompletedNotCancelledOnRealError`: `Error:` → `Failed:` (brief's `✗ Failed: <err>`).
- Signatures: `FormatFormulaInfo`/`FormatCaskInfo` (+icons), `tapTrustStory` (+width), `fetchTabContentCmd` (+icons); call sites updated.
- go.mod/go.sum: `github.com/charmbracelet/harmonica v0.2.0 // indirect` (progress package's dependency).

## Contradictions and what the next session must know
- STATE row: the brief gives `—` in the commits column, but work/README.md says the column lists commits. Followed the brief; commit ids are in this receipt.
- Dashboard text lines (`statusDashboard` items, `DataLoadedMsg.Dashboard`, `FormatStatusDashboard`) are still built and read only by tests (~15 expectations). No longer drawn. Removing them is a follow-up with those tests.
- `presentation.truncate` is now used only by its tests.
- Cask requirement shapes (`macos` maps, `arch` objects) are not in any real output on this machine; the parser follows brew's documented shapes and is covered only by synthetic tests.
- upgrade-all progress counts brew's `==> Upgrading ` lines, from brew's documented shape, not from an observed upgrade.
- Esc while help is open now closes help (before, it cancelled a running task). Other keys still pass through while help is open, as before.
- At 80×24 the dashboard cards stack; Services and Health sit below the fold and need scrolling. Nothing overflows sideways.
- Operation view has no frames (tests only).
