# runner: lb02-fetch-timeouts

Commits (AC order): AC-01 `0027596` (panel+status bounded) · AC-02 `f7bc287`
(search bounded) · AC-03 `00f73f2` (tests) · AC-04 `16c24f4` (bookkeeping).

**AC-01.** `internal/gui/commands.go`: each read gets its own deadline via
`boundedRead`/`boundedReadIn`. `brewReadTimeout` (30s) caps one read;
`brewStatusTotalTimeout` (45s) caps the whole dashboard and trims each read's
deadline, so one slow source cannot starve the others. The five dashboard reads
run concurrently under the total budget, joined before the command returns.
`Diagnostics.Config`'s error stays ignored (as before) but now has a deadline.
A timeout surfaces as `&brew.TimeoutError{...}` through `DataLoadedMsg.Err` and
the dashboard's aggregated `⚠` lines.
**AC-02.** `executeSearch` and `fetchSelectedSearchInfo` use
`brewSearchTimeout` (15s); a timeout surfaces as `*brew.TimeoutError` through
`SearchDoneMsg.Err` and `SearchInfoLoadedMsg.Err`.
**AC-03.** `internal/gui/fetch_timeouts_test.go`, six tests: one per path plus
two for the status total budget. The mock blocks on `ctx.Done()` then fails with
`"signal: killed"` (not `context.DeadlineExceeded`), so a timeout can only be
reported by detecting the deadline from the context's own error. Each test
asserts the call returns (every wait is bounded), the error is a
`*brew.TimeoutError` via `errors.As`, and the message names the command and the
duration. Mutation-checked: detect the timeout from the error's type instead of
`ctx.Err()` → 5 tests fail; drop the total budget →
`TestFetchStatusDataTotalBudgetBoundsTheAggregate` fails; drop the per-read cap
→ the panel/search tests hang and fail.
**AC-04.** `work/backlog.md`: LB-02 `closed` + evidence note; new row LB-07
(low, active, "found by lb02-fetch-timeouts"). `work/STATE.md`: row
`| lb02-fetch-timeouts | — | implemented | awaiting reviewer |`.
**Gate.** `make build` PASS · `make lint` PASS · `go test ./... -race -count=1`
PASS (all packages ok) · `make cover-check` PASS (brew 64.3%, presentation
95.4%, gui 57.4%, modal 41.4%).
**Caps.** 30s per panel/status read, 15s per search read, 45s for the whole
status dashboard — deliberately larger than a single read's cap, so one slow
source eats at most 30s of it and the other four keep theirs.
**Notes.** The brief's LB-07 line numbers (951 / 1045) were pre-slice; the two
reads are now at `commands.go:1034` and `commands.go:1128`, cited in the row.
Tab-content path untouched (`git diff 1ec84ec -- internal/gui/commands.go`).
`context.Background()` remains only in the bounded-read helper, the tab-content
path, the two LB-07 reads, and the `serviceCleanup` write path.
