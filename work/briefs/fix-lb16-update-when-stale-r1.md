# Fix brief — lb16-update-when-stale, round r1 (before review)

> Slice contract: `work/briefs/lb16-update-when-stale.md`. This brief adds one
> change from ruling `Q-lb16-r-key-auto`. One commit, staged by name.

## F-r1-01 — `R` runs `brew update` under `auto`

In `internal/gui/keys.go`, the `R` case runs `brew update` only when
`UpdateOnStart == config.UpdateAlways`. Change the gate to
`!= config.UpdateNever` (and `!m.isUpdating`). The update path through
`task.Manager` and the load-flagged `RefreshMsg` stay as they are.

Tests, written first (show the failing run in the receipt):
1. `auto` + `R` enqueues exactly one `brew update` and, when it completes, the
   Outdated count is loaded and known.
2. `never` + `R` enqueues no `brew update` and only reloads (existing behaviour).
3. `R` while `isUpdating` enqueues nothing.
Update the help text or footer wording if it says `R` updates only with
`update_on_start: true`; README likewise. Name any golden that changes.

Gate: `make test-fast PKG=./internal/gui/... ./internal/config/...` for the
commit, `make ci` before the receipt. Receipt: append a "Fix r1" section (≤ 15
lines) to `work/inbox/runner-lb16-update-when-stale.md`, commit it
(`inbox: runner fix r1 lb16-update-when-stale`), reply with one done line.
