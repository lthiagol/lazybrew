verdict: changes-required

# Review r1 — slice v2-05s-services (wip 08e7653..06ddd77)

**V0 safety:** `systemctl --user show sh.brew.moshi-hook` after the review's own harness run:
MainPID=800, NRestarts=0, ActiveEnterTimestamp=Wed 2026-11-18 11:37:48 -03 — same as the
orchestrator's pre-slice record. journalctl since 2026-11-18 11:00: last event is the 11:37:48
start; nothing in the slice window. Harness diff vs v2-05 harness: only `6`,`j`,shot,`S`,shot,`esc`
added; no `y`/enter after `S`/`r`. No test builds a real runner (grep `NewRunner|NewDefaultRunner`
in `*_test.go`: none). No blocker.
**V1 gate (CGO_ENABLED=1):** build, lint, `go test ./... -race -count=1` (9 pkgs ok),
cover-check (brew 76.5, presentation 98.2, gui 82.9, modal 90.1) — all pass.
**V2 frames:** harness re-run against real brew into scratch, `diff -r` vs `after/`: identical.
40 `.txt` all exactly W×H. `before/` == `v2-05-content-polish/after/`. `internal/gui/dump` removed.

| AC | verdict | evidence |
|---|---|---|
| AC-01 | ok | fixture == live `brew services info --all --json` byte-for-byte; `*int` pid/exit_code/interval, `map` cron; `KeyServicesInfo` with `servicesTTL`; `TestServicesWritesInvalidateServicesInfo` covers start/stop/restart/run/cleanup; `fetchPanelData` reads via `boundedRead(brewReadTimeout, …)`; refresh invalidates it (`invalidatePanelCache`) |
| AC-02 | ok (see F-r1-01) | tab `Info`; frame 19 fields, null fields omitted, key line + footer `S stop  r restart`; list fallback `serviceListDetail` + test |
| AC-03 | changes-required | toasts, confirms (frame 20: STOP chip, danger, consequence), s/f no confirm, reload on success/failure/cancel, cursor by name with reorder test — all ok; but the guard is wider than "stopped": F-r1-01 |
| AC-04 | ok | `serviceWriteError` reads err + stderr for sudo/requires root/permission denied; toast text exact; no sudo exec anywhere; others unchanged (`failureText`); mock tests |
| AC-05 | ok | result.md explains 19/20 and 01–18 unchanged; STATE row present; LB-09 parked and named in receipt |

V8: no new key, `panelHints` untouched; 31 tests added, none removed; the 4 changed bodies keep
their intent (restart still asserted as `brew services restart redis` after confirm). One commit per AC.
Toast wrap / icon decisions: implemented as ruled (3 rows then `…`; short toasts byte-identical in frames 10/13; glyphs from the icon set).

## Findings

F-r1-01 (major) — internal/gui/details.go:181 (`serviceRunning`), used by commands.go:834 guard
and `serviceActionHints` — "running" is `Status == started`, so `error` and `scheduled` services
are treated as stopped: `S`/`r` toast "<name> is not running" and enqueue nothing, and Info/footer
offer `s start  f run`. — Evidence: brew's formula_wrapper.rb status: `error` and `scheduled` are
*loaded* units (`:none` is the only not-loaded state); cli.rb `stop` acts on any `loaded?` service.
Before this slice `r`/`S` ran on them; now a crashed service cannot be restarted and a scheduled one
cannot be stopped from lazybrew — a regression of `S`/`r` meaning (invariant "keep their meaning"),
and AC-03.1 only asks to refuse S/r on a *stopped* service. No test covers error/scheduled actions. —
Fix must achieve: `S` and `r` on `error` and `scheduled` open their confirms and run; the Info key
line and footer offer stop/restart for them (start/run too if wanted); tests for both states.

F-r1-02 (minor) — internal/gui/gui.go:408 — a failed services task now sends `RefreshMsg`, which
reloads every loaded panel and ends with a green "Data refreshed" toast stacked right after the red
error/needs-sudo toast. — Evidence: `RefreshMsg` sets `m.refreshing`; last `DataLoadedMsg` toasts
"Data refreshed" (gui.go:558-562). — Fix: after a failed/cancelled services task, reload Services
(list + info) without the success toast, or state in result.md why the toast is wanted.

F-r1-03 (minor, visual) — after/80x24/19-services-info.txt — the Info key line is under the three
stale toasts carried from frames 10/13 (`S stop  r restart  c c…`), so the 80×24 frame does not show
what AC-02.2 adds. — Fix: dismiss/expire toasts before frame 19 in the harness (or note it in result.md).

140×40 `.ansi` (19, 20) read well; no other visual issue.
