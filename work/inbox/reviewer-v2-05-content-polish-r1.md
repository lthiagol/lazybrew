verdict: changes-required

# reviewer — v2-05-content-polish r1

Reviewed 1192a40..67b8df2 (6 AC commits + receipt), one commit per AC, in order.

| AC | verdict | evidence |
|---|---|---|
| AC-01 | met | `presentation.Detail` (detail.go) used by Formula/Cask Info, Outdated, Tap Info, Services, search info, Config; empty omitted, Muted labels, Info URLs, prose wraps, rest cut `…`; snapshot_test.go unchanged; presentation 97.9 ≥ 91 |
| AC-02 | met | depviews.go icons from loaded list; `casks_installed.json` is **byte-identical in content** to my own read-only `brew info --json=v2 --cask --installed` (both casks, every key equal); macos/arch shapes match brew's `{">=":[…]}` / `[{"type","bits"}]`, synthetic tests only (no such cask here — honestly stated) |
| AC-03 | met, see F-r1-02, F-r1-04 | five cards, gauge, `not loaded — press R` never 0, doctorView blocks, stack < 72 cols |
| AC-04 | met, see F-r1-03 | help.go scroll keys only inside help; `1–37 of 40` label; `TestHelpScrollReachesEveryLineAt80x24` passes |
| AC-05 | **not met**, F-r1-01 | steps/elapsed (spinner tick, no new goroutine)/final states/bar ok; upgrade-all n/total is wrong |
| AC-06 | met | harness re-run vs real brew: 54 files identical to `after/`; 36 `.txt` frames exactly W×H; `internal/gui/dump` removed; result.md explains every frame |

Gate (my run): `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` all ok · `make cover-check` all floors met (brew 76.1, presentation 97.9, gui 81.5, modal 88.3).
Invariants: no new brew command; `internal/brew/**` only cask depends_on; only `bubbles/v2/progress` (+ its indirect harmonica); no test func deleted; three changed expectations named in commits.

## Judgements asked for

- **Esc while help is open.** Before (46f5506 gui.go:684), Esc with a running task called `CancelCurrent()` and left help on screen — while the help footer advertised `?/esc close`. So Esc-to-close-help silently killed an install/upgrade. Now the first Esc closes help, a second Esc cancels. That is the brief's AC-04.1 (`?/esc close`) and safer; **not a regression**. It is untested → F-r1-03.
- **Dead dashboard text.** Should go → F-r1-02.
- **Fixtures.** Cask fixture: honest (real output). Upgrade transcript: honestly labelled as hand-written, and its shape matches brew — but the upgrade-all test fixture leaves out the very header line the AC-05.4 transcript includes, which hides a bug → F-r1-01.

## Findings

F-r1-01 (major) — internal/gui/operation.go:131 — upgrade-all progress counts brew's `==> Upgrading N outdated packages:` header as a package begun. — Real `brew upgrade` prints that header first (the slice's own `upgradeTranscript`, operation_test.go:17, has it); `TestUpgradeAllCountsPackagesBegun` (operation_test.go:174) feeds lines without it. With the header, 1 package outdated and bash still upgrading: begun=2, running −1 → bar reads `1/1` full mid-run; every mid-run count is one high. Also gui.go:304 takes `BatchTotal` from `Outdated.N`, which counts pinned packages `brew upgrade` skips, so a finished run can end at e.g. `3/5`. — Fix: count only per-package `Upgrading` steps (exclude the `outdated package(s):` header), make the test fixture start with the real header and assert mid-run and final counts; take the total from the header's N (or exclude pinned) so a finished run reads total/total.

F-r1-02 (major) — internal/gui/commands.go:1777-1795, messages.go:26, presentation/formatters.go:100,231,239 — `DataLoadedMsg.Dashboard`, the `items` built in `statusDashboard`, `FormatStatusDashboard`, `FormatDoctorStatus`, `OutdatedInfo` and `truncate` are no longer drawn, but ~15 expectations still assert on them (fetch_timeouts_test.go:108/114/147, outdated_count_test.go:104, gui_test.go:1948–2385). The invariants those tests guard — a timed-out read surfaces as ⚠, outdated unknown is never `0`, doctor exit 1 is warnings not failure, doctor failure is not "no issues" — are now checked against text no user sees; the live path (`summary.Problems`/`Outdated`/`Doctor` → Packages/Health cards) has only hand-built-summary unit tests. That is a weakened guard in effect, and dead code kept alive by tests. — Fix: retarget each of those expectations to `dMsg.Summary` and/or the rendered cards (`renderDashboard`), e.g. timeout → Health card shows `⚠ formulae: brew list timed out after 50ms`; then delete the dead text path and helpers and their now-pointless tests; presentation floor kept; name every moved expectation.

F-r1-03 (minor) — internal/gui/gui.go:675 — the Esc behaviour change has no test. `TestHelpEscClose` (gui_test.go:213) has no running task. — Fix: a test with help open and a running `opState`: first Esc closes help and does not cancel (task still running / `CancelCurrent` not called); second Esc cancels.

F-r1-04 (minor) — frame `01-status` 80x24 — when stacked, Health (doctor + failed-source ⚠ lines) is the last card and sits below the fold, so at 80×24 a timeout or doctor problem is invisible without scrolling; when doctor failed it reads `press d to run doctor` although doctor did run. — Fix: when cards stack, put Health before Taps/Services (or under Homebrew), and say `doctor unavailable` when the doctor read failed (Problems has a doctor entry).

## .ansi read (140x40)

Read the 140x40 and 80x24 frames 01, 03, 04, 11, 17, 18 as text: no further polish issues beyond F-r1-04. Help frames 08/16 not read in the ANSI form.
