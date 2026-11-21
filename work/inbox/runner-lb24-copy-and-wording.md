# lb24-copy-and-wording — runner receipt

Commits, in AC order:
- AC-01 `9beaafe` lb24 AC-01: y copies the shown text to the clipboard
- AC-02 `755f7de` lb24 AC-02: the sudo failure row says needs sudo; the command goes below

## AC-01 — y copies (9beaafe)
- Test first: before the fix, `go test ./internal/gui/ -run TestY` did not build:
  `copy_test.go:18:4: m.setClipboard undefined`.
- Code: `clipboard.go` (copyScreen, copyText, plainText), `case "y"` in `keys.go`,
  `keyCopy` in `keybindings.go`, `Model.setClipboard` (New sets `tea.SetClipboard`).
- Evidence: `go test ./internal/gui/ -run TestY -v` → PASS ×5: overlay as plain
  text, main panel (Formulae Info) as plain text and uncut, focused box types `y`,
  Search unfocused types `y`, empty copy → "Nothing to copy".
- Choices: overlay copy = title + every `op.Lines` line, ANSI-stripped (status row
  left out); main panel = `renderContent` at width 400, ANSI-stripped, trailing
  spaces trimmed; toast `Copied N lines` as the brief words it.
- Search: `y` types, like g/G/h/l. So `y` is on the footer and help of Formulae,
  Casks, Services, Taps, Outdated, Status, not on Search. Filed `Q-lb24-y-search`.
- Goldens: no existing expected text changed. Two agreement tests now skip `y` as
  a panel key, as they skip j/k, h/l, c, a: `TestServicesFooterAndInfoAgreeOnActions`,
  `TestTapsFooterAndInfoAgreeOnActions`.
- Docs: README key row and mouse paragraph; `gui.mouse` sample comment.

## AC-02 — sudo failure row (755f7de)
- Test first: `TestFailedOverlaySaysWhatTheRetryRunsOnItsOwnLine` failed before the fix:
  `terminal_retry_test.go:404: 80x24: status row lacks "Failed: needs sudo · Enter  run in this terminal · Esc  close":`
  `✗ Failed: needs sudo — run "brew ins… · Enter  run in this terminal · Esc  close`
- After the fix it passes at 80x24 and 140x40: the row holds the full hint, and
  the line under it is `brew install --cask foo`.
- Other commands: unchanged path. At 80 columns their message is cut before the
  Esc hint, as it was before; at 140 it shows in full.
- Goldens: none changed.

## Gate
- `make ci` → exit 0 (fmt-check, vet, `test ./... -race -count=1`, cover-check
  all floors met, build). The gui floor is 83 %, measured 90.9 %.
- Per commit: `make test-fast PKG="./internal/gui/... ./internal/config/..."`
  (AC-01) and `make test-fast PKG="./internal/gui/..."` (AC-02): ok.
- `GOOS=darwin go build ./...`: ok.
- `deadcode ./cmd/...`: no output.
- `staticcheck ./...`: 5 findings, all present at `718ea0b`: keymap_test.go:39
  and :431 (SA1019 ParseDir), launch_check_test.go:180, :220, :453 (SA4006).
  None is in a line this slice touched.

## Seen, not done
- Overlay open on Search: `y` types into the search box, not the overlay (see
  `Q-lb24-y-search`).
- The copy of an Operation omits its status row (elapsed time, Failed text, hint).
- Not checked in a real terminal: OSC 52 copy on macOS. Tests cover the payload
  through the `setClipboard` hook only.
- No TUI launch; no brew write commands run.

## Fix r2 — Q-lb24-y-search (6b2e627)
- Open overlay: `y` copies it on every panel, Search too; a focused box still types. No overlay on Search: `y` types.
- Overlay copy: title, status row without the `y copy` hint, the retry command line when shown, then every output line.
- Status row ends `· y copy` where the row holds it: 80x24 and 140x40 retry rows yes; at 50 columns the close hint stays whole and no hint. Left out where it would cut the failure text (refused command at 80, long errors); their text is unchanged.
- Goldens changed: `TestFinishedRunsEndWithTheCloseHint` (+ `y copy`); `TestYCopiesTheOverlayAsPlainText` (status row, "Copied 4 lines"); `TestYWithNothingToCopySaysSo` replaced by `TestYCopiesAnOverlayWithNoOutputByItsStatusRow`.
- Revert check: overlay-first check removed → `TestYOnSearchCopiesAFailedOverlay` and `TestYCopiesTheRetryCommandUnderTheStatusRow` fail (`y copied []`).
- "Nothing to copy" is unreachable from the UI now (every panel renders text; every overlay has a status row). Branch kept as a guard.
- Gate: `make ci` exit 0; `GOOS=darwin go build ./...` ok.

## Fix r3 (F-r1-01 … F-r1-04)
- F-r1-01 `22a1a87`: 80x24 row `✗ Failed: needs sudo · y copy`, keys on the line under it, command below. 140x40 unchanged. Before the fix `TestFailedRetryRowInTheFrame` failed: `✗ Fa… · Enter…`. The width-only tests are replaced by View tests at 80x24 and 140x40; the 50-cell case is dropped.
- F-r1-02 `b05e324`: `copying` flag on the copy's Model; `fitLines` skips the cut for Logs and Config. Before: `TestYCopiesALongLogLineWhole` and `TestYCopiesALongConfigRowWhole` failed (`…` in the copy).
- F-r1-03 `9762f16`: y does nothing while help is open (no help text copied). Before: `TestYWithHelpOpenCopiesNothing` copied the hidden panel.
- F-r1-04 `ccd6fed`: README y row says Search types y unless an Operation is open.
- Also `2907e72`: five unused `m =` reassignments in copy_test.go (SA4006) removed.
- Goldens: none changed. Gate: `make ci` exit 0; `GOOS=darwin go build ./...` ok; deadcode no output; staticcheck back to the 5 baseline findings.
- Seen, not done: the 80x24 footer shows no `y copy`; the Logs copy omits the stream/file header line.
