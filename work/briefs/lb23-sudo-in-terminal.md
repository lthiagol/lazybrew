# Slice: lb23-sudo-in-terminal — run a step that needs sudo in the real terminal (LB-11)

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb22-update-default-and-leftovers` closed. User request
(2026-11-21): build it; the user tests on macOS later (this machine is Linux:
test with a mock and a fake `brew` script). Never launch the TUI against the real
brew; no brew write commands. Background: `Q-v2-11-b` in `work/questions.md`.

## Today

brew runs in its own session (`Setsid`) with no controlling terminal, so a cask
`.pkg` step that calls `sudo` fails at once; `sudoRunner` turns it into a
`NeedsSudoError` (`internal/brew/sudo.go`) and the failed Operation tells the
user to run the command in another terminal.

## Behaviour

When a task fails with a `NeedsSudoError` for an **install, reinstall, upgrade
or uninstall** (formula or cask, single or batch member), the failed Operation
overlay shows `Enter  run in this terminal · Esc  close`. Pressing Enter:

1. suspends the TUI and runs **the same brew command** attached to the real
   terminal with `tea.ExecProcess`, so sudo can ask for the password;
2. when it exits, the TUI resumes and the result shows as a toast: success, or
   `exited with status N` (a failure keeps the Operation text);
3. then refreshes exactly as a normal finished mutation does (cache
   invalidation for that command, the `RefreshMsg` path): the lists show the new
   state.

Esc still closes the overlay, as today. Nothing else changes for other failures.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — `NeedsSudoError` carries the command's args
- [ ] AC-02 — Enter on the failed overlay runs it with `tea.ExecProcess`
- [ ] AC-03 — result, refresh, docs

### AC-01

`NeedsSudoError` (brew package) gains `Args []string` (the failing command's
brew args, not a joined string only); `sudoError` fills it. Tests: the error
carries `["install","--cask","foo"]` for a fake runner failing with sudo's
terminal words; `Command` is unchanged for the existing message.

### AC-02

1. The Model gets a field `execProcess func(*exec.Cmd, tea.ExecCallback)
   tea.Cmd`, defaulting to `tea.ExecProcess`, so tests capture the command.
   Name the field and where it is set.
2. On the failed overlay with a sudo error for an allowed command (above
   only; never `update`, `services`, `tap`, `trust`, `cleanup`, `autoremove`),
   Enter builds `exec.Command(brewPath, args...)` with the **same environment
   the runner uses** (`HOMEBREW_NO_AUTO_UPDATE=1`, `HOMEBREW_NO_ASK=1`, inherited
   env), **no** `Setsid`/`Setpgid`, stdin/stdout/stderr left to ExecProcess, and
   returns `m.execProcess(cmd, cb)`. The brew path is the runner's `BrewPath()`.
3. This write does not go through `task.Manager`. That is deliberate and
   AGENTS.md rule 2 must say so: add a sentence to rule 2 ("the one exception
   is the user-confirmed retry of a step that needs sudo, which runs with the TUI
   suspended; the manager is idle then and nothing else can run"), and add a
   decision-log entry in `work/questions.md` (`Q-lb23-execprocess`). Guard it:
   it only runs when `m.tasks` has nothing running or queued and no Operation
   is in progress; otherwise Enter does nothing. Test that guard.
4. Tests (mock runner returning a `NeedsSudoError`, fake `execProcess`):
   - allowed command → Enter returns the exec cmd with the right path, args
     and env; the TUI model is otherwise unchanged;
   - not allowed (e.g. `services start`, `update`, `tap`) → Enter does nothing;
   - a batch member that needs sudo → the same, for that member's args;
   - the overlay hint line shows `Enter  run in this terminal` only then
     (frame test, both widths), and Esc closes as before.

### AC-03

1. The callback maps the exit to a message (`SudoRunDoneMsg{Err}`); the Update
   case toasts the result and issues the same refresh the finished mutation
   does (reuse the existing completion path; do not copy it). Tests: success
   toast + refresh msgs for the command's cache group; failure toast keeps the
   Operation closed and shows the exit status.
2. A real-process test with a fake `brew` shell script in a temp dir (reads
   `/dev/tty`, writes a marker file, exits 0 or 3), run under a pty through the
   real `execProcess` path: the script gets the tty (not Setsid), the exit
   status reaches `SudoRunDoneMsg`. If a pty helper is not available without a
   new dependency, drive it with `script(1)` and skip with a clear message when
   absent; say which in the receipt.
3. README (key + when it appears), the help page for the overlay if there is
   one, and AGENTS.md rule 2 as above. Name each golden that changes.
4. The receipt includes a "macOS manual test" section (≤ 8 lines) the user can
   follow with a real `.pkg` cask.

## Done means

`make ci` green; `GOOS=darwin go build ./...`; floors not lowered; deadcode and
staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Bubble Tea rules otherwise held (no `program.Send`, no raw
goroutines). Receipt `work/inbox/runner-lb23-sudo-in-terminal.md` (≤ 40 lines,
committed, not edited afterwards), one commit per AC, tests first.
