# Receipt: cr5-dead-code, fix round r1 (runner)

Source: `work/inbox/reviewer-cr5-dead-code-r1.md`. Comments, test names and this receipt only; no production code, golden or testdata changed.

**F-r1-01 (2b46654):** the `renderMainPanelIn` and `contentWidthIn` comments in `render.go` now carry the text deleted with `renderMainPanel` and `mainContentWidth` in c978a89 (read at `c978a89^`). A grep of `internal/` and `cmd/` for the removed names in the review's list finds no comment naming one now. `DESIGN.md:108,123` still say TypedCache; that is a dated plan and decision log, left as history.

**F-r1-02 (adb681f):** `task/manager_test.go` now names `TestManager_CancelledCompletion` and says what it asserts. The `gui_test.go` comments and failure messages at the three sites no longer name the fast path. Renamed, bodies unchanged:
- `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply` → `TestUsesTabReissuesFetchOnDataLoaded`
- `TestDepsFastPathRendersCachedDependencies` → `TestDepsRendersDependenciesWithoutFetch`
- `TestDepsFastPathNoDependenciesShowsExplicitMessage` → `TestDepsNoDependenciesShowsEmptyState`
- `TestDepsTabViewLeavesLoadingAfterFastPath` → `TestDepsTabViewLeavesLoadingWithoutFetch`

The renamed Used By test is not a copy: it sets `formulaUsesTab`, while `TestDepsTabPopulatedAfterDataLoaded` sets `formulaDepsTab`.

**F-r1-03 (this commit):** the AC-07.2 mapping, from `git show --stat -M d0a6a86` (seven renames at 100% similarity):
- `internal/gui/review_r1_test.go` → `gui_misc_test.go`
- `internal/gui/review_r2_test.go` → `batch_upgrade_test.go`
- `internal/gui/review_r3_test.go` → `batch_queue_test.go`
- `internal/gui/review_r4_test.go` → `batch_selection_test.go`
- `internal/gui/regression_test.go` → `tab_content_selection_test.go`
- `internal/brew/runner_stream_r1_test.go` → `runner_stream_errors_test.go`
- `internal/brew/runner_stream_r2_test.go` → `runner_stream_cancel_test.go`

**Gate, tree at adb681f:** `make ci` exit 0 (fmt, vet, race tests, all 10 coverage floors met: gui 89.9, style 89.5, task 96.2, build). `GOOS=darwin go build ./...` exit 0. `make test-fast` on `./internal/gui/...` and `./internal/gui/task/...` green after each commit.

**Tools after the fix:** `deadcode ./cmd/lazybrew` → `logger.go:23:6: unreachable func: SetDebug` (kept on purpose). `staticcheck ./...` → exit 1, only SA1019 at keymap_test.go:38 and :434 (left, as asked).

**Seen, not done:** nothing new; the first receipt's backlog stands.

**Status:** nothing failed; commits are local, nothing pushed.
