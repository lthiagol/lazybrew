# Backlog — parked work

Work that is real, known, and **not** in any active slice. When a slice picks an
item up, move it into that slice's brief and drop the row here.

The `source` column points at where the item came from — the old `mp` plan, which
is now a frozen local archive (untracked by git, history only, not an input to any
decision). The source is provenance, not a spec.

| id | work | source | priority | status |
|---|---|---|---|---|
| LB-01 | Verify the M13 Deps-tab fix by running the Deps test suite (`TestDepsTabLoads`, `TestDepsSelectionChangeUpdatesContent`) | M13 (execution done, independent review never passed) | medium | closed |
| LB-02 | Apply context timeouts to the four brew fetch paths: `fetchPanelData`, `fetchStatusData`, `fetchSelectedSearchInfo`, `executeSearch` | B-02 (residual from M13 F-02/F-07) | medium | closed |
| LB-03 | Status dashboard outdated count stays 0 until the first refresh | B-03 (M09 F-02 residual) | low | closed |
| LB-04 | Chunk `brew tap-info --json` batches when the tap count risks ARG_MAX — slice `lb04-lb10-small-fixes` | B-04 (M11 F-01 residual) | low | closed |
| LB-05 | Full interactive human smoke before the stable tag — **human gate**, not an agent task; checklist `SMOKE-CHECKLIST.md` *Run 2*; deferred by the user 2026-11-20 (`Q-backlog-next`) | B-05 | medium | closed (human smoke Run 2, 2026-11-21) |
| LB-06 | The M13 independent review itself — closed by verification together with LB-01; all 7 findings were already fixed, so the review is not left silently dropped | M13 | medium | closed-by-LB-01 |
| LB-07 | Apply context timeouts to the two brew reads still unbounded: the uninstall dependency pre-check (`internal/gui/commands.go:1034`, `Formulae.Uses`) and `toggleLeaves` (`internal/gui/commands.go:1128`, `Formulae.Leaves`) | found by lb02-fetch-timeouts | low | closed |
| LB-08 | Visual overhaul (change-gate look, overlays, harmonica animations) — slices v2-01..v2-06 + v2-05s closed 2026-11-20; checkpoint passed 2026-11-20; follow-up slices v2-07..v2-11 (Q-v2-next) closed 2026-11-20; code-complete; human smoke passed 2026-11-21 | orchestrator review 2026-11-19 | medium | closed |
| LB-09 | Services Logs tab: tail a service's stdout/stderr log files from lazybrew — slice `lb09-service-logs` | user, 2026-11-19 | low | closed |
| LB-10 | Confirm dialog: when a long dependents list is cut, the closing `…` row is centred instead of placed in the list column (`dialogMaxRows` cut in `modal/frame.go`, > ~12 body rows) — slice `lb04-lb10-small-fixes` | reviewer v2-04 r4 | low | closed |
| LB-11 | Run brew steps that need a terminal (macOS `.pkg` casks asking for sudo) with the TUI suspended (`tea.ExecProcess`), so the user can type the password; today they fail fast with the needs-sudo text — the user ruled that message is enough for now (2026-11-20) | orchestrator + user, 2026-11-20 (Q-v2-11-b) | low | closed (lb23-sudo-in-terminal r2; macOS manual test pending) |
| LB-DROP-01 | Dead header fix in the `master-plan-bkp` directory | B-01 | n/a | dropped |

**LB-DROP-01 reason:** the target directory does not exist on this machine —
so the fix cannot be applied and there is nothing to preserve.

**LB-01 closed** — the M13 review is closed by this verification; the evidence
is in `work/verification/lb01-m13-deps.md`, carried by commit `789d506`.

**LB-02 closed** — the four brew fetch paths are bounded by context
timeouts: `internal/gui/commands.go` (commits `0027596`, `f7bc287`) and the
tests in `internal/gui/fetch_timeouts_test.go` (commit `00f73f2`); the
receipt is `work/inbox/runner-lb02-fetch-timeouts.md`.

**LB-07 closed** — the last two unbounded brew reads (the uninstall dependency pre-check and `toggleLeaves`) are bounded by `brewReadTimeout`, and a timed-out dependency check now says so in the uninstall confirm dialog instead of looking like "nothing depends on this package"; receipt `work/inbox/runner-lb07-bound-remaining-reads.md`.

**LB-03 closed** — the status dashboard now says the outdated count is not loaded until the user presses R, which is the only path that loads it; receipt `work/inbox/runner-lb03-dashboard-outdated-count.md`.

**LB-12 (closed, lb12-cask-token r1)** — cask rows carry brew's display name (`Claude Code`), not
the token (`claude-code`), so every cask action sends the wrong name.
`parseCask` should read `token` (fall back to `name`), with a fixture in the
current brew shape. Raised by cr1 (`Q-cr1-cask-name`).

**LB-13 (closed, lb21-backlog-sweep r1)** — the Status, Outdated and Taps footers leave `[ ] tabs` off
(Formulae, Casks and Services show it), and Status `j/k` are help-only although
Status lists rows. Showing them changes footer frames and goldens. Raised by
cr4 (`Q-cr4-gui-dedup-r2-footer-tabs`); the keymap test records the omission in
`footerExempt`.

**LB-14 (closed, lb21-backlog-sweep r1)** — `TestSingleflightPropagatesError` now waits 20 ms for the
second goroutine to queue, because `sfCall.waiters` (the v2-11 AC-02 hook) was
deleted by cr5 AC-03. A slow `-race` run could flake. Needs a race-free hook if
it ever does. Raised by the cr5 r1 review.

**LB-15 (closed, lb21-backlog-sweep r1)** — `brew.SetDebug` is kept only for `app/debug_log_test.go`
(a cross-package logger reset); its `true` branch logs to stderr, the C13
hazard. Narrow it to a reset-only helper later. Raised by the cr5 r1 review.

**LB-17 (closed, lb20-cask-followups r1)** — installed casks get an empty `Cask.FullName`: `brew info`
writes `full_token`, not `full_name`, so `installedFromTap` never lists casks
and the untap warning misses them. Read `full_token`. Raised by the lb12 r1
review.

**LB-18 (closed, lb21-backlog-sweep r1)** — `TestSearchTabCyclesAllPanelsFlow` takes about 21 s under
`-race` (15 keys, a `View()` check after each). Make it cheaper (fewer waits,
or drive the model without a teatest program) if `make ci` time matters. Raised
by the lb19 r2 review.

**LB-19 (closed, lb20-cask-followups r1)** — Search → info runs `brew info --json=v2 <name>` with no
`--cask`; for a name that is both a formula and a cask brew may show the
formula. The row's type is now certain, so pass `--cask` for cask rows. Raised
by the lb19 r1 review.

**LB-20 (closed, lb22 r2)** — two edges from the lb19 r2 review: (a) if one typed search
read exits 1 (no results) and the other fails, the user sees an empty list and no
error, which reads as "no results"; (b) two writes that finish back to back can
land their Search re-marks out of order, so a slow first re-mark could restore a
stale mark (the name guard does not cover it).

**LB-21 (closed, lb22 r2)** — leftovers from the lb21 r1 review: three other singleflight
tests keep 20 ms sleeps (same synctest fix as LB-14); `EnableFileLogging` never
closes its log file; the `ResetLogger` doc in `internal/brew/logger.go` says
nothing in production calls it, but `init()` does (wording).

## Planned — user requests, 2026-11-21 (macOS smoke follow-ups), in this order

**LB-22 (closed, lb25-search-polish)** — Search rows on the cell system: name
left, `formula`/`cask` right-aligned; a green mark before the name replaces the word
`installed`; drop the duplicate `I` (Search `i` only). Decided: a green check ✔ (like brew) marks installed on Formulae, Casks and Search alike.

**LB-23 (closed, lb26-list-filter)** — `/` on Formulae and Casks filters the
loaded list locally as you type (generic panel filter; Outdated later). Global search
stays on `7` then `/`, and `/` on the other panels.

**LB-24 (closed, lb27-status-tabs)** — `m` (brew missing) and `v` (brew vulns)
become Status tabs beside Config and Doctor (Config | Doctor | Missing | Vulns): run on
demand, cached, kept across panel switches; `m`/`v` jump to the tab and refresh.

**LB-25 (closed, lb28-themes)** — themes: Catppuccin Mocha/Macchiato/Frappe/Latte,
Dracula, Nord; `gui.theme` takes the names (`dark`/`light` stay as aliases for Mocha/
Latte). Config only: no live switching (user ruling 2026-11-21).

**LB-26 (closed, lb29-footer-redesign)** — footer: groups (panel actions |
global), dim `·` between hints, one arrow hint `←↑↓→ navigate` (user ruling), `tab next`,
`1-7 jump`; `h/j/k/l` and `shift+tab` leave the footer and stay on `?`. Touches every
golden and the keymap test, so it goes after the key changes above.

**LB-27 (parked)** — Latte (`light`) roles under the contrast floors: Muted 2.30:1 (floor
3), Success 2.96, Warning 2.31, Info 4.34, Subtle 4.37 (floor 4.5). Remap them to darker
official Latte tokens (e.g. Subtext1/Overlay2 for Muted, Green/Yellow shades) and update
the `light` goldens, if the user finds light mode hard to read. Raised by lb28 (`Q-lb28-latte-contrast`).

## Planned — user requests, 2026-11-21 (after the macOS smoke)

**LB-28 (closed, lb30-taps-and-trust)** — the Taps pane becomes `Taps & Trust`: beside the
taps it lists the individually trusted formulae, casks and commands (`brew trust --json=v1`),
with an Info tab and `t` to untrust one. Fix the stale `ListTrusted` JSON shape. User ruling:
name `Taps & Trust`, commands shown too.

**LB-29 (closed, lb31-cleanup-tabs)** — Status tabs `Cleanup` (brew cleanup -n: list, sizes,
total; `c` runs it after a confirm) and `Unused` (brew autoremove -n; `A` runs it after a confirm),
loaded on open and refreshed after the run. Seven Status tabs must still fit at 80 columns.

**LB-30 (closed, lb32-vulns-view)** — the Vulns tab renders `brew vulns --json
--max-summary 0` (real schema checked once against the live command): a summary strip with
severity chips, one block per package (worst first) with a table of ID / severity / full title,
theme-coloured severities (CRITICAL/HIGH error, MEDIUM warning, LOW info, UNKNOWN muted), a
per-package next step (upgrade available → version from Outdated, no fix released, fixed
upstream not released), commit-hash lists collapsed (kept in `y` copy), dimmed footer for
patched/skipped counts, text fallback when the JSON cannot be read. User request 2026-11-21.

**LB-31 (parked)** — one unexplained `GOFLAGS=-p=1 make ci` failure (exit 2 at `make test`, log lost)
right after the lb31 r3 builder exited; five later full runs (one `make ci`, one `go test ./... -race`,
four `-count=4` on gui/brew/config) passed. Suspect a timing-sensitive test under load. If it shows
again, keep the log and file the test name.
2026-11-22: seen twice in one session as `TestTapsTrustedItemFlow` (internal/gui/flows) — it presses `q` once the
"Data refresh" toast shows while the view still reads "refreshing…"; fix: wait for the reloaded `octo/tools` row.

**LB-35 (parked)** — `Missing` (internal/brew/doctor.go) parses stdout only on exit 1, like `Doctor` did before ecc6a51 (brew doctor prints warnings on stderr). Capture real `brew missing` output with a missing dep before changing it.
