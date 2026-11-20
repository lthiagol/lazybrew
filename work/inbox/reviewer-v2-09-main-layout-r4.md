verdict: changes-required

# Reviewer — v2-09-main-layout, round r4

Range reviewed: `31e0206..ac5ff67` (40e09b1 ruling + fix brief r3, a0f5b42, 7e9b214, b2ac4f6, 5d22e50, 5d04885, f1ab719, 0a04a1a, ac5ff67).
Gate per Q-v2-builder-speed and its amendment.

## Verify

- **V1 gate (final tree):** `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` every package ok, 0 FAIL · `make cover-check` "All coverage floors met".
- **Per-commit (amendment):** b2ac4f6 and 5d04885, each checked out alone: `go build ./...`, `go vet ./...`, `go test -count=1 ./internal/gui/...` exit 0. 5d22e50 is frames and docs only.
- **V2 frames:** harness re-run against real brew, 47.8 s, reads only. All 60 files are byte-identical to the committed `after/`. All 40 `.txt` frames are exactly W×H. The 80x24 changes are only the Log row (`│ No commands executed yet`). `internal/gui/dump` removed; worktree clean.
- **Log padding:** `renderCommandLogBox` pads the body by `mainPadX` either side. `TestLogBoxHasInnerPadding` covers empty and with-lines at 140x40 and 80x24. Met.

## Earlier findings

| Finding | Status |
|---|---|
| F-r1-01 | holds — `table.go` unchanged since r3; `TestLongDescriptionIsCutNotDropped` passes |
| F-r1-02 | holds — `result.md` commit list still in git order with the breach named |
| F-r1-03 | holds — `pkgRows` unchanged |
| F-r2-01 | holds — `TestDescriptionKeepsItsFloorBesideALongName` passes; 80x24 Deps frames unchanged |

## Q-v2-09-a — the split at 140x40, read in the terminal

The threshold is right: `splitWidth = 82` (inner 84), with tests at 82/81 and `TestInfoSplitsAt140x40`. I read the 140x40 `.ansi` frames 03, 06, 17 and 19:

- **03 Formulae Info, 17 Cask Info: read well.** Fields on the left. The Description wraps at spaces with the hanging indent under its value: `Bourne-Again SHell, a` / `UNIX command interpreter`.
- **06 Tap Info: does not read well.** Remote: `https://github.com/danielfoehrkn/homeb` / `rew-switch`. In r3's single column at this same width it was whole on one line.
- **19 Services Info: does not read well.** All four paths on the right break mid-token: `…/systemd/user/sh` / `.brew.moshi-hook.service`, `…/opt/mosh` / `i-hook/bin/moshi-hook serve`, `…/var/log/` / `moshi-hook.log`. Meanwhile the left column (Name … Registered) uses about 25 of its 39 cells. In r3 each path was whole on one line.

## Findings

**F-r4-01 (major) — internal/gui/presentation/detail.go:41–63 (`Info`), internal/gui/details.go:65,163 — at 140x40 the split makes Tap Info and Services Info harder to read than the single column it replaces.**
The fix brief requires that every split Info view at 140x40 "reads well". Prose wraps at spaces, but paths and URLs have no spaces, so `Detail` breaks them at the column edge. Values that were whole at this width before the split are now broken mid-word: one in 06, four in 19 (frames quoted above). Moving the Remote and the Unit file right avoided a `…` cut on the left, but traded it for a mid-token break.
**The fix must achieve:** at 140x40, no path or URL in a split Info view breaks mid-token when the single column showed it whole. Two ways to get there:
- a view splits only when its right-column non-prose values fit their column, and otherwise keeps today's single column; or
- the column widths are chosen so those values fit (e.g. a narrow left column sized to its short fields).

Formulae and Cask Info must keep splitting at 140x40. Add a test on the Services Info and Tap Info fixtures at 87 cells: no right-column path is split across rows. Regenerate frames and name the changes.

## Notes (not findings)

- Q-v2-09-a took two commits: b2ac4f6, then 5d22e50 for the 200x60 frames left unstaged. That breaches "one commit per item", but the receipt says so truthfully.
- The Services actions line still sits one column right of the fields (`  S stop`). It did before this slice too.
