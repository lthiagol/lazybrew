# Receipt — runner, lb04-lb10-small-fixes

Brief: `work/briefs/lb04-lb10-small-fixes.md`. Result: `work/verification/lb04-lb10-small-fixes/result.md`.

## Commits (one per AC, in order)

- AC-01 (LB-04): `920887d` brew: chunk tap-info by name count and byte budget
- AC-02 (LB-10): `ac065b8` modal: cut dialog body row sits in the list column
- AC-03: `e6ec90e` docs(plan): result and STATE row
- Receipt: `a03db94`; `8a85f88` trims it to the line limit.

## Acceptance criteria

- **AC-01.** Chunks bounded by `tapInfoChunkMax` (100) and `tapInfoChunkBytes`
  (64 KiB, NUL counted), overridable by tests. Three taps: one call, same argv
  (`TestTapInfoBatchFewTapsIsOneCallWithSameArgs`). Seven names at bound 3: 3+3+1,
  each once, input order (`…SplitsByNameCount`). Byte budget: 2+2+1, oversize
  name alone (`…SplitsByByteBudget`). Cancel stops early
  (`…StopsBeforeNextChunkOnCancel`). Failed second chunk: name-only
  `TrustUnknown`, not cached (`TestTapsListSecondChunkFailureFallsBackAndDoesNotCache`).
- **AC-02.** `cutRow` keeps the row's indent and width. `TestCutBodyEllipsisSitsInListColumn`
  failed before the fix ("column 35, list at 27") and passes after. Frames in
  `work/verification/lb04-lb10-small-fixes/{before,after}/`, 80x24 and 140x40:
  the `…` moves from column 35 to the list column. Exact size; box 20 rows.
- **AC-03.** `result.md` written; STATE row `implemented | awaiting reviewer`.

## Gate (once, after AC-02)

- `make ci`: exit 0. fmt-check, vet, `go test ./... -race -count=1` all ok;
  cover-check "All coverage floors met" (brew 80.0, presentation 98.5, gui 85.9,
  modal 90.2); build ok. `GOOS=darwin go build ./...`: ok.
- Per commit: `make test-fast` on `./internal/brew/...` and `./internal/gui/...`: ok.

## Notes

- No brew write command ran; tests and capture use the mock runner only.
- No brief/repo contradiction; `work/questions.md` untouched.
- Behaviour change, unreachable in production: an empty name list makes no
  tap-info call (it made one with no names).
- Not pushed. Worktree clean.
