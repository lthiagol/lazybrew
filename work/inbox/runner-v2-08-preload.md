# Receipt — runner, v2-08-preload

## Commits (one per AC, in order)

- AC-01  1471c38  feat(gui): a preload scheduler loads the panels after Status
- AC-02  2237be9  feat(gui): the dashboard fills in as preloaded panels land
- AC-03  fd55e63  test(gui): queued panels say loading, and the v2-08 frames

## Acceptance criteria

- **AC-01.1** config + env: `internal/config/preload_test.go` (4 pass:
  default on, file, env overrides, env rejects `yes`). README and DESIGN document both.
- **AC-01.2** order: `TestPreloadOrderAndTwoInFlight` — landing order is
  Formulae, Outdated, Casks, Taps, Services; Search never queued.
- **AC-01.3** 2 in flight: same test checks `inFlight` and the pending count
  after every landing (fails above 2); `TestPreloadStartsWithTwoReads`. Reads are
  `fetchPanelData` (boundedRead). No goroutine; no `program.Send` added
  (the one match in `internal/` is a doc comment in `task/doc.go`).
- **AC-01.4** dedupe: `TestPreloadDedupesWithAUserSwitch` — the switch promotes
  the queued panel, starts no second read; exactly one Taps landing.
- **AC-01.5** off: `TestPreloadOffStaysLazy` — no read at Status, lazy read on switch.
- **AC-01.6** helpers: gui tests use `testConfig()` (preload off) and testutil
  turns it off; the harness sets it on. `TestProductionDefaultPreloads` and
  `TestPreloadDefaultIsOn` pin the default.
- **AC-01.7** unit tests: order, cap, dedupe, failed read, off,
  update_on_start (`TestPreloadWaitsForUpdateOnStart`).
- **AC-02.1** `TestPreloadKeysWorkMidPreload`, `TestPreloadMouseWorksMidPreload`
  (assert on `View()`).
- **AC-02.2** `TestPreloadFailedReadShowsItsErrorOnItsPanelOnly`,
  `TestPreloadFailedReadDoesNotStopTheOthers`.
- **AC-02.3** `TestPreloadFillsTheDashboardCounts` (`loading…`, then `0 packages`);
  `TestPreloadOffDashboardSaysPressR`.
- **AC-03.1** `TestPreloadSidebarSaysLoadingNotNotLoaded`. Loading state is the
  existing spinner, not a `loading…` text in Subtle (judgement, in result.md).
- **AC-03.2** `work/verification/v2-08-preload/{before,after}/` (before = v2-07
  after, copied). Real read-only run, `ok`. Sidebar changed on 16 screens per
  size, each named in result.md; main area changed on 01 and 15 (Outdated
  count), and on 06 and 14 (degraded Taps read — see below).
- **AC-03.3** `result.md` written; STATE row `v2-08-preload | — | implemented | awaiting reviewer`.

## Gate (final tree, after AC-03)

- `make build` — pass
- `make lint` (go vet ./...) — pass
- `CGO_ENABLED=1 go test ./... -race -count=1` — every package `ok`
- `make cover-check` — "All coverage floors met" (gui 85.1%)

## Next session must know

- **Brew is slow on this machine now.** `brew tap-info` takes ~87 s (over the 30 s
  read cap), `brew outdated --formula` ~21 s. The Taps read then falls back to
  name-only, silently. Frames 06 and 14 carry that. Lazy mode does the same
  (checked). Q-v2-08-b asks whether it should surface as an error.
- Q-v2-08-a: a failed preloaded panel retries on switch (existing policy, kept).
- Search is never preloaded, so it still says `not loaded` (by design).
- No brew write ran. The harness never confirms a write; `S` is cancelled.
- Harness takes ~150 s (real brew). Its file is
  `work/verification/v2-08-preload/frame_dump_test.go.txt`.
