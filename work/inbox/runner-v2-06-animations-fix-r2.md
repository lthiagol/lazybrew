# fix round r2 — v2-06-animations

Commit: `2a135d8` fix(gui): the modal backdrop is dimmed at once, no fade (F-r2-01).

F-r2-01: fallback taken, the backdrop fade is dropped.
- Per-cell blend not taken. The resting frame must be the all-muted backdrop,
  and Mocha's inactive borders (`#45475a`) are darker than the muted role
  (`#6c7086`). Any blend to muted brightens those cells at some step, so
  "steps never brighten" and "resting frames unchanged" cannot both hold.
- Now: `dimBackdrop` at once on the modal's first frame; the modal slide stays.
- Removed: `backdropColor`, `fadeBackdrop`, `modalFade`, `backdropFadeAt`.
- Tests: `TestBackdropIsDimAtOnce` runs under dark and light (Latte) themes:
  the first frame's backdrop equals the resting backdrop. The fade tests
  (`TestBackdropFadeNeverBrightens`, `TestModalFadeSpringIsMonotone`) are gone.
- The brief's border-and-text test (fade step 1 not brighter than replaced
  frame) is not written as specified, since it fails on border cells by design
  of the resting frame. Orchestrator to rule if the per-cell path is wanted.

Gate (HEAD `2a135d8`): `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`
all ok; `make cover-check` all floors met.

Frames: `diff -rq` of `after/` against `v2-05s-services/after` is empty. The
motion re-dump changed only `motion-modal-01..05.ansi`.
`result.md` updated (fix round r2 section; slide wording).

Worktree clean. No push. No brew write run.
