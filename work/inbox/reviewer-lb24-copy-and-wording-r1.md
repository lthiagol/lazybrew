verdict: changes-required

Reviewed `9beaafe..02ea524` (AC-01, AC-02, fix r2 `6b2e627` per `Q-lb24-y-search`) on `wip`.

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 copy | partial | OSC 52 real: a tea.Program running the cmd wrote `ESC]52;c;<b64>BEL`, decodes to the payload. Overlay copies (failed/running/done/cancelled) ANSI-free, uncut. Confirm open: `y` says yes, no copy. Add-tap input: `y` typed, no copy. Keymap: dropping `keyCopy` from Formulae hints, from help, or the `case "y"` dispatch each fails `TestPanelKeysAreOnFooterAndHelp`. F-r1-02, F-r1-03, F-r1-04 |
| AC-02 wording | partial | 140x40 frame: `✗ Failed: needs sudo · Enter  run in this terminal · Esc  close · y copy`, command below. `update` and `services start` sudo failures keep the old text. F-r1-01 |
| Fix r2 | ok | Overlay-first check reverted → `TestYOnSearchCopiesAFailedOverlay`, `TestYCopiesTheRetryCommandUnderTheStatusRow` fail. No overlay on Search: `y` types |
| User case | ok | Scratch harness: Search results → `i` → `y` (confirm) → mock stream fails with sudo's words → failed overlay, box not focused → `y` copied 6 lines: `Install wireshark-app` / `✗ Failed: needs sudo · Enter  run in this terminal · Esc  close` / `brew install --cask wireshark-app` / every output line, a 300-cell URL line intact; search box stays empty |

Gate: `make ci` exit 0 (all floors met, gui 90.9 % ≥ 83 %); `GOOS=darwin go build ./...` ok. No `program.Send`, no new goroutines; the copy is a `tea.Cmd` through the `setClipboard` hook. One commit per AC plus the ruled fix commit. README and `gui.mouse` sample comment agree.
View diff base `718ea0b` vs HEAD at 80x24 and 140x40 (every panel, help, 8 overlay states): only the overlay status rows, the retry command line, footers (`y copy`; Taps/Services lose `1-7 jump` at 80) and help differ.

## Findings

F-r1-01 (medium) — internal/gui/terminal_retry_test.go:401, internal/gui/copy_test.go:169 — the "80x24" tests call `renderOperation(80, …)`: a panel 80 wide, not an 80x24 frame. In the real 80x24 frame (`View()` after `WindowSizeMsg{80,24}`) the main panel is ~50 cells and the retry row reads `✗ Fa… · Enter  run in this terminal · Esc  close`: "needs sudo" is cut, and there is no `y copy` hint (the receipt says the 80x24 retry row has one). The footer there has no `y copy` either. — Fix must achieve: in the 80x24 frame the failed row shows `Failed: needs sudo` uncut and the Enter/Esc hint uncut (e.g. the hint on its own line); tests assert on `View()` at 80x24 and 140x40 frames; the receipt states where `y copy` is left out.

F-r1-02 (medium) — internal/gui/clipboard.go:15, internal/gui/logs.go:263, internal/gui/render.go:505 — the panel copy renders at `copyWidth = 400` and the Logs tab (and Status Config) pass through `cutLines`, so a log line wider than 400 cells is copied cut with `…`. Harness: Services Logs with a 610-cell line → copy holds `…7890123…`, not the line. — Fix must achieve: Logs and Config copies hold every line whole (no `…`); a test with a > 400-cell log line.

F-r1-03 (low) — internal/gui/keys.go:188 — with help open (`?`), `y` copies the panel hidden behind it (`Copied N lines`) while the screen shows help; V2 says the copy matches what the user sees. — Fix must achieve: `y` with help open does not copy the hidden panel (no copy, or copy the help text); a test.

F-r1-04 (low) — README.md:74 — the `y` row says it copies the text on screen, but on Search with no overlay `y` types into the box (the ruling). — Fix must achieve: the README row says `y` types on Search unless an Operation is open.

Not checked: OSC 52 in a real macOS terminal (no TUI launch). No brew write commands run.
