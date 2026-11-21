# Fix brief — lb22-update-default-and-leftovers, round r2

Findings: `work/inbox/reviewer-lb22-update-default-and-leftovers-r1.md` (read it in
full). Commit subjects end with the id in brackets (`[LB-21]`); no amends.

- **F-r1-01 (medium):** convert `TestSingleflightCoalescesConcurrent`,
  `TestSingleflightKeysAreIndependent` and `TestSingleflightPanicUnblocksWaiters`
  in `internal/brew/singleflight_test.go` to `synctest.Test` + `synctest.Wait`, no
  sleeps. `CGO_ENABLED=1 go test ./internal/brew -run Singleflight -race
  -count=100` passes, and each test fails when `singleflight.Do` stops sharing
  (panic test: still unblocks waiters). Keep the earlier conversions. Show the
  mutation runs in the receipt.
- **F-r1-02 (low):** the `testConfig` doc in `internal/gui/main_test.go` names the
  real test, `TestProductionDefaultUpdatesOnEveryLaunch`.
- **F-r1-03:** nothing to do (process note).

Gate: `make test-fast PKG=./internal/brew/... ./internal/gui/...`, then `make ci`.
Append "Fix r2" (≤ 8 lines) to `work/inbox/runner-lb22-update-default-and-leftovers.md`,
commit (`inbox: runner fix r2 lb22-update-default-and-leftovers`), do not edit it
afterwards, reply with one done line. Never launch the TUI.
