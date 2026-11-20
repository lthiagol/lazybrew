# Receipt: runner-v2-09-main-layout

Slice `v2-09-main-layout`, brief `work/briefs/v2-09-main-layout.md`. Full
evidence in `work/verification/v2-09-main-layout/result.md`.

## Commits, in AC order

- AC-01 fd1bde3 — main area padding (mainPadX, mainContentWidth, mouse offset)
- AC-02 12a3b1b — table.go helper; Deps, Used By, Casks Deps, Search deps, Taps Formulae, Files
- AC-03 95766a6 — presentation.Info split; 7f6e93b — left column sized to its fields (refinement; the first commit's fixed 2/5 cut the Unit file path. Not amended, per the git rules.)
- AC-04 539a515 — logHeight: 3 idle-empty / min(bodyH/5,7) idle-with-lines / bodyH/5 running / 0 collapsed
- AC-05 d61ea76 — harness, before/, after/ (140x40, 80x24, 200x60), result, STATE row, Q-v2-09-a

## Evidence per AC

- AC-01: `TestMainAreaHasInnerPadding` — every body row of the main box has a blank column at both inner edges, every panel state, 140x40 and 80x24. The tab row shares the padding.
- AC-02: `TestTable*`, `TestPackageRows*` pass. Drop order: Description (drop 2) then Version (drop 1); icon and name never dropped. 80x24 Deps frame: Description gone, Version kept.
- AC-03: `presentation/info_test.go`, `info_split_test.go`. Split at content width ≥ 98 (box inner ≥ 100). 200x60 frames show the split on 03, 06, 17, 19.
- AC-04: `TestLogBoxHeightByState` at 80x24, 140x40, 200x60 (all four states); main + Log = body height.
- AC-05: harness run, 48 s, no brew write. Sidebar identical in every before/after pair.

## Gate (final tree, code identical to 7f6e93b; d61ea76 is docs only)

- `make build` ok; `make lint` (go vet) ok; `CGO_ENABLED=1 go test ./... -race -count=1` ok; `make cover-check`: "All coverage floors met"

## Expectations changed (named, not weakened)

- depviews_test: `"✓ ncurses"` checks → `hasRow` (icon and name are separate columns); renderFiles takes width; Files header is line 2.
- gui_test: Deps tests read `renderContent`, not the stored `tabContent` string.
- services_view_test: at width 100 the log path wraps; checks both halves.
- review_r3_test: idle, empty expanded log → main 15 rows, offset 15 (was 14 / 16).
- scroll_test: fixture 60 → 90 lines (main area is taller).

## Judgement calls and open items

- The Log box is not padded: its text touches its left border. Not a main-area view per the brief.
- Q-v2-09-a: the ≥ 100 threshold never fires at 140x40 with the default sidebar (inner 89). Implemented as written; the 200x60 set shows the split. Orchestrator to rule.
- STATE row uses `—` for commits, as the brief's exact row says.
- Taps Formulae and Files are not in the harness frames; unit tests only.
