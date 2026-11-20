# Receipt — runner, v2-10-tap-actions fix round r1

Source: `work/inbox/reviewer-v2-10-tap-actions-r1.md`, brief `work/briefs/fix-v2-10-tap-actions-r1.md`.

## Commit

- F-r1-01 — 8dd6f81 — fix(gui): the Taps Trust tab names t and T as they act (v2-10 F-r1-01)

## Finding

- **F-r1-01 (minor), Trust tab key hint.** `tapTrustStory` (internal/gui/details.go)
  said "Press t to trust or untrust the tap, or one of its formulae or casks." for every
  state. Now: a known non-official tap gets "Press t to trust or untrust the tap, and T to
  trust one of its formulae or casks."; an unread tap keeps "Press R to read the trust state
  again."; an official tap gets no key line, since t and T both refuse on it.
- Evidence: `TestTapTrustTabNamesTheKeysThatAct` (internal/gui/tap_actions_test.go) draws the
  Trust tab with `View()` for untrusted, trusted, official and unread taps and asserts the
  sentence per state, and that official and unread show no `Press t`/`Press T`. It fails on
  the old text for the official case.

## Frames

- Regenerated the v2-10 harness into the scratchpad and diffed it against
  `work/verification/v2-10-tap-actions/after/`. All 48 `.txt` and 24 `.ansi` frames are
  byte-identical. The Trust tab is not in the frame set, so no frame changed and none was
  copied in.

## Gate

Per-commit (Q-v2-builder-speed amendment): `go test -count=1 ./internal/gui/` passed after the
fix and the test, and `go vet ./internal/gui/` was clean.

Full gate, once, before this receipt:

- `make build` — ok
- `make lint` — ok
- `CGO_ENABLED=1 go test ./... -race -count=1` — every package ok
- `make cover-check` — All coverage floors met (brew 79.6%, gui 85.9%, presentation 98.5%, modal 90.1%)

## Safety

No brew write was run. The frame regeneration is read-only, as in the v2-10 harness.

## For the next session

- Nothing open from F-r1-01. The worktree is clean after this receipt commit.
