verdict: ok

# Review — lb04-lb10-small-fixes, round r1

Reviewed `d2e2737..32cdb31` on `wip`: `920887d` (AC-01), `ac065b8` (AC-02),
`e6ec90e` (AC-03), receipt `a03db94` + trims `8a85f88`, `32cdb31`.

## V1 gate

- `make ci` (CGO_ENABLED=1, race): exit 0. fmt-check, vet, all packages `ok`;
  cover-check "All coverage floors met" (brew 80.0, presentation 98.5, gui 85.9,
  modal 90.2); build ok.
- `GOOS=darwin go build ./...`: ok.

## AC table

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 | met | `chunkTapNames` cuts on `i-start >= tapInfoChunkMax` (100) **or** `size+n > tapInfoChunkBytes` (64 KiB, NUL counted); slices in order, each name once; oversize name gets its own chunk. `ctx.Err()` checked before each chunk; any chunk error returns `nil, err`. `List` fallback (`TrustUnknown`, no cache) untouched (taps.go:93-101). Bounds are package vars, restored via `t.Cleanup`, no `t.Parallel`. Test (a) runs at production bounds: one call, argv `tap-info --json a/x b/y c/z`. |
| AC-02 | met | `cutRow` keeps the replaced row's indent and pads to its width; only the cut branch changed, so an uncut body is byte-identical. Frames: `…` moved from col 35 to the list column (before/after diff is one row only, 80x24 and 140x40). Every frame row exactly 80 / 140 cells, 24 / 40 rows. `TestCutBodyEllipsisSitsInListColumn` checks rows ≤ `dialogMaxRows`, width = `confirmWidth`, `…` column = list column. Existing `confirm_face_test.go` cases unchanged (diff is additions only). |
| AC-03 | met | `result.md` present; STATE row `| lb04-lb10-small-fixes | — | implemented | awaiting reviewer |`. |

## Mutation checks (in a scratch copy, repo untouched)

| Mutation | Caught by |
|---|---|
| drop the last chunk | SplitsByNameCount, SplitsByByteBudget, FewTaps…, TapsServiceList/Get |
| remove ctx check | StopsBeforeNextChunkOnCancel |
| ignore byte budget | SplitsByByteBudget |
| ignore count bound | SplitsByNameCount, …SecondChunkFailure… |
| swallow chunk error | …SecondChunkFailure…, TrustUnknownOnTapInfoTimeout |
| revert `cutRow` to bare `"…"` | TestCutBodyEllipsisSitsInListColumn |

## V4 invariants

- Code diff limited to `taps.go`, `taps_test.go`, `frame.go`,
  `confirm_face_test.go`; no new deps; no test deleted or weakened.
- One commit per AC, in order. No brew write commands run in this review.
- Empty name list now makes no tap-info call (receipt notes it): unreachable,
  since `List` guards `len(names) > 0` and `Get` passes one name. Not a finding.

## Findings

None.

Note (not a finding): `work/verification/lb04-lb10-small-fixes/uninstall-*.txt`
are byte-identical copies of `after/`; harmless duplication.
