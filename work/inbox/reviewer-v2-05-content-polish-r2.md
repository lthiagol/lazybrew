verdict: changes-required

# reviewer — v2-05-content-polish r2

Reviewed a676563..2947d62: fix commits f818b81, 0a03277, 7f8d584, 730c6ce, receipt 3b83b39/a730d04/2947d62.

| AC | verdict | evidence |
|---|---|---|
| AC-01…04, AC-06 | met | as r1; F-r1-02/03/04 fixed (below) |
| AC-05 | **not met** | F-r2-01: upgrade-all n/total still wrong for a real `brew upgrade` |

Gate (my run, HEAD 2947d62): `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` all ok · `make cover-check` brew 76.1 ≥ 62, presentation 98.2 ≥ 91, gui 81.9 ≥ 36, modal 88.3 ≥ 41.
Harness re-run vs real brew (read-only): all 54 `after/` files byte-identical; 36 `.txt` frames exactly W×H; `internal/gui/dump` removed; worktree clean.

## Earlier findings

- **F-r1-01 — partly fixed, see F-r2-01.** Header no longer counted for a formulae-only run; total from brew's header; pinned skip and 1-package cases tested.
- **F-r1-02 — fixed.** Dead text path, `FormatStatusDashboard`, `FormatDoctorStatus`, `OutdatedInfo`, `truncate` gone; each moved expectation named in 0a03277. Checks now read `dashboardCards` (Update + `renderDashboard`) or `healthWords` (`healthBody`). Mutation proof, in a throwaway worktree (removed after):
  - M1 `outdatedField` unknown → `"0 packages"`: FAIL `TestStatusDashboardDoesNotShellCallOutdated`, `…PartialOutdatedCacheIsNotKnown`, `…ExplicitLoadFailureFallsBackToNotLoaded`, `TestOutdatedUnknownIsNotLoadedNeverZero`, `TestStatusDashboardTabShowsTheFullDashboard`.
  - M2 Health drops `s.Problems`: FAIL `TestFetchStatusDataTimesOut`, `…ExplicitLoadFailureFallsBackToNotLoaded`.
  - M3 `DoctorErr` case disabled: FAIL `TestStatusDashboardSurfacesDoctorError`, `TestStackedCardsPutHealthNearTheTop`.
  - M4 `"1 package"` → `"1 packages"`: **all pass** → F-r2-02.
- **F-r1-03 — fixed.** `TestHelpEscClosesHelpBeforeCancellingATask` runs a real task through `task.Manager`: first Esc closes help and its ctx stays live, second Esc cancels it.
- **F-r1-04 — fixed.** Order Homebrew, Health, Packages, Taps, Services; `doctor unavailable: <reason>`; `TestStackedCardsPutHealthNearTheTop`. Frame 01 re-captured at both sizes (80x24: Health on rows 9–11), 15 at 140x40; matches my run.

## Findings

F-r2-01 (major) — internal/gui/operation.go:135 and :166 — `operationProgress` keeps only the **last** `Upgrading N outdated package(s):` header, and counts brew's dependents header as a package. Upgrade All runs plain `brew upgrade` (formulae.go:282), which prints a formulae header (Homebrew `cmd/upgrade.rb:458`) and then a separate casks header (`cask/upgrade.rb:107`), plus `Upgrading N dependent(s) of upgraded formula(e):` (`upgrade.rb:400`) followed by per-dependent `Upgrading x` steps. Probe test (throwaway worktree): 2 formulae then 1 cask, firefox in progress → `1/1` (want `2/3`); 1 formula + 1 dependent, zsh in progress → `1/1`, bar full while brew still works. — Fix: total = sum of every `outdated package(s):` header's N, plus the N of a dependents header; the dependents header is not a package step; tests with a mixed formula+cask transcript and a dependents transcript, both built from those brew source strings, asserting mid-run and final counts.

F-r2-02 (minor) — internal/gui/gui_test.go:2003,2100,2123,2238, outdated_count_test.go:105 — `Contains(joined, "Outdated:  1 package")` also matches `1 packages`, so the singular wording the fix introduced is unguarded (mutation M4 passes). — Fix: assert the whole row (e.g. `Outdated:  1 package ` followed by padding/border, or a regexp `1 package\b`) in at least one of these, and show M4 then fails.

F-r2-03 (minor) — work/verification/v2-05-content-polish/result.md:17,42 — still say the cards go Homebrew, Packages, Taps, Services, Health and that at 80×24 Health is below the fold; the fix-round section (line 54) says the opposite. — Fix: correct lines 17 and 42 to the current frames.

Note (no fix asked): the fix receipt was committed three times (3b83b39, a730d04, 2947d62); the fix brief says committed once. Next receipt: one commit.
