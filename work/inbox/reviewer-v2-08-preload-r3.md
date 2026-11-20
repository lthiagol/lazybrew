verdict: changes-required

# Reviewer — v2-08-preload, round r3

Range: `a37d56c..ea2f4a8` (fix brief 5a58503; 3eed39a, f6107b4; receipt ea2f4a8).
Brew ran at normal speed (harness 47.8 s, including its two 7 s toast waits). No brew
write ran: the harness's added `RefreshMsg` only reads (status and loaded panels).

## Earlier findings

| Id | Status | Evidence |
|---|---|---|
| F-r1-01 | holds | `TestEarlySwitchCountReachesTheDashboard` passes (both orders); `absorbLanded`/`fillUnknown` unchanged on the success path |
| F-r1-02 | holds | STATE.md v2-08 paragraph unchanged since b23a9d8 |
| F-r2-01 | fixed in code; docs F-r3-01 | harness injection now carries `Taps` consistent with `Items`; a `RefreshMsg` re-lands the real Taps read after frame 13; frames 15/18 (140x40) show dashboard `Taps 3`, Taps box `3` |
| F-r2-02 | fixed | a failed landing calls `forgetPanel` on `m.landed` and `m.status` and invalidates that panel's cache; `TestFailedReadForgetsItsCount` asserts on `View()` before and after a Status landing (`press R`, no `1 package`). My r2 probe scenario is now covered by that test |

## Verify

| Check | Result | Evidence |
|---|---|---|
| V1 gate | pass | build, vet ok; `CGO_ENABLED=1 go test ./... -race -count=1` 11 `ok`; cover-check "All coverage floors met" (gui 84.9%) |
| V2 frames | pass on size/repro | harness rerun `ok`; 40/40 frames byte-identical to committed `after/`, all exactly W×H; `internal/gui/dump` removed |
| V3–V5 | met | unchanged from r2; all `TestPreload*` pass |
| V6 | met | no new brew command; no test deleted or weakened; one commit per finding |

## Findings

F-r3-01 (minor) — work/verification/v2-08-preload/result.md ("Recapture after
F-r2-01", last paragraph and per-frame list) — the record of what changed against
`before/` (v2-07) is wrong in two places:
(a) It says the machine's three taps are what "v2-07 also showed". v2-07 frames 15–20
show the injected one-tap list (`before/140x40/15-menu-brewfile.txt` Taps box:
`▸ homebrew/cask-versions  official`; `before/80x24/17-cask-info.txt` likewise). The
three-tap Taps box in 15–20 is a new difference from `before/`, caused by the
harness's refresh.
(b) The toasts are gone from 15, 17 and 18 (140x40) and 17 (80x24). `before/` and
e3c4e08 each carry 3 toast lines there (`Install bash completed`, `Upgrade bash`,
`Cannot untap official taps`); the new frames carry 0. The per-frame list does not
name this, although it says "nothing else changed". Harness step 2 mentions only that
toasts are dropped.
Both come from the harness change, not production code. — Fix must achieve:
result.md names, per frame against `before/`, every difference the harness change
causes (Taps box rows in 15–20, toasts removed in 15/17/18), and drops the claim that
v2-07 showed the three taps. No code or frame change needed.
