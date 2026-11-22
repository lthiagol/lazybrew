# GUI rules (Bubble Tea)

All GUI state changes go through `Update(msg tea.Msg)`. The stack is Charm v2: import
`charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2` (not
`github.com/charmbracelet/...`); keys are `tea.KeyPressMsg` (`Code`, `Text`); `View()`
returns a `tea.View` (`.Content`).

1. **No `program.Send` in production code.** Return a `tea.Cmd` that yields a message.
2. **Every brew write is a task.** Build it with `errTask` or `streamTask` and start it
   with `m.enqueue(t)` (or `m.tryEnqueue` to react to a full queue), in `task_build.go` —
   not `m.tasks.Enqueue` directly. A write is any verb that changes brew state: install,
   uninstall, reinstall, upgrade, update, pin/unpin, tap/untap/`tap --repair`,
   trust/untrust, services start/stop/restart/run/cleanup, cleanup, autoremove. The one
   exception is the user-confirmed retry of a step that needs sudo (`terminal_retry.go`):
   it runs with the TUI suspended, only when the manager is idle and no batch is in flight.
3. **No raw goroutines in handlers.** Use `tea.Cmd` closures or a `streamTask`.
4. **The Operation overlay owns streaming output.** `TaskOutputMsg` lines append to
   `m.opState`; `TaskCompletedMsg` marks it `opSuccess`/`opError`/`opCancelled` and
   toasts. Success closes itself after 2 s; failed and cancelled runs stay until
   Esc or a panel switch. There is no progress modal.
5. **Model fields are the source of truth.** Never read `m.program` to infer state.
6. **Call a mutating method before the return.** `Update` returns the Model by value,
   and Go does not order the read of `m` against a call in the same return. Write
   `cmd := m.switchPanel(p); return m, cmd`, never `return m, m.switchPanel(p)`
   (lb09 F-r1-01; the sites left are LB-38).

## Things that are easy to get wrong

- **The Log** (the command-log box, `command_log.go`). `LoggingRunner.OnCommand`
  (`internal/brew/runner.go`) is one hook per brew command: it appends one running row
  and returns the finisher that sets its final status. Which exit 1 counts as a report
  (Warn) or a result (OK) lives in `brew.ExitOneReport` (see [brew.md](brew.md)).
- **Refresh.** `m.refreshing` counts the reads of pending refreshes (`+=` per refresh,
  `gui.go`); "Data refreshed" toasts at 0 (`land.go`). Known bug, parked as LB-36: any
  non-preloaded landing also decrements it, so the toast can come early. A loaded
  sidebar panel keeps its rows while it re-reads; the main pane shows "Loading..."
  (ASCII dots) over cached content, pinned by `TestLoadingOverCachedFilesTabShowsLoading`.
- **Taps & Trust rows** are built once, by `setTapRows` in `trusted.go`, when taps
  or trust entries land; every index lookup goes through `tapRowAt`. A test that
  sets `p.taps` / `p.trusted` by hand must call `setTapRows()`.
- **`/` on Formulae and Casks filters the installed list;** it does not search
  brew. Brew search is the Search panel (`7`, then `/`).
