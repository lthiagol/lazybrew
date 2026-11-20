# runner — v2-05-content-polish, fix round r2

Commits: F-r2-01 0831cd8 · F-r2-02 da7d80e · F-r2-03 c9acd20.

- **F-r2-01** total = sum of N over every header that names packages:
  `Upgrading N outdated package(s):` (formulae, then casks) and
  `Upgrading N dependent(s) of upgraded formula(e):`. Headers are not package
  steps; `Upgrading <x> dependency: <y> (v)` is excluded (no header names it).
  Strings from Homebrew's source: cmd/upgrade.rb:458, cask/upgrade.rb:102,434,
  upgrade.rb:246,400,424. Tests: `TestUpgradeAllCountsFormulaeThenCask` (2/3
  mid-run, 3/3 end), `TestUpgradeAllCountsDependents` (1/2, 2/2),
  `TestUpgradeAllBarNeverFullWhileRunning` (every prefix),
  `TestUpgradeAllIgnoresDependencyInstalls`. r1 cases still pass.
- **F-r2-02** singular checks use `Outdated: +1 package\b` (`oneOutdated`).
  Mutation M4 (one package read as "packages") now fails five tests:
  TestStatusDashboardReportsCachedOutdatedCount,
  TestStatusDashboardExplicitLoadShellsOutdatedOnce,
  TestStatusDashboardExplicitLoadShellsOneCallPerOutdatedClass,
  TestRefreshMsgWithLoadOutdatedShellsOutdated, TestStatusRefreshKeepsOutdatedCache.
  Mutation reverted; worktree clean.
- **F-r2-03** result.md lines 17 and 42 describe the current card order and the
  80×24 fold; the fix-round section now gives Health on rows 9–10 (was "row 5").

Gate at da7d80e: `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race
-count=1` all ok; cover-check brew 76.1 ≥ 62, presentation 98.2 ≥ 91, gui
81.9 ≥ 36, modal 88.3 ≥ 41. c9acd20 is documentation only.

Notes: the bar reads the latest package step as in progress, so it can show
one short while brew is between packages. The frames do not capture the
Operation view, so no re-run was needed. The r1 receipt was committed three
times; this one is committed once.
