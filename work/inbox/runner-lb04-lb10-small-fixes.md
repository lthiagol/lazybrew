# Receipt — runner, lb04-lb10-small-fixes

Brief: `work/briefs/lb04-lb10-small-fixes.md`. Result: `work/verification/lb04-lb10-small-fixes/result.md`.

## Commits (one per AC, in order)

- AC-01 (LB-04): `920887d` brew: chunk tap-info by name count and byte budget
- AC-02 (LB-10): `ac065b8` modal: cut dialog body row sits in the list column
- AC-03: `e6ec90e` docs(plan): result and STATE row

Receipt commit: this file, committed once after AC-03.

## Acceptance criteria

**AC-01 — tap-info is chunked.** `fetchTapInfoBatch` runs one `tap-info --json`
per chunk, bounded by `tapInfoChunkMax` (100 names) and `tapInfoChunkBytes`
(64 KiB, each name counted with its NUL). The bounds are package variables that
tests lower. Three taps make one call with argv `tap-info --json a/x b/y c/z`
(`TestTapInfoBatchFewTapsIsOneCallWithSameArgs`). Seven names with bound 3 make
3+3+1 calls, each name sent once, entries in input order
(`TestTapInfoBatchSplitsByNameCount`). Five 30-byte names with a 70-byte budget
make 2+2+1; a 50-byte name over a 40-byte budget gets its own call
(`TestTapInfoBatchSplitsByByteBudget`). A cancel during chunk 1 stops the batch
with `context.Canceled` after one call (`TestTapInfoBatchStopsBeforeNextChunkOnCancel`).
A failed second chunk makes `List` name-only with `TrustUnknown`, and the next
`List` retries uncached (`TestTapsListSecondChunkFailureFallsBackAndDoesNotCache`).

**AC-02 — the cut row sits in the list column.** `cutRow` keeps the replaced
row's indent and pads it to the row's width. Uncut bodies take no new code path.
`TestCutBodyEllipsisSitsInListColumn` failed before the fix ("ellipsis starts at
column 35, the list at column 27") and passes after. Frames under
`work/verification/lb04-lb10-small-fixes/{before,after}/`, 80x24 and 140x40:
the `…` is centred at column 35 before and sits under the list after. Each frame
is exactly its terminal size, and the box is 20 rows. The capture harness is
`capture_frames_test.go.txt`, mock runner only.

**AC-03 — bookkeeping.** `result.md` written. STATE row:
`| lb04-lb10-small-fixes | — | implemented | awaiting reviewer |`.

## Gate (run once, after AC-02, before AC-03)

- `make ci`: exit 0. fmt-check clean; `go vet ./...` clean; `CGO_ENABLED=1 go test ./... -race -count=1`
  all `ok`; `make cover-check` "All coverage floors met" (brew 80.0 ≥ 62,
  presentation 98.5 ≥ 91, gui 85.9 ≥ 36, modal 90.2 ≥ 41); build ok.
- `GOOS=darwin go build ./...`: ok.
- Per commit: `make test-fast PKG=./internal/brew/...` ok (AC-01);
  `make test-fast PKG=./internal/gui/...` ok (AC-02).

## Notes for the next session

- No brew write command was run. Every test and the frame capture use the mock runner.
- No brief/repo contradiction and no open question; `work/questions.md` untouched.
- Behaviour change, unreachable in production: an empty name list now makes no
  tap-info call (it made one call with no names). `List` only batches when names exist and `Get` passes one.
- The capture harness is a `.txt`, not a committed `_test.go`, so `go test ./...` does not run it.
- Not pushed, per the brief. The worktree is clean.
