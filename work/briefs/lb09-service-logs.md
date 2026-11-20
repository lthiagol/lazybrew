# Slice: lb09-service-logs — a Logs tab that tails a service's log files

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. If you find something in the repo that contradicts the brief,
> follow the brief and report the contradiction — do not stop to ask.

## Goal

Backlog LB-09, the user's design answers `Q-lb09-design` in
`work/questions.md`. Services gets a second main-view tab, **Logs**, next to
Info (`panelTabs[PanelServices]`, `internal/gui/panel.go`). It shows the last
lines of the selected service's log file and keeps adding new lines while it
is open, like `tail -f`. `brew services info --json` already gives both paths
(`brew.ServiceInfo.LogPath`, `.ErrorLogPath`, `internal/brew/types.go` ~:183).

## ⚠ Safety

Never run brew write commands on this machine (`brew install|uninstall|
upgrade|reinstall|tap|untap|trust|untrust|pin|unpin|cleanup|autoremove|update`,
`brew services start|stop|restart|run|cleanup|kill`). `moshi-hook` is a live
service: do **not** start/stop/restart it to make log lines. Tests use temp
files and the mock runner. Reading the real moshi-hook log file for a frame
capture is allowed (read-only).

## Acceptance criteria

One commit per criterion, in order, staged by name. Every commit builds and
passes `make test-fast PKG=<touched packages>`; the full gate (with `-race`)
runs once, before the receipt (ruling `Q-v2-builder-speed`).

### AC-01 — reading the tail of a file

1. A pure, tested function (in `internal/brew` or a small new package under
   `internal/`, your choice — say which) that returns the last `n` lines of a
   file, reading at most a bounded number of bytes from the end (a named
   constant, e.g. 256 KiB) — never the whole file. It reports: lines, the
   file's size (so a later read can tell growth, truncation or rotation), and
   distinct results for "path empty", "file does not exist", other errors.
2. Tests on temp files: short file, file longer than `n` lines, a file bigger
   than the byte bound (only the end is read, first partial line dropped), no
   trailing newline, empty file, missing file, CRLF, invalid UTF-8 / control
   bytes made safe for the terminal (no raw escape sequences reach the view).

### AC-02 — the Logs tab

1. `panelTabs[PanelServices]` is `Info`, `Logs`; `[ ]` switches tabs and the
   Services footer lists `[ ] tabs`. Info is unchanged.
2. Logs shows the last 200 lines (named constant) of the selected service's
   **stdout** log (`LogPath`). `e` toggles to **stderr** (`ErrorLogPath`) and
   back; the tab's header line names the stream and the path
   (e.g. `stdout · /…/moshi-hook.log`), cut to fit. The toggle resets to
   stdout when the selected service changes. `e` is listed in the footer only
   while Logs is active, and in help.
3. Empty states, each one line of muted text, never an error toast: path unset
   (`no stdout log for this service`), file missing (`<path> does not exist
   yet`), stderr path equals stdout path (`stderr goes to the same file as
   stdout`, still showing the lines), read error (the error text).
4. Long lines are cut to the main area's width (no wrap); the frame is
   exactly the terminal size at 80x24 and 140x40.
5. Scrolling uses the existing main-view keys (`pgup/pgdown/home/end/g/G`).

### AC-03 — follow

1. While Services is the active panel **and** Logs the active tab, the view
   re-reads the tail every 1 s (named, test-overridable) via `tea.Tick` →
   message → `Update` → `tea.Cmd` read. No raw goroutines, no
   `program.Send`, no file watchers (AGENTS.md Bubble Tea rules).
2. A tick carries a generation (or equivalent): switching panel/tab/service
   or stream stops the loop, and a stale tick or stale read result is
   dropped — never two loops at once, never a frame showing another service's
   log.
3. When at the bottom, the view stays pinned to the newest line as lines
   arrive; if the user scrolled up, new lines do not move the view (until
   `end`/`G`). A shrunk file (truncated/rotated) is re-read from its new end.
4. No re-render work when the file did not change (size and mod time equal).

### AC-04 — tests and evidence

1. GUI tests (unit + a `teatest` flow in `internal/gui/flows`, asserting on
   `View()`): open Logs, see the last lines of a temp log; append lines and
   deliver a tick → they appear; toggle `e` → stderr; switch service → stdout
   of the new one, no stale lines; leave the tab → no further reads (count
   reads through an injectable reader); every empty state of AC-02.3.
2. Frames at 80x24 and 140x40 of the Logs tab (a temp log via the mock runner,
   plus the real moshi-hook log read-only) under
   `work/verification/lb09-service-logs/`, and `result.md` (tests, what each
   proves, gate summary).
3. `work/STATE.md` row `| lb09-service-logs | — | implemented | awaiting reviewer |`.

## Invariants

- No change to other panels, the Info tab, or service actions. No new
  dependencies. Coverage floors unchanged (raise none, lower none). Every
  frame exactly the terminal size.
- No test deleted, skipped or weakened; changed expectations named (footer
  goldens that gain `[ ] tabs` / `e` are expected — name them).
- One commit per AC, in order, staged by name. No amend/rebase/revert/push.

## Gate

`make ci`; also `GOOS=darwin go build ./...`.

## Report protocol

`work/inbox/runner-lb09-service-logs.md` (≤ 40 lines), committed once;
worktree clean; reply with exactly one line:
`done: work/inbox/runner-lb09-service-logs.md`. Do not touch other panes.

## Fix rounds

`work/briefs/fix-lb09-service-logs-rN.md` → fix every finding, one commit each
(subject names the id), gate, receipt
`work/inbox/runner-lb09-service-logs-fix-rN.md` (≤ 30 lines, committed once).

## Questions

Append to `work/questions.md` and keep going on unaffected ACs.
