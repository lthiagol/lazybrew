# Receipt — runner, lb04-lb10-small-fixes

Brief: `work/briefs/lb04-lb10-small-fixes.md`. Result: `work/verification/lb04-lb10-small-fixes/result.md`.

## Commits (one per AC, in order)

- AC-01 (LB-04): `920887d` brew: chunk tap-info by name count and byte budget
- AC-02 (LB-10): `ac065b8` modal: cut dialog body row sits in the list column
- AC-03: `e6ec90e` docs(plan): result and STATE row
- Receipt: `a03db94`; this trim is the follow-up commit (the first receipt was 56 lines).

## Acceptance criteria

- **AC-01.** Chunks are bounded by `tapInfoChunkMax` (100) and `tapInfoChunkBytes`
  (64 KiB, NUL counted), overridable by tests. Three taps: one call, same argv
  (`TestTapInfoBatchFewTapsIsOneCallWithSameArgs`). Seven names at bound 3: calls
  3+3+1, each name once, input order (`TestTapInfoBatchSplitsByNameCount`). Byte
  budget splits 2+2+1, and an oversize name gets its own call
  (`TestTapInfoBatchSplitsByByteBudget`). Cancel stops before the next chunk
  (`TestTapInfoBatchStopsBeforeNextChunkOnCancel`). A failed second chunk gives
  the name-only `TrustUnknown` list and no cache (`TestTapsListSecondChunkFailureFallsBackAndDoesNotCache`).
- **AC-02.** `cutRow` keeps the replaced row's indent and width. The new test
  `TestCutBodyEllipsisSitsInListColumn` failed before the fix ("column 35, list
  at column 27") and passes after. Frames in `work/verification/lb04-lb10-small-fixes/{before,after}/`
  at 80x24 and 140x40: the `…` moves from column 35 to the list column. Frames are
  exact terminal size; box stays 20 rows.
- **AC-03.** `result.md` written. STATE row `| lb04-lb10-small-fixes | — | implemented | awaiting reviewer |`.

## Gate (once, after AC-02)

- `make ci`: exit 0. fmt-check clean; vet clean; `CGO_ENABLED=1 go test ./... -race -count=1`
  all ok; `make cover-check` "All coverage floors met" (brew 80.0 ≥ 62, presentation
  98.5 ≥ 91, gui 85.9 ≥ 36, modal 90.2 ≥ 41); build ok.
- `GOOS=darwin go build ./...`: ok.
- Per commit: `make test-fast` on `./internal/brew/...` (AC-01) and `./internal/gui/...` (AC-02): ok.

## Notes

- No brew write command ran; tests and the capture use the mock runner only.
- No brief/repo contradiction; `work/questions.md` untouched.
- Behaviour change, unreachable in production: an empty name list now makes no
  tap-info call (it made one with no names).
- Capture harness is `capture_frames_test.go.txt`, not a committed `_test.go`.
- Not pushed. Worktree clean after the trim.
