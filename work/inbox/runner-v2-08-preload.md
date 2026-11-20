# Receipt — runner, v2-08-preload

## Commits (one per AC, in order)
- AC-01  1471c38  feat(gui): a preload scheduler loads the panels after Status
- AC-02  2237be9  feat(gui): the dashboard fills in as preloaded panels land
- AC-03  fd55e63  test(gui): queued panels say loading, and the v2-08 frames

## Acceptance criteria

- **AC-01.1–.7**: `internal/config/preload_test.go` (4 pass); scheduler tests in
  `internal/gui/preload_test.go`: `TestPreloadOrderAndTwoInFlight` (order, and
  `inFlight` ≤ 2 after every landing), `TestPreloadDedupesWithAUserSwitch`,
  `TestPreloadOffStaysLazy`, `TestPreloadFailedReadDoesNotStopTheOthers`,
  `TestPreloadWaitsForUpdateOnStart`. Reads are `fetchPanelData`; no goroutine;
  the one `program.Send` match in `internal/` is a doc comment. Helpers and
  harness set the preload explicitly (result.md).
- **AC-02.1** `TestPreloadKeysWorkMidPreload`, `TestPreloadMouseWorksMidPreload`.
  **AC-02.2** `TestPreloadFailedReadShowsItsErrorOnItsPanelOnly`.
  **AC-02.3** `TestPreloadFillsTheDashboardCounts`, `TestPreloadOffDashboardSaysPressR`.
- **AC-03.1** `TestPreloadSidebarSaysLoadingNotNotLoaded`. The loading state is
  the existing spinner, not Subtle `loading…` text (judgement, in result.md).
- **AC-03.2** `work/verification/v2-08-preload/{before,after}/`; real read-only
  run, `ok`. Sidebar changed on 16 screens per size, each named in result.md.
  Main area changed on 01 and 15 (Outdated count), and on 06 and 14 (degraded
  Taps read, see below).
- **AC-03.3** `result.md`; STATE row `v2-08-preload | — | implemented | awaiting reviewer`.

## Gate (final tree, after AC-03)

- `make build`: pass. `make lint` (go vet ./...): pass.
- `CGO_ENABLED=1 go test ./... -race -count=1`: every package `ok`.
- `make cover-check`: "All coverage floors met" (gui 85.1%).

## Next session must know

- Brew is slow here now: `brew tap-info` ~87 s (over the 30 s cap), so Taps
  falls back to name-only with no error. Lazy mode does the same. Q-v2-08-b.
- Q-v2-08-a: a failed preloaded panel retries on switch (existing policy, kept).
- Search is never preloaded, so it still says `not loaded` (by design).
- No brew write ran; the harness never confirms a write.
