verdict: changes-required

# Reviewer — v2-08-preload, round r2

Range: `4fc12b3..1f46f4c` (fix brief be30d4e; 253c3ba, e3c4e08, b23a9d8; receipt
1f46f4c). Brew ran at normal speed (`brew tap-info` 0.91 s). No brew write ran.

## Earlier findings

| Id | Status | Evidence |
|---|---|---|
| F-r1-01 | fixed | `absorbLanded` runs for every non-Status landing with preload on; `fillUnknown` keeps landed counts when a later Status lands. `TestEarlySwitchCountReachesTheDashboard` covers both landing orders on `View()`. My r1 probe scenario now shows the count. Note: two commits carry F-r1-01 (253c3ba code, e3c4e08 frames). The fix brief says one per finding, but the frames commit is evidence, not code. |
| F-r1-02 | fixed | STATE.md v2-08 paragraph describes the a8a1c38 frames and marks the slow capture superseded; Q-v2-08-b open; earlier receipt superseded by the fix receipt, as the brief asked. |

## Verify

| Check | Result | Evidence |
|---|---|---|
| V1 gate | pass | build, vet ok; `CGO_ENABLED=1 go test ./... -race -count=1` 11 `ok`; cover-check "All coverage floors met" (gui 85.0%) |
| V2 frames | pass on size/repro; see F-r2-01 | harness rerun on real brew `ok` 27.5 s, 40/40 frames byte-identical to committed `after/`, all exactly W×H; `internal/gui/dump` removed |
| V3–V5 | met | unchanged from r1 except F-r1-01 path; `TestPreload*` all pass |
| V6 | met | no new brew command; no test deleted or weakened |

## Taps 0 on frames 15 and 18: harness artifact, not a production defect

Production success emitters (`internal/gui/commands.go:1588-1644`) always carry the
typed slice that `Items` is built from, so `absorbPanel`'s `len(msg.Taps)` matches
the list. A failed `brew tap` returns `Err`, which `absorbLanded` skips. A tap-info
timeout returns the name-only list (`internal/brew/taps.go:89`), so the count is still
right. A real empty landing gives 0 only when brew really has 0. The harness instead
injects `DataLoadedMsg{PanelID: PanelTaps, Items: [...]}` with no `Taps` slice
(`frame_dump_test.go.txt:191`), a shape production never sends. So the frame's dashboard
says `Taps 0` while the sidebar Taps box says `1`, and the machine has 3 taps.

## Findings

F-r2-01 (minor) — work/verification/v2-08-preload/frame_dump_test.go.txt:191,
after/140x40/15-menu-brewfile.txt:8, 18-doctor.txt — frames 15 and 18 show a dashboard
state production cannot reach (Taps 0 / Total 0 / Third-party 0, against a sidebar box
of 1 and 3 real taps), and result.md:126 still says 15 changed only for Outdated.
Evidence: rerun frame 15, line 8 `Taps 0`, line 28 Taps box `1`. — Fix must achieve:
the injected landing has a production shape (a `Taps` slice consistent with `Items`), or
the real Taps read is re-landed after frame 13, so 15 and 18 show the machine's
dashboard counts. result.md names any change that remains in 15 and 18.

F-r2-02 (minor) — internal/gui/preload.go (absorbLanded), internal/gui/gui.go (landData
`fillUnknown`) — `m.landed` keeps a count after its source fails later, and a Status
landing then shows that stale count as known. Evidence: a temporary test (removed)
preloaded Outdated with 1 formula, then landed a failed Outdated read and a Status
landing with Outdated unknown. The Outdated panel shows `timeout`, but the dashboard
shows `Outdated:  1 package` (`status.Outdated={N:1 Known:true}`). Before 253c3ba
that read `not loaded — press R`. This is reachable after a write: the refresh's
Outdated read fails and the Status read finds the outdated cache invalidated. — Fix must
achieve: a failed landing forgets that source's count in `m.landed`, so an unknown count
is never filled from a read older than the source's last failure. Add a test for this
on `View()`.
