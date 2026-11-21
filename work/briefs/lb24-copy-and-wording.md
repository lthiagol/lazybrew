# Slice: lb24-copy-and-wording — a copy key, and a clearer sudo failure line

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb23-sudo-in-terminal` closed. User report (2026-11-21, macOS
smoke): text cannot be selected in the TUI because the mouse is captured. Never
launch the TUI against the real brew; no brew write commands.

## AC-01 — `y` copies what is on screen

1. A key `y` copies plain text to the terminal clipboard with
   `tea.SetClipboard` (OSC 52; Bubble Tea v2). What it copies:
   - with an Operation overlay open (running, done, failed): the overlay's full
     text, title line first, then every output line, **ANSI-free** and
     untruncated (not the clipped view);
   - otherwise, the main panel's content for the selected item (the current
     tab's text, ANSI-free, untruncated: Info, Deps, Logs, doctor output…).
   - not while a text input is focused (`y` types there).
2. A toast confirms: `Copied N lines`; nothing to copy: `Nothing to copy`.
3. `y` is in the single keymap (`keybindings.go`): footer on the panels where it
   applies and the help page; the keymap test passes; check `y` collides with no
   binding. Update README (keys, and a sentence that mouse capture blocks the
   terminal's own selection: hold Option/Shift, or set `gui.mouse: false`) and
   the config sample comment for `gui.mouse` to say the same.
4. Tests: the copied text for a failed overlay equals the overlay lines without
   escape codes (assert on the `SetClipboard` cmd's payload through a Model hook
   `setClipboard func(string) tea.Cmd`, defaulting to `tea.SetClipboard`, as
   lb23 did for `execProcess`); a main-panel copy (Formulae Info); the focused
   input case does nothing; the toast text. Name each golden that changes
   (footers/help).

## AC-02 — sudo failure line without the redundant clause

When the terminal retry is offered (`terminalRetry()` true), the failed line is
`Failed: needs sudo · Enter  run in this terminal · Esc  close` (drop
`— run "brew … " in a terminal`). The command stays visible: put it on the
line below, `brew install --cask wireshark-app`, in the subtle style, so the
user can read it. When the retry is not offered (other commands), the old text
stays unchanged. Frame tests at 80x24 and 140x40: no `…` clipping of the hint.
Name each golden that changes.

## Done means

`make ci` green; `GOOS=darwin go build ./...`; floors not lowered; deadcode and
staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Receipt `work/inbox/runner-lb24-copy-and-wording.md` (≤ 40
lines, committed, not edited afterwards), one commit per AC, tests first.
