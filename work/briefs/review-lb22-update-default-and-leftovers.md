# Review brief — slice `lb22-update-default-and-leftovers`

> Contract: `work/briefs/lb22-update-default-and-leftovers.md` (with
> `cr-common.md`). You review; you do **not** edit code, tests or commits. Review
> everything on `wip` after the commit that added the slice brief
> (`git log --oneline -- work/briefs/lb22-update-default-and-leftovers.md`) up to
> HEAD. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate. No
> brew write commands; never launch the TUI; scratch under /home/agent/.cache;
> do not touch other panes.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 AC-01:** `config.Default()` is `true`; the sample, README and DESIGN agree
  and the sample still loads to `Default()`; an old config with `auto` or
  `false` keeps its meaning; a config file that omits the key gets `true`. With
  the mock runner: default launch enqueues exactly one `brew update` through
  `task.Manager`, Outdated loads after it, and `R` still behaves per lb16.
  Offline failure leaves a usable TUI. No test builds a default Model that
  accidentally updates (grep the helper); frames: only the header badge and
  intended goldens differ (own `View()` diff base vs HEAD, 80x24 and 140x40).
  Races: `-race -count=20` on the launch/update tests.
- **V3 AC-02:** (a) one read exit 1 + other real error + no rows → error toast;
  with rows → rows kept, no error; (b) a stale generation re-mark is dropped
  (revert the generation check: the test must fail). Revert each fix to see red.
- **V4 AC-03:** the singleflight tests have no sleeps and pass `-race
  -count=100`, fail when the sharing is reverted; the log file is closed at
  exit (test) and nothing writes to a closed file afterwards; the doc is right.
  deadcode/staticcheck from the installed binaries: nothing new vs the base.
- **V5:** one commit per AC; floors not lowered; no new goroutines in handlers.

## Report
`work/inbox/reviewer-lb22-update-default-and-leftovers-rN.md` (≤ 50 lines): first
line `verdict: ok` or `verdict: changes-required`; AC table; findings `F-rN-01
(severity) — file:line — what — evidence — fix must achieve`. Commit once
(`inbox: reviewer reports lb22-update-default-and-leftovers rN`), worktree clean,
reply with one line: `done: work/inbox/reviewer-lb22-update-default-and-leftovers-rN.md`.
