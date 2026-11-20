# Receipt — runner, v2-10-tap-actions fix round r1

Source: `work/inbox/reviewer-v2-10-tap-actions-r1.md`; brief `work/briefs/fix-v2-10-tap-actions-r1.md`.
Correction: 1b64e4b was 46 lines and cc62776 was 34; this version (≤30) replaces both.

## Commit

- F-r1-01 — 8dd6f81 — fix(gui): the Taps Trust tab names t and T as they act (v2-10 F-r1-01)

## F-r1-01

- The Trust tab said "Press t to trust or untrust the tap, or one of its formulae or casks."
  in every state. Now a known non-official tap names t (tap) and T (formula or cask); an unread
  tap names R; an official tap names no key.
- Evidence: `TestTapTrustTabNamesTheKeysThatAct` (internal/gui/tap_actions_test.go), `View()` per state.

## Frames

- Harness regenerated to the scratchpad: all 48 `.txt` and 24 `.ansi` frames byte-identical to
  `after/`. The Trust tab is not in the frame set; none copied.

## Gate

- Per commit: `go test -count=1 ./internal/gui/` and `go vet` clean.
- Full, once: `make build`, `make lint` ok; `CGO_ENABLED=1 go test ./... -race -count=1` all ok;
  `make cover-check` All coverage floors met (gui 85.9%).

## Safety

No brew write was run.
