# Slice: lb16-update-when-stale — refresh brew's metadata at launch when it is stale

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb12-cask-token` closed. User ruling (2026-11-21): *update
when stale*.

## Problem

lazybrew runs brew with `HOMEBREW_NO_AUTO_UPDATE=1` (`internal/brew/runner.go`)
and `brew.update_on_start` defaults to `false`. A fresh launch therefore reads
tap metadata from the last `brew update`, so `brew outdated` misses recent
releases. The user saw 2 of 4 outdated packages; after an external
`brew update` all 4 appeared. They want to manage brew from the TUI alone.

## Behaviour

At launch, when `brew.update_on_start` is `auto` (new default), run the
existing `startBrewUpdate` path only if brew's metadata is older than
`brew.update_stale_hours` (default 6). When it finishes, the existing
`RefreshMsg` reloads Outdated. Otherwise do what Init does today.

`update_on_start` stays a bool in YAML for `true`/`false`; add the string
`auto`. `true` = always, `false` = never (today's behaviour), `auto` = when
stale. Default `auto`. Say how you parse it (custom `UnmarshalYAML` on a small
type, or a second key `update_when_stale: true`; pick the smaller change and
say which).

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — read brew's last-update time
- [ ] AC-02 — config: `auto` mode and `update_stale_hours`
- [ ] AC-03 — Init updates when stale; header shows it
- [ ] AC-04 — README and the update badge

### AC-01 — last-update time

1. Add `Diagnostics.LastUpdate(ctx) (time.Time, error)` or a plain function in
   `internal/brew`: the newest mtime of `$(brew --repository)/.git/FETCH_HEAD`
   (resolve the repository from the runner's `--repository`, cached; on a
   missing file return the zero time, which counts as stale). Do not shell out
   to `git`. Test with a temp dir and a fake runner.
2. A zero time or any error means "stale".

### AC-02 — config

1. `update_on_start`: `true` | `false` | `auto`, default `auto`.
   `update_stale_hours`: int, default 6, minimum 1, validated like the other
   fields. Tests: defaults, each value, and an invalid string is a validation
   error.

### AC-03 — Init

1. In `Init`: with `auto`, issue a command that reads `LastUpdate`; its
   result message decides: stale → `StartUpdateMsg`, fresh → fetch as today.
   While that read is in flight, still fetch the active panel so the screen
   is not blank. When the update finishes, `RefreshMsg{LoadOutdated: true,
   ReloadDiagnostics: true}` (check it already does; the earlier race between
   `brew update` and the `R` refresh, where "1 outdated" sometimes went
   missing, must not reappear: write a test that the Outdated count is
   present after the update ends).
2. `updateStatusText`: `never updated` becomes `updating…` while the stale
   update runs, and a fresh launch shows `updated Nh ago` from `LastUpdate`
   rather than `never updated`.
3. Tests: stale → one `brew update` enqueued through `task.Manager` and the
   Outdated list reloaded after; fresh → no `brew update`; error reading the
   time → treated as stale. teatest flow asserting on `View()`.

### AC-04 — docs

1. README config block: document `auto`, `update_stale_hours`, and that
   lazybrew never updates behind your back in `false` mode.

## Done means

- `make ci` green; receipt lists each golden that changed and why.
- Invariants: the update goes through `task.Manager`; no `program.Send`; no raw
  goroutines in handlers; floors not lowered.
