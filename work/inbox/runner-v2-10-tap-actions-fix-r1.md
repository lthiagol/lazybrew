# Receipt — runner, v2-10-tap-actions fix round r1

Source: `work/inbox/reviewer-v2-10-tap-actions-r1.md`; brief `work/briefs/fix-v2-10-tap-actions-r1.md`.

Correction: the first receipt commit (1b64e4b) was 46 lines, over the 30-line limit. This
trimmed version replaces it in a new commit; the content of the finding is unchanged.

## Commit

- F-r1-01 — 8dd6f81 — fix(gui): the Taps Trust tab names t and T as they act (v2-10 F-r1-01)

## Finding F-r1-01

- `tapTrustStory` (internal/gui/details.go) said "Press t to trust or untrust the tap, or one
  of its formulae or casks." in every state. Now: a known non-official tap names t (tap) and T
  (formula or cask); an unread tap names R; an official tap names no key (t and T refuse on it).
- Evidence: `TestTapTrustTabNamesTheKeysThatAct` (internal/gui/tap_actions_test.go) draws the
  Trust tab with `View()` per state and asserts the sentence, and that official and unread show
  no `Press t`/`Press T`.

## Frames

- Regenerated the v2-10 harness to the scratchpad and diffed against `after/`: all 48 `.txt`
  and 24 `.ansi` frames are byte-identical. The Trust tab is not in the frame set; none copied.

## Gate

- Per commit: `go test -count=1 ./internal/gui/` and `go vet ./internal/gui/` clean.
- Full, once, before the receipt: `make build` ok; `make lint` ok; `CGO_ENABLED=1 go test ./...
  -race -count=1` every package ok; `make cover-check` All coverage floors met (gui 85.9%).

## Safety

No brew write was run.
