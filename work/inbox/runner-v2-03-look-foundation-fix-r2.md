# Runner fix receipt — v2-03-look-foundation, round r2

**Commit:** `3677383` fix(gui): F-r2-01 a sidebar count always shows whole; the title yields.

**F-r2-01.** A count is now marked (`borderLabel.count`) and always shown whole. The
title yields first its icon (`borderTitle.bare`), then its name, cut with `…`. Labels
with a short form still yield to the title as before. Comments in `box.go` and
`render.go` reflowed.

**Tests.** `TestCountsStayWholeAt80Columns`: 1234 and 12345 items on Formulae,
Outdated and Services at 80×24; the border ends with the whole count and the row is
exactly 24 cells. `TestBorderTitleYieldsToCount`: a 7-digit count keeps its whole
form and the name is cut. Mutation: dropping `count: true` from the sidebar labels
fails the first test with the reviewer's symptom,
`╭─ 2 ∘ Formulae ───────╮` (no count); restored, it passes.

**Gate** (`CGO_ENABLED=1` for `-race`): `make build` 0 · `make lint` 0 ·
`go test ./... -race -count=1 -v` 0, PASS 615 / FAIL 0 / SKIP 0 (was 613; +2 new,
none deleted) · `make cover-check` "All coverage floors met".

**Frames.** Re-run with the same harness: 0 frames changed (22 `.txt` and 11
`.ansi` byte-identical to `after/`). No `after/` or `result.md` change.

Worktree clean; no push; no brew write ran; `internal/zzharness` removed.
