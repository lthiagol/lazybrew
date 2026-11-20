# runner — v2-05-content-polish, fix round r1

## Commits (finding order)
- F-r1-01 f818b81 · F-r1-02 0a03277 · F-r1-03 7f8d584 · F-r1-04 730c6ce
- F-r1-04's code (Health second; "doctor unavailable") landed in 0a03277, because
  the Health checks were retargeted in that same pass. 730c6ce adds its tests and
  the re-captured frames. Not amended: history stays as it was committed.

## F-r1-01 — upgrade-all counts packages, not brew's header
- Total = N of the header `Upgrading N outdated package(s):`; pinned packages are
  not in it. Done = begun package steps (header excluded), minus one while running
  or failed; success reads total/total. No bar until the header comes.
- Tests: `TestUpgradeAllCountsPackagesBegun` (six cases: no header, one package
  running/finished, three-of-three mid-run/failed/finished with a pinned skip),
  `TestUpgradeAllBarShowsOnlyOnceTheTotalIsKnown`.

## F-r1-02 — dead dashboard text removed; checks retargeted
- Removed: `DataLoadedMsg.Dashboard`, statusDashboard's text items,
  `FormatStatusDashboard`, `FormatDoctorStatus`, `OutdatedInfo`, `truncate`, and
  their tests (presentation: TestFormatStatusDashboard, ...OutdatedNotLoaded,
  ...OutdatedKnownZero, TestFormatDoctorStatus, TestTruncateShort/Long/UTF8,
  TestSnapshotDashboard, TestSnapshotDashboardOutdatedNotLoaded).
- New: `statusSummary.DoctorErr` (a failed doctor's reason); helpers
  `dashboardCards` and `healthWords`. Presentation coverage 98.2 ≥ 91.
- Moved checks (old → new, all in gui tests):
  - fetch_timeouts_test.go:108/114 (TestFetchStatusDataTimesOut) → healthWords + Summary.Problems
  - fetch_timeouts_test.go:147 (TestFetchStatusDataRespectsTotalBudget) → healthWords non-empty
  - outdated_count_test.go:104 → dashboardCards "Outdated:  1 package"
  - gui_test.go:1948 → "Outdated:  not loaded — press R" (was "(press R to load)")
  - gui_test.go:2002 → "1 package" (was "1 packages"; now singular)
  - gui_test.go:2050 → "Outdated:  0 packages" on cards
  - gui_test.go:2074 → "1 package" absent on cards
  - gui_test.go:2099, 2122, 2237 → "Outdated:  1 package" on cards
  - gui_test.go:2150 → "⚠ outdated:" in healthWords
  - gui_test.go:2342/2350 (TestStatusDashboardSurfacesDoctorError) → healthWords "⚠ doctor unavailable: brew doctor exited with code 2"; "ready to brew" absent
  - gui_test.go:2378/2385 (TestStatusDashboardDoesNotWarnOnDoctorExitCode1) → healthWords "⚠ 1 warning"; no doctor failure
  - status_test.go:91 literal Dashboard field removed; card assertions stand

## F-r1-03 — Esc closes help before cancelling
- `TestHelpEscClosesHelpBeforeCancellingATask`: a running task's context is
  captured from its Run; first Esc closes help and the context is not cancelled;
  second Esc cancels it. (The old behaviour would fail the first check.)

## F-r1-04 — Health near the top when stacked; doctor failure named
- Card order is Homebrew, Health, Packages, Taps, Services. Health lines wrap to
  the card, so a timeout reads whole.
- Tests: `TestStackedCardsPutHealthNearTheTop` (Health before Packages at 50
  columns; title on row 4), `TestSurfacesDoctorError` now checks the reason.
- Frames 01 (140x40, 80x24) and 15 (140x40) re-captured; `after/` and result.md
  updated (fix round r1 section). Other frames byte-identical.
- Harness run from a temporary package, removed afterwards.

## Gate (after 730c6ce)
- `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` all ok
- `make cover-check`: brew 76.1 ≥ 62 · presentation 98.2 ≥ 91 · gui 81.9 ≥ 36 · modal 88.3 ≥ 41

## Notes for the next session
- The singular "1 package" is a small wording change in the cards.
- At 80×24 Packages, Taps and Services are below the fold and need scrolling.
