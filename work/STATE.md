# State — the resume record

One row per slice. This file is the resume record: a cold session reads it to
learn what is done, what is in flight, and where to pick up.

| slice | commits | state | verdict |
|---|---|---|---|
| detach-master-plan | 3a770a7, b95af56, 7428739, 3431b9d, 180d398, 17b1b8d, dce1468, 26f44cf, 88dbed9 | closed | ok |
| lb01-m13-deps-verify | 789d506, a7d5498, 196575b, b7dec88, 544e0ed | closed | ok |
| lb02-fetch-timeouts | 0027596, f7bc287, 00f73f2, 16c24f4, 8647165, 0cfb213 | closed | ok |
| lb03-dashboard-outdated-count | 8a7b191, 4ae32bd, 99e54b7, 45fc3f8, 9fd23dd, adb8418, 80065c4, b92ff9d, 93a036b | closed | ok |
| lb07-bound-remaining-reads | bd6dfd7, b5c1cc5, ea71699, 7b21fc0 | closed | ok |
| v2-01-charm-v2 | 1861585, 0a89f8e, 0f65ba6, 10aade4, 6eec052, 7f32f2f | closed | ok |
| v2-02-layout-bugs | 2611e00, a9ccbbe, a56fd8e, 2498d26, 23a031c, 5525dfd, 9046165, 8d9af6b, 690e5dd, 4a649cf | closed | ok |
| v2-03-look-foundation | a6f22ab, c569ee0, 85f5e3f, 0d4593e, e56d97c, d58fb96, 4aaa3b0, f6c950a, 02fcbae, 18796ad, b79b0e3, 96a5508, 6736abd, 56e78c0, f4f4abf, 3677383 | closed | ok |
| v2-03b-sidebar-lists | 0fa39ff, 4b16b65, c30158e, 33f6348, 01fb663, 9ba0f80, d68d66f, 3c38d91, 8d2c171, 752583b, 3fefcf8, 9c91b5d, b501d1f, 020011f | closed | ok |
| v2-04-overlays-styled | f8a8b10, 54fe613, aa432d6, b8c110d, 6b01ecb, 215aadd, 23dbb79, 40588cf, f6807c4, f4b4127, 64bc9d8 | closed | ok |
| v2-05-content-polish | 1192a40, ca14a11, c1b4b07, dfb79fb, 158e76e, f818b81, 0a03277, 7f8d584, 730c6ce, 0831cd8, da7d80e | closed | ok |
| v2-05s-services | e8add34, a4bc483, 421c0ba, 7d8b51c, befbf7e, a97a13a | closed | ok |
| v2-06-animations | 46736ea, 4f54fc0, 7c79a51, 0ebc253, d5fb165, 16358cf, ee7d782, 789caf6, 7a66ae2, 2a135d8 | closed | ok |
| v2-07-op-dismiss | 0dfd9ff, ede3959, 21b3b21, 21fb1a1 | closed | ok |
| v2-08-preload | 1471c38, 2237be9, fd55e63, a8a1c38, 253c3ba, e3c4e08, b23a9d8, 3eed39a, f6107b4, 39f34d2 | closed | ok |
| v2-09-main-layout | fd1bde3, 12a3b1b, 95766a6, 539a515, 7f6e93b, 801054e, 77f0e2c, d2b31bd, e8bde9a, b2ac4f6, 5d22e50, 5d04885, a3d3d93 | closed | ok |
| v2-10-tap-actions | 5f91936, 1e91fad, a4eaf97, 6d5c46a, d1ee781, f05b443, 9c4afce, 8dd6f81 | closed | ok |
| v2-11-cancel-and-test-speed | d76b8de, f11d2a6, 02e9cac, 74c4438, 1872c96, 5174d37 | closed | ok |

**detach-master-plan, closed.** The five slice commits implement the detach
(`AGENTS.md`, `README.md`, `DESIGN.md`, the `work/` lane-flow structure,
`SMOKE-CHECKLIST.md`, and the untracking of `master-plan/`). The four fix
commits answer the first review's findings F-01..F-04. The re-review returned
`ok`: the gate passes (`make build`, `make lint`, `go test ./... -race
-count=1`, `make cover-check`), the archive is byte-identical on disk, and no
live `mp` instruction survives. Receipts: `work/inbox/runner-detach-master-plan.md`,
`work/inbox/reviewer-detach-master-plan.md`,
`work/inbox/runner-fix-detach-master-plan.md`,
`work/inbox/reviewer-fix-detach-master-plan.md`.

**Accepted, not fixed:** reviewer finding F-05 (minor) — the fix receipt
explained a command's output wrongly; its recorded output was right and its
explanation was not. No acceptance criterion was affected, so the receipt was
left as the historical record and the wrong sentence was corrected only in
`work/briefs/review-fix-detach-master-plan.md`.

**lb01-m13-deps-verify, closed.** `LB-01` is closed and `LB-06` with it: the
M13 Deps-tab fix (`ac5bac6`) is verified in all five of its parts, and
`work/verification/lb01-m13-deps.md` carries the mutation evidence — for each
of the five invariants, the exact mutation, the command, and the verbatim
failing output — proving that **nine** tests across the five invariants fail
when the invariant they cover is broken. The first review found no blocker;
its two remaining findings (F-01, F-05) were fixed by `b7dec88` and
`544e0ed`, and the re-review returned `ok` after reproducing every mutation
byte-for-byte. Receipts: `work/inbox/runner-lb01-m13-deps-verify.md`,
`work/inbox/reviewer-lb01-m13-deps-verify.md`,
`work/inbox/runner-lb01-m13-deps-fix.md`,
`work/inbox/reviewer-lb01-m13-deps-fix.md`.

**Accepted, not fixed:** F-03 (minor) — the runner receipt was committed twice
(`5841cc9`, then `7cc8084` tightening it to the brief's line limit). The
second commit was content-harmless and history is not rewritten.

**lb02-fetch-timeouts, closed.** `LB-02` is closed. The four brew read paths
that used `context.Background()` — `fetchPanelData`, `fetchStatusData`,
`fetchSelectedSearchInfo`, `executeSearch` — are now bounded: 30s per
panel/status read, 15s per search read, and a separate 45s **total** budget for
the status dashboard so it resolves even when every source hangs. A timeout
surfaces as `*brew.TimeoutError` — the error type that already existed at
`internal/brew/errors.go:42` and had no user until now — detected from the
context's own error, because `exec.CommandContext` kills the brew process and
returns `"signal: killed"` rather than `context.DeadlineExceeded`. The
dashboard's six reads now run concurrently under the total budget; the reviewer
judged that sound (race-clean over five runs, deterministic ⚠ order, the shared
brew cache is mutex-guarded) and **required** by the total-budget guarantee.
The tab-content path was not touched. Review verdict `ok` after five mutations
proved the new tests bite, then one fix cycle for a negative duration the
reviewer found in the clamp (`brew list timed out after -1s` → now `0s`, and a
spent parent no longer starts the read). Receipts:
`work/inbox/runner-lb02-fetch-timeouts.md`,
`work/inbox/reviewer-lb02-fetch-timeouts.md`,
`work/inbox/runner-lb02-timeout-budget-fix.md`,
`work/inbox/reviewer-lb02-timeout-budget-fix.md`.

**Accepted, not changed:** reviewer finding F-02 (minor) — the dashboard's
fan-out puts both heavy `brew info --json=v2` loads and `brew doctor` in the
same instant, a higher peak than the sequential shape would have. The reviewer
judged it acceptable (Init and manual `R` only; auto-refresh is off by
default) and it is a shape note, not a defect. Reconsider only if the dashboard
feels slow on a cold cache.

**lb07-bound-remaining-reads, closed.** `LB-07` is closed: **no brew read in
the GUI is left unbounded.** The last two — the uninstall dependency check
(`Formulae.Uses`) and `toggleLeaves` (`Formulae.Leaves`) — now go through the
same `boundedRead` helper with the existing 30s `brewReadTimeout`; no new
timeout var was added and no timeout value changed. The important part was
G-2: the uninstall check used to **discard its error**, so a hung `brew uses`
produced a confirmation dialog indistinguishable from "nothing depends on this
package" — the user then confirmed an uninstall that might break other
packages, never told the check had not run. It now opens the dialog as before
but says *"Dependency check timed out after 30s — the list of packages that
depend on X may be incomplete."* Any other error keeps the old behaviour
exactly. `toggleLeaves` surfaces its timeout through the existing error toast:
`leaves: brew leaves timed out after 30s`. Review verdict `ok`, no findings;
four mutations (remove the notice, uncap either read, make the notice fire on
any error) each broke the named test. Receipts:
`work/inbox/runner-lb07-bound-remaining-reads.md`,
`work/inbox/reviewer-lb07-bound-remaining-reads.md`.

**lb03-dashboard-outdated-count, closed.** `LB-03` is closed. The status
dashboard no longer reports `Outdated:  0 packages` when it means *not
loaded*: it renders `Outdated:  — not loaded (press R to load)` until the
count is known, and a genuine zero still renders `Outdated:  0 packages` —
the formatter gained an explicit `presentation.OutdatedInfo{Count, Known}`,
because a plain `int` cannot tell "zero" from "unknown". **Only the `R` key
loads the count** (`RefreshMsg.LoadOutdated`); Init, the auto-refresh tick,
the install and uninstall paths and the task-completion path all stay
cache-only, so M9's rule — the dashboard never shells out to `brew outdated`
on its own — survives, now pinned by a test on every emitter. A status
refresh also stopped discarding the outdated cache, so a valid 30-minute
result is no longer thrown away. The first review found no defect in the
implementation but two unprotected behaviours and one edge case (with
`brew.update_on_start: true` the first `R` emitted no refresh at all); the
fix cycle answered all three, and the re-review returned `ok` with no
findings — four mutations, each breaking the named test. Q-04 (a cold load
costs one `brew outdated` call **per outdated class**, two in total, because
brew has separate formulae and casks readers) was raised by the runner, ruled
on by the orchestrator, and recorded in `work/questions.md`. Receipts:
`work/inbox/runner-lb03-dashboard-outdated-count.md`,
`work/inbox/reviewer-lb03-dashboard-outdated-count.md`,
`work/inbox/runner-lb03-fix.md`, `work/inbox/reviewer-lb03-fix.md`.

**Next slice: none open.** `LB-04` (chunk `brew tap-info --json` batches for
ARG_MAX) is **parked by the maintainer** — ARG_MAX needs thousands of
arguments and typical installs have far fewer taps; revisit only if it is hit
in the wild. `LB-05` is a human gate for the stable tag. The release itself
is deferred: `wip` is 80+ commits ahead of `stable`, `v1.0.0-rc1` is still the
only tag, `M22.1b` / `M22.4` in `AGENTS.md` remain open, the signed-off smoke
checklist does not cover the mouse-support commit, and the work is unpushed
so CI has not yet seen it.

**How a cold session resumes work.** Read `work/briefs/<slice>.md` for the
contract of the slice in the row above, then read the receipts in `work/inbox/`
for that slice — the receipt is the record of what was actually done and what the
evidence was. The `commits` column tells you which commits to inspect, in AC
order. If a lane reported a blocker, it is named in the receipt and the
corresponding question (if any) is in `work/questions.md`. If the state is
`implementing` but the receipt is missing, the lane died mid-slice: re-read the
brief, check `git log`, and redo only the criteria without a commit. Do not
re-derive the plan from anywhere else — the brief is the contract, and
`master-plan/` is a frozen, untracked, local-only archive that is not an input to
any decision.

**v2-01-charm-v2, closed.** lazybrew now runs on the Charm v2 stack
(`charm.land/bubbletea/v2 v2.1.2`, `lipgloss/v2 v2.1.1`, `bubbles/v2 v2.2.1`,
`x/exp/teatest/v2`); no v1 module remains. The user sees **no change**: the
18 ANSI-stripped frames in `work/verification/v2-01-charm-v2/baseline/` (nine
screens × 140×40 and 80×24, real Homebrew) are byte-identical after the port,
reproduced by the reviewer in both rounds. Lip Gloss v2 counts borders inside
`Width`/`Height`, so box, help, log and modal sizes were adjusted in code, not
in the baseline. Terminal setup (alt screen, `WindowTitle: "lazybrew"`, mouse
mode from `gui.mouse`) moved into `View()`. A `tea.PasteMsg` case keeps v1
paste behaviour, pinned by three tests. CI reads its Go version from `go.mod`
(v2 needs Go 1.26). Tests 406 → 410, none removed or weakened. Review r1
returned two findings (CI Go pin, paste untested), both fixed; r2 returned
`ok`. Receipts: `work/inbox/runner-v2-01-charm-v2.md`,
`work/inbox/reviewer-v2-01-charm-v2-r1.md`,
`work/inbox/runner-v2-01-charm-v2-fix-r1.md`,
`work/inbox/reviewer-v2-01-charm-v2-r2.md`.

**Note for later slices:** on this machine `go env CGO_ENABLED` is 0, so the
race gate runs as `CGO_ENABLED=1 go test ./... -race -count=1`. Real-terminal
mouse, keys and paste are left to the human smoke gate (LB-05).

**v2-02-layout-bugs, closed.** Every frame is now exactly the terminal size in
every state (`TestEveryFrameIsTerminalSize`: 18 states × 140×40, 80×24,
200×60, plus too-small) — the guard for the rest of the overhaul. Modals and
toasts are Lip Gloss v2 compositor layers over a full-size base frame (no more
modal below the screen, no layout shift on a toast); modals draw their content
on the first frame (the `Loading...` gate is gone). A never-fetched panel says
`not loaded`, a fetched empty one `(empty)` (explicit `loaded` flag). The Log
is a sibling box below the main panel. The Formulae Deps tab shows real
dependencies: brew's `runtime_dependencies` objects carry `full_name`, which
`flexStrings` now reads (`bash` → `ncurses`, `readline`). Two extra bugs found
and fixed by the runner: the main list was capped by the sidebar's row count,
and help overflowed its box at 80×24. Review r1 found the main list still
scrolled on the sidebar's window (major) and a weak first-frame modal test
(minor); both fixed, r2 `ok`. Tests 410 → 426. Frames:
`work/verification/v2-02-layout-bugs/after/`. Questions raised and ruled:
Q-v2-hints, Q-v2-small (→ v2-03), Q-v2-help, Q-v2-cask-deps (→ v2-05).

**v2-03-look-foundation, closed.** The new look's foundation, in the style of
the change-gate operator TUI: the full Catppuccin palette (Mocha, and real
Latte for the light theme) behind semantic roles, with the app painting its
own canvas; Nerd Font icons by default with a Unicode set
(`gui.icons` / `LAZYBREW_ICONS`) and `show_icons: false` honoured; one panel
primitive with the title in the top border and a state label on the right
(count, `–` not loaded, `…` loading, `!` error — counts always whole, the
title yields first); a one-row header (app, Homebrew version, running
operation, count/outdated/update badges, unknowns omitted); a one-row footer
with the panel's keys first and `? more` when cut. Review r1: a regression test
that no longer guarded R3-4 (blocker), plus four minors, plus orchestrator
rulings Q-v2-03-a/c/d — all fixed; r2: a 4-digit count vanished at 80 columns
(major) — fixed; r3 `ok`. Tests 562 → 615 (count by `-v` PASS lines).
Frames: `work/verification/v2-03-look-foundation/after/` (`.ansi` at 140×40).

**v2-03b-sidebar-lists, closed.** User ruling Q-v2-sidebar (option B,
lazygit-style) is in: the sidebar holds the lists, built from typed data
(icon, name, right-aligned version, `1.2 → 1.3` in the warning role, pin
mark; full-row highlight in the focused box only; 35 % wide, cap 60); the main
area shows only the selected item's details and tabs — never the list again.
Formulae lost its `List` tab (Info is the default and shows the tap; tab
indices are named constants); Outdated `Versions` shows the selected item and
the batch summary; Taps `Tap Info` the selected tap, `Trust` only its trust
story; Services `Status` the selected service. The Status data is typed, so
the header shows formulae/casks counts on first launch. The empty viewport was
replaced by a per-item scroll offset: wheel follows the pointer, PgUp/PgDn
page, Home/End (and g/G outside Search) jump. Review r1: three minors plus
rulings Q-v2-03b-a/c/d (batch marks only on Outdated, Trust tab, page keys);
r2: `g`/`G` had stolen type-to-search on the Search panel (major) and a plural
(minor); r3 `ok`. Tests 615 → 633 PASS lines (some table tests removed with
named replacements). Frames: `work/verification/v2-03b-sidebar-lists/after/`.

**v2-04-overlays-styled, closed.** Overlays in the change-gate style: every
confirm carries an action kind (danger red, upgrade yellow, install blue,
neutral mauve) chosen at the call site; the dialog has a thick accent border,
an action chip (` UNINSTALL `), the question, a consequence line on every call
site, and `[y]es [n]o esc cancel`; input and menu share the frame. Prose wraps,
single-line parts are cut (ruling Q-v2-04-wrap). The frame behind an open
modal is muted at the compositing step. Toasts are a stack of up to 3 icon
boxes, bottom-right, each its own layer; 3 s (errors 6 s); a click dismisses.
Review: r1 toast padding blanked frame cells and most confirms lacked the
consequence line (both major) + 3 minors; r2 the input modal was off-centre;
r3 the per-row centring scattered the dependents list; r4 `ok`. Frames:
`work/verification/v2-04-overlays-styled/after/` (incl. 12–15).

**v2-05-content-polish, closed.** The main area now says things well: one
label/value helper for every detail view (Muted labels, Info-role URLs,
empty fields omitted, prose wraps); dependency lists mark installed vs not;
cask `depends_on` objects are parsed (formulae, casks, macOS/arch
requirements); the Status dashboard is cards with an outdated gauge (never `0`
when unknown) and Health right after Homebrew when stacked; Doctor lists each
warning; help scrolls and lines up; the Operation view shows steps from
brew's `==>` lines, elapsed time and, for batch/upgrade-all, a progress bar
whose total sums every header brew prints (formulae, casks, dependents). The
dead text dashboard was deleted and its invariant tests retargeted to the
cards. Review: r1 progress counted brew's header as a package and the
dashboard tests checked invisible text (both major) + 2 minors; r2 progress
ignored the casks/dependents headers (major) + 2 minors; r3 `ok`. Known limit
(not a defect): a dependency upgraded inside another package's step has no
step of its own, so the bar reads low mid-run and lands on total/total.
Frames: `work/verification/v2-05-content-polish/after/`.

**v2-05s-services, closed.** The user's Services request: `brew services info
--all --json` is read into a cached, typed `ServiceInfo` (invalidated by every
services write, read through `boundedRead`); the Services `Info` tab shows
status (icon + word), PID, user, loaded/registered, unit file, command,
working dir, log paths and schedule, with the actions that apply now as a key
line and in the footer. `S`/`r` confirm (danger/upgrade); start/stop/restart
are state-aware following brew's own loaded rule (`started`, `error`,
`scheduled` are loaded); after a services task the list and info reload with
the selection kept by name; a root-needing service gets a plain "needs sudo —
run …" toast (lazybrew never runs sudo). Long toasts wrap to ≤ 3 rows.
Review: r1 `error`/`scheduled` services could no longer be stopped/restarted
(major) + 2 minors; r2 `ok`. **Safety:** `moshi-hook` (live on this machine)
was never touched — MainPID 800, NRestarts 0, active since 2026-11-18 before
and after the slice. Not verified live: a real needs-sudo refusal (no system
service here). Logs tab stays parked (LB-09).

**v2-06-animations, closed.** Spring motion with harmonica (critically damped):
the tab underline slides, sidebar boxes grow/shrink with focus (heights always
sum to the sidebar), modals drop in, toasts slide in from the right and
re-stack, the batch progress bar springs. One 60 fps tick only while something
moves; no goroutines. `gui.animations` (default on) and `LAZYBREW_ANIMATIONS`
switch it; test helpers and the harness build still models. Resting frames
are byte-identical to `v2-05s-services/after/`; mid-motion frames in
`work/verification/v2-06-animations/motion/`. Review: r1 test helpers relied on
the caller's env and app tests read the real config (major), motion work ran
on every message even with animations off (187 allocs → fixed to 2), a
backdrop flash, and a red intermediate commit (recorded); r2 the fade still
brightened Mocha's dark borders — the fade was dropped (no monotone fade can
reach the muted resting backdrop), the backdrop now dims at once while the
modal still slides; r3 `ok`.

**LB-08 (visual overhaul) — all slices closed:** v2-01 … v2-06 plus v2-05s.
Next: the user checkpoint, then LB-05 (human smoke gate) and the `wip` →
`stable` PR. Not verified anywhere in the overhaul: real-terminal feel (mouse,
keys, paste, motion), a live brew write (install/upgrade/services), a real
needs-sudo refusal.

**v2-07-op-dismiss — closed (r1 `ok`).** A successful operation closes
itself after 2 s (a tick that names its operation; never mid-batch); failed
and cancelled runs stay with ` · Esc to close`; the footer leads with
`esc close`; switching panel closes a finished operation. Resting frames equal
v2-06 except moshi-hook's live PID (ruling Q-v2-07-pid).

**v2-08-preload — implemented, awaiting reviewer (fix round r1 applied).** `gui.preload`
(default on; `LAZYBREW_PRELOAD` overrides) loads every panel but Search in the
background after the Status dashboard, two reads at a time, through the scheduler in
`internal/gui/preload.go`. With the preload on, every panel landing (preloaded or
opened early) fills the dashboard count (F-r1-01). The sidebar shows the spinner, not
`not loaded`, while a read is queued or in flight. The committed frames are the
recapture in `a8a1c38`, taken with brew at normal speed: Taps shows its trust data,
and Outdated shows the one real outdated package. Frames 15 and 18 were regenerated
after F-r1-01 (see that commit). The first capture, taken while brew was slow, is
superseded: its Taps frames were degraded, and that text is no longer current. Q-v2-08-b
stays open: a slow `brew tap-info` still degrades the Taps panel without an error. The
judgement calls are in `work/verification/v2-08-preload/result.md`.

**v2-08-preload — closed (r4 `ok`).** With `gui.preload` on (default; env
`LAZYBREW_PRELOAD`), Status loads first, then Formulae → Outdated → Casks →
Taps → Services in the background, at most 2 reads at once, through the
existing bounded reads; boxes show a spinner while queued/in flight; the
dashboard fills in as panels land, and a failed read clears its count. Review:
r1 an early switch lost its dashboard count, stale STATE text; r2 a failed read
kept a stale count, harness showed an impossible `Taps 0`; r3 result.md
misstated harness-caused differences; r4 `ok`. Q-v2-08-b (silent name-only
Taps list) moved into v2-10 as AC-00.

**v2-09-main-layout — closed (r5 `ok`, the last allowed round).** The main
area has one column of padding (Log box too); lists are full-width tables
(columns cut with `…`, a column dropped only below its minimum width,
Description then Version); Info views split into fields | prose when the main
box's inner width is ≥ 84 (140x40) **and** every path/URL in the right
column fits whole — so Formulae/Cask Info split at 140x40, Tap/Services Info
stay single-column there and split wider; the Log box shrinks when idle.
Review: r1 a long description dropped the whole column, commit order, per-row
map; r2 a kept column could shrink to `G…`; r3 ok → orchestrator ruling
Q-v2-09-a lowered the split threshold (100 never fired at 140x40) and padded
the Log; r4 split broke paths mid-token; r5 `ok`.

**v2-11-cancel-and-test-speed — closed (r2 `ok`).** brew runs in its own
session (`Setsid`), so a cancel kills its whole process tree (Esc on an
install no longer waits for curl/git) and a step that wants a sudo password
fails at once with `needs sudo — run "brew …" in a terminal` instead of
hanging. Tests no longer wait on the wall clock (race run 38.9 s → 20.8 s);
CI runs `-race`. Review: r1 `ok` with an observation (sudo prompt would stop
silently under `Setpgid`) → orchestrator ruling Q-v2-11-b; r2 `ok`.
Accepted trade-off: any tty prompt brew triggers (sudo, git credentials) fails
fast instead of drawing over the TUI.

**v2-10-tap-actions — closed (r2 `ok`).** A Taps list read without tap-info
says `trust unknown — press R` and refuses `t`/`T` (Q-v2-08-b); Tap Info and
the footer show the actions that apply (official: `r repair` only); `t` trusts
an untrusted tap at once and confirms an untrust; `T` opens the per-formula /
per-cask trust menu; the untap confirm states the consequence and lists
installed packages from the tap; after a tap task the list reloads with the
cursor on the same tap. Review: orchestrator ruling Q-v2-10-a before review;
r1 the Trust tab's key hint was stale; r2 `ok`.

**Q-v2-next run (v2-07 … v2-11) — all closed.** Next: the user checkpoint,
then LB-05 (human smoke gate) and the `wip` → `stable` PR. Not verified by
agents: real-terminal feel, any live brew write (install/upgrade/tap/untap/
trust/services), a real sudo refusal.
