verdict: ok

# Reviewer — v2-08-preload, round r4

Range: `eafcac4..63e8830` (fix brief 483dd59; 39f34d2; receipt 63e8830). This round
is docs only: `git diff --stat 3eed39a..HEAD -- internal` and
`git diff eafcac4..HEAD -- work/verification/v2-08-preload/after` are both empty. So
the r3 harness run (on real brew at normal speed, byte-identical, 40/40 frames W×H)
still stands for the committed frames. It was not rerun. No brew write ran.

## Earlier findings

| Id | Status | Evidence |
|---|---|---|
| F-r1-01 | holds | `TestEarlySwitchCountReachesTheDashboard` passes; no code change since r3 |
| F-r1-02 | holds | STATE.md v2-08 paragraph unchanged |
| F-r2-01 | holds | frames unchanged since f6107b4; 140x40 15/18 dashboard `Taps 3` |
| F-r2-02 | holds | `TestFailedReadForgetsItsCount` passes |
| F-r3-01 | fixed | result.md drops "which v2-07 also showed" and adds a per-frame table against `before/`. Checked against the files: toasts (`Install bash completed` / `Upgrade bash` / `Cannot untap`) are 3 → 0 in 15, 17, 18 at 140x40 and in 17, 18 at 80x24, and visible ones → 0 in 80x24/15. They are kept in 13 and absent in 19/20 in both sets. The Taps box shows the injected tap in v2-07 15–20 and the machine's taps now. Dashboard Taps is 3 in both v2-07 and now at 140x40 15 and 18 (row moved down one for the Outdated row) |

## Verify

| Check | Result | Evidence |
|---|---|---|
| V1 gate | pass | build, vet ok; `CGO_ENABLED=1 go test ./... -race -count=1` 11 `ok`; cover-check "All coverage floors met" (gui 84.9%) |
| V2 frames | pass | unchanged since r3's rerun (byte-identical, all W×H) |
| V3–V6 | met | unchanged from r3 |

## AC table

| AC | Commits | Verdict |
|---|---|---|
| AC-01 | 1471c38 (+ 253c3ba, 3eed39a) | met |
| AC-02 | 2237be9 (+ 253c3ba, 3eed39a) | met |
| AC-03 | fd55e63 (+ a8a1c38, e3c4e08, f6107b4, 39f34d2) | met |

## Findings

None.

Note, not a finding: at 80x24 frame 15, v2-07 shows only two of the three toasts in
part (`✗ U…`, `⚠ Cannot untap…`); the Brewfile menu covers the third. result.md's
"the same three lines" for that cell is a little loose, but its substance (toasts
removed) is right. Q-v2-08-a and Q-v2-08-b remain open for the orchestrator.
