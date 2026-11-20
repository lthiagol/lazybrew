# runner — v2-05-content-polish, fix round r1

Commits: F-r1-01 f818b81 · F-r1-02 0a03277 · F-r1-03 7f8d584 · F-r1-04 730c6ce.
F-r1-04's code (Health second, "doctor unavailable") landed in 0a03277 with
F-r1-02, since the same Health checks were retargeted; 730c6ce adds its tests
and frames.

- **F-r1-01** total = N of brew's `Upgrading N outdated package(s):` header;
  done = package steps begun (header excluded), minus one while running or
  failed; success reads total/total; no bar before the header.
  Tests: `TestUpgradeAllCountsPackagesBegun` (six cases incl. a pinned skip),
  `TestUpgradeAllBarShowsOnlyOnceTheTotalIsKnown`.
- **F-r1-02** removed `DataLoadedMsg.Dashboard`, the statusDashboard text items,
  FormatStatusDashboard, FormatDoctorStatus, OutdatedInfo, truncate and their
  tests. New `statusSummary.DoctorErr`, helpers `dashboardCards`, `healthWords`.
  Moved checks (old test:line → new):
  - fetch_timeouts_test.go:108,114 → healthWords, Summary.Problems
  - fetch_timeouts_test.go:147 → healthWords non-empty
  - outdated_count_test.go:104 → cards "Outdated:  1 package"
  - gui_test.go:1948 → "Outdated:  not loaded — press R"
  - gui_test.go:2002 → cards "1 package" (was "1 packages")
  - gui_test.go:2050 → cards "Outdated:  0 packages"
  - gui_test.go:2074 → cards "1 package" absent
  - gui_test.go:2099,2122,2237 → cards "Outdated:  1 package"
  - gui_test.go:2150 → healthWords "⚠ outdated:"
  - gui_test.go:2342,2350 → healthWords "⚠ doctor unavailable: brew doctor exited with code 2"
  - gui_test.go:2378,2385 → healthWords "⚠ 1 warning", no doctor failure
  - status_test.go:91 → Dashboard literal removed; card asserts stand
- **F-r1-03** `TestHelpEscClosesHelpBeforeCancellingATask`: first esc closes
  help, task context not cancelled; second esc cancels.
- **F-r1-04** cards: Homebrew, Health, Packages, Taps, Services; Health lines
  wrap. Tests: `TestStackedCardsPutHealthNearTheTop`. Frames 01 (both sizes)
  and 15 (140x40) re-captured; result.md fix-round section added.

Gate after 730c6ce: build, lint, `CGO_ENABLED=1 go test ./... -race -count=1`
all ok; cover-check: brew 76.1 ≥ 62, presentation 98.2 ≥ 91, gui 81.9 ≥ 36,
modal 88.3 ≥ 41.

Notes: "1 package" is a wording change. At 80×24 Packages, Taps and Services
sit below the fold.
