# Slice: v2-05s-services — a real Services Info view and safe start / stop / restart

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Seventh slice of `LB-08`, asked for by the user (ruling `Q-v2-services` in
`work/questions.md`): a proper **Info view** for a service and **start / stop
/ restart** as first-class, obvious and safe actions. The keys already exist
(`s` start, `S` stop, `r` restart, `f` run, `c` cleanup — `internal/gui/gui.go`
around the `case "s":` block, `serviceAction` in `commands.go`) and go through
`task.Manager`; this slice makes them visible, state-aware and confirmed, and
feeds the main area from `brew services info`.

## ⚠ Safety — read first

This machine runs **real** brew services (`moshi-hook` is live and must stay
up). **Never run `brew services start|stop|restart|run|cleanup|kill`** — not
in a test, not in the harness, not to "check". Tests use the mock runner only.
The frame harness may press `S`/`r` **only** to open a confirm dialog and must
then press `esc`; never `y` or `enter` on a services confirm. Read-only
commands allowed: `brew services list --json`, `brew services info <name>
--json`, `brew services info --all --json`.

## Real data shape

`brew services info --all --json` on this machine (read-only) returns, per
service: `name`, `service_name`, `running`, `loaded`, `schedulable`, `pid`
(null when stopped), `exit_code` (null), `user`, `status` (`started`, `none`,
`stopped`, `error`, `scheduled`, `unknown`), `file`, `registered`,
`loaded_file`, `command`, `working_dir`, `root_dir`, `log_path`,
`error_log_path`, `interval`, `cron`. Run it yourself and build the fixture
from its real output (keep `moshi-hook`'s fields as they are; it is not
secret).

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes the tests; `TestEveryFrameIsTerminalSize` passes at every commit.

### AC-01 — `brew services info` in the brew layer

1. `internal/brew`: a `ServicesReader.Info(ctx)` (all services, one call:
   `services info --all --json`) returning a typed `ServiceInfo` with the
   fields above (nullable ints/strings as pointers or zero+flag — your choice,
   documented). Cached under its own key with the services TTL; invalidated by
   every services write exactly like `KeyServicesList` (start, stop, restart,
   run, cleanup).
2. Tests with the real-shape fixture: parsing (running and not-running
   entries, null pid/exit_code), cache hit, invalidation by each write.
3. GUI reads it through `boundedRead` with `brewReadTimeout` (the read rule
   from `lb02`/`lb07`), when the Services panel loads and on refresh.

### AC-02 — the Services Info view

1. Services main tab renamed `Info` (was `Status`); it renders the selected
   service with the v2-05 detail-field helper: status with icon and role
   (`● running` Success, `○ stopped` Muted, `✗ error` Error, `⏲ scheduled`
   Info), PID, user, loaded / registered, unit file, command, working dir,
   log paths, schedule (interval / cron) — empty fields omitted.
2. Below the fields, the **actions that apply now** as a key line: running →
   `S stop  r restart`; not running → `s start  f run`; always `c cleanup`
   when relevant. The same state-aware set leads the footer for this panel.
3. Before `brew services info` has loaded, fall back to the list's fields
   (status, user, file, exit code) — never a blank pane.

### AC-03 — actions are state-aware and confirmed

1. `s` on a running service and `S`/`r` on a stopped one do **not** enqueue a
   task; they show an info toast (`moshi-hook is already running`).
2. `S` (stop) and `r` (restart) open a confirm dialog (v2-04 style): stop →
   danger kind, chip `STOP`, consequence `the service stops until you start
   it again`; restart → upgrade kind, chip `RESTART`, consequence `the
   service stops and starts again`. `s` and `f` run without a confirm (as
   today). Confirm → the existing `serviceAction` path through `task.Manager`.
3. After any services task completes (success or error), the Services list
   **and** the info reload, and the selection stays on the **same service by
   name** (not by index).

### AC-04 — clear errors

1. When a services task fails because brew needs root (system services —
   detect from brew's error text, e.g. `requires root`/`sudo`/`Permission
   denied` — read `internal/brew/errors.go` for existing mapping), the error
   toast says so plainly: `<name>: needs sudo — run "sudo brew services
   <action> <name>" in a terminal`. lazybrew never runs sudo.
2. Other failures keep the existing error toast text.
3. Tests with the mock runner returning those errors.

### AC-05 — evidence and bookkeeping

1. Harness copied to `work/verification/v2-05s-services/`; same screens plus
   `19-services-info` (Services → `moshi-hook` selected → Info) and
   `20-services-stop-confirm` (press `S` on `moshi-hook`, capture, then
   **`esc`** — never confirm). `before/` = `v2-05-content-polish/after/`;
   `after/` with `.ansi` 140×40.
2. `result.md`; `work/STATE.md` row
   `| v2-05s-services | — | implemented | awaiting reviewer |`;
   `work/backlog.md` LB-09 (Logs tab) stays parked — mention it in the receipt.

## Invariants

- The **Safety** section above. A brew services write on this machine is a
  blocker no matter what it was for.
- No change to how other panels work; no restyle of v2-03..v2-05 helpers or
  overlays; no animation (v2-06).
- No new keybinding: `s S r f c` keep their meaning; only the confirm and the
  state guard are added.
- Every frame exactly the terminal size. No test deleted, skipped or
  weakened; changed expectations named. Coverage floors unchanged. No new
  dependencies.
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.

## Report protocol

`work/inbox/runner-v2-05s-services.md` (≤ 40 lines), committed once; state in
it that no brew services write was run. Worktree clean; reply with exactly one
line: `done: work/inbox/runner-v2-05s-services.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-v2-05s-services-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-v2-05s-services-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
