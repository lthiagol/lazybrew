# v2-08-preload — result

Preload on: every panel but Search loads in the background after the Status
dashboard lands, two reads at a time, in the order Formulae, Outdated, Casks,
Taps, Services. Frames below are from a real run, read-only against this
machine's brew. No brew write ran: the harness never sends `y` or `enter` on a
confirm, and `S` is cancelled with `esc`.

## Commits (one per AC, in order)

| AC | commit | subject |
|---|---|---|
| AC-01 | 1471c38 | feat(gui): a preload scheduler loads the panels after Status (v2-08 AC-01) |
| AC-02 | 2237be9 | feat(gui): the dashboard fills in as preloaded panels land (v2-08 AC-02) |
| AC-03 | see the receipt | sidebar test, frames, this file, STATE row |

## What the code does

- **Config** (`internal/config/config.go`): `gui.preload` (default `true`),
  `LAZYBREW_PRELOAD=0|1` overrides it, anything else is an error (same as
  `LAZYBREW_ANIMATIONS`). README and DESIGN document both.
- **Scheduler** (`internal/gui/preload.go`): a `preloadState` on the model. The
  first `DataLoadedMsg` for Status starts it (with `update_on_start` that is the
  load after the brew update). Each panel that is neither loaded nor loading is
  queued, and `loading` is set while it waits, so the dedupe with `lazyLoadPanel`
  is the existing flag. At most `preloadWidth` (2) reads are in flight; each
  landing (a `preloadedMsg`) frees a slot and starts the next. Every read is
  `fetchPanelData`, so it goes through `boundedRead` and `brewReadTimeout`. No
  goroutine and no `program.Send`.
- **User switch to a queued panel**: `lazyLoadPanel` moves it to the front of the
  queue (`promotePreload`) and starts it when a slot frees. It never starts a
  second read, and the cap holds.
- **Dashboard fill-in** (`status.go` `absorbPanel`, `preload.go` `absorbPreload`):
  each successful preloaded read writes its count into the Status summary and the
  Status sidebar rows. A failed read leaves its count unknown and shows its error
  on its own panel only (AC-02.2).
- **Outdated row** (`dashboard.go`): `loading…` while the preload read is pending;
  `not loaded — press R` when the preload is off or the read failed.

## Test helpers and the frame harness (AC-01.6)

- **gui tests** (`internal/gui/*_test.go`): every `config.Default()` became
  `testConfig()` (`main_test.go`), which is the production default with
  `GUI.Preload = false`. Reason: these tests drive each panel's read themselves
  and count the reads they expect; a preload running behind them would add reads
  to every count. The preload tests in `preload_test.go` turn it on by hand.
- **testutil** (`internal/gui/testutil/program.go`, used by `flows`): preload off,
  for the same reason.
- **Frame harness** (`frame_dump_test.go.txt`): preload **on**, set explicitly, so
  the frames do not depend on the default.
- `TestProductionDefaultPreloads` (gui) and `TestPreloadDefaultIsOn` (config) pin
  the production default.

## Frames (AC-03.2)

`after/140x40/` and `after/80x24/`: 20 screens each, the `.txt` frames plus `.ansi`
at 140x40. `before/` is `work/verification/v2-07-op-dismiss/after/`, copied.

Run: `frame_dump_test.go.txt` copied to `internal/gui/dump/dump_test.go`,
`DUMP_DIR=<this dir>/after CGO_ENABLED=1 go test ./internal/gui/dump/ -count=1 -v`,
then the copy removed. Result: `ok`, 147 s. The harness settles every command
until the preload is done, and checks that at most one sidebar box (Search, which
is never preloaded) says `not loaded`.

Changes from the v2-07 harness, all in `settle` and the setup:
1. `settle` replaces `drain` for `Init` and the resize. A preload read starts
   from a landing, so `drain`'s depth 6 would drop the chain; the new cut is 64.
2. A command is waited for up to 120 s. The status dashboard's budget is 45 s
   and a brew read's is 30 s; a shorter wait gave up before the app did (a first
   run with 20 s did exactly that, and every box read `not loaded`).
3. Spinner ticks are not followed (they run for as long as the program does).
4. `gui.preload` is set to `true` explicitly.

### Sidebar: every screen whose sidebar changed, and why

Sidebar changed on 16 of the 20 screens at both sizes: 01 status, 02 formulae,
03 formulae-info, 04 formulae-deps, 05 outdated, 06 taps, 07 confirm-modal,
09 search, 10 toast, 11 deps-bash, 12 confirm-upgrade, 13 toast-stack,
14 input-modal, 15 menu-brewfile, 17 cask-info, 18 doctor.

Why: in v2-07 a panel's box said `not loaded` until the user visited it. Here the
preload has loaded it before the first frame, so its border shows its count.
Example, 140x40 `02-formulae`:

```
before                                         after
╭─ 3 ▣ Casks ────────────────── not loaded ─╮   ╭─ 3 ▣ Casks ─────────────────────── 2 ─╮
╭─ 4 ↑ Outdated ─────────────── not loaded ─╮   ╭─ 4 ↑ Outdated ──────────────────── 0 ─╮
╭─ 5 ⊕ Taps ─────────────────── not loaded ─╮   ╭─ 5 ⊕ Taps ──────────────────────── 3 ─╮
╭─ 6 ◎ Services ─────────────── not loaded ─╮   ╭─ 6 ◎ Services ──────────────────── 3 ─╮
╭─ 7 ⌕ Search ───────────────── not loaded ─╮   ╭─ 7 ⌕ Search ───────────── not loaded ─╮
```

At 80x24 the same change reads `– ` before (Outdated, Services) and `not loaded`
before (Casks, Taps), and the counts after. Search still says `not loaded`
(`–` at 80x24): it is never preloaded (AC-01.2).

Not changed: `08-help`, `16-help-scrolled` (help overlays), and `19-services-info`,
`20-services-stop-confirm`. In v2-07 those screens came after every panel had
been visited, so their sidebar was already loaded.

### Main area: the three screens that changed

- `01-status` (both sizes) and `15-menu-brewfile` (140x40 only): the dashboard's
  Packages card shows `Outdated: 0 packages` (before: `not loaded — press R`),
  and the gauge under it, because the Outdated count now fills in from the
  preloaded read (AC-02.3). At 80x24 the brewfile menu covers that row.
- `06-taps` and `14-input-modal` (both sizes): **not a preload effect**, see
  the environment note below. The Taps header reads `0 trusted · 3 untrusted`
  where v2-07 read `2 trusted · 1 untrusted`, and the selected tap's details
  show `Formulae: 0 / Casks: 0` instead of its remote.

## Environment: brew is slow on this machine right now

Measured with read-only brew commands during this slice:

| command | time | note |
|---|---|---|
| `brew info --json=v2 --installed` | 0.50 s | |
| `brew services list --json` | 0.36 s | |
| `brew config` | 0.28 s | |
| `brew doctor` | 1.50 s | |
| `brew outdated --formula --json=v2` | 20.8 s | prints "Downloading Homebrew API data… falling back to cached version" |
| `brew outdated --cask --json=v2` | 0.25 s | |
| `brew tap-info --json <3 taps>` | 87.4 s | rc 0, full data |

The tap-info batch is over the 30 s `brewReadTimeout`. When it times out, the
tap reader returns the name-only list and does not cache it (`taps.go`), so the
Taps panel shows no trust and no remote and no error. Preload is not the cause:
with `gui.preload` off, visiting Taps in this environment gives the same
`0 trusted · 3 untrusted` line (checked with a throwaway test, reverted). The v2-07
frames were captured while tap-info was fast. Two runs of the v2-08 harness gave
the same degraded Taps read.

The frames are kept as captured. Taps is loaded, and the sidebar count (3) is
right; only the Taps main area is degraded. A re-run when brew is quick again
should bring frames 06 and 14 back to the v2-07 content.

## Judgement calls (for the orchestrator to redirect)

- **Loading label (AC-03.1)**: the queued and in-flight box draws the existing
  spinner (accent colour), not a `loading…` text in Subtle. The brief allows
  "spinner or loading…"; the spinner is what the sidebar already draws for a
  loading panel, so I did not restyle it.
- **The 2-read cap** counts preload reads (including a user-opened queued panel,
  which is promoted, not started alongside). An `R` refresh's reads are the
  existing refresh path and are not counted.
- **Failed preloaded panel, switch to it**: `lazyLoadPanel`'s existing policy
  retries a panel with no items when it is switched to. So the error shows on the
  box, and switching to that panel retries it (its error returns if it fails
  again). I kept the existing policy. Test: `TestPreloadFailedReadShowsItsErrorOnItsPanelOnly`.
- **Refresh count**: a preloaded landing does not decrement `refreshing`, so a
  `R` refresh that overlaps the preload still reports "Data refreshed" at the
  right time.

## Tests

`internal/gui/preload_test.go` (new, 15 tests) and `internal/config/preload_test.go`
(new, 4 tests). Names: `TestPreloadOrderAndTwoInFlight`, `TestPreloadStartsWithTwoReads`,
`TestPreloadSkipsLoadedPanels`, `TestPreloadDedupesWithAUserSwitch`,
`TestPreloadFailedReadDoesNotStopTheOthers`, `TestPreloadOffStaysLazy`,
`TestPreloadWaitsForUpdateOnStart`, `TestPreloadLandingDoesNotCountAsRefresh`,
`TestPreloadKeysWorkMidPreload`, `TestPreloadMouseWorksMidPreload`,
`TestPreloadFailedReadShowsItsErrorOnItsPanelOnly`, `TestPreloadFillsTheDashboardCounts`,
`TestPreloadOffDashboardSaysPressR`, `TestPreloadSidebarSaysLoadingNotNotLoaded`,
`TestProductionDefaultPreloads`; config: `TestPreloadDefaultIsOn`,
`TestPreloadFromFile`, `TestPreloadEnvOverridesFile`, `TestPreloadEnvRejectsOtherValues`.

No existing test was deleted, skipped or weakened. The changed expectations are
the test helpers' config only (`config.Default()` → `testConfig()`, preload off).

## Not verified here

- The 2-read cap under real concurrency: it is a structural cap, checked by state
  (`inFlight`, `queue`) in `TestPreloadOrderAndTwoInFlight`, not by timing.
- Real-terminal feel of the background load (spinner, key latency, mouse).
- Behaviour when brew is slow (see the environment note): the reads are bounded,
  but a slow tap-info shows a degraded Taps panel with no error.
