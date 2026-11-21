# Slice: lb18-config-sample — `--config-sample` and `--config-create`

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb16-update-when-stale` closed. User request (2026-11-21).

## Goal

A user can get a complete, documented config file without reading the README:

```
lazybrew --config-sample                       # prints the sample to stdout
lazybrew --config-sample >> ~/.config/lazybrew/config.yml
lazybrew --config-create                       # writes it to the default path
lazybrew --config-create --config ~/x.yml      # or to the path given
```

## The sample

- Every option in `config.Config` appears, with a one-line comment saying what it
  does and what values it accepts.
- **The default value is uncommented**; each other accepted value is shown as a
  comment on the same line or the line below (`theme: dark   # dark or light`).
  Options whose default is empty or zero (`brew.path`, `gui.auto_refresh_seconds`)
  are still written uncommented with that default.
- The environment overrides (`LAZYBREW_ICONS`, `LAZYBREW_ANIMATIONS`,
  `LAZYBREW_PRELOAD`) are listed in a comment block at the top.
- A header comment says: delete a line to follow lazybrew's default, and the
  file is optional.
- **It is generated from `config.Default()`**, not a hand-typed block, so a
  changed default cannot drift from the sample. Keep the prose in one place (a
  table of key → comment or a template) next to the config package.

## Flags

- `--config-sample`: print the sample to stdout and exit 0. No brew, no config
  load, no TUI.
- `--config-create`: write the sample to the config path (`--config PATH` if
  given, else the default `~/.config/lazybrew/config.yml`; reuse the existing
  default-path function), creating the directory (0755) and the file (0644).
  If the file **already exists**, refuse: print one line naming the path and
  saying it was not touched, exit non-zero. Never overwrite. On success print
  the path written, exit 0.
- Both flags exit before the app starts, and are documented in `--help` and the
  README flag table and Configuration section.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — `config.Sample()` generated from `Default()`
- [ ] AC-02 — the two flags in `cmd/lazybrew`
- [ ] AC-03 — README

### AC-01 — `config.Sample()`

1. `config.Sample() string` per the rules above.
2. Tests (in `internal/config`):
   - the sample loads through the real loader with no error and yields exactly
     `Default()` (so uncommented defaults equal the real defaults, TTLs
     included);
   - **every yaml key of `Config` is in the sample**, found by reflecting on
     the `yaml` tags (nested too): adding a field without a sample line fails
     the test;
   - every key appears once uncommented;
   - the header lists all three env overrides (`grep` the constants);
   - changing a default (e.g. temporarily set `DefaultCasksTTL`) changes the
     sample: prove it by a test that builds a sample from a modified
     `Default()` copy if the API allows, otherwise say why not.

### AC-02 — flags

1. Add `--config-sample` and `--config-create`; put the logic in a function
   that takes a writer and path, so it is testable without the process
   exiting (`run(args, stdout, stderr) int` style if `main` allows; follow the
   existing code).
2. Tests: stdout equals `Sample()`; create into a temp dir makes the dir and
   file with the sample; create over an existing file leaves it byte-identical
   and returns non-zero; create with `--config` honours the path; neither flag
   starts the app or touches brew (use the mock runner or assert no runner is
   built).
3. `./bin/lazybrew --config-sample | head` and `--config-create --config
   <tmp>/c.yml` run by hand: paste the output in the receipt.

### AC-03 — README

Document both flags in the flag table, and add a "Create a config" paragraph
to Configuration. Replace the hand-written YAML block with a pointer plus the
sample's first lines, or keep the block if it still matches; say which.

## Done means

- `make ci` green; `GOOS=darwin go build ./...` ok.
- Invariants: no change to how the config is loaded; no new dependency;
  coverage floors not lowered; no golden changes.
