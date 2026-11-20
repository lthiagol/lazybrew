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

Captured twice in this slice. The first run (147 s) had brew slow and its Taps
frames degraded; that is recorded below. The committed frames are the second
capture, with brew at normal speed (`brew tap-info` 0.8 s, `brew outdated` 1.6 s,
measured just before the run). The run took 28 s.

Run: `frame_dump_test.go.txt` copied to `internal/gui/dump/dump_test.go`,
`DUMP_DIR=<this dir>/after CGO_ENABLED=1 go test ./internal/gui/dump/ -count=1 -v`,
then the copy removed. Result: `ok`. The harness settles every command until the
preload is done, and checks that at most one sidebar box (Search, which is never
preloaded) says `not loaded`.

Changes from the v2-07 harness, all in `settle` and the setup:
1. `settle` replaces `drain` for `Init` and the resize. A preload read starts
   from a landing, so `drain`'s depth 6 would drop the chain; the new cut is 64.
2. A command is waited for up to 120 s. The status dashboard's budget is 45 s
   and a brew read's is 30 s; a shorter wait gave up before the app did (a first
   run with 20 s did exactly that, and every box read `not loaded`).
3. Spinner ticks are not followed (they run for as long as the program does).
4. `gui.preload` is set to `true` explicitly.

### Data differs from v2-07 in one way that is not the preload

Brew now reports one outdated formula, `git-delta 0.21.1 → 0.21.2`. v2-07 had
none. So the Outdated count is `1`, not `0`, and the header shows `1 outdated`.
That data change appears on every screen, and it is why the Outdated box and the
Status rows differ from `before/` on screens where the preload did nothing else.
The frames show the real state of the machine at capture time; nothing was edited.

### Sidebar: every screen whose sidebar changed, and why

Sidebar changed on 18 of the 20 screens at both sizes: 01 status, 02 formulae,
03 formulae-info, 04 formulae-deps, 05 outdated, 06 taps, 07 confirm-modal,
09 search, 10 toast, 11 deps-bash, 12 confirm-upgrade, 13 toast-stack,
14 input-modal, 15 menu-brewfile, 17 cask-info, 18 doctor, 19 services-info,
20 services-stop-confirm.

Why: in v2-07 a panel's box said `not loaded` until the user visited it. Here the
preload has loaded it before the first frame, so its border shows its count.
Example, 140x40 `02-formulae` (copied from the frame files):

```
before                                              after
╭─ 3 ▣ Casks ────────────────────── not loaded ─╮   ╭─ 3 ▣ Casks ─────────────────────────────── 2 ─╮
╭─ 4 ↑ Outdated ─────────────────── not loaded ─╮   ╭─ 4 ↑ Outdated ──────────────────────────── 1 ─╮
╭─ 5 ⊕ Taps ─────────────────────── not loaded ─╮   ╭─ 5 ⊕ Taps ──────────────────────────────── 3 ─╮
╭─ 6 ◎ Services ─────────────────── not loaded ─╮   ╭─ 6 ◎ Services ──────────────────────────── 3 ─╮
╭─ 7 ⌕ Search ───────────────────── not loaded ─╮   ╭─ 7 ⌕ Search ───────────────────── not loaded ─╮
```

At 80x24 the same change reads `–` before (Outdated, Services) and `not loaded`
before (Casks, Taps), and the counts after. Search still says `not loaded`
(`–` at 80x24): it is never preloaded (AC-01.2).

19 and 20 changed in the sidebar only for the Outdated count (1, from the data
change above); their services rows are the same as v2-07, including the moshi-hook
PID row.

Not changed: `08-help` and `16-help-scrolled`. Both are help overlays that cover
the sidebar.

### Main area: the three screens that changed

- `01-status` (both sizes): the dashboard's Packages card shows
  `Outdated: 1 package` with the gauge (before: `not loaded — press R`). That is
  the preload's count fill-in (AC-02.3), and the 1 is the data change above.
- `05-outdated` (both sizes): the Outdated list now has `git-delta`, and its
  Info tab shows it. Data change only; the preload did not alter the panel.
- `15-menu-brewfile` (140x40 only): the Brewfile menu covers part of the
  dashboard, and the rows it leaves show `Outdated: 1 package` as in `01`. At 80x24
  the menu covers the dashboard, so no dashboard row changes there.

`06-taps` and `14-input-modal` no longer differ in the main area: the Taps header
reads `2 trusted · 1 untrusted` and the selected tap shows its remote, as in v2-07.

### Recapture after F-r2-01 (the injected Taps landing)

The harness's step 13 injects a Taps landing for the untap refusal. It had no
`Taps` slice, a shape production never sends, so frames 15 and 18 showed
`Taps 0` beside a Taps box of one tap (r2 finding). Two changes to the harness:

1. The injected landing carries `Taps` consistent with `Items` (one official tap,
   `homebrew/cask-versions`), the shape production sends.
2. After frame 13, `gui.RefreshMsg{}` is sent and settled, and its toasts are
   dropped the same way as before frame 19. The refresh reads the loaded panels
   and the status (reads only), so the machine's own three taps return for
   frames 15 on. Frame 13 is the only frame that shows the injected tap.

Frames regenerated with the same harness (`DUMP_DIR` as before; `ok`, 48 s). Against
the committed frames of e3c4e08, 18 files changed. The differences from `before/`
(v2-07) that the harness change causes are listed per frame below (F-r3-01). They come
from the harness, not from production code or from brew's data.

Taps dashboard counts are not among them: at 140x40, where the dashboard is visible,
the Status Taps count is 3 in v2-07 and in these frames. The row it sits on moved down
by one because of the new Outdated row, which is brew data.

| frame | 140x40 | 80x24 |
|---|---|---|
| `13-toast-stack` | Taps box row is the injected tap in its production shape (`▸ ✓ homebrew/cask-versions official`, with its trust icon); the Taps main panel shows that tap's trust, official, formula and cask rows where v2-07 showed `No custom taps`. Toasts kept. | the same Taps box and main-panel change (truncated). Toasts kept. |
| `14-input-modal` | no harness effect (taken before the injection); differences from v2-07 are brew data and the preload's counts. | same. |
| `15-menu-brewfile` | Taps box: the injected row is replaced by the machine's taps (`⚠ danielfoehrkn/switch untrusted`, `✓ lthiagol/tap trusted`, ...). Toasts removed: `Install bash completed`, `Upgrade bash: brew upgrade failed`, `Cannot untap official taps`. | Taps box first row replaced by the machine's first tap. Toasts removed (the same three lines). The inactive Formulae box shows its cursor row `▸ bash`, where v2-07 showed `alsa-lib` at that row; frame 13 (before the refresh) shows the Formulae rows unchanged. Selection is still bash. |
| `16-help-scrolled` | none (help overlay, sidebar hidden). | none. |
| `17-cask-info` | Taps box rows replaced by the machine's taps. Toasts removed (the same three). | Taps box row replaced. Toasts removed. Formulae cursor row, as in `15`. |
| `18-doctor` | Taps box rows replaced by the machine's taps. Toasts removed (the same three). | Taps box row replaced. Toasts removed. Formulae cursor row, as in `15`. |
| `19-services-info` | Taps box rows replaced by the machine's taps. No toasts were on screen in v2-07, so none removed. | Taps box row replaced. Formulae cursor row, as in `15`. |
| `20-services-stop-confirm` | Taps box rows replaced by the machine's taps, outside the confirm overlay. | Taps box first row replaced (`▸ h…` in v2-07, `▸ ⚠…` now: the machine's first tap); the rest of the row is under the overlay. Formulae cursor row, as in `15`. |

Why the toasts went: v2-07 had no wait between frame 13 and frame 15, so the toasts
of frame 13 were on screen at 15, 17 and 18. The harness change adds a 7 s wait and
a toast tick after frame 13, which drops them before 15. The same wait is what
drops them before 19 in v2-07, so they are gone from 19 in both sets.

Why the Formulae cursor row moved at 80x24: the refresh reloads the Formulae list
and recomputes its window; the cursor item is the same (bash). Frame 13 is taken
before the refresh and its Formulae rows are unchanged.

## Environment: the first capture ran with brew slow

Measured during the first capture (brew then):

| command | time | note |
|---|---|---|
| `brew info --json=v2 --installed` | 0.50 s | |
| `brew services list --json` | 0.36 s | |
| `brew config` | 0.28 s | |
| `brew doctor` | 1.50 s | |
| `brew outdated --formula --json=v2` | 20.8 s | "Downloading Homebrew API data… falling back to cached version" |
| `brew outdated --cask --json=v2` | 0.25 s | |
| `brew tap-info --json <3 taps>` | 87.4 s | rc 0, full data |

The tap-info batch went over the 30 s `brewReadTimeout`. When it times out, the
tap reader returns the name-only list and does not cache it (`taps.go`), so the
Taps panel showed no trust and no remote and no error. Preload was not the cause:
with `gui.preload` off, visiting Taps during that slow period gave the same
`0 trusted · 3 untrusted` line (checked with a throwaway test, reverted).

Brew was fast again at the second capture (tap-info 0.84 s, outdated formula 1.6 s,
cask 0.23 s), and the committed frames are from that run. The first run's Taps
frames are not kept; they were replaced.

Notes that predate this recapture and are now superseded: the STATE.md paragraph
and `work/inbox/runner-v2-08-preload.md` describe the first capture's degraded
Taps frames (06, 14) and the slow brew. Both are out of this commit's scope.
Q-v2-08-b (the silent fallback in `tapsReader`) still stands as a question: a
slow tap-info still degrades the Taps panel without an error.

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
