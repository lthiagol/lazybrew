# lb04-lb10-small-fixes — result

Brief: `work/briefs/lb04-lb10-small-fixes.md`. Commits, one per AC, in order:

| AC | Commit | Subject |
|---|---|---|
| AC-01 (LB-04) | 920887d | brew: chunk tap-info by name count and byte budget (LB-04) |
| AC-02 (LB-10) | ac065b8 | modal: cut dialog body row sits in the list column (LB-10) |
| AC-03 | (this commit) | bookkeeping: this file and the STATE row |

## Tests added (AC-01, `internal/brew/taps_test.go`)

- `TestTapInfoBatchFewTapsIsOneCallWithSameArgs` — three taps make one
  `tap-info --json a/x b/y c/z` call, the same argv as before; entries come
  back in input order.
- `TestTapInfoBatchSplitsByNameCount` — seven names with the count bound set to
  3 make three calls of 3, 3 and 1 names; each name is sent once; the entries
  merge in input order.
- `TestTapInfoBatchSplitsByByteBudget` — five 30-byte names with a 70-byte
  budget (each name costs 31 with its NUL) make calls of 2, 2 and 1, each within
  the budget. A 50-byte name over a 40-byte budget goes in a call of its own.
- `TestTapInfoBatchStopsBeforeNextChunkOnCancel` — the ctx is cancelled during
  the first chunk; the batch returns `context.Canceled` with one call made.
- `TestTapsListSecondChunkFailureFallsBackAndDoesNotCache` — with two names per
  chunk, the second chunk fails: `List` returns every tap name-only with
  `TrustUnknown` set. The next `List` retries (4 calls in all) and is enriched
  with `TrustUnknown` clear, so nothing was cached.

The bounds are package variables (`tapInfoChunkMax`, `tapInfoChunkBytes`); the
tests lower them through `setTapInfoChunkBounds`, which restores them on
cleanup. Production values are 100 names and 64 KiB.

The existing tap tests are unchanged, including `TestTapsBatchLoadIsSingleInvocation`,
`TestTapsBatchLoadFallsBackToNameOnlyOnError`, `TestTapsBatchFailureDoesNotPoisonCache`
and `TestTapsTrustUnknownOnTapInfoTimeout`.

## Test added (AC-02, `internal/gui/modal/confirm_face_test.go`)

- `TestCutBodyEllipsisSitsInListColumn` — a confirm with 30 dependents is cut.
  The `…` row's first text column must equal the list rows' first text column,
  the box must fit `dialogMaxRows`, and every row must be `confirmWidth` cells.
  Before the fix it failed: `the ellipsis starts at column 35, the list at column 27`.
  After the fix it passes.

The uncut path is not touched by the change. The existing `confirm_face_test.go`
expectations are unchanged, and nothing was skipped or weakened.

## Frames (AC-02.3)

`work/verification/lb04-lb10-small-fixes/before/` and `after/` hold
`uninstall-80x24.txt` and `uninstall-140x40.txt`. Each is the uninstall confirm
for `testformula` with 30 dependents, after `2` then `x`. They were taken by
`capture_frames_test.go.txt` in this directory, which drives the real `gui.Model`
over a `brew.MockRunner`. No brew command ran, and `esc` closes the confirm.

- Every frame is exactly its terminal size: 24 rows of 80 cells, 40 rows of 140 cells.
- `before/`, the pre-fix `frame.go`: the `…` row is centred alone at column 35.
- `after/`: the `…` row sits at the list's column, under the dependent names.
- The box is 20 rows, the same height as before.

## Gate

Run once on the code tree, after AC-02 and before this commit.

- `make ci` — exit 0.
  - `make fmt-check` (gofmt): clean.
  - `go vet ./...`: clean.
  - `CGO_ENABLED=1 go test ./... -race -count=1`: every package `ok` (app, brew,
    config, gui, gui/anim, gui/flows, gui/modal, gui/presentation, gui/style,
    gui/task, gui/testutil).
  - `make cover-check`: "All coverage floors met" — brew 80.0% ≥ 62%,
    presentation 98.5% ≥ 91%, gui 85.9% ≥ 36%, modal 90.2% ≥ 41%.
  - `go build -o bin/lazybrew ./cmd/lazybrew`: ok.
- `GOOS=darwin go build ./...`: ok.
- Per commit: `make test-fast PKG=./internal/brew/...` after AC-01 and
  `make test-fast PKG=./internal/gui/modal/...` after AC-02, plus
  `make test-fast PKG=./internal/gui/...` after AC-02, all `ok`.

## Notes

- An empty name list now makes no `tap-info` call. Before, it made one call with
  no names. Neither caller can reach that: `List` only batches when there are
  names, and `Get` passes one name. Flagged here, not changed.
- The byte budget counts each name with its NUL terminator, so the count is a
  little over the bytes of the names alone.
- No brew write command was run. The capture and every test use the mock runner.
